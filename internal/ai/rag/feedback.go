package rag

import (
	"context"
	"fmt"
	"strings"

	aistate "github.com/cy77cc/OpsPilot/internal/ai/state"
)

type Feedback struct {
	IsEffective bool   `json:"is_effective"`
	Comment     string `json:"comment,omitempty"`
}

type QAExtractor interface {
	Extract(ctx context.Context, sessionID string) (KnowledgeEntry, error)
}

type FeedbackCollector interface {
	Collect(ctx context.Context, sessionID, namespace string, feedback Feedback) (*KnowledgeEntry, error)
}

type SessionFeedbackCollector struct {
	indexer   Indexer
	extractor QAExtractor
}

func NewFeedbackCollector(indexer Indexer, extractor QAExtractor) *SessionFeedbackCollector {
	return &SessionFeedbackCollector{indexer: indexer, extractor: extractor}
}

func (c *SessionFeedbackCollector) Collect(ctx context.Context, sessionID, namespace string, feedback Feedback) (*KnowledgeEntry, error) {
	if c == nil || c.indexer == nil || c.extractor == nil {
		return nil, fmt.Errorf("feedback collector is not initialized")
	}
	if !feedback.IsEffective {
		return nil, nil
	}
	entry, err := c.extractor.Extract(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return nil, err
	}
	entry.Source = SourceFeedback
	entry.Namespace = strings.TrimSpace(namespace)
	if err := c.indexer.Index(ctx, []KnowledgeEntry{entry}); err != nil {
		return nil, err
	}
	return &entry, nil
}

type SessionSnapshotLoader interface {
	Load(ctx context.Context, sessionID string) (*aistate.SessionSnapshot, error)
}

type SessionQAExtractor struct {
	loader SessionSnapshotLoader
}

func NewSessionQAExtractor(loader SessionSnapshotLoader) *SessionQAExtractor {
	return &SessionQAExtractor{loader: loader}
}

func (e *SessionQAExtractor) Extract(ctx context.Context, sessionID string) (KnowledgeEntry, error) {
	if e == nil || e.loader == nil {
		return KnowledgeEntry{}, fmt.Errorf("session qa extractor is not initialized")
	}
	snapshot, err := e.loader.Load(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return KnowledgeEntry{}, err
	}
	if snapshot == nil || len(snapshot.Messages) == 0 {
		return KnowledgeEntry{}, fmt.Errorf("session snapshot not found")
	}
	var question, answer string
	for i := len(snapshot.Messages) - 1; i >= 0; i-- {
		msg := snapshot.Messages[i]
		switch msg.Role {
		case "assistant":
			if answer == "" {
				answer = strings.TrimSpace(msg.Content)
			}
		case "user":
			if question == "" {
				question = strings.TrimSpace(msg.Content)
			}
		}
		if question != "" && answer != "" {
			break
		}
	}
	if question == "" || answer == "" {
		return KnowledgeEntry{}, fmt.Errorf("session does not contain a complete qa pair")
	}
	return KnowledgeEntry{Question: question, Answer: answer}, nil
}
