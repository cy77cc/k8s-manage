package rag

import (
	"testing"

	"github.com/cy77cc/OpsPilot/internal/ai/rewrite"
)

func TestBuildRewriteQueryEnvelopeUsesRewriteSemanticFields(t *testing.T) {
	envelope := BuildRewriteQueryEnvelope(rewrite.Output{
		RawUserInput:      "帮我看看 payment-api 最近告警和历史 case",
		NormalizedGoal:    "查看 payment-api 最近告警并参考历史 case",
		RetrievalIntent:   "incident_lookup",
		RetrievalQueries:  []string{"payment-api 最近告警 历史 case", "payment-api 最近告警 历史 case"},
		RetrievalKeywords: []string{"payment-api", "alert", "runbook", "runbook"},
		KnowledgeScope:    []string{"incident_history", "service_runbooks", "incident_history"},
		RequiresRAG:       true,
		ResourceHints: rewrite.ResourceHints{
			ServiceName: "payment-api",
			ServiceID:   42,
		},
	})

	if envelope.Intent != "incident_lookup" {
		t.Fatalf("Intent = %q, want incident_lookup", envelope.Intent)
	}
	if len(envelope.Queries) != 1 || envelope.Queries[0] != "payment-api 最近告警 历史 case" {
		t.Fatalf("Queries = %#v", envelope.Queries)
	}
	if len(envelope.Keywords) != 3 {
		t.Fatalf("Keywords = %#v, want 3 unique items", envelope.Keywords)
	}
	if len(envelope.KnowledgeScope) != 2 {
		t.Fatalf("KnowledgeScope = %#v, want 2 unique items", envelope.KnowledgeScope)
	}
	if !envelope.RequiresRAG {
		t.Fatalf("RequiresRAG = false, want true")
	}
	if envelope.ResourceHints.ServiceID != 42 {
		t.Fatalf("ResourceHints.ServiceID = %d, want 42", envelope.ResourceHints.ServiceID)
	}
}

func TestBuildRewriteQueryEnvelopeFallsBackToRewriteGoal(t *testing.T) {
	envelope := BuildRewriteQueryEnvelope(rewrite.Output{
		RawUserInput:   "怎么排查 cilium pod 重启",
		NormalizedGoal: "排查 cilium pod 重启",
		DomainHints:    []string{"kubernetes", "networking"},
		KnowledgeScope: []string{"runbooks"},
	})

	if len(envelope.Queries) != 1 || envelope.Queries[0] != "排查 cilium pod 重启" {
		t.Fatalf("Queries = %#v, want normalized goal fallback", envelope.Queries)
	}
	if len(envelope.Keywords) != 3 {
		t.Fatalf("Keywords = %#v, want domain hints + knowledge scope", envelope.Keywords)
	}
	if envelope.RequiresRAG {
		t.Fatalf("RequiresRAG = true, want false when rewrite did not request rag")
	}
}
