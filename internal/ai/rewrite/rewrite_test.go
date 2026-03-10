package rewrite

import (
	"context"
	"testing"
)

type stubRunner struct {
	output string
	err    error
}

func (s stubRunner) Run(_ context.Context, _ string) (string, error) {
	return s.output, s.err
}

func TestHeuristicRewriteTreatsAllHostsAsExplicitTarget(t *testing.T) {
	out := heuristicRewrite(Input{Message: "查看所有主机的状态"})
	if out.OperationMode != "query" {
		t.Fatalf("OperationMode = %s, want query", out.OperationMode)
	}
	if len(out.AmbiguityFlags) != 0 {
		t.Fatalf("AmbiguityFlags = %v, want empty", out.AmbiguityFlags)
	}
	if len(out.DomainHints) != 1 || out.DomainHints[0] != "hostops" {
		t.Fatalf("DomainHints = %v, want [hostops]", out.DomainHints)
	}
}

func TestRewriteConstrainsModelOutputForAllHosts(t *testing.T) {
	r := New(stubRunner{
		output: `{"normalized_goal":"查看所有主机的状态","operation_mode":"query","domain_hints":["k8s","hostops","delivery"],"ambiguity_flags":["resource_target_not_explicit"],"narrative":"bad narrative"}`,
	})
	out, err := r.Rewrite(context.Background(), Input{Message: "查看所有主机的状态"})
	if err != nil {
		t.Fatalf("Rewrite() error = %v", err)
	}
	if len(out.DomainHints) != 1 || out.DomainHints[0] != "hostops" {
		t.Fatalf("DomainHints = %v, want [hostops]", out.DomainHints)
	}
	if len(out.AmbiguityFlags) != 0 {
		t.Fatalf("AmbiguityFlags = %v, want empty", out.AmbiguityFlags)
	}
	if out.Narrative == "" || out.Narrative == "bad narrative" {
		t.Fatalf("Narrative = %q, want rebuilt narrative", out.Narrative)
	}
}
