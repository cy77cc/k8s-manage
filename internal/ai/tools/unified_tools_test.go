package tools

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newUnifiedToolsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Service{}, &model.AlertEvent{}, &model.MetricPoint{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func TestK8sQueryMissingResource(t *testing.T) {
	res, err := k8sQuery(context.Background(), PlatformDeps{}, K8sQueryInput{})
	if err != nil {
		t.Fatalf("k8sQuery returned error: %v", err)
	}
	if res.OK {
		t.Fatalf("expected invalid result for missing resource")
	}
	if res.ErrorCode != "missing_param" {
		t.Fatalf("expected missing_param, got %s", res.ErrorCode)
	}
}

func TestHostExecMissingCommand(t *testing.T) {
	res, err := hostExec(context.Background(), PlatformDeps{}, HostExecInput{HostID: 1})
	if err != nil {
		t.Fatalf("hostExec returned error: %v", err)
	}
	if res.OK {
		t.Fatalf("expected invalid result for missing command")
	}
	if res.ErrorCode != "missing_param" {
		t.Fatalf("expected missing_param, got %s", res.ErrorCode)
	}
}

func TestServiceStatusReturnsServiceSummary(t *testing.T) {
	db := newUnifiedToolsTestDB(t)
	svc := model.Service{
		ProjectID:   1,
		TeamID:      1,
		Name:        "gateway",
		Type:        "stateless",
		Image:       "nginx:1.27",
		Status:      "running",
		Env:         "prod",
		RuntimeType: "k8s",
		Replicas:    3,
	}
	if err := db.Create(&svc).Error; err != nil {
		t.Fatalf("create service: %v", err)
	}

	res, err := serviceStatus(context.Background(), PlatformDeps{DB: db}, ServiceStatusInput{ServiceID: int(svc.ID)})
	if err != nil {
		t.Fatalf("serviceStatus returned error: %v", err)
	}
	if !res.OK {
		t.Fatalf("expected ok result, got %s", res.Error)
	}
	payload, ok := res.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", res.Data)
	}
	if payload["status"] != "running" {
		t.Fatalf("unexpected status: %#v", payload["status"])
	}
}

func TestMonitorAlertReturnsFiringAlerts(t *testing.T) {
	db := newUnifiedToolsTestDB(t)
	now := time.Now()
	if err := db.Create(&model.AlertEvent{
		Title:       "CPU high",
		Message:     "cpu > 90%",
		Metric:      "cpu_usage",
		Severity:    "critical",
		Source:      "service:1",
		Status:      "firing",
		TriggeredAt: now,
	}).Error; err != nil {
		t.Fatalf("create alert event: %v", err)
	}

	res, err := monitorAlert(context.Background(), PlatformDeps{DB: db}, MonitorAlertInput{Severity: "critical", ServiceID: 1})
	if err != nil {
		t.Fatalf("monitorAlert returned error: %v", err)
	}
	if !res.OK {
		t.Fatalf("expected ok result, got %s", res.Error)
	}
	payload, ok := res.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", res.Data)
	}
	if payload["total"] != 1 {
		t.Fatalf("expected one alert, got %#v", payload["total"])
	}
}

func TestMonitorMetricReturnsPoints(t *testing.T) {
	db := newUnifiedToolsTestDB(t)
	now := time.Now()
	if err := db.Create(&model.MetricPoint{
		Metric:    "cpu_usage",
		Source:    "service:1",
		Value:     87.2,
		Collected: now.Add(-5 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("create metric point: %v", err)
	}

	res, err := monitorMetric(context.Background(), PlatformDeps{DB: db}, MonitorMetricInput{Query: "cpu_usage", TimeRange: "1h"})
	if err != nil {
		t.Fatalf("monitorMetric returned error: %v", err)
	}
	if !res.OK {
		t.Fatalf("expected ok result, got %s", res.Error)
	}
	payload, ok := res.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", res.Data)
	}
	if payload["count"] != 1 {
		t.Fatalf("expected one metric point, got %#v", payload["count"])
	}
}
