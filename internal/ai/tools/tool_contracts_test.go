package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type stubMemoryAccessor struct {
	params map[string]map[string]any
}

func (s *stubMemoryAccessor) GetLastToolParams(toolName string) map[string]any {
	if s == nil {
		return nil
	}
	return s.params[toolName]
}

func (s *stubMemoryAccessor) SetLastToolParams(toolName string, params map[string]any) {
	if s.params == nil {
		s.params = map[string]map[string]any{}
	}
	s.params[toolName] = params
}

func TestApprovalAndConfirmationErrors(t *testing.T) {
	approval := &ApprovalRequiredError{Message: "need approval"}
	if approval.Error() != "need approval" {
		t.Fatalf("unexpected approval error text: %q", approval.Error())
	}
	if (&ApprovalRequiredError{}).Error() != "approval required" {
		t.Fatalf("expected default approval message")
	}
	if got, ok := IsApprovalRequired(approval); !ok || got != approval {
		t.Fatalf("expected approval required detection")
	}

	confirm := &ConfirmationRequiredError{
		Token:     "cf-1",
		Tool:      "service_deploy",
		Preview:   map[string]any{"service": "demo"},
		ExpiresAt: time.Now(),
	}
	if confirm.Error() != "confirmation required" {
		t.Fatalf("expected default confirmation error text")
	}
	if got, ok := IsConfirmationRequired(confirm); !ok || got != confirm {
		t.Fatalf("expected confirmation required detection")
	}
}

func TestToolContextHelpers(t *testing.T) {
	ctx := context.Background()
	calls := 0
	var emittedEvent string
	var emittedPayload any

	ctx = WithToolPolicyChecker(ctx, func(_ context.Context, meta ToolMeta, params map[string]any) error {
		calls++
		if meta.Name != "host_exec" || params["target"] != "localhost" {
			t.Fatalf("unexpected checker input: meta=%#v params=%#v", meta, params)
		}
		return nil
	})
	ctx = WithToolEventEmitter(ctx, func(event string, payload any) {
		emittedEvent = event
		emittedPayload = payload
	})
	ctx = WithToolUser(ctx, 42, "approval-token")
	ctx = WithToolRuntimeContext(ctx, map[string]any{"scene": "ops"})
	mem := &stubMemoryAccessor{}
	ctx = WithToolMemoryAccessor(ctx, mem)

	if err := CheckToolPolicy(ctx, ToolMeta{Name: "host_exec"}, map[string]any{"target": "localhost"}); err != nil {
		t.Fatalf("check policy: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected checker to run once, got %d", calls)
	}
	EmitToolEvent(ctx, "tool_call", map[string]any{"tool": "host_exec"})
	if emittedEvent != "tool_call" {
		t.Fatalf("unexpected emitted event: %q", emittedEvent)
	}
	if payload, ok := emittedPayload.(map[string]any); !ok || payload["tool"] != "host_exec" {
		t.Fatalf("unexpected emitted payload: %#v", emittedPayload)
	}

	uid, token := ToolUserFromContext(ctx)
	if uid != 42 || token != "approval-token" {
		t.Fatalf("unexpected tool user context: uid=%d token=%q", uid, token)
	}
	if ToolRuntimeContextFromContext(ctx)["scene"] != "ops" {
		t.Fatalf("unexpected runtime context")
	}
	if ToolMemoryAccessorFromContext(ctx) != mem {
		t.Fatalf("expected memory accessor round-trip")
	}

	if err := CheckToolPolicy(context.Background(), ToolMeta{}, nil); err != nil {
		t.Fatalf("expected nil policy checker to no-op: %v", err)
	}
	EmitToolEvent(context.Background(), "noop", nil)
	if ToolMemoryAccessorFromContext(context.Background()) != nil {
		t.Fatalf("expected nil memory accessor from empty context")
	}
}

func TestMarshalToolResultAndInputErrors(t *testing.T) {
	raw, err := MarshalToolResult(ToolResult{
		OK:        true,
		Data:      map[string]any{"service": "demo"},
		Source:    "unit-test",
		LatencyMS: 12,
	})
	if err != nil {
		t.Fatalf("marshal tool result: %v", err)
	}
	var decoded ToolResult
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("decode marshaled tool result: %v", err)
	}
	if !decoded.OK || decoded.Source != "unit-test" {
		t.Fatalf("unexpected decoded tool result: %#v", decoded)
	}

	err = NewMissingParam("target", "target is required")
	if ie, ok := AsToolInputError(err); !ok || ie.Code != "missing_param" {
		t.Fatalf("expected missing param input error, got %#v", err)
	}
	if NewInvalidParam("limit", "").Error() != "invalid_param: limit" {
		t.Fatalf("unexpected invalid param fallback text")
	}
	if NewParamConflict("namespace", "conflict").Error() != "conflict" {
		t.Fatalf("unexpected param conflict text")
	}
	if _, ok := AsToolInputError(errors.New("plain error")); ok {
		t.Fatalf("did not expect plain error to unwrap as tool input error")
	}
}
