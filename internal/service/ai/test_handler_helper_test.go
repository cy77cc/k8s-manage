package ai

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCommandTestHandler(t *testing.T) *handler {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_busy_timeout=5000", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
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
		&model.CICDAuditEvent{},
		&model.AIChatSession{},
		&model.AIChatMessage{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return newHandler(&svc.ServiceContext{DB: db})
}
