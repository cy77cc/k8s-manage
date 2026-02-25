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

	"github.com/cloudwego/eino/schema"
	ai2 "github.com/cy77cc/k8s-manage/internal/ai"
	"github.com/gin-gonic/gin"
)

type toolSummary struct {
	Calls          int      `json:"calls"`
	Results        int      `json:"results"`
	Missing        []string `json:"missing"`
	MissingCallIDs []string `json:"missing_call_ids,omitempty"`
}

type streamErrorPayload struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
}

type toolEventTracker struct {
	mu      sync.Mutex
	calls   map[string]int
	results map[string]int
	callIDs map[string]string
	doneIDs map[string]struct{}
}

func newToolEventTracker() *toolEventTracker {
	return &toolEventTracker{
		calls:   map[string]int{},
		results: map[string]int{},
		callIDs: map[string]string{},
		doneIDs: map[string]struct{}{},
	}
}

func (t *toolEventTracker) noteCall(callID, tool string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	name := strings.TrimSpace(tool)
	if name == "" {
		name = "unknown"
	}
	cid := strings.TrimSpace(callID)
	if cid == "" {
		cid = fmt.Sprintf("legacy-%s-%d", name, t.calls[name]+1)
	}
	t.callIDs[cid] = name
	t.calls[name]++
}

func (t *toolEventTracker) noteResult(callID, tool string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cid := strings.TrimSpace(callID)
	if cid != "" {
		t.doneIDs[cid] = struct{}{}
	}
	name := strings.TrimSpace(tool)
	if name == "" {
		name = "unknown"
	}
	t.results[name]++
}

func (t *toolEventTracker) summary() toolSummary {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := toolSummary{}
	for _, count := range t.calls {
		out.Calls += count
	}
	for _, count := range t.results {
		out.Results += count
	}
	hasCallID := len(t.callIDs) > 0
	for callID, tool := range t.callIDs {
		if _, ok := t.doneIDs[callID]; ok {
			continue
		}
		out.MissingCallIDs = append(out.MissingCallIDs, callID)
		if tool != "" {
			out.Missing = append(out.Missing, tool)
		}
	}
	if !hasCallID {
		for tool, callCount := range t.calls {
			missing := callCount - t.results[tool]
			for i := 0; i < missing; i++ {
				out.Missing = append(out.Missing, tool)
			}
		}
	}
	return out
}

func resolveStreamState(fatalErr *streamErrorPayload, summary toolSummary) string {
	if fatalErr != nil {
		return "failed"
	}
	if len(summary.MissingCallIDs) > 0 || len(summary.Missing) > 0 {
		return "partial"
	}
	return "ok"
}

func (h *handler) chat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "message is required"}})
		return
	}
	if h.svcCtx.AI == nil || h.svcCtx.AI.Runnable == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": gin.H{"message": "ai agent not initialized"}})
		return
	}
	uid, ok := uidFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"message": "unauthorized"}})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "streaming not supported"}})
		return
	}
	turnID := fmt.Sprintf("turn-%d", time.Now().UnixNano())
	writer := newSSEWriter(c, flusher, turnID)
	emit := writer.Emit
	var finalOnce sync.Once
	emitFinal := func(event string, payload gin.H) bool {
		sent := false
		finalOnce.Do(func() {
			sent = emit(event, payload)
		})
		return sent
	}
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
		emitFinal("error", gin.H{"message": err.Error()})
		return
	}
	if !emit("meta", gin.H{"sessionId": session.ID, "createdAt": session.CreatedAt}) {
		return
	}

	approvalToken := strings.TrimSpace(toString(req.Context["approval_token"]))
	tracker := newToolEventTracker()
	streamCtx := h.buildToolContext(c.Request.Context(), uid, approvalToken, scene, req.Context, emit, tracker)
	prompt := msg
	if len(req.Context) > 0 {
		prompt = msg + "\n\n上下文:\n" + mustJSON(req.Context)
	}
	inputMessages := h.buildConversationMessages(session.Messages, msg, prompt)
	stream, err := h.svcCtx.AI.Stream(streamCtx, inputMessages)
	if err != nil {
		emitFinal("error", gin.H{"message": err.Error()})
		return
	}
	defer stream.Close()

	var fatalErr *streamErrorPayload
	var assistantContent strings.Builder
	var reasoningContent strings.Builder
	var streamErr error
	for {
		item, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			if apErr, ok := ai2.IsApprovalRequired(recvErr); ok {
				_ = emit("approval_required", gin.H{
					"tool":           apErr.Tool,
					"approval_token": apErr.Token,
					"expiresAt":      apErr.ExpiresAt,
					"message":        apErr.Error(),
				})
				streamErr = recvErr
				break
			}
			streamErr = recvErr
			break
		}
		if item == nil {
			continue
		}
		if item.ReasoningContent != "" {
			reasoningContent.WriteString(item.ReasoningContent)
			_ = emit("thinking_delta", gin.H{"contentChunk": item.ReasoningContent})
		}
		if item.Content != "" {
			assistantContent.WriteString(item.Content)
			if !emit("delta", gin.H{"contentChunk": item.Content}) {
				return
			}
		}
	}
	if streamErr != nil && !errors.Is(streamErr, io.EOF) {
		if _, ok := ai2.IsApprovalRequired(streamErr); !ok {
			fatalErr = &streamErrorPayload{
				Code:        "stream_interrupted",
				Message:     streamErr.Error(),
				Recoverable: true,
			}
		}
	}

	summary := tracker.summary()
	if len(summary.MissingCallIDs) > 0 {
		toolErr := &streamErrorPayload{
			Code:        "tool_result_missing",
			Message:     fmt.Sprintf("tool result missing for %d call(s)", len(summary.MissingCallIDs)),
			Recoverable: true,
		}
		_ = emitFinal("error", gin.H{
			"code":         toolErr.Code,
			"message":      toolErr.Message,
			"recoverable":  toolErr.Recoverable,
			"tool_summary": summary,
		})
	}

	content := strings.TrimSpace(assistantContent.String())
	if content == "" {
		switch {
		case fatalErr != nil:
			content = fmt.Sprintf("本轮执行未完整结束：%s", fatalErr.Message)
		case len(summary.MissingCallIDs) > 0:
			content = "本轮工具调用结果不完整。"
		default:
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
	h.refreshSuggestions(uid, scene, content)
	streamState := resolveStreamState(fatalErr, summary)
	if fatalErr != nil {
		_ = emitFinal("error", gin.H{
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

func (h *handler) buildToolContext(ctx context.Context, uid uint64, approvalToken, scene string, runtime map[string]any, emit func(event string, payload gin.H) bool, tracker *toolEventTracker) context.Context {
	ctx = ai2.WithToolUser(ctx, uid, approvalToken)
	ctx = ai2.WithToolRuntimeContext(ctx, runtime)
	ctx = ai2.WithToolMemoryAccessor(ctx, &toolMemoryAccessor{
		store: h.store,
		uid:   uid,
		scene: scene,
	})
	ctx = ai2.WithToolPolicyChecker(ctx, h.toolPolicy)
	ctx = ai2.WithToolEventEmitter(ctx, func(event string, payload any) {
		switch event {
		case "tool_call", "tool_result":
			pm := toPayloadMap(payload)
			toolName := strings.TrimSpace(toString(pm["tool"]))
			callID := strings.TrimSpace(toString(pm["call_id"]))
			switch event {
			case "tool_call":
				tracker.noteCall(callID, toolName)
			case "tool_result":
				tracker.noteResult(callID, toolName)
			}
			_ = emit(event, gin.H{
				"tool":             toolName,
				"call_id":          callID,
				"payload":          pm,
				"ts":               time.Now().UTC().Format(time.RFC3339Nano),
				"retry":            pm["retry"],
				"param_resolution": pm["param_resolution"],
			})
		}
	})
	return ctx
}

func toPayloadMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{"raw": v}
}

func mustJSON(v any) string {
	raw, _ := jsonMarshal(v)
	return raw
}

func (h *handler) buildConversationMessages(history []map[string]any, originalMsg, finalPrompt string) []*schema.Message {
	if len(history) == 0 {
		return []*schema.Message{schema.UserMessage(finalPrompt)}
	}
	start := 0
	if len(history) > 20 {
		start = len(history) - 20
	}
	out := make([]*schema.Message, 0, len(history[start:])+1)
	for i := start; i < len(history); i++ {
		role := strings.TrimSpace(toString(history[i]["role"]))
		content := toString(history[i]["content"])
		if role == "assistant" && strings.TrimSpace(content) != "" {
			out = append(out, schema.AssistantMessage(content, nil))
			continue
		}
		if role == "user" {
			if i == len(history)-1 && content == originalMsg {
				out = append(out, schema.UserMessage(finalPrompt))
			} else {
				out = append(out, schema.UserMessage(content))
			}
		}
	}
	if len(out) == 0 {
		out = append(out, schema.UserMessage(finalPrompt))
	}
	return out
}
