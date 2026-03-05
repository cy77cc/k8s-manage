package ai

import (
	"errors"
	"fmt"
	"io"
	"strings"

	adkcore "github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/gin-gonic/gin"
)

func (h *handler) processADKEvent(
	emit func(event string, payload gin.H) bool,
	tracker *toolEventTracker,
	event *adkcore.AgentEvent,
	assistantContent *strings.Builder,
	reasoningContent *strings.Builder,
) error {
	if event == nil {
		return nil
	}
	if event.Err != nil {
		return event.Err
	}

	if event.Output != nil && event.Output.MessageOutput != nil {
		if err := h.handleStream(emit, tracker, event.Output.MessageOutput, assistantContent, reasoningContent); err != nil {
			return err
		}
	}

	if event.Action != nil {
		if err := h.handleAction(emit, event.Action); err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) handleStream(
	emit func(event string, payload gin.H) bool,
	tracker *toolEventTracker,
	output *adkcore.MessageVariant,
	assistantContent *strings.Builder,
	reasoningContent *strings.Builder,
) error {
	if output == nil {
		return nil
	}
	if output.Message != nil {
		h.applyMessageChunk(emit, tracker, output.Message, assistantContent, reasoningContent)
		return nil
	}
	if output.MessageStream == nil {
		return nil
	}
	defer output.MessageStream.Close()
	for {
		chunk, err := output.MessageStream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if chunk == nil {
			continue
		}
		h.applyMessageChunk(emit, tracker, chunk, assistantContent, reasoningContent)
	}
}

func (h *handler) applyMessageChunk(
	emit func(event string, payload gin.H) bool,
	tracker *toolEventTracker,
	msg *schema.Message,
	assistantContent *strings.Builder,
	reasoningContent *strings.Builder,
) {
	if msg.ReasoningContent != "" {
		reasoningContent.WriteString(msg.ReasoningContent)
		_ = emit("thinking_delta", gin.H{"contentChunk": msg.ReasoningContent})
	}
	if msg.Content != "" {
		assistantContent.WriteString(msg.Content)
		emitDeltaChunks(emit, msg.Content)
	}
	for _, tc := range msg.ToolCalls {
		toolName := strings.TrimSpace(tc.Function.Name)
		callID := strings.TrimSpace(tc.ID)
		tracker.noteCall(callID, toolName)
		_ = emit("tool_call", gin.H{
			"tool":    toolName,
			"call_id": callID,
			"payload": tc,
		})
	}
	if msg.Role == schema.Tool {
		tracker.noteResult("", "")
		_ = emit("tool_result", gin.H{"content": msg.Content})
	}
}

func emitDeltaChunks(emit func(event string, payload gin.H) bool, content string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	const maxRunesPerChunk = 24
	rs := []rune(content)
	for i := 0; i < len(rs); i += maxRunesPerChunk {
		end := i + maxRunesPerChunk
		if end > len(rs) {
			end = len(rs)
		}
		_ = emit("delta", gin.H{"contentChunk": string(rs[i:end])})
	}
}

func (h *handler) handleInterrupt(emit func(event string, payload gin.H) bool, info *adkcore.InterruptInfo) error {
	if info == nil {
		return nil
	}
	switch data := info.Data.(type) {
	case *aitools.ApprovalInfo:
		_ = emit("approval_required", gin.H{
			"tool":      data.ToolName,
			"arguments": data.ArgumentsInJSON,
			"risk":      data.Risk,
			"preview":   data.Preview,
		})
	case *aitools.ReviewEditInfo:
		_ = emit("review_required", gin.H{
			"tool":      data.ToolName,
			"arguments": data.ArgumentsInJSON,
		})
	default:
		if len(info.InterruptContexts) > 0 {
			_ = emit("interrupt_required", gin.H{"contexts": info.InterruptContexts})
		} else {
			_ = emit("interrupt_required", gin.H{"message": fmt.Sprintf("interrupt: %v", data)})
		}
	}
	return nil
}

func (h *handler) handleAction(emit func(event string, payload gin.H) bool, action *adkcore.AgentAction) error {
	if action == nil {
		return nil
	}
	if action.Interrupted != nil {
		return h.handleInterrupt(emit, action.Interrupted)
	}
	if action.Exit {
		_ = emit("done", gin.H{"reason": "agent_exit"})
	}
	if action.TransferToAgent != nil {
		_ = emit("agent_transfer", gin.H{"dest_agent": action.TransferToAgent.DestAgentName})
	}
	return nil
}
