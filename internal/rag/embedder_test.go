package rag

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cy77cc/OpsPilot/internal/config"
)

type fakeEinoEmbedder struct{}

func (fakeEinoEmbedder) EmbedStrings(_ context.Context, texts []string, _ ...embedding.Option) ([][]float64, error) {
	out := make([][]float64, 0, len(texts))
	for i := range texts {
		out = append(out, []float64{float64(i) + 0.1, float64(i) + 0.2})
	}
	return out, nil
}

func TestEmbedderEmbedBatch(t *testing.T) {
	e := NewEmbedderWithComponent(config.Embedder{Provider: "openai"}, fakeEinoEmbedder{})
	out, err := e.EmbedBatch(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("embed batch: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(out))
	}
	if len(out[0]) != 2 || len(out[1]) != 2 {
		t.Fatalf("unexpected vector dimensions: %#v", out)
	}
}

func TestEmbedderEmbed(t *testing.T) {
	e := NewEmbedderWithComponent(config.Embedder{Provider: "openai"}, fakeEinoEmbedder{})
	row, err := e.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatalf("embed single: %v", err)
	}
	if len(row) != 2 {
		t.Fatalf("expected vector dim=2, got %d", len(row))
	}
}

func TestEmbedderMissingAPIKeyForOpenAI(t *testing.T) {
	e := NewEmbedder(config.Embedder{Provider: "openai"})
	_, err := e.EmbedBatch(context.Background(), []string{"a"})
	if err == nil {
		t.Fatalf("expected missing api key error")
	}
	if !strings.Contains(err.Error(), "api key") {
		t.Fatalf("unexpected error: %v", err)
	}
}
