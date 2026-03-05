package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"testing"
	"time"

	prominfra "github.com/cy77cc/k8s-manage/internal/infra/prometheus"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDashboardQueryP95Under50ms(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"cpu_usage"},"values":[[1710000000,"11.2"],[1710000060,"12.6"]]}]}}`))
	}))
	defer ts.Close()

	promClient, err := prominfra.NewClient(prominfra.Config{Address: ts.URL, Timeout: 2 * time.Second, RetryCount: 1})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	dbPath := filepath.Join(t.TempDir(), "perfmetrics.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	logic := NewLogic(&svc.ServiceContext{DB: db, Prometheus: promClient})

	samples := make([]float64, 0, 30)
	for i := 0; i < 30; i++ {
		start := time.Now()
		_, err := logic.GetMetrics(context.Background(), MetricQuery{
			Metric:         "cpu_usage",
			Start:          time.Unix(1710000000, 0),
			End:            time.Unix(1710000200, 0),
			GranularitySec: 60,
		})
		if err != nil {
			t.Fatalf("get metrics: %v", err)
		}
		samples = append(samples, float64(time.Since(start).Milliseconds()))
	}
	sort.Float64s(samples)
	p95 := samples[int(float64(len(samples))*0.95)-1]
	if p95 >= 50 {
		t.Fatalf("expected p95 < 50ms, got %.2fms", p95)
	}
}

func TestAlertNotificationLatencyUnder30s(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "perfnotify.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.AlertEvent{}, &model.AlertNotificationChannel{}, &model.AlertNotificationDelivery{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Create(&model.AlertNotificationChannel{Name: "default-log", Type: "log", Provider: "log", Enabled: true}).Error; err != nil {
		t.Fatalf("seed channel: %v", err)
	}
	gw := NewNotificationGateway(&svc.ServiceContext{DB: db})
	start := time.Now()
	_, err = gw.HandleWebhook(context.Background(), AlertmanagerWebhook{
		Status: "firing",
		Alerts: []AlertmanagerAlert{{
			Status:      "firing",
			Fingerprint: "perf-fp-1",
			Labels:      map[string]string{"alertname": "CPUHigh", "severity": "warning", "rule_id": "1"},
			StartsAt:    time.Now(),
		}},
	})
	if err != nil {
		t.Fatalf("handle webhook: %v", err)
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		var deliveries int64
		_ = db.Model(&model.AlertNotificationDelivery{}).Count(&deliveries).Error
		if deliveries > 0 {
			latency := time.Since(start)
			if latency >= 30*time.Second {
				t.Fatalf("expected latency < 30s, got %s", latency)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("delivery did not complete within 30s")
}

func TestRuleSyncLatencyUnder5s(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "perfrulesync.db")
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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	service := &RuleSyncService{
		db:        db,
		rulesFile: t.TempDir() + "/alerting_rules.yml",
		reloadURL: ts.URL + "/-/reload",
		client:    &http.Client{Timeout: 2 * time.Second},
	}
	start := time.Now()
	_, err = service.SyncRules(context.Background())
	if err != nil {
		t.Fatalf("sync rules: %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 5*time.Second {
		t.Fatalf("expected sync < 5s, got %s", elapsed)
	}
}
