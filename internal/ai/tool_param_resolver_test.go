package ai

import (
	"context"
	"testing"
)

type fakeAccessor struct {
	last map[string]any
}

func (f *fakeAccessor) GetLastToolParams(_ string) map[string]any {
	return f.last
}

func (f *fakeAccessor) SetLastToolParams(_ string, params map[string]any) {
	f.last = params
}

func TestResolveToolParamsPriority(t *testing.T) {
	ctx := context.Background()
	ctx = WithToolRuntimeContext(ctx, map[string]any{
		"namespace": "ns-runtime",
		"target":    "10.0.0.10",
	})
	ctx = WithToolMemoryAccessor(ctx, &fakeAccessor{last: map[string]any{
		"namespace": "ns-memory",
		"limit":     20,
	}})

	meta := ToolMeta{
		Name:        "k8s_get_events",
		DefaultHint: map[string]any{"namespace": "ns-default", "limit": 50},
	}
	resolved, _ := resolveToolParams(ctx, meta, map[string]any{}, "")
	if resolved["namespace"] != "ns-runtime" {
		t.Fatalf("expected runtime namespace, got %v", resolved["namespace"])
	}
	if resolved["limit"] != 20 {
		t.Fatalf("expected memory limit, got %v", resolved["limit"])
	}
	if resolved["target"] != "10.0.0.10" {
		t.Fatalf("expected runtime target, got %v", resolved["target"])
	}
}
