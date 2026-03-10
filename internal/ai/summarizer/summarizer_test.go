package summarizer

import (
	"context"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/ai/executor"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/runtime"
)

func TestSummarizerMarksNeedMoreInvestigation(t *testing.T) {
	s := New(nil)
	out, err := s.Summarize(context.Background(), Input{
		Message: "check payment-api",
		Plan: &planner.ExecutionPlan{
			PlanID: "plan-1",
			Goal:   "check payment-api",
		},
		State: runtime.ExecutionState{
			PlanID: "plan-1",
		},
		Steps: []executor.StepResult{
			{StepID: "step-1", Status: runtime.StepFailed, Summary: "service check failed"},
		},
	})
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}
	if !out.NeedMoreInvestigation {
		t.Fatalf("NeedMoreInvestigation = false, want true")
	}
	if out.ReplanHint == nil {
		t.Fatalf("ReplanHint = nil, want non-nil")
	}
}

func TestSummarizerExplainsApprovalWait(t *testing.T) {
	s := New(nil)
	out, err := s.Summarize(context.Background(), Input{
		Message: "deploy payment-api",
		State: runtime.ExecutionState{
			PlanID: "plan-2",
			PendingApproval: &runtime.PendingApproval{
				PlanID: "plan-2",
				StepID: "step-2",
				Title:  "发布服务",
				Status: "pending",
			},
		},
	})
	if err != nil {
		t.Fatalf("Summarize() error = %v", err)
	}
	if out.NeedMoreInvestigation {
		t.Fatalf("NeedMoreInvestigation = true, want false")
	}
	if out.Conclusion == "" {
		t.Fatalf("Conclusion is empty")
	}
}
