package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
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
		&model.Node{},
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

func TestEnforceHostOperationPolicy_BlockedByMaintenanceAndDenylist(t *testing.T) {
	h := newCommandTestHandler(t)
	if err := h.svcCtx.DB.Create(&model.Node{
		ID:     11,
		Name:   "node-maintenance",
		IP:     "10.0.0.11",
		Port:   22,
		Status: "maintenance",
	}).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	cc := commandContext{
		Intent: "host.batch.exec.apply",
		Params: map[string]any{
			"host_ids": []any{11},
			"command":  "echo ok",
		},
	}
	if err := h.enforceHostOperationPolicy(context.Background(), cc); err == nil {
		t.Fatalf("expected maintenance host to be blocked")
	}

	if err := h.svcCtx.DB.Model(&model.Node{}).Where("id = ?", 11).Update("status", "online").Error; err != nil {
		t.Fatalf("update node status: %v", err)
	}
	cc.Params["command"] = "rm -rf /tmp/demo"
	if err := h.enforceHostOperationPolicy(context.Background(), cc); err == nil {
		t.Fatalf("expected denylist command to be blocked")
	}
}

func TestExecuteCommand_HostMutationRequiresApprovalToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newCommandTestHandler(t)
	if err := h.svcCtx.DB.Exec("INSERT INTO users(id, username, password_hash, email, phone, status, create_time, update_time) VALUES(?,?,?,?,?,?,?,?)",
		1, "admin", "x", "admin@example.com", "13800138000", 1, 1, 1,
	).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if err := h.svcCtx.DB.Create(&model.Node{
		ID:     21,
		Name:   "node-online",
		IP:     "10.0.0.21",
		Port:   22,
		Status: "online",
	}).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}

	body := map[string]any{
		"command": "host.batch.exec.apply",
		"scene":   "scene:hosts",
		"confirm": true,
		"params": map[string]any{
			"host_ids": []any{21},
			"command":  "echo health",
		},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/commands/execute", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", uint64(1))

	h.executeCommand(c)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", w.Code)
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != 2004 {
		t.Fatalf("expected forbidden code 2004, got %d (%s)", resp.Code, resp.Msg)
	}
	if resp.Msg != "approval token required" {
		t.Fatalf("expected approval-token-required message, got %q", resp.Msg)
	}
}

func TestRedactSensitiveText(t *testing.T) {
	out := redactSensitiveText("token=abc123\npassword=demo\nauthorization: bearer x")
	if out == "" {
		t.Fatalf("expected non-empty output")
	}
	if out == "token=abc123\npassword=demo\nauthorization: bearer x" {
		t.Fatalf("expected redaction to mutate sensitive content")
	}
	if containsAny(out, "abc123", "demo", "bearer x") {
		t.Fatalf("expected sensitive raw tokens to be hidden, got: %s", out)
	}
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if n != "" && bytes.Contains([]byte(s), []byte(n)) {
			return true
		}
	}
	return false
}
