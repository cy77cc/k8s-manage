package rag

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeVectorStore struct {
	rowsByCollection map[string][]interface{}
	flushes          map[string]int
}

func newFakeVectorStore() *fakeVectorStore {
	return &fakeVectorStore{
		rowsByCollection: make(map[string][]interface{}),
		flushes:          make(map[string]int),
	}
}

func (f *fakeVectorStore) InsertRows(_ context.Context, collection string, rows []interface{}) error {
	f.rowsByCollection[collection] = append(f.rowsByCollection[collection], rows...)
	return nil
}

func (f *fakeVectorStore) FlushCollection(_ context.Context, collection string) error {
	f.flushes[collection]++
	return nil
}

type fakeEmbedder struct{}

func (f fakeEmbedder) EmbedStrings(_ context.Context, texts []string, _ ...embedding.Option) ([][]float64, error) {
	out := make([][]float64, 0, len(texts))
	for i := range texts {
		out = append(out, []float64{float64(i + 1), 0.5})
	}
	return out, nil
}

func setupIngestionDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", t.Name(), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.AICommandExecution{},
		&model.CMDBCI{},
		&model.Service{},
		&model.Cluster{},
	); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	return db
}

func TestIngestToolExamples(t *testing.T) {
	db := setupIngestionDB(t)
	store := newFakeVectorStore()
	svc := NewIngestionService(db, store, fakeEmbedder{})

	now := time.Now().UTC()
	rows := []model.AICommandExecution{
		{ID: "1", Status: "succeeded", CommandText: "deployment.release service_id=1", Intent: "deployment.release", ParamsJSON: `{"service_id":1}`, ExecutionSummary: "ok", UpdatedAt: now.Add(-2 * time.Minute)},
		{ID: "2", Status: "failed", CommandText: "host.exec id=1", Intent: "host.exec", ParamsJSON: `{}`, ExecutionSummary: "failed", UpdatedAt: now.Add(-1 * time.Minute)},
		{ID: "3", Status: "approved", CommandText: "cicd.release.approve id=2", Intent: "", ParamsJSON: `{"id":2}`, ExecutionSummary: "approved", UpdatedAt: now},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed commands: %v", err)
	}

	stats, err := svc.IngestToolExamples(context.Background(), now.Add(-90*time.Second))
	if err != nil {
		t.Fatalf("ingest tool examples: %v", err)
	}
	if stats.Total != 1 || stats.Inserted != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if len(store.rowsByCollection[CollectionToolExamples]) != 1 {
		t.Fatalf("expected 1 inserted row, got %d", len(store.rowsByCollection[CollectionToolExamples]))
	}
	if store.flushes[CollectionToolExamples] != 1 {
		t.Fatalf("expected flush on collection %s", CollectionToolExamples)
	}
}

func TestIngestPlatformAssets(t *testing.T) {
	db := setupIngestionDB(t)
	store := newFakeVectorStore()
	svc := NewIngestionService(db, store, fakeEmbedder{})

	now := time.Now().UTC()
	if err := db.Create(&model.CMDBCI{CIUID: "ci-1", CIType: "host", Name: "hk-host-1", Source: "manual", Status: "active", UpdatedAt: now.Add(-2 * time.Hour)}).Error; err != nil {
		t.Fatalf("seed cmdb ci: %v", err)
	}
	if err := db.Create(&model.Service{Name: "api-service", Type: "stateless", Status: "running", Env: "staging", RuntimeType: "k8s", UpdatedAt: now.Add(-30 * time.Minute)}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	if err := db.Create(&model.Cluster{Name: "prod-cluster", Type: "kubernetes", Status: "ready", EnvType: "production", UpdatedAt: now.Add(-10 * time.Minute)}).Error; err != nil {
		t.Fatalf("seed cluster: %v", err)
	}

	stats, err := svc.IngestPlatformAssets(context.Background(), now.Add(-45*time.Minute))
	if err != nil {
		t.Fatalf("ingest assets: %v", err)
	}
	if stats.Total != 2 || stats.Inserted != 2 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if got := len(store.rowsByCollection[CollectionPlatformAssets]); got != 2 {
		t.Fatalf("expected 2 inserted asset rows, got %d", got)
	}
}

func TestIngestTroubleshootingCasesAndIncremental(t *testing.T) {
	db := setupIngestionDB(t)
	store := newFakeVectorStore()
	svc := NewIngestionService(db, store, fakeEmbedder{})

	fixedNow := time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)
	svc.nowFn = func() time.Time { return fixedNow }

	if err := db.Create(&model.AICommandExecution{ID: "cmd-1", Status: "succeeded", CommandText: "ops.aggregate.status", Intent: "ops.aggregate.status", UpdatedAt: fixedNow.Add(-time.Minute)}).Error; err != nil {
		t.Fatalf("seed command: %v", err)
	}

	cases := []TroubleshootingCaseInput{
		{Title: "Nginx 502", Symptom: "502 bad gateway", Diagnosis: "upstream timeout", Solution: "restart upstream", UpdatedAt: fixedNow.Add(-2 * time.Minute)},
		{Title: "Pod OOM", Symptom: "oom killed", Diagnosis: "limit too low", Solution: "increase limit", UpdatedAt: fixedNow.Add(-10 * time.Second)},
	}

	first, err := svc.IngestIncremental(context.Background(), cases)
	if err != nil {
		t.Fatalf("first incremental ingest: %v", err)
	}
	if first[checkpointToolExamples].Inserted != 1 {
		t.Fatalf("expected first tool ingest to insert 1, got %+v", first[checkpointToolExamples])
	}
	if first[checkpointTroubleshooting].Inserted != 2 {
		t.Fatalf("expected first troubleshooting ingest to insert 2, got %+v", first[checkpointTroubleshooting])
	}

	if got := svc.Checkpoint(checkpointToolExamples); got.IsZero() {
		t.Fatalf("expected non-zero tool checkpoint")
	}
	if got := svc.Checkpoint(checkpointTroubleshooting); got.IsZero() {
		t.Fatalf("expected non-zero case checkpoint")
	}

	second, err := svc.IngestIncremental(context.Background(), cases)
	if err != nil {
		t.Fatalf("second incremental ingest: %v", err)
	}
	if second[checkpointToolExamples].Inserted != 0 {
		t.Fatalf("expected second tool ingest to insert 0, got %+v", second[checkpointToolExamples])
	}
	if second[checkpointTroubleshooting].Inserted != 0 {
		t.Fatalf("expected second case ingest to insert 0, got %+v", second[checkpointTroubleshooting])
	}

	if len(store.rowsByCollection[CollectionTroubleshooting]) != 2 {
		t.Fatalf("expected 2 troubleshooting rows in store, got %d", len(store.rowsByCollection[CollectionTroubleshooting]))
	}
}

func TestExtractToolNameFallback(t *testing.T) {
	row := model.AICommandExecution{Intent: "", CommandText: "host.batch.exec.apply host_ids=1"}
	if got := extractToolName(row); got != "host.batch.exec.apply" {
		t.Fatalf("unexpected tool name: %s", got)
	}

	if got := extractToolName(model.AICommandExecution{}); got != "unknown" {
		t.Fatalf("unexpected fallback tool name: %s", got)
	}
}

func TestBatchInsertRowsBatches(t *testing.T) {
	store := newFakeVectorStore()
	rows := make([]interface{}, 0, 300)
	for i := 0; i < 300; i++ {
		rows = append(rows, fmt.Sprintf("row-%d", i))
	}
	if err := batchInsertRows(context.Background(), store, CollectionToolExamples, rows); err != nil {
		t.Fatalf("batch insert rows: %v", err)
	}
	if len(store.rowsByCollection[CollectionToolExamples]) != 300 {
		t.Fatalf("expected 300 rows inserted, got %d", len(store.rowsByCollection[CollectionToolExamples]))
	}
	if store.flushes[CollectionToolExamples] != 1 {
		t.Fatalf("expected single flush, got %d", store.flushes[CollectionToolExamples])
	}
}
