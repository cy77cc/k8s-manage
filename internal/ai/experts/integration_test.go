package experts

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/tool"
)

func buildTestTools(t *testing.T) map[string]tool.InvokableTool {
	t.Helper()
	return map[string]tool.InvokableTool{
		"host_list_inventory":      mustTool(t, "host_list_inventory"),
		"host_ssh_exec_readonly":   mustTool(t, "host_ssh_exec_readonly"),
		"k8s_list_resources":       mustTool(t, "k8s_list_resources"),
		"k8s_get_events":           mustTool(t, "k8s_get_events"),
		"service_get_detail":       mustTool(t, "service_get_detail"),
		"service_deploy_preview":   mustTool(t, "service_deploy_preview"),
		"monitor_metric_query":     mustTool(t, "monitor_metric_query"),
		"audit_log_search":         mustTool(t, "audit_log_search"),
		"permission_check":         mustTool(t, "permission_check"),
		"deployment_target_detail": mustTool(t, "deployment_target_detail"),
	}
}

func TestHybridMOEPipelineE2E(t *testing.T) {
	reg, err := NewExpertRegistry(context.Background(), "configs/experts.yaml", buildTestTools(t), nil)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	router, err := NewHybridRouter(reg, "configs/scene_mappings.yaml")
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	orch := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))

	decision := router.Route(context.Background(), &RouteRequest{
		Message: "请排查服务发布失败",
		Scene:   "scene:services:deploy",
	})
	if decision == nil || decision.Source != "scene" || decision.PrimaryExpert == "" {
		t.Fatalf("unexpected route decision: %#v", decision)
	}

	result, err := orch.Execute(context.Background(), &ExecuteRequest{
		Message:  "请排查服务发布失败",
		Decision: decision,
		RuntimeContext: map[string]any{
			"timeout_ms": 3000,
		},
	})
	if err != nil {
		t.Fatalf("orchestrator execute: %v", err)
	}
	if result == nil || !strings.Contains(result.Response, decision.PrimaryExpert) {
		t.Fatalf("unexpected orchestration result: %#v", result)
	}
}

func TestHybridMOEPipelineErrorCase(t *testing.T) {
	reg, err := NewExpertRegistry(context.Background(), "configs/experts.yaml", buildTestTools(t), nil)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	orch := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))
	_, err = orch.Execute(context.Background(), &ExecuteRequest{
		Message: "test",
		Decision: &RouteDecision{
			PrimaryExpert: "non_exist_expert",
			Strategy:      StrategySingle,
			Source:        "scene",
		},
	})
	if err == nil {
		t.Fatalf("expected error for unknown expert")
	}
}

func TestMultiExpertCollaborationScenario(t *testing.T) {
	reg := &fakeRegistry{
		experts: map[string]*Expert{
			"service_expert":  {Name: "service_expert"},
			"k8s_expert":      {Name: "k8s_expert"},
			"monitor_expert":  {Name: "monitor_expert"},
			"topology_expert": {Name: "topology_expert"},
		},
	}
	orch := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))
	out, err := orch.Execute(context.Background(), &ExecuteRequest{
		Message: "发布后服务不可用，需要从依赖、监控、拓扑联合排查",
		Decision: &RouteDecision{
			PrimaryExpert:   "service_expert",
			OptionalHelpers: []string{"k8s_expert", "monitor_expert", "topology_expert"},
			Strategy:        StrategyParallel,
			Source:          "scene",
		},
		RuntimeContext: map[string]any{"timeout_ms": 5000},
	})
	if err != nil {
		t.Fatalf("execute multi expert collaboration: %v", err)
	}
	if out == nil {
		t.Fatalf("expected execute result")
	}
	if got, want := len(out.Traces), 4; got != want {
		t.Fatalf("expected %d traces, got %d", want, got)
	}
	for _, expertName := range []string{"service_expert", "k8s_expert", "monitor_expert", "topology_expert"} {
		if !strings.Contains(out.Response, expertName) {
			t.Fatalf("expected response includes expert %s, got: %s", expertName, out.Response)
		}
	}
}

func TestComplexSequentialScenarioTraceability(t *testing.T) {
	reg := &fakeRegistry{
		experts: map[string]*Expert{
			"primary_expert": {Name: "primary_expert"},
			"helper_1":       {Name: "helper_1"},
			"helper_2":       {Name: "helper_2"},
			"helper_3":       {Name: "helper_3"},
		},
	}
	orch := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))
	out, err := orch.Execute(context.Background(), &ExecuteRequest{
		Message: "复杂场景：跨域排查并输出可追踪的协作路径",
		Decision: &RouteDecision{
			PrimaryExpert:   "primary_expert",
			OptionalHelpers: []string{"helper_1", "helper_2", "helper_3"},
			Strategy:        StrategySequential,
			Source:          "scene",
		},
	})
	if err != nil {
		t.Fatalf("execute complex sequential scenario: %v", err)
	}
	if out == nil {
		t.Fatalf("expected execute result")
	}
	if got, want := len(out.Traces), 4; got != want {
		t.Fatalf("expected %d traces, got %d", want, got)
	}
	if out.Traces[0].Role != "primary" {
		t.Fatalf("expected first trace role=primary, got=%s", out.Traces[0].Role)
	}
	for i := 1; i < len(out.Traces); i++ {
		if out.Traces[i].Role != "helper" {
			t.Fatalf("expected helper role at index %d, got=%s", i, out.Traces[i].Role)
		}
	}
	if out.Metadata["strategy"] != StrategySequential {
		t.Fatalf("unexpected strategy metadata: %#v", out.Metadata)
	}
}
