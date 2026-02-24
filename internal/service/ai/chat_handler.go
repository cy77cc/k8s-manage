package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	ai2 "github.com/cy77cc/k8s-manage/internal/ai"
	"github.com/gin-gonic/gin"
)

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
		_ = writeSSE(c, flusher, "error", gin.H{"message": err.Error()})
		return
	}
	if !writeSSE(c, flusher, "meta", gin.H{"sessionId": session.ID, "createdAt": session.CreatedAt}) {
		return
	}

	approvalToken := strings.TrimSpace(toString(req.Context["approval_token"]))
	streamCtx := h.buildToolContext(c.Request.Context(), uid, approvalToken, flusher, c)
	prompt := msg
	if len(req.Context) > 0 {
		prompt = msg + "\n\n上下文:\n" + mustJSON(req.Context)
	}
	inputMessages := h.buildConversationMessages(session.Messages, msg, prompt)
	stream, err := h.svcCtx.AI.Stream(streamCtx, inputMessages)
	if err != nil {
		_ = writeSSE(c, flusher, "error", gin.H{"message": err.Error()})
		return
	}
	defer stream.Close()

	var assistantContent strings.Builder
	var reasoningContent strings.Builder
	for {
		item, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			if apErr, ok := ai2.IsApprovalRequired(recvErr); ok {
				_ = writeSSE(c, flusher, "approval_required", gin.H{
					"tool":           apErr.Tool,
					"approval_token": apErr.Token,
					"expiresAt":      apErr.ExpiresAt,
					"message":        apErr.Error(),
				})
				break
			}
			_ = writeSSE(c, flusher, "error", gin.H{"message": recvErr.Error()})
			break
		}
		if item == nil {
			continue
		}
		if len(item.ToolCalls) > 0 {
			_ = writeSSE(c, flusher, "tool_call", gin.H{"tool_calls": item.ToolCalls})
		}
		if item.ReasoningContent != "" {
			reasoningContent.WriteString(item.ReasoningContent)
			_ = writeSSE(c, flusher, "thinking_delta", gin.H{"contentChunk": item.ReasoningContent})
		}
		if item.Content != "" {
			assistantContent.WriteString(item.Content)
			if !writeSSE(c, flusher, "delta", gin.H{"contentChunk": item.Content}) {
				return
			}
		}
	}

	content := strings.TrimSpace(assistantContent.String())
	if content == "" {
		content = "已完成。"
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
		_ = writeSSE(c, flusher, "error", gin.H{"message": err.Error()})
		return
	}
	h.refreshSuggestions(uid, scene, content)
	_ = writeSSE(c, flusher, "done", gin.H{"session": session})
}

func (h *handler) buildToolContext(ctx context.Context, uid uint64, approvalToken string, flusher http.Flusher, c *gin.Context) context.Context {
	ctx = ai2.WithToolUser(ctx, uid, approvalToken)
	ctx = ai2.WithToolPolicyChecker(ctx, h.toolPolicy)
	ctx = ai2.WithToolEventEmitter(ctx, func(event string, payload any) {
		switch event {
		case "tool_call", "tool_result":
			_ = writeSSE(c, flusher, event, payload)
		}
	})
	return ctx
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
