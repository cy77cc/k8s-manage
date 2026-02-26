package ai

import (
	"context"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCommandTestHandler(t *testing.T) *handler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:ai_command_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.Service{},
		&model.CICDRelease{},
		&model.AlertEvent{},
		&model.CMDBRelation{},
		&model.CMDBCI{},
		&model.AICommandExecution{},
		&model.CICDAuditEvent{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return newHandler(&svc.ServiceContext{DB: db})
}

func TestBuildCommandContext_MissingParams(t *testing.T) {
	h := newCommandTestHandler(t)
	cc, err := h.buildCommandContext("deployment.release service_id=1", "scene:cicd", nil)
	if err != nil {
		t.Fatalf("build command context: %v", err)
	}
	if cc.Intent != "deployment.release" {
		t.Fatalf("expected deployment.release, got %s", cc.Intent)
	}
	if len(cc.Missing) == 0 {
		t.Fatalf("expected missing params")
	}
	if cc.Risk != commandRiskLow {
		t.Fatalf("expected low risk, got %s", cc.Risk)
	}
}

func TestExecuteAggregate(t *testing.T) {
	h := newCommandTestHandler(t)
	_ = h.svcCtx.DB.Create(&model.Service{ProjectID: 1, TeamID: 1, Name: "svc-a", Type: "stateless", Image: "nginx:latest", RuntimeType: "k8s", Status: "active"}).Error
	_ = h.svcCtx.DB.Create(&model.CICDRelease{ServiceID: 1, DeploymentID: 1, RuntimeType: "k8s", Version: "v1", Strategy: "rolling", Status: "succeeded"}).Error
	_ = h.svcCtx.DB.Create(&model.AlertEvent{Title: "cpu high", Status: "firing", Metric: "cpu"}).Error
	_ = h.svcCtx.DB.Create(&model.CMDBRelation{FromCIID: 1, ToCIID: 2, RelationType: "depends_on"}).Error

	cc, err := h.buildCommandContext("ops.aggregate.status limit=5 max_parallel=2 timeout_sec=3", "scene:dashboard", nil)
	if err != nil {
		t.Fatalf("build command context: %v", err)
	}
	out, err := executeAggregate(context.Background(), h, 1, cc, "")
	if err != nil {
		t.Fatalf("execute aggregate: %v", err)
	}
	details, ok := out["details"].(map[string]any)
	if !ok {
		t.Fatalf("expected details map")
	}
	if len(details) == 0 {
		t.Fatalf("expected aggregate details")
	}
}

func TestSaveAndLoadCommandRecord(t *testing.T) {
	h := newCommandTestHandler(t)
	cc, err := h.buildCommandContext("service.status service_id=1", "scene:service", nil)
	if err != nil {
		t.Fatalf("build command context: %v", err)
	}
	if err := h.store.saveCommandRecord(7, cc, "previewed", map[string]any{"ok": true}, nil, "preview ok"); err != nil {
		t.Fatalf("save command record: %v", err)
	}
	row, err := h.store.getCommandRecord(7, cc.CommandID)
	if err != nil {
		t.Fatalf("get command record: %v", err)
	}
	if row.Intent != cc.Intent {
		t.Fatalf("intent mismatch")
	}
}
