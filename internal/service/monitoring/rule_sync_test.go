package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cy77cc/OpsPilot/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConvertRuleToPrometheus(t *testing.T) {
	rule := model.AlertRule{
		ID:          1,
		Name:        "CPU high",
		Metric:      "cpu_usage",
		Operator:    "gt",
		Threshold:   85,
		DurationSec: 300,
		Severity:    "warning",
		Source:      "host",
	}
	pr, err := convertRuleToPrometheus(rule)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if pr.Expr != "cpu_usage > 85" {
		t.Fatalf("unexpected expr: %s", pr.Expr)
	}
	if pr.For != "5m0s" {
		t.Fatalf("unexpected for: %s", pr.For)
	}
	if pr.Labels["rule_id"] != "1" {
		t.Fatalf("unexpected labels: %+v", pr.Labels)
	}
}

func TestRuleSyncServiceSyncRules(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "rulesync.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.AlertRule{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Create(&model.AlertRule{Name: "CPU high", Metric: "cpu_usage", Operator: "gt", Threshold: 80, Enabled: true, Severity: "warning", DurationSec: 120}).Error; err != nil {
		t.Fatalf("seed rule: %v", err)
	}

	reloadCalls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reloadCalls++
		if r.URL.Path != "/-/reload" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	rulesFile := filepath.Join(t.TempDir(), "alerting_rules.yml")
	svc := &RuleSyncService{
		db:        db,
		rulesFile: rulesFile,
		reloadURL: ts.URL + "/-/reload",
		client:    &http.Client{Timeout: 2 * time.Second},
	}
	n, err := svc.SyncRules(context.Background())
	if err != nil {
		t.Fatalf("sync rules: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected synced=1, got %d", n)
	}
	if reloadCalls != 1 {
		t.Fatalf("expected 1 reload call, got %d", reloadCalls)
	}
	b, err := os.ReadFile(rulesFile)
	if err != nil {
		t.Fatalf("read rules file: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("rules file is empty")
	}
}
