package ai

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
	"gorm.io/gorm"
)

type ControlPlane struct {
	db      *gorm.DB
	runtime *logic.RuntimeStore
	ai      *AIAgent
}

func NewControlPlane(db *gorm.DB, runtime *logic.RuntimeStore, ai *AIAgent) *ControlPlane {
	return &ControlPlane{db: db, runtime: runtime, ai: ai}
}

func (c *ControlPlane) ToolPolicy(ctx context.Context, meta tools.ToolMeta, params map[string]any) error {
	uid, _ := tools.ToolUserFromContext(ctx)
	if meta.Permission == "" || uid == 0 || c.db == nil {
		return nil
	}
	if !httpx.HasAnyPermission(c.db, uid, meta.Permission) {
		return ErrPermissionDenied
	}
	_ = params
	return nil
}

func (c *ControlPlane) HasPermission(uid uint64, code string) bool {
	if code == "" || c.db == nil {
		return true
	}
	return httpx.HasAnyPermission(c.db, uid, code)
}

func (c *ControlPlane) IsAdmin(uid uint64) bool {
	if c.db == nil {
		return false
	}
	return httpx.IsAdmin(c.db, uid)
}

func (c *ControlPlane) FindMeta(name string) (tools.ToolMeta, bool) {
	if c == nil || c.ai == nil {
		return tools.ToolMeta{}, false
	}
	return c.ai.FindMeta(name)
}
