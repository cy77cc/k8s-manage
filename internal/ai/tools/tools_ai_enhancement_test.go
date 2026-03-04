package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newToolsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:ai_tools_enhancement_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Service{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.Node{},
		&model.DeploymentTarget{},
		&model.ClusterCredential{},
		&model.CICDServiceCIConfig{},
		&model.Job{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func TestServiceCatalogList(t *testing.T) {
	db := newToolsTestDB(t)
	_ = db.Create(&model.Service{ProjectID: 1, TeamID: 1, Name: "svc-a", Type: "stateless", Image: "nginx", ServiceKind: "business", Visibility: "team"}).Error
	_ = db.Create(&model.Service{ProjectID: 1, TeamID: 1, Name: "redis", Type: "stateful", Image: "redis", ServiceKind: "middleware", Visibility: "public"}).Error

	res, err := serviceCatalogList(context.Background(), PlatformDeps{DB: db}, ServiceCatalogListInput{CategoryID: 1, Limit: 10})
	if err != nil {
		t.Fatalf("serviceCatalogList error: %v", err)
	}
	if !res.OK {
		t.Fatalf("expected ok result, got error: %s", res.Error)
	}
	payload, _ := res.Data.(map[string]any)
	if payload == nil {
		t.Fatalf("expected map payload")
	}
	if payload["total"] == nil {
		t.Fatalf("expected total field")
	}
}

func TestPermissionCheck(t *testing.T) {
	db := newToolsTestDB(t)
	_ = db.Exec("INSERT INTO users(id, username, password_hash, email, phone, status, create_time, update_time) VALUES(?,?,?,?,?,?,?,?)",
		100, "testuser", "x", "u@example.com", "13800138000", 1, 1, 1,
	).Error
	_ = db.Create(&model.Role{ID: 200, Name: "reader", Code: "reader", Status: 1, CreateTime: 1, UpdateTime: 1}).Error
	_ = db.Create(&model.Permission{ID: 300, Name: "service-read", Code: "service.read", Resource: "service", Action: "read", Status: 1, CreateTime: 1, UpdateTime: 1}).Error
	_ = db.Create(&model.UserRole{UserID: 100, RoleID: 200}).Error
	_ = db.Create(&model.RolePermission{RoleID: 200, PermissionID: 300}).Error

	res, err := permissionCheck(context.Background(), PlatformDeps{DB: db}, PermissionCheckInput{UserID: 100, Resource: "service", Action: "read"})
	if err != nil {
		t.Fatalf("permissionCheck error: %v", err)
	}
	if !res.OK {
		t.Fatalf("expected ok result, got error: %s", res.Error)
	}
	raw, _ := json.Marshal(res.Data)
	if !json.Valid(raw) {
		t.Fatalf("expected valid json data")
	}
}

func TestResolveToolParamHints(t *testing.T) {
	db := newToolsTestDB(t)
	_ = db.Create(&model.Service{ProjectID: 1, TeamID: 1, Name: "svc-hint", Type: "stateless", Image: "nginx"}).Error

	meta := ToolMeta{
		Name:       "service_get_detail",
		Required:   []string{"service_id"},
		Schema:     map[string]any{"properties": map[string]any{"service_id": map[string]any{"type": "integer"}}},
		EnumSources: map[string]string{"service_id": "service_list_inventory"},
	}
	resp := ResolveToolParamHints(context.Background(), PlatformDeps{DB: db}, meta)
	if resp.Tool != "service_get_detail" {
		t.Fatalf("unexpected tool name: %s", resp.Tool)
	}
	item, ok := resp.Params["service_id"]
	if !ok {
		t.Fatalf("expected service_id hint")
	}
	if item.EnumSource != "service_list_inventory" {
		t.Fatalf("unexpected enum source: %s", item.EnumSource)
	}
	if len(item.Values) == 0 {
		t.Fatalf("expected enum values")
	}
}
