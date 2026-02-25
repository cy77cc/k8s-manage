package cicd

import (
	"context"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
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
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return NewLogic(&svc.ServiceContext{DB: db, Rdb: redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})})
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
	_, err := logic.UpsertDeploymentCDConfig(ctx, 2, 201, UpsertDeploymentCDConfigReq{
		Env:              "prod",
		RuntimeType:      "k8s",
		Strategy:         "rolling",
		StrategyConfig:   map[string]any{"batch": 1},
		ApprovalRequired: true,
	})
	if err != nil {
		t.Fatalf("upsert cd config: %v", err)
	}
	release, err := logic.TriggerRelease(ctx, 3, TriggerReleaseReq{ServiceID: 101, DeploymentID: 201, Env: "prod", RuntimeType: "k8s", Version: "v1.0.0"})
	if err != nil {
		t.Fatalf("trigger release: %v", err)
	}
	if release.Status != "pending_approval" {
		t.Fatalf("expected pending_approval, got %s", release.Status)
	}
	approved, err := logic.ApproveRelease(ctx, 4, release.ID, "looks good")
	if err != nil {
		t.Fatalf("approve release: %v", err)
	}
	if approved.Status != "succeeded" {
		t.Fatalf("expected succeeded after approval flow, got %s", approved.Status)
	}
	rolled, err := logic.RollbackRelease(ctx, 5, release.ID, "v0.9.0", "rollback")
	if err != nil {
		t.Fatalf("rollback release: %v", err)
	}
	if rolled.Status != "rolled_back" {
		t.Fatalf("expected rolled_back, got %s", rolled.Status)
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
