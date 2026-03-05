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
	decision := router.Route(context.Background(), &RouteRequest{
		Message: "请排查服务发布失败",
		Scene:   "scene:services:deploy",
	})
	if decision == nil || decision.Source != "scene" || decision.PrimaryExpert == "" {
		t.Fatalf("unexpected route decision: %#v", decision)
	}

	exec := NewExpertExecutor(reg)
	result, err := exec.ExecuteStep(context.Background(), &ExecutionStep{
		ExpertName: decision.PrimaryExpert,
		Task:       "执行主专家分析",
	}, &ExecuteRequest{Message: "请排查服务发布失败", Decision: decision}, nil)
	if err != nil {
		t.Fatalf("execute step: %v", err)
	}
	if result == nil || !strings.Contains(result.Output, "专家模型未初始化") {
		t.Fatalf("unexpected execution result: %#v", result)
	}
}

func TestHybridMOEPipelineErrorCase(t *testing.T) {
	reg, err := NewExpertRegistry(context.Background(), "configs/experts.yaml", buildTestTools(t), nil)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	exec := NewExpertExecutor(reg)
	_, err = exec.ExecuteStep(context.Background(), &ExecutionStep{
		ExpertName: "non_exist_expert",
		Task:       "test",
	}, &ExecuteRequest{Message: "test"}, nil)
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
	exec := NewExpertExecutor(reg)
	steps := []ExecutionStep{
		{ExpertName: "service_expert", Task: "主专家分析"},
		{ExpertName: "k8s_expert", Task: "k8s 协助"},
		{ExpertName: "monitor_expert", Task: "监控协助"},
		{ExpertName: "topology_expert", Task: "拓扑协助"},
	}
	results := make([]ExpertResult, 0, len(steps))
	for i := range steps {
		out, err := exec.ExecuteStep(context.Background(), &steps[i], &ExecuteRequest{Message: "联合排查"}, results)
		if err != nil {
			t.Fatalf("execute step %d: %v", i, err)
		}
		results = append(results, *out)
	}
	if got, want := len(results), 4; got != want {
		t.Fatalf("expected %d results, got %d", want, got)
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
	exec := NewExpertExecutor(reg)
	prior := make([]ExpertResult, 0, 4)
	seq := []ExecutionStep{
		{ExpertName: "primary_expert", Task: "主分析"},
		{ExpertName: "helper_1", Task: "helper_1", ContextFrom: []int{0}},
		{ExpertName: "helper_2", Task: "helper_2", ContextFrom: []int{0, 1}},
		{ExpertName: "helper_3", Task: "helper_3", ContextFrom: []int{0, 1, 2}},
	}
	for i := range seq {
		out, err := exec.ExecuteStep(context.Background(), &seq[i], &ExecuteRequest{Message: "复杂场景"}, prior)
		if err != nil {
			t.Fatalf("execute sequential step %d: %v", i, err)
		}
		prior = append(prior, *out)
	}
	if got, want := len(prior), 4; got != want {
		t.Fatalf("expected %d traces, got %d", want, got)
	}
}
