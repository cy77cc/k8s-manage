package model

import "testing"

func TestAIApprovalTaskJSONHelpers(t *testing.T) {
	t.Parallel()

	task := &AIApprovalTask{}
	err := task.SetTaskDetail(TaskDetail{
		Summary: "Restart service",
		Steps: []ExecutionStep{
			{Title: "Drain traffic"},
			{Title: "Restart pods"},
		},
		RiskAssessment: RiskAssessment{Level: "medium", Summary: "short interruption"},
		RollbackPlan:   "Scale previous replica set",
	})
	if err != nil {
		t.Fatalf("SetTaskDetail() error = %v", err)
	}
	err = task.SetToolCalls([]ApprovalToolCall{{Name: "service_restart", Arguments: map[string]any{"service_id": 1}}})
	if err != nil {
		t.Fatalf("SetToolCalls() error = %v", err)
	}

	detail, err := task.TaskDetail()
	if err != nil {
		t.Fatalf("TaskDetail() error = %v", err)
	}
	if detail.Summary != "Restart service" {
		t.Fatalf("TaskDetail().Summary = %q", detail.Summary)
	}
	calls, err := task.ToolCalls()
	if err != nil {
		t.Fatalf("ToolCalls() error = %v", err)
	}
	if len(calls) != 1 || calls[0].Name != "service_restart" {
		t.Fatalf("ToolCalls() = %+v", calls)
	}
}
