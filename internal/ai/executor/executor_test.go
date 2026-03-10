package executor

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/runtime"
)

func newExecutionStore(t *testing.T) *runtime.ExecutionStore {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
		mr.Close()
	})
	return runtime.NewExecutionStore(client, "ai:test:execution:")
}

func TestExecutorApprovalResumeFlow(t *testing.T) {
	store := newExecutionStore(t)
	exec := New(store)
	ctx := context.Background()

	result, err := exec.Run(ctx, Request{
		TraceID:   "trace-1",
		SessionID: "session-1",
		Message:   "deploy payment-api",
		Plan: planner.ExecutionPlan{
			PlanID: "plan-1",
			Goal:   "deploy payment-api",
			Steps: []planner.PlanStep{
				{
					StepID: "step-1",
					Title:  "发布服务",
					Expert: "service",
					Mode:   "mutating",
					Risk:   "high",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.PendingApproval == nil {
		t.Fatalf("expected pending approval")
	}
	if got := result.State.Status; got != runtime.ExecutionStatusWaitingApproval {
		t.Fatalf("state status = %s, want %s", got, runtime.ExecutionStatusWaitingApproval)
	}

	resumed, err := exec.Resume(ctx, ResumeRequest{
		SessionID: "session-1",
		PlanID:    "plan-1",
		StepID:    "step-1",
		Approved:  true,
	})
	if err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if got := resumed.State.Status; got != runtime.ExecutionStatusCompleted {
		t.Fatalf("resumed state status = %s, want %s", got, runtime.ExecutionStatusCompleted)
	}
	if got := resumed.State.Steps["step-1"].Status; got != runtime.StepCompleted {
		t.Fatalf("step status = %s, want %s", got, runtime.StepCompleted)
	}
}
