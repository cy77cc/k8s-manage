package ai

import (
	"context"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
)

func seedPermissionTestData(t *testing.T, h *handler) {
	t.Helper()
	db := h.svcCtx.DB
	if err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.Service{},
		&model.Node{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	_ = db.Exec("DELETE FROM role_permissions").Error
	_ = db.Exec("DELETE FROM user_roles").Error
	_ = db.Exec("DELETE FROM permissions").Error
	_ = db.Exec("DELETE FROM roles").Error
	_ = db.Exec("DELETE FROM users").Error
	_ = db.Exec("DELETE FROM services").Error
	_ = db.Exec("DELETE FROM nodes").Error

	_ = db.Create(&model.User{ID: 1, Username: "admin001", PasswordHash: "x", Email: "admin001@example.com"}).Error
	_ = db.Create(&model.User{ID: 2, Username: "approver1", PasswordHash: "x", Email: "approver1@example.com"}).Error
	_ = db.Create(&model.User{ID: 3, Username: "owner001", PasswordHash: "x", Email: "owner001@example.com"}).Error

	_ = db.Create(&model.Role{ID: 1, Code: "admin", Name: "admin"}).Error
	_ = db.Create(&model.Role{ID: 2, Code: "approver", Name: "approver"}).Error
	_ = db.Create(&model.Role{ID: 3, Code: "deployer", Name: "deployer"}).Error

	_ = db.Create(&model.Permission{ID: 1, Code: "service:approve", Name: "service approve"}).Error
	_ = db.Create(&model.Permission{ID: 2, Code: "service:deploy", Name: "service deploy"}).Error

	_ = db.Create(&model.RolePermission{RoleID: 2, PermissionID: 1}).Error
	_ = db.Create(&model.RolePermission{RoleID: 3, PermissionID: 2}).Error
	_ = db.Create(&model.UserRole{UserID: 1, RoleID: 1}).Error
	_ = db.Create(&model.UserRole{UserID: 2, RoleID: 2}).Error
	_ = db.Create(&model.UserRole{UserID: 3, RoleID: 3}).Error

	_ = db.Create(&model.Service{ID: 10, Name: "svc", Type: "stateless", Image: "nginx:latest", ProjectID: 1, OwnerUserID: 3}).Error
}

func TestPermissionCheckerToolResourceMapping(t *testing.T) {
	h := newCommandTestHandler(t)
	checker := NewPermissionChecker(h.svcCtx.DB)
	rule, ok := checker.Rule("service_deploy_apply")
	if !ok {
		t.Fatalf("expected mapping for service_deploy_apply")
	}
	if rule.ResourceType != "service" {
		t.Fatalf("unexpected resource type: %s", rule.ResourceType)
	}
	if rule.WritePermission == "" || rule.ApprovePermission == "" {
		t.Fatalf("mapping permissions should not be empty")
	}
}

func TestPermissionCheckerCheckPermission(t *testing.T) {
	h := newCommandTestHandler(t)
	seedPermissionTestData(t, h)
	checker := NewPermissionChecker(h.svcCtx.DB)

	ok, err := checker.CheckPermission(context.Background(), 3, "service_deploy_apply")
	if err != nil {
		t.Fatalf("check permission: %v", err)
	}
	if !ok {
		t.Fatalf("expected deployer to have service deploy permission")
	}

	ok, err = checker.CheckPermission(context.Background(), 2, "service_deploy_apply")
	if err != nil {
		t.Fatalf("check permission: %v", err)
	}
	if ok {
		t.Fatalf("approver should not have service deploy permission")
	}
}

func TestPermissionCheckerFindApprovers(t *testing.T) {
	h := newCommandTestHandler(t)
	seedPermissionTestData(t, h)
	checker := NewPermissionChecker(h.svcCtx.DB)

	approvers, err := checker.FindApprovers(context.Background(), "service_deploy_apply", "10")
	if err != nil {
		t.Fatalf("find approvers: %v", err)
	}
	if len(approvers) == 0 {
		t.Fatalf("expected approvers list")
	}
	contains := func(uid uint64) bool {
		for _, item := range approvers {
			if item == uid {
				return true
			}
		}
		return false
	}
	if !contains(2) {
		t.Fatalf("expected permission approver user 2 in result: %v", approvers)
	}
	if !contains(3) {
		t.Fatalf("expected owner user 3 in result: %v", approvers)
	}
	if !contains(1) {
		t.Fatalf("expected admin user 1 in result: %v", approvers)
	}
}
