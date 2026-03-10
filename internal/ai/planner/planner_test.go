package planner

import (
	"context"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/ai/rewrite"
)

func TestPlanFallsBackToMinimalPlan(t *testing.T) {
	out, err := New(nil).Plan(context.Background(), Input{
		Message: "查看所有主机的状态",
		Rewrite: rewrite.Output{
			NormalizedGoal: "查看所有主机的状态",
			OperationMode:  "query",
			NormalizedRequest: rewrite.NormalizedRequest{
				Intent: "service_health_check",
				Targets: []rewrite.RequestTarget{
					{Type: "host", Name: "all"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if out.Type != DecisionPlan {
		t.Fatalf("Type = %s, want %s", out.Type, DecisionPlan)
	}
	if out.Plan == nil || len(out.Plan.Steps) != 1 {
		t.Fatalf("Plan = %#v", out.Plan)
	}
	if out.Plan.Steps[0].Expert != "hostops" {
		t.Fatalf("Expert = %q, want hostops", out.Plan.Steps[0].Expert)
	}
	if out.Plan.Steps[0].Task != "查看所有主机的状态" {
		t.Fatalf("Task = %q", out.Plan.Steps[0].Task)
	}
}

func TestPlanFallsBackToClarifyWhenRewriteStillAmbiguous(t *testing.T) {
	out, err := New(nil).Plan(context.Background(), Input{
		Message: "帮我看看状态",
		Rewrite: rewrite.Output{
			AmbiguityFlags: []string{"resource_target_not_explicit"},
		},
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if out.Type != DecisionClarify {
		t.Fatalf("Type = %s, want %s", out.Type, DecisionClarify)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("Candidates = %#v", out.Candidates)
	}
}

func TestNormalizeDecisionDoesNotPanicWhenBaseHasNoPlan(t *testing.T) {
	base := Decision{
		Type:      DecisionClarify,
		Message:   "need more info",
		Narrative: "clarify first",
	}
	parsed := Decision{
		Type: DecisionPlan,
		Plan: &ExecutionPlan{
			Goal: "check mysql-0",
		},
	}

	out := normalizeDecision(base, parsed)
	if out.Plan == nil {
		t.Fatalf("Plan is nil")
	}
	if out.Plan.Goal != "check mysql-0" {
		t.Fatalf("Goal = %q", out.Plan.Goal)
	}
}
