package ai

import (
	"testing"

	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestShouldSkipToolApproval(t *testing.T) {
	meta := aitools.ToolMeta{Name: "service_deploy", Mode: aitools.ToolModeMutating, Risk: aitools.ToolRiskHigh}

	if !shouldSkipToolApproval(meta, map[string]any{"preview": true, "apply": false}) {
		t.Fatal("expected preview-only service_deploy to skip approval")
	}
	if shouldSkipToolApproval(meta, map[string]any{"preview": true, "apply": true}) {
		t.Fatal("expected apply service_deploy to require approval")
	}
	if shouldSkipToolApproval(aitools.ToolMeta{Name: "host_batch"}, map[string]any{"preview": true}) {
		t.Fatal("expected non-service_deploy tool to keep approval")
	}
}
