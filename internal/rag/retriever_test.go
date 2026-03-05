package rag

import (
	"context"
	"errors"
	"strings"
	"testing"

	milvusclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type fakeSearcher struct {
	results map[string][]milvusclient.SearchResult
	err     error
	calls   []string
}

func (f *fakeSearcher) SearchVectors(_ context.Context, collection, _ string, _ []entity.Vector, _ int) ([]milvusclient.SearchResult, error) {
	f.calls = append(f.calls, collection)
	if f.err != nil {
		return nil, f.err
	}
	return f.results[collection], nil
}

type fakeRetrieverEmbedder struct{}

func (f fakeRetrieverEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	return [][]float32{{0.1, 0.2, 0.3}}, nil
}

func TestRAGRetrieverRetrieve(t *testing.T) {
	searcher := &fakeSearcher{results: map[string][]milvusclient.SearchResult{
		CollectionToolExamples: {
			buildSearchResult([]float32{0.9, 0.7}, map[string][]string{
				"tool_name":      {"deployment.release", "deployment.release"},
				"intent":         {"deploy api", "deploy api"},
				"params_json":    {`{"service_id":1}`, `{"service_id":1}`},
				"result_summary": {"ok", "ok"},
			}),
		},
		CollectionPlatformAssets: {
			buildSearchResult([]float32{0.8}, map[string][]string{
				"asset_type":    {"service"},
				"asset_id":      {"100"},
				"name":          {"api-service"},
				"status":        {"running"},
				"metadata_json": {`{"env":"prod"}`},
			}),
		},
		CollectionTroubleshooting: {
			buildSearchResult([]float32{0.85}, map[string][]string{
				"title":     {"Nginx 502"},
				"symptom":   {"502"},
				"diagnosis": {"upstream timeout"},
				"solution":  {"restart"},
			}),
		},
	}}

	r := NewRAGRetriever(searcher, fakeRetrieverEmbedder{})
	ctx, err := r.Retrieve(context.Background(), "deploy api service", 6)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if len(ctx.ToolExamples) != 1 {
		t.Fatalf("expected deduped tool examples = 1, got %d", len(ctx.ToolExamples))
	}
	if len(ctx.RelatedAssets) != 1 || len(ctx.TroubleshootingCases) != 1 {
		t.Fatalf("unexpected rag context: %+v", ctx)
	}
	if len(searcher.calls) != 3 {
		t.Fatalf("expected 3 parallel collection searches, got %d", len(searcher.calls))
	}
}

func TestRAGRetrieverBuildAugmentedPrompt(t *testing.T) {
	r := NewRAGRetriever(&fakeSearcher{}, fakeRetrieverEmbedder{})
	prompt := r.BuildAugmentedPrompt("how to deploy", &RAGContext{
		ToolExamples: []ToolExample{{ToolName: "deployment.release", Intent: "deploy", ParamsJSON: `{}`}},
		RelatedAssets: []PlatformAsset{{AssetType: "service", Name: "api", AssetID: "1", Status: "running"}},
		TroubleshootingCases: []TroubleshootingCase{{Title: "502", Solution: "restart"}},
	})
	for _, token := range []string{"Related Tool Examples", "Related Platform Assets", "Related Troubleshooting Cases", "how to deploy"} {
		if !strings.Contains(prompt, token) {
			t.Fatalf("prompt missing token %q: %s", token, prompt)
		}
	}
}

func TestRAGRetrieverRetrieveError(t *testing.T) {
	r := NewRAGRetriever(&fakeSearcher{err: errors.New("search failed")}, fakeRetrieverEmbedder{})
	_, err := r.Retrieve(context.Background(), "query", 6)
	if err == nil {
		t.Fatalf("expected retrieve error")
	}
}

func buildSearchResult(scores []float32, fields map[string][]string) milvusclient.SearchResult {
	resultFields := make(milvusclient.ResultSet, 0, len(fields))
	for name, values := range fields {
		resultFields = append(resultFields, entity.NewColumnVarChar(name, values))
	}
	return milvusclient.SearchResult{
		ResultCount: len(scores),
		Fields:      resultFields,
		Scores:      scores,
	}
}
