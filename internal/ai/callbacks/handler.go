package callbacks

import (
	"context"
	"time"
)

// AIEventHandler emits unified AI runtime events.
type AIEventHandler struct {
	emitter EventEmitter
}

func NewAIEventHandler(emitter EventEmitter) *AIEventHandler {
	if emitter == nil {
		emitter = NopEmitter
	}
	return &AIEventHandler{emitter: emitter}
}

func (h *AIEventHandler) OnToolCallStart(ctx context.Context, tool, callID string, args map[string]any) context.Context {
	h.emitter.Emit("tool_call", ToolCallEvent{
		Tool:      tool,
		CallID:    callID,
		Arguments: args,
		Timestamp: time.Now().UTC(),
	})
	return ctx
}

func (h *AIEventHandler) OnToolCallEnd(ctx context.Context, tool, callID string, result any, err error, duration time.Duration) context.Context {
	event := ToolCallEvent{
		Tool:      tool,
		CallID:    callID,
		Result:    result,
		Timestamp: time.Now().UTC(),
		Duration:  duration,
	}
	if err != nil {
		event.Error = err.Error()
	}
	h.emitter.Emit("tool_result", event)
	return ctx
}

func (h *AIEventHandler) OnExpertStart(ctx context.Context, expert, task string) context.Context {
	h.emitter.Emit("expert_progress", ExpertProgressEvent{
		Expert: expert,
		Status: "running",
		Task:   task,
	})
	return ctx
}

func (h *AIEventHandler) OnExpertEnd(ctx context.Context, expert string, duration time.Duration, err error) context.Context {
	event := ExpertProgressEvent{
		Expert:     expert,
		Status:     "done",
		DurationMs: duration.Milliseconds(),
	}
	if err != nil {
		event.Status = "failed"
		event.Error = err.Error()
	}
	h.emitter.Emit("expert_progress", event)
	return ctx
}

func (h *AIEventHandler) OnStreamDelta(ctx context.Context, content string) context.Context {
	h.emitter.Emit("stream", StreamEvent{
		Type:      "delta",
		Content:   content,
		Timestamp: time.Now().UTC(),
	})
	return ctx
}
