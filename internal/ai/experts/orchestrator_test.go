package experts

import (
	"context"
	"strings"
	"testing"
)

func TestOrchestratorExecuteSequential(t *testing.T) {
	reg := &fakeRegistry{
		experts: map[string]*Expert{
			"service_expert": {Name: "service_expert"},
			"k8s_expert":     {Name: "k8s_expert"},
		},
	}
	orch := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))
	out, err := orch.Execute(context.Background(), &ExecuteRequest{
		Message: "service unavailable",
		Decision: &RouteDecision{
			PrimaryExpert:   "service_expert",
			OptionalHelpers: []string{"k8s_expert"},
			Strategy:        StrategySequential,
			Source:          "scene",
		},
	})
	if err != nil {
		t.Fatalf("execute sequential: %v", err)
	}
	if out == nil || !strings.Contains(out.Response, "service_expert") {
		t.Fatalf("unexpected execute result: %#v", out)
	}
	if len(out.Traces) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(out.Traces))
	}
}

func TestOrchestratorExecuteParallel(t *testing.T) {
	reg := &fakeRegistry{
		experts: map[string]*Expert{
			"host_expert":    {Name: "host_expert"},
			"monitor_expert": {Name: "monitor_expert"},
		},
	}
	orch := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))
	out, err := orch.Execute(context.Background(), &ExecuteRequest{
		Message: "host issue",
		Decision: &RouteDecision{
			PrimaryExpert:   "host_expert",
			OptionalHelpers: []string{"monitor_expert"},
			Strategy:        StrategyParallel,
			Source:          "keyword",
		},
		RuntimeContext: map[string]any{"timeout_ms": 5000},
	})
	if err != nil {
		t.Fatalf("execute parallel: %v", err)
	}
	if out == nil || len(out.Traces) != 2 {
		t.Fatalf("unexpected execute result: %#v", out)
	}
}

func TestOrchestratorExecuteSingle(t *testing.T) {
	reg := &fakeRegistry{
		experts: map[string]*Expert{
			"general_expert": {Name: "general_expert"},
		},
	}
	orch := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))
	out, err := orch.Execute(context.Background(), &ExecuteRequest{
		Message: "quick check",
		Decision: &RouteDecision{
			PrimaryExpert: "general_expert",
			Strategy:      StrategySingle,
			Source:        "default",
		},
	})
	if err != nil {
		t.Fatalf("execute single: %v", err)
	}
	if out == nil || len(out.Traces) != 1 {
		t.Fatalf("unexpected single execute result: %#v", out)
	}
}

func TestOrchestratorExecutePrimaryLedFallback(t *testing.T) {
	reg := &fakeRegistry{
		experts: map[string]*Expert{
			"service_expert": {Name: "service_expert"},
			"k8s_expert":     {Name: "k8s_expert"},
		},
	}
	orch := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))
	out, err := orch.Execute(context.Background(), &ExecuteRequest{
		Message: "service issue",
		Decision: &RouteDecision{
			PrimaryExpert:   "service_expert",
			OptionalHelpers: []string{"k8s_expert"},
			Strategy:        StrategyPrimaryLed,
			Source:          "scene",
		},
	})
	if err != nil {
		t.Fatalf("execute primary led fallback: %v", err)
	}
	if out == nil || len(out.Traces) != 2 {
		t.Fatalf("unexpected primary led execute result: %#v", out)
	}
}

func TestOrchestratorTracePathIsTrackable(t *testing.T) {
	reg := &fakeRegistry{
		experts: map[string]*Expert{
			"primary_expert": {Name: "primary_expert"},
			"helper_a":       {Name: "helper_a"},
			"helper_b":       {Name: "helper_b"},
		},
	}
	orch := NewOrchestrator(reg, NewResultAggregator(AggregationTemplate, nil))
	out, err := orch.Execute(context.Background(), &ExecuteRequest{
		Message: "trace this request path",
		Decision: &RouteDecision{
			PrimaryExpert:   "primary_expert",
			OptionalHelpers: []string{"helper_a", "helper_b"},
			Strategy:        StrategySequential,
			Source:          "scene",
		},
	})
	if err != nil {
		t.Fatalf("execute sequential trace path: %v", err)
	}
	if out == nil {
		t.Fatalf("expected execute result")
	}
	if got, want := len(out.Traces), 3; got != want {
		t.Fatalf("expected %d traces, got %d", want, got)
	}
	if out.Traces[0].Role != "primary" || out.Traces[0].ExpertName != "primary_expert" {
		t.Fatalf("unexpected primary trace: %#v", out.Traces[0])
	}
	if out.Traces[1].Role != "helper" || out.Traces[2].Role != "helper" {
		t.Fatalf("expected helper roles in trace path: %#v", out.Traces)
	}
	if out.Metadata["strategy"] != StrategySequential {
		t.Fatalf("unexpected strategy metadata: %#v", out.Metadata)
	}
}
