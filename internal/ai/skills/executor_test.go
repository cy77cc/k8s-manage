package skills

import (
	"context"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestExecutorExecute(t *testing.T) {
	skill := Skill{
		Name: "deploy_service",
		Parameters: []SkillParameter{
			{Name: "service_id", Type: "int", Required: true},
			{Name: "env", Type: "string", Required: false, Default: "staging"},
		},
		Steps: []SkillStep{
			{Name: "preview", Type: "tool", Tool: "deployment.release.preview", ParamsTemplate: map[string]any{"service_id": "{{params.service_id}}", "env": "{{params.env}}"}},
			{Name: "approval", Type: "approval"},
			{Name: "execute", Type: "tool", Tool: "deployment.release.apply", ParamsTemplate: map[string]any{"service_id": "{{params.service_id}}", "env": "{{params.env}}"}},
		},
	}

	toolCalls := 0
	approvalCalls := 0
	exec := NewExecutor(
		func(_ context.Context, toolName string, params map[string]any) (tools.ToolResult, error) {
			toolCalls++
			if params["service_id"].(int) != 10 {
				t.Fatalf("unexpected service_id: %#v", params["service_id"])
			}
			if params["env"].(string) == "" {
				t.Fatalf("unexpected env")
			}
			return tools.ToolResult{OK: true, Data: map[string]any{"tool": toolName}}, nil
		},
		func(_ context.Context, _ Skill, _ SkillStep, _ ExecutionState) error {
			approvalCalls++
			return nil
		},
		nil,
	)

	result, err := exec.Execute(context.Background(), skill, map[string]any{"service_id": "10"})
	if err != nil {
		t.Fatalf("execute skill: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected execution success")
	}
	if toolCalls != 2 {
		t.Fatalf("expected 2 tool calls, got %d", toolCalls)
	}
	if approvalCalls != 1 {
		t.Fatalf("expected 1 approval call, got %d", approvalCalls)
	}
}

func TestExecutorExecuteResolverStep(t *testing.T) {
	skill := Skill{
		Name: "batch_exec",
		Parameters: []SkillParameter{{Name: "host_ids", Type: "array<int>", Required: true}},
		Steps: []SkillStep{{Name: "resolve", Type: "resolver"}, {Name: "run", Type: "tool", Tool: "host.batch.exec.apply", ParamsTemplate: map[string]any{"host_ids": "{{params.host_ids}}", "approval_token": "{{params.approval_token}}"}}},
	}
	exec := NewExecutor(
		func(_ context.Context, _ string, params map[string]any) (tools.ToolResult, error) {
			if params["approval_token"].(string) != "token-123" {
				t.Fatalf("expected approval token from resolver")
			}
			return tools.ToolResult{OK: true}, nil
		},
		nil,
		func(_ context.Context, _ Skill, _ SkillStep, _ ExecutionState) (map[string]any, error) {
			return map[string]any{"approval_token": "token-123"}, nil
		},
	)
	if _, err := exec.Execute(context.Background(), skill, map[string]any{"host_ids": "1,2"}); err != nil {
		t.Fatalf("execute with resolver: %v", err)
	}
}
