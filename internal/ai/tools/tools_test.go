package tools

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	approvaltools "github.com/cy77cc/OpsPilot/internal/ai/tools/approval"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
)

func TestNewAllToolsIncludesMutatingToolsWrappedWithApprovalGate(t *testing.T) {
	tools := NewAllTools(context.Background(), common.PlatformDeps{})
	if len(tools) == 0 {
		t.Fatalf("toolset is empty")
	}

	assertWrappedMutatingToolPresent(t, tools, "service_deploy_apply")
	assertWrappedMutatingToolPresent(t, tools, "host_batch_exec_apply")
	assertWrappedMutatingToolPresent(t, tools, "cicd_pipeline_trigger")
}

func TestInferToolModeFallsBackToMutatingPatterns(t *testing.T) {
	mode, risk := inferToolSpec("cluster_restart_workload")
	if mode != "mutating" {
		t.Fatalf("mode = %q, want mutating", mode)
	}
	if risk == "" {
		t.Fatalf("risk = empty, want inferred risk")
	}
}

func assertWrappedMutatingToolPresent(t *testing.T, tools []tool.BaseTool, target string) {
	t.Helper()
	for _, candidate := range tools {
		invokable, ok := candidate.(tool.InvokableTool)
		if !ok {
			continue
		}
		info, err := invokable.Info(context.Background())
		if err != nil || info == nil || info.Name != target {
			continue
		}
		if _, ok := candidate.(*approvaltools.Gate); !ok {
			t.Fatalf("tool %q is %T, want *approval.Gate", target, candidate)
		}
		return
	}
	t.Fatalf("tool %q not found in toolset", target)
}
