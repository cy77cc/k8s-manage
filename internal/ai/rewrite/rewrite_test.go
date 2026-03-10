package rewrite

import (
	"context"
	"testing"
)

func TestRewriteFallsBackToStandardizedPassthrough(t *testing.T) {
	out, err := New(nil).Rewrite(context.Background(), Input{Message: "查看所有主机的状态"})
	if err != nil {
		t.Fatalf("Rewrite() error = %v", err)
	}
	if out.RawUserInput != "查看所有主机的状态" {
		t.Fatalf("RawUserInput = %q", out.RawUserInput)
	}
	if out.NormalizedGoal != "查看所有主机的状态" {
		t.Fatalf("NormalizedGoal = %q", out.NormalizedGoal)
	}
	if out.OperationMode != "query" {
		t.Fatalf("OperationMode = %q, want query", out.OperationMode)
	}
	if out.NormalizedRequest.Intent != "user_request" {
		t.Fatalf("Intent = %q, want user_request", out.NormalizedRequest.Intent)
	}
	if len(out.Assumptions) != 1 || out.Assumptions[0] != "rewrite_runner_unavailable" {
		t.Fatalf("Assumptions = %v, want [rewrite_runner_unavailable]", out.Assumptions)
	}
}

func TestNormalizeOutputKeepsModelSemanticsAndFillsBaseFields(t *testing.T) {
	base := buildBaseOutput(Input{
		Message: "查看状态",
		SelectedResources: []SelectedResource{
			{Type: "service", Name: "payment-api"},
		},
	})
	parsed := Output{
		OperationMode: "investigate",
		DomainHints:   []string{"service", "observability"},
		NormalizedRequest: NormalizedRequest{
			Intent: "service_health_check",
		},
		Assumptions: []string{"llm_inferred_scope"},
	}

	out := normalizeOutput(base, parsed)
	if out.RawUserInput != "查看状态" {
		t.Fatalf("RawUserInput = %q", out.RawUserInput)
	}
	if out.NormalizedGoal != "查看状态" {
		t.Fatalf("NormalizedGoal = %q", out.NormalizedGoal)
	}
	if out.ResourceHints.ServiceName != "payment-api" {
		t.Fatalf("ResourceHints.ServiceName = %q, want payment-api", out.ResourceHints.ServiceName)
	}
	if len(out.NormalizedRequest.Targets) != 1 || out.NormalizedRequest.Targets[0].Name != "payment-api" {
		t.Fatalf("Targets = %#v", out.NormalizedRequest.Targets)
	}
	if out.NormalizedRequest.Intent != "service_health_check" {
		t.Fatalf("Intent = %q", out.NormalizedRequest.Intent)
	}
	if out.OperationMode != "investigate" {
		t.Fatalf("OperationMode = %q, want investigate", out.OperationMode)
	}
}

func TestRewriteDetectsNumericResourceIDsFromSelection(t *testing.T) {
	out, err := New(nil).Rewrite(context.Background(), Input{
		Message: "查看 default 命名空间 mysql-0 的日志",
		SelectedResources: []SelectedResource{
			{Type: "cluster", ID: "12", Name: "prod-cluster"},
			{Type: "service", ID: "34", Name: "payment-api"},
			{Type: "host", ID: "56", Name: "node-a"},
		},
	})
	if err != nil {
		t.Fatalf("Rewrite() error = %v", err)
	}
	if out.ResourceHints.ClusterID != 12 {
		t.Fatalf("ClusterID = %d, want 12", out.ResourceHints.ClusterID)
	}
	if out.ResourceHints.ServiceID != 34 {
		t.Fatalf("ServiceID = %d, want 34", out.ResourceHints.ServiceID)
	}
	if out.ResourceHints.HostID != 56 {
		t.Fatalf("HostID = %d, want 56", out.ResourceHints.HostID)
	}
}
