package rag

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	ollamaembedding "github.com/cloudwego/eino-ext/components/embedding/ollama"
	openaicl "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cy77cc/OpsPilot/internal/config"
)

const (
	DefaultEmbeddingModel = "text-embedding-3-small"
	DefaultEmbeddingDim   = 1536
)

type Embedder struct {
	cfg config.Embedder

	mu   sync.Mutex
	impl embedding.Embedder
}

func NewEmbedder(cfg config.Embedder) *Embedder {
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = DefaultEmbeddingModel
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 20 * time.Second
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	return &Embedder{cfg: cfg}
}

func NewEmbedderWithComponent(cfg config.Embedder, component embedding.Embedder) *Embedder {
	e := NewEmbedder(cfg)
	e.impl = component
	return e
}

func (e *Embedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	impl, err := e.component(ctx)
	if err != nil {
		return nil, err
	}
	return impl.EmbedStrings(ctx, texts, opts...)
}

func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	rows, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("embedding result is empty")
	}
	return rows[0], nil
}

func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	rows64, err := e.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, err
	}
	out := make([][]float32, 0, len(rows64))
	for _, row64 := range rows64 {
		row32 := make([]float32, len(row64))
		for i := range row64 {
			row32[i] = float32(row64[i])
		}
		out = append(out, row32)
	}
	return out, nil
}

func (e *Embedder) component(ctx context.Context) (embedding.Embedder, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.impl != nil {
		return e.impl, nil
	}
	provider := strings.ToLower(strings.TrimSpace(e.cfg.Provider))
	switch provider {
	case "", "openai":
		if strings.TrimSpace(e.cfg.ApiKey) == "" {
			return nil, fmt.Errorf("embedder api key is required for provider=%q", firstNonEmpty(provider, "openai"))
		}
		httpClient := &http.Client{Timeout: e.cfg.Timeout}
		comp, err := openaicl.NewEmbeddingClient(ctx, &openaicl.EmbeddingConfig{
			APIKey:     e.cfg.ApiKey,
			BaseURL:    strings.TrimSpace(e.cfg.BaseURL),
			Model:      e.cfg.Model,
			HTTPClient: httpClient,
		})
		if err != nil {
			return nil, err
		}
		e.impl = comp
		return e.impl, nil
	case "ollama":
		comp, err := ollamaembedding.NewEmbedder(ctx, &ollamaembedding.EmbeddingConfig{
			BaseURL: strings.TrimSpace(e.cfg.BaseURL),
			Model:   e.cfg.Model,
			Timeout: e.cfg.Timeout,
		})
		if err != nil {
			return nil, err
		}
		e.impl = comp
		return e.impl, nil
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", e.cfg.Provider)
	}
}

type legacyTextEmbedder interface {
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

type legacyEmbedderAdapter struct {
	legacy legacyTextEmbedder
}

func (a *legacyEmbedderAdapter) EmbedStrings(ctx context.Context, texts []string, _ ...embedding.Option) ([][]float64, error) {
	rows, err := a.legacy.EmbedBatch(ctx, texts)
	if err != nil {
		return nil, err
	}
	out := make([][]float64, 0, len(rows))
	for _, row := range rows {
		converted := make([]float64, len(row))
		for i := range row {
			converted[i] = float64(row[i])
		}
		out = append(out, converted)
	}
	return out, nil
}
