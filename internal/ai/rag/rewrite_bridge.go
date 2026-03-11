package rag

import (
	"strings"

	"github.com/cy77cc/OpsPilot/internal/ai/rewrite"
)

type RewriteQueryEnvelope struct {
	Intent         string                `json:"intent,omitempty"`
	Goal           string                `json:"goal,omitempty"`
	Queries        []string              `json:"queries,omitempty"`
	Keywords       []string              `json:"keywords,omitempty"`
	KnowledgeScope []string              `json:"knowledge_scope,omitempty"`
	ResourceHints  rewrite.ResourceHints `json:"resource_hints,omitempty"`
	RequiresRAG    bool                  `json:"requires_rag,omitempty"`
}

func BuildRewriteQueryEnvelope(out rewrite.Output) RewriteQueryEnvelope {
	semantic := out.SemanticContract()
	queries := append([]string(nil), semantic.RetrievalQueries...)
	if len(queries) == 0 {
		if goal := strings.TrimSpace(semantic.NormalizedGoal); goal != "" {
			queries = append(queries, goal)
		} else if raw := strings.TrimSpace(semantic.RawUserInput); raw != "" {
			queries = append(queries, raw)
		}
	}
	keywords := append([]string(nil), semantic.RetrievalKeywords...)
	if len(keywords) == 0 {
		keywords = appendKeywords(keywords, semantic.DomainHints...)
		keywords = appendKeywords(keywords, semantic.KnowledgeScope...)
	}
	return RewriteQueryEnvelope{
		Intent:         strings.TrimSpace(semantic.RetrievalIntent),
		Goal:           strings.TrimSpace(semantic.NormalizedGoal),
		Queries:        dedupeStrings(queries),
		Keywords:       dedupeStrings(keywords),
		KnowledgeScope: dedupeStrings(semantic.KnowledgeScope),
		ResourceHints:  semantic.ResourceHints,
		RequiresRAG:    semantic.RequiresRAG || len(semantic.RetrievalQueries) > 0,
	}
}

func appendKeywords(base []string, values ...string) []string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		base = append(base, value)
	}
	return base
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
