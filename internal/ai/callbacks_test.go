package ai

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
)

func TestNewStreamingCallbacks_OnEnd(t *testing.T) {
	var capturedEvent string
	var capturedPayload map[string]any

	emit := func(event string, payload map[string]any) bool {
		capturedEvent = event
		capturedPayload = payload
		return true
	}

	handler := NewStreamingCallbacks(emit)
	ctx := context.Background()

	// Simulate tool callback
	info := &callbacks.RunInfo{
		Name:      "test_tool",
		Component: components.ComponentOfTool,
	}

	// Call OnEnd
	resultCtx := handler.OnEnd(ctx, info, "tool result")

	if resultCtx == nil {
		t.Fatal("OnEnd returned nil context")
	}

	if capturedEvent != "tool_result" {
		t.Fatalf("expected event 'tool_result', got '%s'", capturedEvent)
	}

	if capturedPayload["tool_name"] != "test_tool" {
		t.Fatalf("expected tool_name 'test_tool', got '%s'", capturedPayload["tool_name"])
	}

	if capturedPayload["status"] != "success" {
		t.Fatalf("expected status 'success', got '%s'", capturedPayload["status"])
	}
}

func TestNewStreamingCallbacks_OnError(t *testing.T) {
	var capturedEvent string
	var capturedPayload map[string]any

	emit := func(event string, payload map[string]any) bool {
		capturedEvent = event
		capturedPayload = payload
		return true
	}

	handler := NewStreamingCallbacks(emit)
	ctx := context.Background()

	// Simulate tool callback error
	info := &callbacks.RunInfo{
		Name:      "failing_tool",
		Component: components.ComponentOfTool,
	}

	testErr := fmtError("tool execution failed")
	resultCtx := handler.OnError(ctx, info, testErr)

	if resultCtx == nil {
		t.Fatal("OnError returned nil context")
	}

	if capturedEvent != "tool_result" {
		t.Fatalf("expected event 'tool_result', got '%s'", capturedEvent)
	}

	if capturedPayload["tool_name"] != "failing_tool" {
		t.Fatalf("expected tool_name 'failing_tool', got '%s'", capturedPayload["tool_name"])
	}

	if capturedPayload["status"] != "error" {
		t.Fatalf("expected status 'error', got '%s'", capturedPayload["status"])
	}

	if capturedPayload["error"] != "tool execution failed" {
		t.Fatalf("expected error message, got '%s'", capturedPayload["error"])
	}
}

func TestNewStreamingCallbacks_NilEmit(t *testing.T) {
	handler := NewStreamingCallbacks(nil)
	ctx := context.Background()

	info := &callbacks.RunInfo{
		Name:      "test_tool",
		Component: components.ComponentOfTool,
	}

	// Should not panic with nil emit
	resultCtx := handler.OnEnd(ctx, info, "result")
	if resultCtx == nil {
		t.Fatal("OnEnd returned nil context")
	}
}

func TestNewStreamingCallbacks_NonToolComponent(t *testing.T) {
	var called bool
	emit := func(event string, payload map[string]any) bool {
		called = true
		return true
	}

	handler := NewStreamingCallbacks(emit)
	ctx := context.Background()

	// Non-tool component should not trigger emission
	info := &callbacks.RunInfo{
		Name:      "chat_model",
		Component: components.ComponentOfChatModel,
	}

	handler.OnEnd(ctx, info, "model response")

	if called {
		t.Fatal("emit should not be called for non-tool components")
	}
}

type fmtError string

func (e fmtError) Error() string { return string(e) }
