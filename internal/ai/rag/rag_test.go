package rag

import (
	"context"
	"testing"

	aistate "github.com/cy77cc/OpsPilot/internal/ai/state"
)

type fakeBackend struct {
	last []KnowledgeEntry
}

func (f *fakeBackend) Upsert(_ context.Context, entries []KnowledgeEntry) error {
	f.last = append([]KnowledgeEntry(nil), entries...)
	return nil
}

type fakeExtractor struct {
	entry KnowledgeEntry
}

func (f fakeExtractor) Extract(_ context.Context, _ string) (KnowledgeEntry, error) {
	return f.entry, nil
}

type fakeSnapshotLoader struct {
	snapshot *aistate.SessionSnapshot
}

func (f fakeSnapshotLoader) Load(_ context.Context, _ string) (*aistate.SessionSnapshot, error) {
	return f.snapshot, nil
}

func TestMilvusIndexerAndRetrieverNamespaceFiltering(t *testing.T) {
	indexer := NewMilvusIndexer(&fakeBackend{})
	if _, err := indexer.AddUserKnowledge(context.Background(), "team-a", "restart service", "Use rolling restart"); err != nil {
		t.Fatalf("add user knowledge: %v", err)
	}
	if _, err := indexer.AddUserKnowledge(context.Background(), "team-b", "delete pod", "Delete only stuck pod"); err != nil {
		t.Fatalf("add second namespace knowledge: %v", err)
	}

	retriever := NewNamespaceRetriever(indexer)
	entries, err := retriever.Retrieve(context.Background(), "team-a", "restart", 5)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if len(entries) != 1 || entries[0].Namespace != "team-a" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}

func TestFeedbackCollectorIndexesEffectiveFeedback(t *testing.T) {
	indexer := NewMilvusIndexer(nil)
	collector := NewFeedbackCollector(indexer, fakeExtractor{entry: KnowledgeEntry{
		ID:       "fb-1",
		Question: "How do I restart api?",
		Answer:   "Use the service_restart tool.",
	}})

	entry, err := collector.Collect(context.Background(), "sess-1", "team-a", Feedback{IsEffective: true})
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if entry == nil || entry.Source != SourceFeedback {
		t.Fatalf("unexpected entry: %+v", entry)
	}

	items, err := indexer.List(context.Background(), "team-a")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestSessionQAExtractorExtractsLatestPair(t *testing.T) {
	extractor := NewSessionQAExtractor(fakeSnapshotLoader{snapshot: &aistate.SessionSnapshot{
		SessionID: "sess-1",
		Messages: []aistate.StoredMessage{
			{Role: "user", Content: "How do I restart api?"},
			{Role: "assistant", Content: "Use the service_restart tool."},
		},
	}})
	entry, err := extractor.Extract(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if entry.Question == "" || entry.Answer == "" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
}
