package callbacks

import (
	"context"
	"errors"
	"testing"
	"time"
)

type capturedEvent struct {
	name    string
	payload any
}

type captureEmitter struct {
	events []capturedEvent
}

func (c *captureEmitter) Emit(event string, payload any) bool {
	c.events = append(c.events, capturedEvent{name: event, payload: payload})
	return true
}

func TestHandlerOnToolCallStartAndEnd(t *testing.T) {
	em := &captureEmitter{}
	h := NewAIEventHandler(em)

	ctx := context.Background()
	ctx = h.OnToolCallStart(ctx, "host_list_inventory", "cid-1", map[string]any{"keyword": "hk"})
	_ = h.OnToolCallEnd(ctx, "host_list_inventory", "cid-1", map[string]any{"ok": true}, nil, 150*time.Millisecond)

	if got := len(em.events); got != 2 {
		t.Fatalf("expected 2 events, got %d", got)
	}
	if em.events[0].name != "tool_call" {
		t.Fatalf("expected first event tool_call, got %s", em.events[0].name)
	}
	start, ok := em.events[0].payload.(ToolCallEvent)
	if !ok {
		t.Fatalf("expected ToolCallEvent payload")
	}
	if start.Tool != "host_list_inventory" || start.CallID != "cid-1" {
		t.Fatalf("unexpected start event: %+v", start)
	}
	if em.events[1].name != "tool_result" {
		t.Fatalf("expected second event tool_result, got %s", em.events[1].name)
	}
	end, ok := em.events[1].payload.(ToolCallEvent)
	if !ok {
		t.Fatalf("expected ToolCallEvent payload")
	}
	if end.Duration <= 0 || end.Error != "" {
		t.Fatalf("unexpected end event: %+v", end)
	}
}

func TestHandlerOnExpertStartAndEnd(t *testing.T) {
	em := &captureEmitter{}
	h := NewAIEventHandler(em)

	ctx := context.Background()
	ctx = h.OnExpertStart(ctx, "sre_expert", "check cluster")
	_ = h.OnExpertEnd(ctx, "sre_expert", 80*time.Millisecond, errors.New("timeout"))

	if got := len(em.events); got != 2 {
		t.Fatalf("expected 2 events, got %d", got)
	}
	start, ok := em.events[0].payload.(ExpertProgressEvent)
	if !ok {
		t.Fatalf("expected ExpertProgressEvent payload")
	}
	if start.Status != "running" || start.Expert != "sre_expert" {
		t.Fatalf("unexpected expert start event: %+v", start)
	}
	end, ok := em.events[1].payload.(ExpertProgressEvent)
	if !ok {
		t.Fatalf("expected ExpertProgressEvent payload")
	}
	if end.Status != "failed" || end.Error == "" {
		t.Fatalf("unexpected expert end event: %+v", end)
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()
	em := &captureEmitter{}
	ctx = WithEmitter(ctx, em)
	if got := EmitterFromContext(ctx); got == nil {
		t.Fatalf("expected emitter from context")
	}
	h := HandlerFromContext(ctx)
	_ = h.OnStreamDelta(ctx, "chunk")
	if len(em.events) != 1 || em.events[0].name != "stream" {
		t.Fatalf("unexpected stream event: %+v", em.events)
	}
}
