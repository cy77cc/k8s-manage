package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type KnowledgeSource string

const (
	SourceUserInput KnowledgeSource = "user_input"
	SourceFeedback  KnowledgeSource = "feedback"
)

type KnowledgeEntry struct {
	ID        string          `json:"id"`
	Source    KnowledgeSource `json:"source"`
	Namespace string          `json:"namespace"`
	Question  string          `json:"question"`
	Answer    string          `json:"answer"`
	CreatedAt time.Time       `json:"created_at"`
}

type Indexer interface {
	Index(ctx context.Context, entries []KnowledgeEntry) error
	AddUserKnowledge(ctx context.Context, namespace, question, answer string) (KnowledgeEntry, error)
	List(ctx context.Context, namespace string) ([]KnowledgeEntry, error)
}

type MilvusBackend interface {
	Upsert(ctx context.Context, entries []KnowledgeEntry) error
}

type MilvusIndexer struct {
	backend MilvusBackend
	nowFn   func() time.Time

	mu      sync.RWMutex
	entries map[string][]KnowledgeEntry
}

func NewMilvusIndexer(backend MilvusBackend) *MilvusIndexer {
	return &MilvusIndexer{
		backend: backend,
		nowFn:   time.Now,
		entries: make(map[string][]KnowledgeEntry),
	}
}

func (i *MilvusIndexer) Index(ctx context.Context, entries []KnowledgeEntry) error {
	if i == nil {
		return fmt.Errorf("indexer is nil")
	}
	prepared := make([]KnowledgeEntry, 0, len(entries))
	for idx, entry := range entries {
		entry.Namespace = strings.TrimSpace(entry.Namespace)
		if entry.Namespace == "" {
			return fmt.Errorf("knowledge entry namespace is required")
		}
		entry.Question = strings.TrimSpace(entry.Question)
		entry.Answer = strings.TrimSpace(entry.Answer)
		if entry.Question == "" && entry.Answer == "" {
			return fmt.Errorf("knowledge entry content is empty")
		}
		if strings.TrimSpace(entry.ID) == "" {
			entry.ID = fmt.Sprintf("%s-%d", entry.Namespace, idx+1)
		}
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = i.nowFn()
		}
		prepared = append(prepared, entry)
	}
	if len(prepared) == 0 {
		return nil
	}
	if i.backend != nil {
		if err := i.backend.Upsert(ctx, prepared); err != nil {
			return err
		}
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, entry := range prepared {
		items := append(i.entries[entry.Namespace], entry)
		sort.Slice(items, func(a, b int) bool { return items[a].CreatedAt.After(items[b].CreatedAt) })
		i.entries[entry.Namespace] = items
	}
	return nil
}

func (i *MilvusIndexer) AddUserKnowledge(ctx context.Context, namespace, question, answer string) (KnowledgeEntry, error) {
	entry := KnowledgeEntry{
		ID:        fmt.Sprintf("user-%d", i.nowFn().UnixNano()),
		Source:    SourceUserInput,
		Namespace: strings.TrimSpace(namespace),
		Question:  strings.TrimSpace(question),
		Answer:    strings.TrimSpace(answer),
		CreatedAt: i.nowFn(),
	}
	return entry, i.Index(ctx, []KnowledgeEntry{entry})
}

func (i *MilvusIndexer) List(_ context.Context, namespace string) ([]KnowledgeEntry, error) {
	if i == nil {
		return nil, fmt.Errorf("indexer is nil")
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	items := append([]KnowledgeEntry(nil), i.entries[strings.TrimSpace(namespace)]...)
	return items, nil
}
