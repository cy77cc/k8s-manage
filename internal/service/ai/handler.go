package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	adkcore "github.com/cloudwego/eino/adk"
	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

func (h *handler) chatWithADK(c *gin.Context, req chatRequest, uid uint64, msg string) {
	if h.svcCtx.AI == nil || h.svcCtx.AI.ADKAgent == nil {
		httpx.Fail(c, xcode.ServerError, "ai adk agent not initialized")
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		httpx.Fail(c, xcode.ServerError, "streaming not supported")
		return
	}
	turnID := fmt.Sprintf("turn-%d", time.Now().UnixNano())
	writer := newSSEWriter(c, flusher, turnID)
	emit := writer.Emit
	defer writer.Close()

	stopHeartbeat := make(chan struct{})
	go heartbeatLoop(stopHeartbeat, emit)
	defer close(stopHeartbeat)

	sid := strings.TrimSpace(req.SessionID)
	scene := normalizeScene(toString(req.Context["scene"]))
	if sid == "" {
		if session, ok := h.store.currentSession(uid, scene); ok {
			sid = session.ID
		} else {
			sid = fmt.Sprintf("sess-%d", time.Now().UnixNano())
		}
	}

	userTime := time.Now()
	session, err := h.store.appendMessage(uid, scene, sid, map[string]any{
		"id":        fmt.Sprintf("u-%d", userTime.UnixNano()),
		"role":      "user",
		"content":   msg,
		"timestamp": userTime,
	})
	if err != nil {
		_ = emit("error", gin.H{"message": err.Error()})
		return
	}
	if !emit("meta", gin.H{"sessionId": session.ID, "createdAt": session.CreatedAt}) {
		return
	}

	approvalToken := strings.TrimSpace(toString(req.Context["approval_token"]))
	tracker := newToolEventTracker()
	h.store.rememberContext(uid, scene, extractResourceContext(req.Context, msg))
	streamCtx := h.buildToolContext(c.Request.Context(), uid, approvalToken, scene, msg, req.Context, emit, tracker)

	prompt := msg
	directive := composePromptDirectives(
		buildToolExecutionDirective(msg, scene),
		buildHelpKnowledgeDirective(msg),
	)
	if directive != "" {
		prompt = directive + "\n\n用户问题:\n" + msg
	}
	if len(req.Context) > 0 {
		prompt = msg + "\n\n上下文:\n" + mustJSON(req.Context)
		if directive != "" {
			prompt = directive + "\n\n用户问题:\n" + msg + "\n\n上下文:\n" + mustJSON(req.Context)
		}
	}

	runner := adkcore.NewRunner(streamCtx, adkcore.RunnerConfig{
		EnableStreaming: true,
		Agent:           h.svcCtx.AI.ADKAgent,
		CheckPointStore: h.svcCtx.AICheckpoints,
	})

	iter := runner.Query(streamCtx, prompt, adkcore.WithCheckPointID(sid))

	var assistantContent strings.Builder
	var reasoningContent strings.Builder
	var fatalErr *streamErrorPayload
	var finalOnce sync.Once
	emitFinal := func(event string, payload gin.H) {
		finalOnce.Do(func() {
			_ = emit(event, payload)
		})
	}

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if err := h.processADKEvent(emit, tracker, event, &assistantContent, &reasoningContent); err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			fatalErr = &streamErrorPayload{Code: "stream_interrupted", Message: err.Error(), Recoverable: true}
			break
		}
	}

	summary := tracker.summary()
	content := strings.TrimSpace(assistantContent.String())
	if content == "" {
		if fatalErr != nil {
			content = fmt.Sprintf("本轮执行未完整结束：%s", fatalErr.Message)
		} else {
			content = "无输出。"
		}
	}

	assistantTime := time.Now()
	session, err = h.store.appendMessage(uid, scene, sid, map[string]any{
		"id":        fmt.Sprintf("a-%d", assistantTime.UnixNano()),
		"role":      "assistant",
		"content":   content,
		"thinking":  strings.TrimSpace(reasoningContent.String()),
		"timestamp": assistantTime,
	})
	if err != nil {
		emitFinal("error", gin.H{"message": err.Error()})
		return
	}

	streamState := resolveStreamState(fatalErr, summary)
	if fatalErr != nil {
		emitFinal("error", gin.H{
			"code":         fatalErr.Code,
			"message":      fatalErr.Message,
			"recoverable":  fatalErr.Recoverable,
			"tool_summary": summary,
		})
	}
	emitFinal("done", gin.H{
		"session":      session,
		"stream_state": streamState,
		"tool_summary": summary,
	})
}

func (h *handler) resumeWithADK(ctx context.Context, checkpointID string, targets map[string]any) (*adkcore.AsyncIterator[*adkcore.AgentEvent], error) {
	if h.svcCtx.AI == nil || h.svcCtx.AI.ADKAgent == nil {
		return nil, fmt.Errorf("ai adk agent not initialized")
	}
	runner := adkcore.NewRunner(ctx, adkcore.RunnerConfig{
		EnableStreaming: true,
		Agent:           h.svcCtx.AI.ADKAgent,
		CheckPointStore: h.svcCtx.AICheckpoints,
	})
	return runner.ResumeWithParams(ctx, checkpointID, &adkcore.ResumeParams{Targets: targets})
}

func (h *handler) resumeADKApproval(c *gin.Context) {
	var req struct {
		CheckpointID string `json:"checkpoint_id" binding:"required"`
		Target       string `json:"target" binding:"required"`
		Data         any    `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	iter, err := h.resumeWithADK(c.Request.Context(), strings.TrimSpace(req.CheckpointID), map[string]any{
		strings.TrimSpace(req.Target): req.Data,
	})
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}

	var output strings.Builder
	for {
		ev, ok := iter.Next()
		if !ok {
			break
		}
		if ev == nil {
			continue
		}
		if ev.Err != nil {
			httpx.Fail(c, xcode.ServerError, ev.Err.Error())
			return
		}
		if ev.Output != nil && ev.Output.MessageOutput != nil && ev.Output.MessageOutput.Message != nil {
			output.WriteString(ev.Output.MessageOutput.Message.Content)
		}
	}
	httpx.OK(c, gin.H{"resumed": true, "content": strings.TrimSpace(output.String())})
}
