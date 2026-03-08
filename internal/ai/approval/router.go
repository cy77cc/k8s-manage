package approval

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/gorm"
)

type ApprovalRouter interface {
	Route(ctx context.Context, task *model.AIApprovalTask) ([]uint64, error)
}

type ResourceOwnerRouter struct {
	db *gorm.DB
}

func NewResourceOwnerRouter(db *gorm.DB) *ResourceOwnerRouter {
	return &ResourceOwnerRouter{db: db}
}

func (r *ResourceOwnerRouter) Route(ctx context.Context, task *model.AIApprovalTask) ([]uint64, error) {
	if task == nil {
		return nil, fmt.Errorf("approval task is nil")
	}
	if r.db == nil {
		return []uint64{task.RequestUserID}, nil
	}
	switch strings.ToLower(strings.TrimSpace(task.TargetResourceType)) {
	case "service":
		return r.routeService(ctx, task)
	case "project":
		return r.routeProject(ctx, task)
	default:
		return []uint64{task.RequestUserID}, nil
	}
}

func (r *ResourceOwnerRouter) routeService(ctx context.Context, task *model.AIApprovalTask) ([]uint64, error) {
	var service model.Service
	if err := r.db.WithContext(ctx).Select("id", "owner_user_id").Where("id = ?", parseUint(task.TargetResourceID)).First(&service).Error; err != nil {
		return nil, err
	}
	if service.OwnerUserID == 0 {
		return []uint64{task.RequestUserID}, nil
	}
	return []uint64{uint64(service.OwnerUserID)}, nil
}

func (r *ResourceOwnerRouter) routeProject(ctx context.Context, task *model.AIApprovalTask) ([]uint64, error) {
	var project model.Project
	if err := r.db.WithContext(ctx).Select("id", "owner_id").Where("id = ?", parseUint(task.TargetResourceID)).First(&project).Error; err != nil {
		return nil, err
	}
	if project.OwnerID == 0 {
		return []uint64{task.RequestUserID}, nil
	}
	return []uint64{uint64(project.OwnerID)}, nil
}

func parseUint(v string) uint64 {
	n, _ := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
	return n
}
