package graph

import (
	"context"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/ai/experts"
)

func TestBuilderBuildAndCompile(t *testing.T) {
	b := NewBuilder()
	g, err := b.Build(context.Background())
	if err != nil {
		t.Fatalf("build graph: %v", err)
	}
	if g == nil {
		t.Fatalf("graph is nil")
	}

	r, err := g.Compile(context.Background())
	if err != nil {
		t.Fatalf("compile graph: %v", err)
	}
	out, err := r.Invoke(context.Background(), &GraphInput{Message: "check cluster"})
	if err != nil {
		t.Fatalf("invoke graph: %v", err)
	}
	if out == nil {
		t.Fatalf("output is nil")
	}
	if out.Response == "" {
		t.Fatalf("empty response")
	}
}

type fakePrimaryRunner struct{}

func (fakePrimaryRunner) RunPrimary(context.Context, *experts.ExecuteRequest) (string, error) {
	return "primary answer", nil
}

type fakeHelperRunner struct{}

func (fakeHelperRunner) RunHelper(_ context.Context, _ *experts.ExecuteRequest, helper experts.HelperRequest) (experts.ExpertResult, error) {
	return experts.ExpertResult{ExpertName: helper.ExpertName, Output: helper.Task}, nil
}

func TestBuilderBranchByStrategy(t *testing.T) {
	b := NewBuilderWithRunners(fakePrimaryRunner{}, fakeHelperRunner{})
	g, err := b.Build(context.Background())
	if err != nil {
		t.Fatalf("build graph: %v", err)
	}
	r, err := g.Compile(context.Background())
	if err != nil {
		t.Fatalf("compile graph: %v", err)
	}

	cases := []struct {
		name     string
		strategy experts.ExecutionStrategy
		wantN    int
	}{
		{name: "parallel helpers", strategy: experts.StrategyParallel, wantN: 2},
		{name: "sequential helpers", strategy: experts.StrategySequential, wantN: 2},
		{name: "skip helpers", strategy: experts.StrategySingle, wantN: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &GraphInput{
				Message:  "diag issue",
				Request:  &experts.ExecuteRequest{Message: "diag issue"},
				Strategy: tc.strategy,
				HelperRequests: []experts.HelperRequest{
					{ExpertName: "sre", Task: "check metrics"},
					{ExpertName: "k8s", Task: "check pods"},
				},
			}
			if tc.wantN == 0 {
				in.HelperRequests = nil
			}
			out, err := r.Invoke(context.Background(), in)
			if err != nil {
				t.Fatalf("invoke graph: %v", err)
			}
			if out.Response != "primary answer" {
				t.Fatalf("unexpected response: %q", out.Response)
			}
			if len(out.Results) != tc.wantN {
				t.Fatalf("expected %d helper results, got %d", tc.wantN, len(out.Results))
			}
		})
	}
}
