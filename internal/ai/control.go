package ai

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
	"gorm.io/gorm"
)

// ControlPlane 是 AI 模块的控制平面。
// 负责工具执行策略检查、权限验证和工具元信息查询。
type ControlPlane struct {
	// db 是数据库连接，用于权限查询。
	db *gorm.DB
	// runtime 是运行时存储。
	runtime *logic.RuntimeStore
	// ai 是 AI Agent 实例，用于查找工具元信息。
	ai *AIAgent
}

// NewControlPlane 创建一个新的控制平面实例。
//
// 参数:
//   - db: 数据库连接。
//   - runtime: 运行时存储。
//   - ai: AI Agent 实例。
//
// 返回:
//   - *ControlPlane: 控制平面实例。
func NewControlPlane(db *gorm.DB, runtime *logic.RuntimeStore, ai *AIAgent) *ControlPlane {
	return &ControlPlane{db: db, runtime: runtime, ai: ai}
}

// ToolPolicy 检查工具执行策略。
// 验证用户是否有权限执行指定工具。
//
// 参数:
//   - ctx: 上下文。
//   - meta: 工具元信息。
//   - params: 工具参数。
//
// 返回:
//   - error: 权限错误（如 ErrPermissionDenied）或 nil。
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

// HasPermission 检查用户是否拥有指定权限码。
//
// 参数:
//   - uid: 用户 ID。
//   - code: 权限码。
//
// 返回:
//   - bool: 是否有权限。
func (c *ControlPlane) HasPermission(uid uint64, code string) bool {
	if code == "" || c.db == nil {
		return true
	}
	return httpx.HasAnyPermission(c.db, uid, code)
}

// IsAdmin 检查用户是否为管理员。
//
// 参数:
//   - uid: 用户 ID。
//
// 返回:
//   - bool: 是否为管理员。
func (c *ControlPlane) IsAdmin(uid uint64) bool {
	if c.db == nil {
		return false
	}
	return httpx.IsAdmin(c.db, uid)
}

// FindMeta 根据名称查找工具元信息。
//
// 参数:
//   - name: 工具名称。
//
// 返回:
//   - tools.ToolMeta: 工具元信息。
//   - bool: 是否找到。
func (c *ControlPlane) FindMeta(name string) (tools.ToolMeta, bool) {
	if c == nil || c.ai == nil {
		return tools.ToolMeta{}, false
	}
	return c.ai.FindMeta(name)
}
