package ai

import (
	"strings"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestPreviewBuilderBuildPreview(t *testing.T) {
	b := NewPreviewBuilder(nil, []tools.ToolMeta{
		{
			Name:        "service_deploy_apply",
			Description: "deploy service",
			Mode:        tools.ToolModeMutating,
			Risk:        tools.ToolRiskHigh,
		},
	})
	out := b.BuildPreview("service_deploy_apply", map[string]any{
		"service_id":      12,
		"cluster_id":      2,
		"current_version": "v1.0.0",
		"target_version":  "v1.1.0",
	})
	if out.ToolName != "service_deploy_apply" {
		t.Fatalf("unexpected tool name: %s", out.ToolName)
	}
	if out.RiskLevel != "high" {
		t.Fatalf("unexpected risk level: %s", out.RiskLevel)
	}
	if out.Mode != "mutating" {
		t.Fatalf("unexpected mode: %s", out.Mode)
	}
	if out.Timeout <= 0 {
		t.Fatalf("timeout not set")
	}
	if len(out.TargetResources) != 2 {
		t.Fatalf("expected 2 target resources, got %d", len(out.TargetResources))
	}
	if !strings.Contains(out.PreviewDiff, "v1.0.0 -> v1.1.0") {
		t.Fatalf("unexpected diff preview: %s", out.PreviewDiff)
	}
}

func TestPreviewBuilderExtractTargetResources(t *testing.T) {
	b := NewPreviewBuilder(nil, nil)
	out := b.extractTargetResources("host_batch_exec_apply", map[string]any{
		"host_ids": []any{1, 2, "3"},
	})
	if got, want := len(out), 3; got != want {
		t.Fatalf("expected %d resources, got %d", want, got)
	}
	for _, item := range out {
		if item.Type != "host" {
			t.Fatalf("expected host type, got %s", item.Type)
		}
	}
}

func TestPreviewBuilderGenerateImpactScope(t *testing.T) {
	b := NewPreviewBuilder(nil, nil)
	scope := b.generateImpactScope("host_batch_exec_apply", []TargetResource{
		{Type: "host", ID: "1"},
		{Type: "host", ID: "2"},
	}, map[string]any{})
	if !strings.Contains(scope, "host:2") {
		t.Fatalf("unexpected scope: %s", scope)
	}
	if !strings.Contains(scope, "批量操作") {
		t.Fatalf("expected batch marker: %s", scope)
	}
}
