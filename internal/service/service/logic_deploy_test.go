package service

import (
	"context"
	"strings"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/model"
	"github.com/cy77cc/OpsPilot/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDeployTestLogic(t *testing.T) *Logic {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:servicedeploy?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Service{}, &model.ServiceDeployTarget{}, &model.DeploymentTarget{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return NewLogic(&svc.ServiceContext{DB: db})
}

func TestResolveDeployTargetFallbackToDeploymentTarget(t *testing.T) {
	logic := newDeployTestLogic(t)
	ctx := context.Background()

	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{ID: 2, Name: "svc", Env: "production", ProjectID: 1, TeamID: 1}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.DeploymentTarget{ID: 10, Name: "k8s-prod", TargetType: "k8s", RuntimeType: "k8s", ClusterID: 88, ProjectID: 1, TeamID: 1, Env: "production", Status: "active", ReadinessStatus: "ready"}).Error; err != nil {
		t.Fatalf("seed deployment target: %v", err)
	}

	resp, err := logic.resolveDeployTarget(ctx, 2, DeployReq{Env: "production"})
	if err != nil {
		t.Fatalf("resolve deploy target: %v", err)
	}
	if resp.ClusterID != 88 {
		t.Fatalf("expected cluster id 88, got %d", resp.ClusterID)
	}
	if resp.DeployTarget != "k8s" {
		t.Fatalf("expected deploy target k8s, got %s", resp.DeployTarget)
	}

	var linked model.ServiceDeployTarget
	if err := logic.svcCtx.DB.WithContext(ctx).Where("service_id = ? AND is_default = 1", 2).First(&linked).Error; err != nil {
		t.Fatalf("expected default service deploy target persisted: %v", err)
	}
}

func TestResolveDeployTargetExplicitCluster(t *testing.T) {
	logic := newDeployTestLogic(t)
	ctx := context.Background()

	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{ID: 3, Name: "svc-3", Env: "staging", ProjectID: 2}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}

	resp, err := logic.resolveDeployTarget(ctx, 3, DeployReq{ClusterID: 99, DeployTarget: "k8s"})
	if err != nil {
		t.Fatalf("resolve explicit target: %v", err)
	}
	if resp.ClusterID != 99 || resp.IsDefault {
		t.Fatalf("expected explicit cluster target, got %+v", resp)
	}
}

func TestResolveDeployTargetUsesServiceDefault(t *testing.T) {
	logic := newDeployTestLogic(t)
	ctx := context.Background()
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{ID: 4, Name: "svc-4", Env: "staging", ProjectID: 1, TeamID: 2}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.ServiceDeployTarget{
		ServiceID:    4,
		ClusterID:    321,
		Namespace:    "default",
		DeployTarget: "k8s",
		IsDefault:    true,
	}).Error; err != nil {
		t.Fatalf("seed service default target: %v", err)
	}
	resp, err := logic.resolveDeployTarget(ctx, 4, DeployReq{Env: "staging"})
	if err != nil {
		t.Fatalf("resolve with service default: %v", err)
	}
	if resp.ClusterID != 321 || !resp.IsDefault {
		t.Fatalf("expected service default target, got %+v", resp)
	}
}

func TestResolveDeployTargetFallbackFailureHasDiagnostics(t *testing.T) {
	logic := newDeployTestLogic(t)
	ctx := context.Background()
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{ID: 5, Name: "svc-5", Env: "production", ProjectID: 8, TeamID: 9}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	_, err := logic.resolveDeployTarget(ctx, 5, DeployReq{Env: "production", DeployTarget: "k8s"})
	if err == nil {
		t.Fatalf("expected error when no fallback target")
	}
	msg := err.Error()
	if !containsAll(msg, []string{"deploy target not configured", "project_id=8", "team_id=9", "env=production", "target_type=k8s"}) {
		t.Fatalf("unexpected diagnostic message: %s", msg)
	}
}

func TestCacheFallbackDefaultTargetSkipsWhenAlreadyExists(t *testing.T) {
	logic := newDeployTestLogic(t)
	ctx := context.Background()
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{ID: 6, Name: "svc-6", Env: "staging", ProjectID: 1}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.ServiceDeployTarget{
		ServiceID:    6,
		ClusterID:    100,
		Namespace:    "default",
		DeployTarget: "k8s",
		IsDefault:    true,
	}).Error; err != nil {
		t.Fatalf("seed existing default: %v", err)
	}

	err := logic.cacheFallbackDefaultTarget(ctx, 6, DeployTargetResp{
		ServiceID:    6,
		ClusterID:    200,
		Namespace:    "prod",
		DeployTarget: "k8s",
	})
	if err != nil {
		t.Fatalf("cache fallback should not fail: %v", err)
	}

	var rows []model.ServiceDeployTarget
	if err := logic.svcCtx.DB.WithContext(ctx).Where("service_id = ?", 6).Find(&rows).Error; err != nil {
		t.Fatalf("query defaults: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one default target row, got %d", len(rows))
	}
	if rows[0].ClusterID != 100 {
		t.Fatalf("existing default should be unchanged, got cluster_id=%d", rows[0].ClusterID)
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
