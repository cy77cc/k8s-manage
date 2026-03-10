package service

import (
	"context"
	"strings"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/model"
	"github.com/cy77cc/OpsPilot/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateRejectsMissingProjectScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:servicescope?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Service{}, &model.ServiceRevision{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	logic := NewLogic(&svc.ServiceContext{DB: db})

	_, err = logic.Create(context.Background(), 1, ServiceCreateReq{
		Name:        "svc-missing-project",
		TeamID:      2,
		Env:         "staging",
		RuntimeType: "k8s",
		ConfigMode:  "standard",
	})
	if err == nil {
		t.Fatalf("expected create to fail when project scope is missing")
	}
	if !strings.Contains(err.Error(), "project_id is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
