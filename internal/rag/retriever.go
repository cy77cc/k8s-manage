package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	einoembedding "github.com/cloudwego/eino/components/embedding"
	einoretriever "github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	milvusclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type vectorSearcher interface {
	SearchVectors(ctx context.Context, collection, vectorField string, vectors []entity.Vector, topK int) ([]milvusclient.SearchResult, error)
}

type ToolExample struct {
	ToolName      string
	Intent        string
	ParamsJSON    string
	ResultSummary string
	Score         float32
}

type PlatformAsset struct {
	AssetType    string
	AssetID      string
	Name         string
	Status       string
	MetadataJSON string
	Score        float32
}

type TroubleshootingCase struct {
	Title     string
	Symptom   string
	Diagnosis string
	Solution  string
	Score     float32
}

type RAGContext struct {
	ToolExamples         []ToolExample
	RelatedAssets        []PlatformAsset
	TroubleshootingCases []TroubleshootingCase
}

type RAGRetriever struct {
	toolRetriever    einoretriever.Retriever
	assetRetriever   einoretriever.Retriever
	caseRetriever    einoretriever.Retriever
}

func NewRAGRetriever(searcher vectorSearcher, embedder any) *RAGRetriever {
	if searcher == nil || embedder == nil {
		return &RAGRetriever{}
	}
	emb := asEinoEmbedder(embedder)
	if emb == nil {
		return &RAGRetriever{}
	}
	return &RAGRetriever{
		toolRetriever:  newMilvusCollectionRetriever(searcher, emb, CollectionToolExamples, "embedding", "intent", []string{"tool_name", "intent", "params_json", "result_summary"}),
		assetRetriever: newMilvusCollectionRetriever(searcher, emb, CollectionPlatformAssets, "embedding", "name", []string{"asset_type", "asset_id", "name", "status", "metadata_json"}),
		caseRetriever:  newMilvusCollectionRetriever(searcher, emb, CollectionTroubleshooting, "embedding", "title", []string{"title", "symptom", "diagnosis", "solution"}),
	}
}

func (r *RAGRetriever) Retrieve(ctx context.Context, query string, topK int) (*RAGContext, error) {
	if r == nil || r.toolRetriever == nil || r.assetRetriever == nil || r.caseRetriever == nil {
		return nil, fmt.Errorf("rag retriever is not fully initialized")
	}
	if strings.TrimSpace(query) == "" {
		return &RAGContext{}, nil
	}
	if topK <= 0 {
		topK = 6
	}
	perCollection := topK / 3
	if perCollection <= 0 {
		perCollection = 1
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		ctxOut   = &RAGContext{}
		firstErr error
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		docs, err := r.toolRetriever.Retrieve(ctx, query, einoretriever.WithTopK(perCollection))
		mu.Lock()
		defer mu.Unlock()
		if err != nil && firstErr == nil {
			firstErr = err
			return
		}
		ctxOut.ToolExamples = docsToToolExamples(docs)
	}()
	go func() {
		defer wg.Done()
		docs, err := r.assetRetriever.Retrieve(ctx, query, einoretriever.WithTopK(perCollection))
		mu.Lock()
		defer mu.Unlock()
		if err != nil && firstErr == nil {
			firstErr = err
			return
		}
		ctxOut.RelatedAssets = docsToAssets(docs)
	}()
	go func() {
		defer wg.Done()
		docs, err := r.caseRetriever.Retrieve(ctx, query, einoretriever.WithTopK(perCollection))
		mu.Lock()
		defer mu.Unlock()
		if err != nil && firstErr == nil {
			firstErr = err
			return
		}
		ctxOut.TroubleshootingCases = docsToCases(docs)
	}()
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	ctxOut.ToolExamples = dedupeToolExamples(ctxOut.ToolExamples)
	ctxOut.RelatedAssets = dedupeAssets(ctxOut.RelatedAssets)
	ctxOut.TroubleshootingCases = dedupeCases(ctxOut.TroubleshootingCases)
	sort.Slice(ctxOut.ToolExamples, func(i, j int) bool { return ctxOut.ToolExamples[i].Score > ctxOut.ToolExamples[j].Score })
	sort.Slice(ctxOut.RelatedAssets, func(i, j int) bool { return ctxOut.RelatedAssets[i].Score > ctxOut.RelatedAssets[j].Score })
	sort.Slice(ctxOut.TroubleshootingCases, func(i, j int) bool { return ctxOut.TroubleshootingCases[i].Score > ctxOut.TroubleshootingCases[j].Score })
	return ctxOut, nil
}

func (r *RAGRetriever) BuildAugmentedPrompt(query string, context *RAGContext) string {
	if context == nil {
		return strings.TrimSpace(query)
	}
	var b strings.Builder
	if len(context.ToolExamples) > 0 {
		b.WriteString("\n[Related Tool Examples]\n")
		for _, item := range context.ToolExamples {
			b.WriteString(fmt.Sprintf("- tool=%s intent=%s params=%s\n", item.ToolName, item.Intent, item.ParamsJSON))
		}
	}
	if len(context.RelatedAssets) > 0 {
		b.WriteString("\n[Related Platform Assets]\n")
		for _, item := range context.RelatedAssets {
			b.WriteString(fmt.Sprintf("- %s %s (id=%s, status=%s)\n", item.AssetType, item.Name, item.AssetID, item.Status))
		}
	}
	if len(context.TroubleshootingCases) > 0 {
		b.WriteString("\n[Related Troubleshooting Cases]\n")
		for _, item := range context.TroubleshootingCases {
			b.WriteString(fmt.Sprintf("- %s | %s\n", item.Title, item.Solution))
		}
	}
	prefix := strings.TrimSpace(b.String())
	if prefix == "" {
		return strings.TrimSpace(query)
	}
	return prefix + "\n\nUser Query:\n" + strings.TrimSpace(query)
}

type milvusCollectionRetriever struct {
	searcher    vectorSearcher
	embedding   einoembedding.Embedder
	collection  string
	vectorField string
	contentField string
	fields      []string
}

func newMilvusCollectionRetriever(searcher vectorSearcher, emb einoembedding.Embedder, collection, vectorField, contentField string, fields []string) einoretriever.Retriever {
	return &milvusCollectionRetriever{
		searcher:     searcher,
		embedding:    emb,
		collection:   collection,
		vectorField:  vectorField,
		contentField: contentField,
		fields:       append([]string(nil), fields...),
	}
}

func (m *milvusCollectionRetriever) Retrieve(ctx context.Context, query string, opts ...einoretriever.Option) ([]*schema.Document, error) {
	if m == nil || m.searcher == nil || m.embedding == nil {
		return nil, fmt.Errorf("collection retriever is not initialized")
	}
	co := einoretriever.GetCommonOptions(&einoretriever.Options{}, opts...)
	topK := 3
	if co.TopK != nil && *co.TopK > 0 {
		topK = *co.TopK
	}
	emb := co.Embedding
	if emb == nil {
		emb = m.embedding
	}
	vecs, err := emb.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 || len(vecs[0]) == 0 {
		return []*schema.Document{}, nil
	}
	queryVector := make([]float32, len(vecs[0]))
	for i := range vecs[0] {
		queryVector[i] = float32(vecs[0][i])
	}
	results, err := m.searcher.SearchVectors(ctx, m.collection, m.vectorField, []entity.Vector{entity.FloatVector(queryVector)}, topK)
	if err != nil {
		return nil, err
	}
	docs := make([]*schema.Document, 0)
	for _, result := range results {
		for i := 0; i < rowCount(result); i++ {
			doc := &schema.Document{
				ID:      getFieldString(result, "id", i),
				Content: getFieldString(result, m.contentField, i),
				MetaData: map[string]any{
					"collection": m.collection,
					"score":      getScore(result, i),
				},
			}
			for _, field := range m.fields {
				doc.MetaData[field] = getFieldString(result, field, i)
			}
			doc.WithScore(float64(getScore(result, i)))
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

func rowCount(result milvusclient.SearchResult) int {
	count := result.ResultCount
	if count <= 0 {
		count = len(result.Scores)
	}
	if count <= 0 {
		count = result.Fields.Len()
	}
	if count > len(result.Scores) && len(result.Scores) > 0 {
		count = len(result.Scores)
	}
	return count
}

func getScore(result milvusclient.SearchResult, idx int) float32 {
	if idx < 0 || idx >= len(result.Scores) {
		return 0
	}
	return result.Scores[idx]
}

func getFieldString(result milvusclient.SearchResult, field string, idx int) string {
	col := result.Fields.GetColumn(field)
	if col == nil || idx < 0 || idx >= col.Len() {
		return ""
	}
	v, err := col.Get(idx)
	if err != nil {
		return ""
	}
	switch value := v.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func docsToToolExamples(docs []*schema.Document) []ToolExample {
	out := make([]ToolExample, 0, len(docs))
	for _, d := range docs {
		if d == nil {
			continue
		}
		out = append(out, ToolExample{
			ToolName:      toString(d.MetaData["tool_name"]),
			Intent:        toString(d.MetaData["intent"]),
			ParamsJSON:    toString(d.MetaData["params_json"]),
			ResultSummary: toString(d.MetaData["result_summary"]),
			Score:         float32(d.Score()),
		})
	}
	return out
}

func docsToAssets(docs []*schema.Document) []PlatformAsset {
	out := make([]PlatformAsset, 0, len(docs))
	for _, d := range docs {
		if d == nil {
			continue
		}
		out = append(out, PlatformAsset{
			AssetType:    toString(d.MetaData["asset_type"]),
			AssetID:      toString(d.MetaData["asset_id"]),
			Name:         toString(d.MetaData["name"]),
			Status:       toString(d.MetaData["status"]),
			MetadataJSON: toString(d.MetaData["metadata_json"]),
			Score:        float32(d.Score()),
		})
	}
	return out
}

func docsToCases(docs []*schema.Document) []TroubleshootingCase {
	out := make([]TroubleshootingCase, 0, len(docs))
	for _, d := range docs {
		if d == nil {
			continue
		}
		out = append(out, TroubleshootingCase{
			Title:     toString(d.MetaData["title"]),
			Symptom:   toString(d.MetaData["symptom"]),
			Diagnosis: toString(d.MetaData["diagnosis"]),
			Solution:  toString(d.MetaData["solution"]),
			Score:     float32(d.Score()),
		})
	}
	return out
}

func dedupeToolExamples(input []ToolExample) []ToolExample {
	seen := make(map[string]struct{}, len(input))
	out := make([]ToolExample, 0, len(input))
	for _, item := range input {
		key := item.ToolName + "|" + item.Intent + "|" + item.ParamsJSON
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func dedupeAssets(input []PlatformAsset) []PlatformAsset {
	seen := make(map[string]struct{}, len(input))
	out := make([]PlatformAsset, 0, len(input))
	for _, item := range input {
		key := item.AssetType + "|" + item.AssetID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func dedupeCases(input []TroubleshootingCase) []TroubleshootingCase {
	seen := make(map[string]struct{}, len(input))
	out := make([]TroubleshootingCase, 0, len(input))
	for _, item := range input {
		key := item.Title + "|" + item.Symptom
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func asEinoEmbedder(embedder any) einoembedding.Embedder {
	if emb, ok := embedder.(einoembedding.Embedder); ok {
		return emb
	}
	if legacy, ok := embedder.(legacyTextEmbedder); ok {
		return &legacyEmbedderAdapter{legacy: legacy}
	}
	return nil
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return fmt.Sprintf("%v", x)
	}
}
