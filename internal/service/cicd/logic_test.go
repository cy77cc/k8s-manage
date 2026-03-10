package cicd

import (
	"context"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/model"
	"github.com/cy77cc/OpsPilot/internal/svc"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestLogic(t *testing.T) *Logic {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:cicdtest?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.CICDServiceCIConfig{},
		&model.CICDServiceCIRun{},
		&model.CICDDeploymentCDConfig{},
		&model.CICDRelease{},
		&model.CICDReleaseApproval{},
		&model.CICDAuditEvent{},
		&model.Service{},
		&model.Cluster{},
		&model.DeploymentTarget{},
		&model.DeploymentRelease{},
		&model.DeploymentReleaseApproval{},
		&model.DeploymentReleaseAudit{},
		&model.ServiceDeployTarget{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return NewLogic(&svc.ServiceContext{DB: db, Rdb: redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})})
}

func TestTriggerReleaseFallsBackToServiceDefaultTarget(t *testing.T) {
	logic := newTestLogic(t)
	ctx := context.Background()

	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{
		ID:          102,
		Name:        "svc-fallback",
		Env:         "production",
		ProjectID:   11,
		TeamID:      12,
		YamlContent: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: svc-fallback\n",
	}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Cluster{
		ID:         2,
		Name:       "cluster-2",
		KubeConfig: "invalid-kubeconfig",
		Status:     "active",
	}).Error; err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.DeploymentTarget{
		ID:              301,
		Name:            "target-fallback",
		TargetType:      "k8s",
		RuntimeType:     "k8s",
		ClusterID:       2,
		ProjectID:       11,
		TeamID:          12,
		Env:             "production",
		Status:          "active",
		ReadinessStatus: "ready",
	}).Error; err != nil {
		t.Fatalf("seed target: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.ServiceDeployTarget{
		ServiceID:    102,
		ClusterID:    2,
		Namespace:    "default",
		DeployTarget: "k8s",
		IsDefault:    true,
	}).Error; err != nil {
		t.Fatalf("seed service deploy target: %v", err)
	}

	release, err := logic.TriggerRelease(ctx, 3, TriggerReleaseReq{
		ServiceID:   102,
		Env:         "production",
		RuntimeType: "k8s",
		Version:     "v1.2.3",
	})
	if err != nil {
		t.Fatalf("trigger release with fallback target: %v", err)
	}
	if release.DeploymentID != 301 {
		t.Fatalf("expected fallback deployment id 301, got %d", release.DeploymentID)
	}
	if release.Status != "pending_approval" {
		t.Fatalf("expected pending_approval, got %s", release.Status)
	}
}

func TestTriggerModeValidation(t *testing.T) {
	logic := newTestLogic(t)
	ctx := context.Background()
	_, err := logic.UpsertServiceCIConfig(ctx, 1, 101, UpsertServiceCIConfigReq{
		RepoURL:        "https://git.example.com/repo.git",
		Branch:         "main",
		BuildSteps:     []string{"go test ./..."},
		ArtifactTarget: "registry.example.com/repo:v1",
		TriggerMode:    "manual",
	})
	if err != nil {
		t.Fatalf("upsert ci config: %v", err)
	}
	if _, err := logic.TriggerCIRun(ctx, 1, 101, TriggerCIRunReq{TriggerType: "source-event"}); err == nil {
		t.Fatalf("expected source-event to be blocked for manual trigger mode")
	}
	if _, err := logic.TriggerCIRun(ctx, 1, 101, TriggerCIRunReq{TriggerType: "manual"}); err != nil {
		t.Fatalf("manual trigger should pass: %v", err)
	}
}

func TestReleaseStateTransitions(t *testing.T) {
	logic := newTestLogic(t)
	ctx := context.Background()
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{
		ID:          101,
		Name:        "svc-ci",
		Env:         "prod",
		YamlContent: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: svc-ci\n",
	}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Cluster{
		ID:         1,
		Name:       "cluster-1",
		KubeConfig: "invalid-kubeconfig",
		Status:     "active",
	}).Error; err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.DeploymentTarget{
		ID:              201,
		Name:            "target-1",
		TargetType:      "k8s",
		RuntimeType:     "k8s",
		ClusterID:       1,
		Env:             "prod",
		Status:          "active",
		ReadinessStatus: "ready",
	}).Error; err != nil {
		t.Fatalf("seed target: %v", err)
	}

	_, err := logic.UpsertDeploymentCDConfig(ctx, 2, 201, UpsertDeploymentCDConfigReq{
		Env:              "production",
		RuntimeType:      "k8s",
		Strategy:         "rolling",
		StrategyConfig:   map[string]any{"batch": 1},
		ApprovalRequired: true,
	})
	if err != nil {
		t.Fatalf("upsert cd config: %v", err)
	}
	release, err := logic.TriggerRelease(ctx, 3, TriggerReleaseReq{ServiceID: 101, DeploymentID: 201, Env: "production", RuntimeType: "k8s", Version: "v1.0.0"})
	if err != nil {
		t.Fatalf("trigger release: %v", err)
	}
	if release.Status != "pending_approval" {
		t.Fatalf("expected pending_approval, got %s", release.Status)
	}
	if release.UnifiedReleaseID == 0 {
		t.Fatalf("expected unified release id")
	}
	rejected, err := logic.RejectRelease(ctx, 4, release.ID, "no go")
	if err != nil {
		t.Fatalf("reject release: %v", err)
	}
	if rejected.Status != "rejected" {
		t.Fatalf("expected rejected, got %s", rejected.Status)
	}
}

func TestComposeCanaryRejected(t *testing.T) {
	logic := newTestLogic(t)
	ctx := context.Background()
	_, err := logic.UpsertDeploymentCDConfig(ctx, 2, 201, UpsertDeploymentCDConfigReq{
		Env:              "staging",
		RuntimeType:      "compose",
		Strategy:         "canary",
		StrategyConfig:   map[string]any{"traffic_percent": 10, "steps": []int{10, 50}},
		ApprovalRequired: false,
	})
	if err == nil {
		t.Fatalf("expected compose canary to be rejected")
	}
}
