package ai

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/cy77cc/OpsPilot/internal/ai/executor"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/runtime"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
)

func newExecutionStoreForOrchestrator(t *testing.T) *runtime.ExecutionStore {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
		mr.Close()
	})
	return runtime.NewExecutionStore(client, "ai:test:execution:")
}

func TestResumeReturnsIdempotentStatus(t *testing.T) {
	store := newExecutionStoreForOrchestrator(t)
	exec := executor.New(store)
	ctx := context.Background()

	_, err := exec.Run(ctx, executor.Request{
		TraceID:   "trace-2",
		SessionID: "session-2",
		Message:   "deploy payment-api",
		Plan: planner.ExecutionPlan{
			PlanID: "plan-2",
			Goal:   "deploy payment-api",
			Steps: []planner.PlanStep{
				{
					StepID: "step-2",
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

	orch := NewOrchestrator(nil, store, common.PlatformDeps{})
	first, err := orch.Resume(ctx, ResumeRequest{
		SessionID: "session-2",
		PlanID:    "plan-2",
		StepID:    "step-2",
		Approved:  true,
	})
	if err != nil {
		t.Fatalf("first Resume() error = %v", err)
	}
	if first.Status == "idempotent" {
		t.Fatalf("first resume unexpectedly idempotent")
	}

	second, err := orch.Resume(ctx, ResumeRequest{
		SessionID: "session-2",
		PlanID:    "plan-2",
		StepID:    "step-2",
		Approved:  true,
	})
	if err != nil {
		t.Fatalf("second Resume() error = %v", err)
	}
	if second.Status != "idempotent" {
		t.Fatalf("second resume status = %s, want idempotent", second.Status)
	}
}
