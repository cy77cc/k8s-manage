package approval

import (
	"context"
	"testing"

	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestResourceOwnerRouterRouteService(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	svc := &model.Service{Name: "payment", OwnerUserID: 42}
	if err := db.Create(svc).Error; err != nil {
		t.Fatalf("seed service: %v", err)
	}

	router := NewResourceOwnerRouter(db)
	got, err := router.Route(context.Background(), &model.AIApprovalTask{
		RequestUserID:      7,
		TargetResourceType: "service",
		TargetResourceID:   "1",
	})
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	if len(got) != 1 || got[0] != 42 {
		t.Fatalf("Route() = %v", got)
	}
}

func TestResourceOwnerRouterFallback(t *testing.T) {
	t.Parallel()

	router := NewResourceOwnerRouter(nil)
	got, err := router.Route(context.Background(), &model.AIApprovalTask{
		RequestUserID:      9,
		TargetResourceType: "unknown",
	})
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	if len(got) != 1 || got[0] != 9 {
		t.Fatalf("Route() = %v", got)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Service{}, &model.Project{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return db
}
