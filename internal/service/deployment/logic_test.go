package deployment

import (
	"context"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDeploymentTestLogic(t *testing.T) *Logic {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:deploytest?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Cluster{},
		&model.Node{},
		&model.Service{},
		&model.DeploymentTarget{},
		&model.DeploymentTargetNode{},
		&model.DeploymentRelease{},
		&model.DeploymentReleaseApproval{},
		&model.DeploymentReleaseAudit{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return NewLogic(&svc.ServiceContext{DB: db})
}

func TestApplyReleaseProductionRequiresApproval(t *testing.T) {
	logic := newDeploymentTestLogic(t)
	ctx := context.Background()
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Node{ID: 2, Name: "n2", IP: "10.0.0.2", SSHUser: "root", Status: "active"}).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{ID: 201, Name: "svc-b", Env: "production", YamlContent: "services:\n  app:\n    image: nginx:latest"}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	target, err := logic.CreateTarget(ctx, 1, TargetUpsertReq{
		Name:       "compose-prod",
		TargetType: "compose",
		Env:        "production",
		Nodes:      []TargetNodeReq{{HostID: 2, Role: "manager", Weight: 100}},
	})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	resp, err := logic.ApplyRelease(ctx, 7, ReleasePreviewReq{ServiceID: 201, TargetID: target.ID, Env: "production", Strategy: "rolling"})
	if err != nil {
		t.Fatalf("apply release: %v", err)
	}
	if !resp.ApprovalRequired {
		t.Fatalf("expected approval required")
	}
	if resp.Status != releaseStatusPendingApproval {
		t.Fatalf("expected pending approval, got %s", resp.Status)
	}
	events, err := logic.ListReleaseTimeline(ctx, resp.ReleaseID)
	if err != nil {
		t.Fatalf("list timeline: %v", err)
	}
	if len(events) == 0 {
		t.Fatalf("expected timeline events")
	}
}

func TestCreateComposeTargetRejectUnavailableNode(t *testing.T) {
	logic := newDeploymentTestLogic(t)
	ctx := context.Background()
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Node{ID: 1, Name: "n1", IP: "10.0.0.1", SSHUser: "root", Status: "offline"}).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	target, err := logic.CreateTarget(ctx, 1, TargetUpsertReq{
		Name:       "edge",
		TargetType: "compose",
		Env:        "staging",
		Nodes: []TargetNodeReq{{
			HostID: 1,
			Role:   "manager",
			Weight: 100,
		}},
	})
	if err == nil {
		t.Fatalf("expected create target to fail, got %+v", target)
	}
}

func TestApplyReleasePersistsFailureDiagnostics(t *testing.T) {
	logic := newDeploymentTestLogic(t)
	ctx := context.Background()
	cluster := model.Cluster{Name: "c1", Endpoint: "https://127.0.0.1:6443", Status: "active", Type: "kubernetes"}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&cluster).Error; err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{ID: 101, Name: "svc-a", Env: "staging", YamlContent: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: a"}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	target, err := logic.CreateTarget(ctx, 1, TargetUpsertReq{Name: "k8s-target", TargetType: "k8s", ClusterID: cluster.ID, Env: "staging"})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Delete(&model.Cluster{}, cluster.ID).Error; err != nil {
		t.Fatalf("delete cluster: %v", err)
	}
	_, err = logic.ApplyRelease(ctx, 1, ReleasePreviewReq{ServiceID: 101, TargetID: target.ID, Env: "staging", Strategy: "rolling"})
	if err == nil {
		t.Fatalf("expected apply release to fail")
	}
	rows, err := logic.ListReleases(ctx, 101, target.ID, "k8s")
	if err != nil {
		t.Fatalf("list releases: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("expected release row")
	}
	if rows[0].Status != releaseStatusFailed {
		t.Fatalf("expected failed status, got %s", rows[0].Status)
	}
	if rows[0].DiagnosticsJSON == "" || rows[0].DiagnosticsJSON == "[]" {
		t.Fatalf("expected diagnostics payload")
	}
}
