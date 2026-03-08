package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type Retriever interface {
	Retrieve(ctx context.Context, namespace, query string, limit int) ([]KnowledgeEntry, error)
}

type NamespaceRetriever struct {
	indexer Indexer
}

func NewNamespaceRetriever(indexer Indexer) *NamespaceRetriever {
	return &NamespaceRetriever{indexer: indexer}
}

func (r *NamespaceRetriever) Retrieve(ctx context.Context, namespace, query string, limit int) ([]KnowledgeEntry, error) {
	if r == nil || r.indexer == nil {
		return nil, fmt.Errorf("retriever is not initialized")
	}
	entries, err := r.indexer.List(ctx, strings.TrimSpace(namespace))
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 4
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		if len(entries) > limit {
			return entries[:limit], nil
		}
		return entries, nil
	}

	type scored struct {
		entry KnowledgeEntry
		score int
	}
	scoredEntries := make([]scored, 0, len(entries))
	for _, entry := range entries {
		score := matchScore(query, entry)
		if score == 0 {
			continue
		}
		scoredEntries = append(scoredEntries, scored{entry: entry, score: score})
	}
	sort.Slice(scoredEntries, func(i, j int) bool {
		if scoredEntries[i].score == scoredEntries[j].score {
			return scoredEntries[i].entry.CreatedAt.After(scoredEntries[j].entry.CreatedAt)
		}
		return scoredEntries[i].score > scoredEntries[j].score
	})
	if len(scoredEntries) > limit {
		scoredEntries = scoredEntries[:limit]
	}
	out := make([]KnowledgeEntry, 0, len(scoredEntries))
	for _, item := range scoredEntries {
		out = append(out, item.entry)
	}
	return out, nil
}

func BuildAugmentedPrompt(query string, entries []KnowledgeEntry) string {
	query = strings.TrimSpace(query)
	if len(entries) == 0 {
		return query
	}
	var b strings.Builder
	b.WriteString("[Knowledge Context]\n")
	for _, entry := range entries {
		b.WriteString("- Q: ")
		b.WriteString(entry.Question)
		b.WriteString("\n  A: ")
		b.WriteString(entry.Answer)
		b.WriteString("\n")
	}
	b.WriteString("\n[User Query]\n")
	b.WriteString(query)
	return b.String()
}

func matchScore(query string, entry KnowledgeEntry) int {
	score := 0
	for _, token := range strings.Fields(query) {
		if token == "" {
			continue
		}
		if strings.Contains(strings.ToLower(entry.Question), token) {
			score += 2
		}
		if strings.Contains(strings.ToLower(entry.Answer), token) {
			score++
		}
	}
	return score
}
