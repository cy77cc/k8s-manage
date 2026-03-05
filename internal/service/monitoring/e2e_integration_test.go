package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	prominfra "github.com/cy77cc/k8s-manage/internal/infra/prometheus"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMetricsQueryEndToEnd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"cpu_usage","source":"host"},"values":[[1710000000,"11.2"],[1710000060,"12.6"]]}]}}`))
	}))
	defer ts.Close()

	promClient, err := prominfra.NewClient(prominfra.Config{Address: ts.URL, Timeout: 2 * time.Second, RetryCount: 1})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	dbPath := filepath.Join(t.TempDir(), "e2emetrics.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	logic := NewLogic(&svc.ServiceContext{DB: db, Prometheus: promClient})

	out, err := logic.GetMetrics(context.Background(), MetricQuery{
		Metric:         "cpu_usage",
		Start:          time.Unix(1710000000, 0),
		End:            time.Unix(1710000200, 0),
		GranularitySec: 60,
		Source:         "host",
	})
	if err != nil {
		t.Fatalf("get metrics: %v", err)
	}
	if len(out.Series) != 2 {
		t.Fatalf("expected 2 points, got %d", len(out.Series))
	}
}

func TestAlertWebhookToNotificationEndToEnd(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "e2ewebhook.db")
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
	_, err = gw.HandleWebhook(context.Background(), AlertmanagerWebhook{
		Status: "firing",
		Alerts: []AlertmanagerAlert{{
			Status:      "firing",
			Fingerprint: "e2e-fp-1",
			Labels: map[string]string{
				"alertname": "CPUHigh",
				"severity":  "warning",
				"rule_id":   "1",
			},
			Annotations: map[string]string{"summary": "cpu high"},
			StartsAt:    time.Now(),
		}},
	})
	if err != nil {
		t.Fatalf("handle webhook: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var deliveries int64
		_ = db.Model(&model.AlertNotificationDelivery{}).Count(&deliveries).Error
		if deliveries > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected delivery record after webhook")
}

func TestRuleSyncEndToEnd(t *testing.T) {
	TestRuleSyncServiceSyncRules(t)
}
