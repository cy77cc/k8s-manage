package deployment

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/config"
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
		&model.EnvironmentInstallJob{},
		&model.EnvironmentInstallJobStep{},
		&model.ClusterCredential{},
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
	if err == nil {
		t.Fatalf("expected apply without preview token to fail")
	}
	preview, err := logic.PreviewRelease(ctx, ReleasePreviewReq{ServiceID: 201, TargetID: target.ID, Env: "production", Strategy: "rolling"})
	if err != nil {
		t.Fatalf("preview release: %v", err)
	}
	resp, err = logic.ApplyRelease(ctx, 7, ReleasePreviewReq{
		ServiceID:    201,
		TargetID:     target.ID,
		Env:          "production",
		Strategy:     "rolling",
		PreviewToken: preview.PreviewToken,
	})
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
	preview, err := logic.PreviewRelease(ctx, ReleasePreviewReq{ServiceID: 101, TargetID: target.ID, Env: "staging", Strategy: "rolling"})
	if err != nil {
		t.Fatalf("preview release: %v", err)
	}
	_, err = logic.ApplyRelease(ctx, 1, ReleasePreviewReq{
		ServiceID:    101,
		TargetID:     target.ID,
		Env:          "staging",
		Strategy:     "rolling",
		PreviewToken: preview.PreviewToken,
	})
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

func TestApplyReleaseRejectsMismatchedPreviewToken(t *testing.T) {
	logic := newDeploymentTestLogic(t)
	ctx := context.Background()
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Node{ID: 3, Name: "n3", IP: "10.0.0.3", SSHUser: "root", Status: "active"}).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{ID: 301, Name: "svc-c", Env: "staging", YamlContent: "services:\n  app:\n    image: nginx:latest"}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	target, err := logic.CreateTarget(ctx, 1, TargetUpsertReq{
		Name:       "compose-staging",
		TargetType: "compose",
		Env:        "staging",
		Nodes:      []TargetNodeReq{{HostID: 3, Role: "manager", Weight: 100}},
	})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	preview, err := logic.PreviewRelease(ctx, ReleasePreviewReq{ServiceID: 301, TargetID: target.ID, Env: "staging", Strategy: "rolling"})
	if err != nil {
		t.Fatalf("preview release: %v", err)
	}
	_, err = logic.ApplyRelease(ctx, 1, ReleasePreviewReq{
		ServiceID:    301,
		TargetID:     target.ID,
		Env:          "staging",
		Strategy:     "canary",
		PreviewToken: preview.PreviewToken,
	})
	if err == nil {
		t.Fatalf("expected mismatched preview token to be rejected")
	}
}

func TestPreviewReleaseRejectsNonReadyTarget(t *testing.T) {
	logic := newDeploymentTestLogic(t)
	ctx := context.Background()
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Node{ID: 6, Name: "n6", IP: "10.0.0.6", SSHUser: "root", Status: "active"}).Error; err != nil {
		t.Fatalf("seed node: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Create(&model.Service{ID: 401, Name: "svc-d", Env: "staging", YamlContent: "services:\n  app:\n    image: nginx:latest"}).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}
	target, err := logic.CreateTarget(ctx, 1, TargetUpsertReq{
		Name:       "compose-pending",
		TargetType: "compose",
		Env:        "staging",
		Nodes:      []TargetNodeReq{{HostID: 6, Role: "manager", Weight: 100}},
	})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	if err := logic.svcCtx.DB.WithContext(ctx).Model(&model.DeploymentTarget{}).Where("id = ?", target.ID).Update("readiness_status", "bootstrap_pending").Error; err != nil {
		t.Fatalf("update readiness: %v", err)
	}
	_, err = logic.PreviewRelease(ctx, ReleasePreviewReq{ServiceID: 401, TargetID: target.ID, Env: "staging", Strategy: "rolling"})
	if err == nil {
		t.Fatalf("expected preview release to fail on non-ready target")
	}
}

func TestImportExternalCredentialEncryptsPayload(t *testing.T) {
	logic := newDeploymentTestLogic(t)
	ctx := context.Background()
	config.CFG.Security.EncryptionKey = "12345678901234567890123456789012"

	kubeconfig := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: local
contexts:
- context:
    cluster: local
    user: local-user
  name: local
current-context: local
users:
- name: local-user
  user:
    token: dummy`
	resp, err := logic.ImportExternalCredential(ctx, 7, ClusterCredentialImportReq{
		Name:        "ext-k8s",
		RuntimeType: "k8s",
		AuthMethod:  "kubeconfig",
		Kubeconfig:  kubeconfig,
	})
	if err != nil {
		t.Fatalf("import credential: %v", err)
	}
	if resp.ID == 0 {
		t.Fatalf("expected credential id")
	}
	var row model.ClusterCredential
	if err := logic.svcCtx.DB.WithContext(ctx).First(&row, resp.ID).Error; err != nil {
		t.Fatalf("query credential: %v", err)
	}
	if row.KubeconfigEnc == "" || row.KubeconfigEnc == kubeconfig {
		t.Fatalf("expected kubeconfig to be encrypted")
	}
}

func TestStartEnvironmentBootstrapRejectsChecksumMismatch(t *testing.T) {
	logic := newDeploymentTestLogic(t)
	ctx := context.Background()
	root := filepath.Join("script", "runtime", "k8s", "v-test-bad")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	if err := os.WriteFile(filepath.Join(root, "runtime-package.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write package: %v", err)
	}
	manifest := `{"runtime":"k8s","version":"v-test-bad","package_file":"runtime-package.txt","sha256":"deadbeef","install_command":"echo install"}`
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	_, err := logic.StartEnvironmentBootstrap(ctx, 1, EnvironmentBootstrapReq{
		Name:           "bad",
		RuntimeType:    "k8s",
		PackageVersion: "v-test-bad",
		ControlPlaneID: 1,
	})
	if err == nil {
		t.Fatalf("expected checksum mismatch error")
	}
}
