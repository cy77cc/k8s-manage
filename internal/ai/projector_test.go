package ai

import (
	"testing"

	"github.com/cy77cc/OpsPilot/internal/ai/events"
)

func TestTurnProjectorEmitsNativeAndCompatibilityStageEvents(t *testing.T) {
	meta := events.EventMeta{
		SessionID: "session-1",
		TraceID:   "trace-1",
		TurnID:    "turn-1",
	}
	var streamed []StreamEvent
	emit := func(evt StreamEvent) bool {
		streamed = append(streamed, evt)
		return true
	}
	projector := newTurnProjector(emit, meta, RolloutConfig{UseTurnBlockStreaming: true})

	projector.Start("rewrite")
	emitStageDelta(emit, projector, meta, "rewrite", "loading", "正在理解你的问题", "", "")
	emitStageDelta(emit, projector, meta, "rewrite", "success", "已完成改写", "", "")

	assertEventSeen(t, streamed, events.TurnStarted)
	assertEventSeen(t, streamed, events.TurnState)
	assertEventSeen(t, streamed, events.StageDelta)
	assertEventSeen(t, streamed, events.BlockOpen)
	assertEventSeen(t, streamed, events.BlockDelta)
	assertEventSeen(t, streamed, events.BlockClose)

	for _, evt := range streamed {
		if evt.Type == events.StageDelta || evt.Type == events.BlockOpen || evt.Type == events.BlockDelta {
			if got := stringValue(evt.Data["turn_id"]); got != "turn-1" {
				t.Fatalf("%s turn_id = %q, want turn-1", evt.Type, got)
			}
		}
	}
}

func TestTurnProjectorDisabledWhenRolloutOff(t *testing.T) {
	meta := events.EventMeta{
		SessionID: "session-1",
		TraceID:   "trace-1",
		TurnID:    "turn-1",
	}
	called := false
	projector := newTurnProjector(func(evt StreamEvent) bool {
		called = true
		return true
	}, meta, RolloutConfig{})

	projector.Start("rewrite")
	projector.StageDelta("rewrite", "loading", "hello", "", "")
	projector.Done("completed", "done")

	if called {
		t.Fatal("expected rollout-disabled projector to suppress native events")
	}
}

func assertEventSeen(t *testing.T, streamed []StreamEvent, name events.Name) {
	t.Helper()
	for _, evt := range streamed {
		if evt.Type == name {
			return
		}
	}
	t.Fatalf("event %s not found in stream: %#v", name, streamed)
}
