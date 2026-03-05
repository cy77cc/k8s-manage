package monitoring

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNotificationGatewayHandleWebhook(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "notigw.db")
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
	processed, err := gw.HandleWebhook(context.Background(), AlertmanagerWebhook{
		Status: "firing",
		Alerts: []AlertmanagerAlert{{
			Status:      "firing",
			Fingerprint: "fp-1",
			Labels: map[string]string{
				"alertname": "CPUHigh",
				"severity":  "warning",
				"rule_id":   "1",
				"metric":    "cpu_usage",
			},
			Annotations: map[string]string{"summary": "cpu high"},
			StartsAt:    time.Now(),
		}},
	})
	if err != nil {
		t.Fatalf("handle webhook: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected processed=1, got %d", processed)
	}

	var alerts int64
	if err := db.Model(&model.AlertEvent{}).Count(&alerts).Error; err != nil {
		t.Fatalf("count alerts: %v", err)
	}
	if alerts != 1 {
		t.Fatalf("expected 1 alert event, got %d", alerts)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var deliveries int64
		_ = db.Model(&model.AlertNotificationDelivery{}).Count(&deliveries).Error
		if deliveries > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("expected at least one delivery record")
}
