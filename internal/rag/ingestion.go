package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	einoembedding "github.com/cloudwego/eino/components/embedding"
	einoindexer "github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"gorm.io/gorm"
)

const (
	checkpointToolExamples    = "tool_examples"
	checkpointPlatformAssets  = "platform_assets"
	checkpointTroubleshooting = "troubleshooting_cases"
)

var successfulCommandStatuses = []string{"succeeded", "success", "approved", "applied", "done"}

type vectorStore interface {
	InsertRows(ctx context.Context, collection string, rows []interface{}) error
	FlushCollection(ctx context.Context, collection string) error
}

type IngestionStats struct {
	Collection   string
	Total        int
	Inserted     int
	Skipped      int
	LatestUpdate time.Time
}

type TroubleshootingCaseInput struct {
	Title     string
	Symptom   string
	Diagnosis string
	Solution  string
	UpdatedAt time.Time
}

type IngestionService struct {
	db *gorm.DB

	toolIndexer  einoindexer.Indexer
	assetIndexer einoindexer.Indexer
	caseIndexer  einoindexer.Indexer

	mu          sync.RWMutex
	checkpoints map[string]time.Time
	nowFn       func() time.Time
}

func NewIngestionService(db *gorm.DB, store vectorStore, embedder any) *IngestionService {
	einoEmb := asEinoEmbedder(embedder)
	return &IngestionService{
		db:          db,
		toolIndexer: newMilvusDocumentIndexer(CollectionToolExamples, store, einoEmb, toolExampleRowBuilder),
		assetIndexer: newMilvusDocumentIndexer(CollectionPlatformAssets, store, einoEmb, platformAssetRowBuilder),
		caseIndexer:  newMilvusDocumentIndexer(CollectionTroubleshooting, store, einoEmb, troubleshootingRowBuilder),
		checkpoints:  make(map[string]time.Time),
		nowFn:        time.Now,
	}
}

func (s *IngestionService) IngestToolExamples(ctx context.Context, since time.Time) (IngestionStats, error) {
	stats := IngestionStats{Collection: CollectionToolExamples}
	if s == nil || s.db == nil || s.toolIndexer == nil {
		return stats, fmt.Errorf("ingestion service is not fully initialized")
	}

	var executions []model.AICommandExecution
	q := s.db.WithContext(ctx).
		Where("status IN ?", successfulCommandStatuses).
		Order("updated_at ASC")
	if !since.IsZero() {
		q = q.Where("updated_at > ?", since)
	}
	if err := q.Find(&executions).Error; err != nil {
		return stats, err
	}
	stats.Total = len(executions)
	if len(executions) == 0 {
		return stats, nil
	}

	docs := make([]*schema.Document, 0, len(executions))
	for _, row := range executions {
		docs = append(docs, &schema.Document{
			ID:      row.ID,
			Content: buildToolEmbeddingText(row),
			MetaData: map[string]any{
				"tool_name":      truncateString(extractToolName(row), 256),
				"intent":         truncateString(firstNonEmpty(row.Intent, row.CommandText), int(defaultVarcharMaxLength)),
				"params_json":    truncateString(row.ParamsJSON, int(defaultVarcharMaxLength)),
				"result_summary": truncateString(row.ExecutionSummary, int(defaultVarcharMaxLength)),
				"updated_at":     row.UpdatedAt,
			},
		})
		if row.UpdatedAt.After(stats.LatestUpdate) {
			stats.LatestUpdate = row.UpdatedAt
		}
	}

	ids, err := s.toolIndexer.Store(ctx, docs)
	if err != nil {
		return stats, err
	}
	stats.Inserted = len(ids)
	stats.Skipped = stats.Total - stats.Inserted
	return stats, nil
}

func (s *IngestionService) IngestPlatformAssets(ctx context.Context, since time.Time) (IngestionStats, error) {
	stats := IngestionStats{Collection: CollectionPlatformAssets}
	if s == nil || s.db == nil || s.assetIndexer == nil {
		return stats, fmt.Errorf("ingestion service is not fully initialized")
	}

	assets, err := s.loadPlatformAssets(ctx, since)
	if err != nil {
		return stats, err
	}
	stats.Total = len(assets)
	if len(assets) == 0 {
		return stats, nil
	}
	docs := make([]*schema.Document, 0, len(assets))
	for _, asset := range assets {
		docs = append(docs, &schema.Document{
			ID:      asset.assetType + ":" + asset.assetID,
			Content: asset.embedText,
			MetaData: map[string]any{
				"asset_type":    truncateString(asset.assetType, 64),
				"asset_id":      truncateString(asset.assetID, 128),
				"name":          truncateString(asset.name, 512),
				"status":        truncateString(asset.status, 64),
				"metadata_json": truncateString(asset.metadataJSON, int(defaultVarcharMaxLength)),
				"updated_at":    asset.updatedAt,
			},
		})
		if asset.updatedAt.After(stats.LatestUpdate) {
			stats.LatestUpdate = asset.updatedAt
		}
	}
	ids, err := s.assetIndexer.Store(ctx, docs)
	if err != nil {
		return stats, err
	}
	stats.Inserted = len(ids)
	stats.Skipped = stats.Total - stats.Inserted
	return stats, nil
}

func (s *IngestionService) IngestTroubleshootingCases(ctx context.Context, cases []TroubleshootingCaseInput, since time.Time) (IngestionStats, error) {
	stats := IngestionStats{Collection: CollectionTroubleshooting}
	if s == nil || s.caseIndexer == nil {
		return stats, fmt.Errorf("ingestion service is not fully initialized")
	}
	if len(cases) == 0 {
		return stats, nil
	}

	selected := make([]TroubleshootingCaseInput, 0, len(cases))
	for _, item := range cases {
		if !since.IsZero() && !item.UpdatedAt.IsZero() && !item.UpdatedAt.After(since) {
			continue
		}
		selected = append(selected, item)
	}
	stats.Total = len(selected)
	if len(selected) == 0 {
		return stats, nil
	}

	docs := make([]*schema.Document, 0, len(selected))
	for i, item := range selected {
		id := fmt.Sprintf("case-%d-%d", i, item.UpdatedAt.UnixNano())
		docs = append(docs, &schema.Document{
			ID:      id,
			Content: strings.TrimSpace(item.Title + " " + item.Symptom + " " + item.Diagnosis + " " + item.Solution),
			MetaData: map[string]any{
				"title":      truncateString(item.Title, 512),
				"symptom":    truncateString(item.Symptom, int(defaultVarcharMaxLength)),
				"diagnosis":  truncateString(item.Diagnosis, int(defaultVarcharMaxLength)),
				"solution":   truncateString(item.Solution, int(defaultVarcharMaxLength)),
				"updated_at": item.UpdatedAt,
			},
		})
		if item.UpdatedAt.After(stats.LatestUpdate) {
			stats.LatestUpdate = item.UpdatedAt
		}
	}
	ids, err := s.caseIndexer.Store(ctx, docs)
	if err != nil {
		return stats, err
	}
	stats.Inserted = len(ids)
	stats.Skipped = stats.Total - stats.Inserted
	return stats, nil
}

func (s *IngestionService) IngestIncremental(ctx context.Context, cases []TroubleshootingCaseInput) (map[string]IngestionStats, error) {
	if s == nil {
		return nil, fmt.Errorf("ingestion service is nil")
	}
	result := make(map[string]IngestionStats, 3)

	toolSince := s.Checkpoint(checkpointToolExamples)
	toolStats, err := s.IngestToolExamples(ctx, toolSince)
	if err != nil {
		return nil, err
	}
	result[checkpointToolExamples] = toolStats
	s.advanceCheckpoint(checkpointToolExamples, toolSince, toolStats.LatestUpdate)

	assetSince := s.Checkpoint(checkpointPlatformAssets)
	assetStats, err := s.IngestPlatformAssets(ctx, assetSince)
	if err != nil {
		return nil, err
	}
	result[checkpointPlatformAssets] = assetStats
	s.advanceCheckpoint(checkpointPlatformAssets, assetSince, assetStats.LatestUpdate)

	caseSince := s.Checkpoint(checkpointTroubleshooting)
	caseStats, err := s.IngestTroubleshootingCases(ctx, cases, caseSince)
	if err != nil {
		return nil, err
	}
	result[checkpointTroubleshooting] = caseStats
	s.advanceCheckpoint(checkpointTroubleshooting, caseSince, caseStats.LatestUpdate)

	return result, nil
}

func (s *IngestionService) Checkpoint(source string) time.Time {
	if s == nil {
		return time.Time{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.checkpoints[source]
}

func (s *IngestionService) SetCheckpoint(source string, ts time.Time) {
	if s == nil || strings.TrimSpace(source) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkpoints[source] = ts
}

func (s *IngestionService) advanceCheckpoint(source string, previous, latest time.Time) {
	next := previous
	if latest.After(next) {
		next = latest
	}
	if next.IsZero() {
		next = s.nowFn().UTC()
	}
	s.SetCheckpoint(source, next)
}

type toolExampleMilvusRow struct {
	ToolName      string             `milvus:"name:tool_name"`
	Intent        string             `milvus:"name:intent"`
	ParamsJSON    string             `milvus:"name:params_json"`
	ResultSummary string             `milvus:"name:result_summary"`
	Embedding     entity.FloatVector `milvus:"name:embedding"`
}

type platformAssetMilvusRow struct {
	AssetType    string             `milvus:"name:asset_type"`
	AssetID      string             `milvus:"name:asset_id"`
	Name         string             `milvus:"name:name"`
	Status       string             `milvus:"name:status"`
	MetadataJSON string             `milvus:"name:metadata_json"`
	Embedding    entity.FloatVector `milvus:"name:embedding"`
}

type troubleshootingMilvusRow struct {
	Title     string             `milvus:"name:title"`
	Symptom   string             `milvus:"name:symptom"`
	Diagnosis string             `milvus:"name:diagnosis"`
	Solution  string             `milvus:"name:solution"`
	Embedding entity.FloatVector `milvus:"name:embedding"`
}

type assetCandidate struct {
	assetType    string
	assetID      string
	name         string
	status       string
	embedText    string
	metadataJSON string
	updatedAt    time.Time
}

func (s *IngestionService) loadPlatformAssets(ctx context.Context, since time.Time) ([]assetCandidate, error) {
	assets := make([]assetCandidate, 0)

	var cis []model.CMDBCI
	ciQ := s.db.WithContext(ctx).Order("updated_at ASC")
	if !since.IsZero() {
		ciQ = ciQ.Where("updated_at > ?", since)
	}
	if err := ciQ.Find(&cis).Error; err != nil {
		return nil, err
	}
	for _, ci := range cis {
		meta := map[string]any{
			"ci_uid":      ci.CIUID,
			"source":      ci.Source,
			"external_id": ci.ExternalID,
			"owner":       ci.Owner,
			"tags_json":   ci.TagsJSON,
			"attrs_json":  ci.AttrsJSON,
		}
		assets = append(assets, assetCandidate{
			assetType:    firstNonEmpty(ci.CIType, "ci"),
			assetID:      ci.CIUID,
			name:         ci.Name,
			status:       ci.Status,
			embedText:    strings.TrimSpace(ci.CIType + " " + ci.Name + " " + ci.Source + " " + ci.Owner),
			metadataJSON: marshalJSON(meta),
			updatedAt:    ci.UpdatedAt,
		})
	}

	var services []model.Service
	svcQ := s.db.WithContext(ctx).Order("updated_at ASC")
	if !since.IsZero() {
		svcQ = svcQ.Where("updated_at > ?", since)
	}
	if err := svcQ.Find(&services).Error; err != nil {
		return nil, err
	}
	for _, svc := range services {
		meta := map[string]any{
			"project_id":   svc.ProjectID,
			"env":          svc.Env,
			"runtime_type": svc.RuntimeType,
			"owner":        svc.Owner,
			"status":       svc.Status,
		}
		assets = append(assets, assetCandidate{
			assetType:    "service",
			assetID:      fmt.Sprintf("%d", svc.ID),
			name:         svc.Name,
			status:       svc.Status,
			embedText:    strings.TrimSpace("service " + svc.Name + " " + svc.Type + " " + svc.Env + " " + svc.RuntimeType),
			metadataJSON: marshalJSON(meta),
			updatedAt:    svc.UpdatedAt,
		})
	}

	var clusters []model.Cluster
	clusterQ := s.db.WithContext(ctx).Order("updated_at ASC")
	if !since.IsZero() {
		clusterQ = clusterQ.Where("updated_at > ?", since)
	}
	if err := clusterQ.Find(&clusters).Error; err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		meta := map[string]any{
			"type":        cluster.Type,
			"version":     cluster.Version,
			"endpoint":    cluster.Endpoint,
			"env_type":    cluster.EnvType,
			"source":      cluster.Source,
			"k8s_version": cluster.K8sVersion,
		}
		assets = append(assets, assetCandidate{
			assetType:    "cluster",
			assetID:      fmt.Sprintf("%d", cluster.ID),
			name:         cluster.Name,
			status:       cluster.Status,
			embedText:    strings.TrimSpace("cluster " + cluster.Name + " " + cluster.Type + " " + cluster.EnvType + " " + cluster.Status),
			metadataJSON: marshalJSON(meta),
			updatedAt:    cluster.UpdatedAt,
		})
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].updatedAt.Before(assets[j].updatedAt)
	})
	return assets, nil
}

type milvusRowBuilder func(doc *schema.Document, vector []float32) (interface{}, error)

type milvusDocumentIndexer struct {
	collection string
	store      vectorStore
	embedding  einoembedding.Embedder
	rowBuilder milvusRowBuilder
}

func newMilvusDocumentIndexer(collection string, store vectorStore, emb einoembedding.Embedder, builder milvusRowBuilder) einoindexer.Indexer {
	if store == nil || emb == nil || builder == nil {
		return nil
	}
	return &milvusDocumentIndexer{
		collection: collection,
		store:      store,
		embedding:  emb,
		rowBuilder: builder,
	}
}

func (m *milvusDocumentIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...einoindexer.Option) ([]string, error) {
	if m == nil || m.store == nil || m.embedding == nil || m.rowBuilder == nil {
		return nil, fmt.Errorf("milvus document indexer is not initialized")
	}
	if len(docs) == 0 {
		return nil, nil
	}
	co := einoindexer.GetCommonOptions(&einoindexer.Options{Embedding: m.embedding}, opts...)
	if co.Embedding == nil {
		return nil, fmt.Errorf("embedding component is required")
	}
	texts := make([]string, 0, len(docs))
	for _, d := range docs {
		if d == nil {
			texts = append(texts, "")
			continue
		}
		texts = append(texts, strings.TrimSpace(d.Content))
	}
	vectors64, err := co.Embedding.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, err
	}
	if len(vectors64) != len(docs) {
		return nil, fmt.Errorf("embedding count mismatch: got %d want %d", len(vectors64), len(docs))
	}
	rows := make([]interface{}, 0, len(docs))
	ids := make([]string, 0, len(docs))
	for i := range docs {
		doc := docs[i]
		if doc == nil {
			continue
		}
		vec := make([]float32, len(vectors64[i]))
		for j := range vectors64[i] {
			vec[j] = float32(vectors64[i][j])
		}
		row, err := m.rowBuilder(doc, vec)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
		if strings.TrimSpace(doc.ID) != "" {
			ids = append(ids, doc.ID)
		} else {
			ids = append(ids, fmt.Sprintf("%s-%d", m.collection, i))
		}
	}
	if err := batchInsertRows(ctx, m.store, m.collection, rows); err != nil {
		return nil, err
	}
	return ids, nil
}

func toolExampleRowBuilder(doc *schema.Document, vector []float32) (interface{}, error) {
	if doc == nil {
		return nil, fmt.Errorf("nil document")
	}
	meta := doc.MetaData
	return toolExampleMilvusRow{
		ToolName:      truncateString(toString(meta["tool_name"]), 256),
		Intent:        truncateString(toString(meta["intent"]), int(defaultVarcharMaxLength)),
		ParamsJSON:    truncateString(toString(meta["params_json"]), int(defaultVarcharMaxLength)),
		ResultSummary: truncateString(toString(meta["result_summary"]), int(defaultVarcharMaxLength)),
		Embedding:     entity.FloatVector(vector),
	}, nil
}

func platformAssetRowBuilder(doc *schema.Document, vector []float32) (interface{}, error) {
	if doc == nil {
		return nil, fmt.Errorf("nil document")
	}
	meta := doc.MetaData
	return platformAssetMilvusRow{
		AssetType:    truncateString(toString(meta["asset_type"]), 64),
		AssetID:      truncateString(toString(meta["asset_id"]), 128),
		Name:         truncateString(toString(meta["name"]), 512),
		Status:       truncateString(toString(meta["status"]), 64),
		MetadataJSON: truncateString(toString(meta["metadata_json"]), int(defaultVarcharMaxLength)),
		Embedding:    entity.FloatVector(vector),
	}, nil
}

func troubleshootingRowBuilder(doc *schema.Document, vector []float32) (interface{}, error) {
	if doc == nil {
		return nil, fmt.Errorf("nil document")
	}
	meta := doc.MetaData
	return troubleshootingMilvusRow{
		Title:     truncateString(toString(meta["title"]), 512),
		Symptom:   truncateString(toString(meta["symptom"]), int(defaultVarcharMaxLength)),
		Diagnosis: truncateString(toString(meta["diagnosis"]), int(defaultVarcharMaxLength)),
		Solution:  truncateString(toString(meta["solution"]), int(defaultVarcharMaxLength)),
		Embedding: entity.FloatVector(vector),
	}, nil
}

func batchInsertRows(ctx context.Context, store vectorStore, collection string, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	const batchSize = 128
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		if err := store.InsertRows(ctx, collection, rows[i:end]); err != nil {
			return err
		}
	}
	return store.FlushCollection(ctx, collection)
}

func extractToolName(row model.AICommandExecution) string {
	if v := strings.TrimSpace(row.Intent); v != "" {
		return v
	}
	parts := strings.Fields(strings.TrimSpace(row.CommandText))
	if len(parts) == 0 {
		return "unknown"
	}
	return parts[0]
}

func buildToolEmbeddingText(row model.AICommandExecution) string {
	payload := strings.TrimSpace(row.Intent + " " + row.CommandText + " " + row.ExecutionSummary)
	if payload == "" {
		return "tool execution"
	}
	return payload
}

func marshalJSON(v map[string]any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func firstNonEmpty(values ...string) string {
	for _, item := range values {
		v := strings.TrimSpace(item)
		if v != "" {
			return v
		}
	}
	return ""
}

func truncateString(input string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(input)
	if len(runes) <= max {
		return input
	}
	return string(runes[:max])
}
