package handler

import (
	"fmt"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRBACDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Role{}, &model.Permission{}, &model.UserRole{}, &model.RolePermission{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSyncUserRolesTxRejectsUnknownRoleCodes(t *testing.T) {
	db := setupRBACDB(t)
	h := NewHandler(&svc.ServiceContext{DB: db})

	if err := db.Create(&model.User{Username: "testuser", PasswordHash: "x", Email: "u@example.com", Status: 1}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&model.Role{Name: "Admin", Code: "admin", Status: 1}).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return h.syncUserRolesTx(tx, 1, []string{"admin", "missing-role"})
	})
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if _, ok := err.(*codeValidationError); !ok {
		t.Fatalf("expected codeValidationError, got %T", err)
	}
}

func TestSyncRolePermissionsTxRejectsUnknownPermissionCodes(t *testing.T) {
	db := setupRBACDB(t)
	h := NewHandler(&svc.ServiceContext{DB: db})

	if err := db.Create(&model.Role{Name: "Admin", Code: "admin", Status: 1}).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	if err := db.Create(&model.Permission{Name: "Read", Code: "rbac:read", Resource: "rbac", Action: "read", Status: 1}).Error; err != nil {
		t.Fatalf("seed permission: %v", err)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		return h.syncRolePermissionsTx(tx, 1, []string{"rbac:read", "rbac:missing"})
	})
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if _, ok := err.(*codeValidationError); !ok {
		t.Fatalf("expected codeValidationError, got %T", err)
	}
}
