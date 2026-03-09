package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	aiapproval "github.com/cy77cc/k8s-manage/internal/ai/approval"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/model"
	aievents "github.com/cy77cc/k8s-manage/internal/service/ai/events"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GatewayRuntime 是 AI 模块的核心网关，提供统一的 API 入口。
// 负责会话管理、工具执行、审批流程和执行记录管理。
type GatewayRuntime struct {
	// db 是数据库连接。
	db *gorm.DB
	// ai 是 AI Agent 实例。
	ai *AIAgent
	// control 是控制平面，处理权限检查。
	control *ControlPlane
	// sessions 是会话存储。
	sessions *logic.SessionStore
	// runtime 是运行时存储，管理执行记录和审批票证。
	runtime *logic.RuntimeStore
	// router 是审批路由器，决定审批人。
	router aiapproval.ApprovalRouter
	// gen 是任务生成器，生成审批任务详情。
	gen *aiapproval.TaskGenerator
	// executor 是审批执行器，执行已批准的任务。
	executor *aiapproval.ApprovalExecutor
}

// NewGatewayRuntime 创建一个新的网关运行时实例。
//
// 参数:
//   - db: 数据库连接。
//   - ai: AI Agent 实例。
//   - control: 控制平面。
//   - sessions: 会话存储。
//   - runtime: 运行时存储。
//
// 返回:
//   - *GatewayRuntime: 网关实例。
func NewGatewayRuntime(db *gorm.DB, ai *AIAgent, control *ControlPlane, sessions *logic.SessionStore, runtime *logic.RuntimeStore) *GatewayRuntime {
	hub := aievents.DefaultApprovalHub()
	return &GatewayRuntime{
		db:       db,
		ai:       ai,
		control:  control,
		sessions: sessions,
		runtime:  runtime,
		router:   aiapproval.NewResourceOwnerRouter(db),
		gen:      aiapproval.NewTaskGenerator(ai.model),
		executor: aiapproval.NewApprovalExecutor(db, ai, hub),
	}
}

// ListSessions 列出用户在指定场景下的所有会话。
//
// 参数:
//   - uid: 用户 ID。
//   - scene: 场景名称。
//
// 返回:
//   - []*logic.AISession: 会话列表。
func (g *GatewayRuntime) ListSessions(uid uint64, scene string) []*logic.AISession {
	return g.sessions.List(uid, scene)
}

// CurrentSession 获取用户在指定场景下的当前（最新）会话。
//
// 参数:
//   - uid: 用户 ID。
//   - scene: 场景名称。
//
// 返回:
//   - *logic.AISession: 会话实例。
//   - bool: 是否存在。
func (g *GatewayRuntime) CurrentSession(uid uint64, scene string) (*logic.AISession, bool) {
	items := g.sessions.List(uid, scene)
	if len(items) == 0 {
		return nil, false
	}
	return items[0], true
}

// GetSession 根据会话 ID 获取会话。
//
// 参数:
//   - uid: 用户 ID。
//   - id: 会话 ID。
//
// 返回:
//   - *logic.AISession: 会话实例。
//   - bool: 是否存在。
func (g *GatewayRuntime) GetSession(uid uint64, id string) (*logic.AISession, bool) {
	return g.sessions.Get(uid, id)
}

// BranchSession 从现有会话创建分支会话。
// 允许用户从历史消息处创建新的对话分支。
//
// 参数:
//   - uid: 用户 ID。
//   - sourceSessionID: 源会话 ID。
//   - anchorMessageID: 锚点消息 ID（当前未使用）。
//   - title: 新会话标题。
//
// 返回:
//   - *logic.AISession: 新创建的分支会话。
//   - error: 创建错误。
func (g *GatewayRuntime) BranchSession(uid uint64, sourceSessionID, anchorMessageID, title string) (*logic.AISession, error) {
	src, ok := g.sessions.Get(uid, sourceSessionID)
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	now := time.Now()
	_ = anchorMessageID
	branched := &logic.AISession{
		ID:        "sess-" + uuid.NewString(),
		UserID:    uid,
		Scene:     src.Scene,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	g.sessions.Put(branched)
	return branched, nil
}

// DeleteSession 删除指定会话。
//
// 参数:
//   - uid: 用户 ID。
//   - id: 会话 ID。
func (g *GatewayRuntime) DeleteSession(uid uint64, id string) {
	g.sessions.Delete(uid, id)
}

// UpdateSessionTitle 更新会话标题。
//
// 参数:
//   - uid: 用户 ID。
//   - id: 会话 ID。
//   - title: 新标题。
//
// 返回:
//   - *logic.AISession: 更新后的会话。
//   - error: 更新错误。
func (g *GatewayRuntime) UpdateSessionTitle(uid uint64, id, title string) (*logic.AISession, error) {
	session, ok := g.sessions.Get(uid, id)
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	session.Title = title
	session.UpdatedAt = time.Now()
	g.sessions.Put(session)
	return session, nil
}

// PreviewTool 预览工具执行，返回工具元信息和审批需求。
// 不会实际执行工具，仅返回预览信息。
//
// 参数:
//   - uid: 用户 ID。
//   - tool: 工具名称。
//   - params: 工具参数。
//
// 返回:
//   - map[string]any: 预览信息，包含风险级别、权限要求、是否需要审批等。
//   - error: 权限检查错误。
func (g *GatewayRuntime) PreviewTool(uid uint64, tool string, params map[string]any) (map[string]any, error) {
	meta, ok := g.control.FindMeta(tool)
	if !ok {
		return nil, ErrToolNotFound
	}
	if !g.control.HasPermission(uid, meta.Permission) {
		return nil, ErrPermissionDenied
	}
	return map[string]any{
		"tool":              meta.Name,
		"risk":              meta.Risk,
		"permission":        meta.Permission,
		"approval_required": meta.Risk == "high" || meta.Risk == "medium",
		"params":            params,
	}, nil
}

// ExecuteTool 执行工具并记录执行结果。
// 会检查用户权限，执行后保存执行记录。
//
// 参数:
//   - ctx: 上下文。
//   - uid: 用户 ID。
//   - tool: 工具名称。
//   - params: 工具参数。
//   - approvalToken: 审批令牌（如果有）。
//
// 返回:
//   - *logic.ExecutionRecord: 执行记录。
//   - error: 执行错误。
func (g *GatewayRuntime) ExecuteTool(ctx context.Context, uid uint64, tool string, params map[string]any, approvalToken string) (*logic.ExecutionRecord, error) {
	meta, ok := g.control.FindMeta(tool)
	if !ok {
		return nil, ErrToolNotFound
	}
	if !g.control.HasPermission(uid, meta.Permission) {
		return nil, ErrPermissionDenied
	}
	ctx = tools.WithToolUser(ctx, uid, approvalToken)
	result, err := g.ai.RunTool(ctx, tool, params)
	if err != nil {
		return nil, err
	}
	rec := &logic.ExecutionRecord{
		ID:        "exe-" + uuid.NewString(),
		Tool:      tool,
		Status:    "succeeded",
		Result:    map[string]any{"ok": result.OK, "data": result.Data, "error": result.Error, "source": result.Source},
		CreatedAt: time.Now(),
	}
	g.runtime.SaveExecution(rec)
	return rec, nil
}

// GetExecution 根据执行 ID 获取执行记录。
//
// 参数:
//   - id: 执行记录 ID。
//
// 返回:
//   - *logic.ExecutionRecord: 执行记录。
//   - bool: 是否存在。
func (g *GatewayRuntime) GetExecution(id string) (*logic.ExecutionRecord, bool) {
	return g.runtime.GetExecution(id)
}

// CreateApproval 创建审批票证。
// 简化版本，内部调用 CreateApprovalTask。
//
// 参数:
//   - uid: 用户 ID。
//   - tool: 工具名称。
//   - params: 工具参数。
//
// 返回:
//   - *logic.ApprovalTicket: 审批票证。
//   - error: 创建错误。
func (g *GatewayRuntime) CreateApproval(uid uint64, tool string, params map[string]any) (*logic.ApprovalTicket, error) {
	task, err := g.CreateApprovalTask(context.Background(), uid, tool, params)
	if err != nil {
		return nil, err
	}
	return approvalTicketFromTask(task), nil
}

// CreateApprovalTask 创建完整的审批任务。
// 包含任务生成、审批路由、数据库持久化等完整流程。
//
// 参数:
//   - ctx: 上下文。
//   - uid: 用户 ID。
//   - tool: 工具名称。
//   - params: 工具参数。
//
// 返回:
//   - *model.AIApprovalTask: 创建的审批任务。
//   - error: 创建错误。
func (g *GatewayRuntime) CreateApprovalTask(ctx context.Context, uid uint64, tool string, params map[string]any) (*model.AIApprovalTask, error) {
	meta, ok := g.control.FindMeta(tool)
	if !ok {
		return nil, ErrToolNotFound
	}
	if !g.control.HasPermission(uid, meta.Permission) {
		return nil, ErrPermissionDenied
	}
	now := time.Now()
	task := &model.AIApprovalTask{
		ID:            "apv-" + uuid.NewString(),
		RequestUserID: uid,
		ApprovalToken: "tok-" + uuid.NewString(),
		ToolName:      tool,
		RiskLevel:     string(meta.Risk),
		Status:        "pending",
		ExpiresAt:     now.Add(24 * time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if raw, err := json.Marshal(params); err == nil {
		task.ParamsJSON = string(raw)
	}
	preview, err := g.PreviewTool(uid, tool, params)
	if err == nil {
		if raw, marshalErr := json.Marshal(preview); marshalErr == nil {
			task.PreviewJSON = string(raw)
		}
	}
	task.TargetResourceType, task.TargetResourceID, task.TargetResourceName = inferApprovalTarget(params)
	if g.gen != nil {
		if detail, err := g.gen.Generate(ctx, task); err == nil {
			_ = task.SetTaskDetail(detail)
		}
	}
	_ = task.SetToolCalls([]model.ApprovalToolCall{{Name: tool, Arguments: params}})
	if g.router != nil {
		if approvers, err := g.router.Route(ctx, task); err == nil && len(approvers) > 0 {
			task.ApproverUserID = approvers[0]
		}
	}
	if g.db != nil {
		if err := g.db.WithContext(ctx).Create(task).Error; err != nil {
			return nil, err
		}
	}
	g.runtime.SaveApproval(approvalTicketFromTask(task))
	publishApprovalUpdate(task, nil)
	return task, nil
}

// ConfirmApproval 确认审批（批准或拒绝）。
// 简化版本，内部调用 ReviewApprovalTask。
//
// 参数:
//   - uid: 审批人用户 ID。
//   - id: 审批任务 ID。
//   - approve: 是否批准。
//
// 返回:
//   - *logic.ApprovalTicket: 更新后的审批票证。
//   - error: 操作错误。
func (g *GatewayRuntime) ConfirmApproval(uid uint64, id string, approve bool) (*logic.ApprovalTicket, error) {
	task, _, err := g.ReviewApprovalTask(context.Background(), uid, id, approve, "")
	if err != nil {
		return nil, err
	}
	return approvalTicketFromTask(task), nil
}

// ListApprovalTasks 列出审批任务。
// 普通用户只能看到自己请求或审批的任务，有审批权限的用户可以看到所有任务。
//
// 参数:
//   - ctx: 上下文。
//   - uid: 用户 ID。
//   - status: 状态过滤（可选）。
//
// 返回:
//   - []model.AIApprovalTask: 审批任务列表。
//   - error: 查询错误。
func (g *GatewayRuntime) ListApprovalTasks(ctx context.Context, uid uint64, status string) ([]model.AIApprovalTask, error) {
	if g.db == nil {
		return nil, nil
	}
	q := g.db.WithContext(ctx).Model(&model.AIApprovalTask{}).Order("created_at DESC")
	if !g.control.HasPermission(uid, "ai:approval:review") {
		q = q.Where("request_user_id = ? OR approver_user_id = ?", uid, uid)
	}
	if strings.TrimSpace(status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(status))
	}
	var items []model.AIApprovalTask
	return items, q.Find(&items).Error
}

// GetApprovalTask 获取单个审批任务详情。
// 检查用户是否有权限查看该任务。
//
// 参数:
//   - ctx: 上下文。
//   - uid: 用户 ID。
//   - id: 审批任务 ID。
//
// 返回:
//   - *model.AIApprovalTask: 审批任务。
//   - error: 查询或权限错误。
func (g *GatewayRuntime) GetApprovalTask(ctx context.Context, uid uint64, id string) (*model.AIApprovalTask, error) {
	if g.db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var task model.AIApprovalTask
	if err := g.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(id)).First(&task).Error; err != nil {
		return nil, err
	}
	if !g.control.HasPermission(uid, "ai:approval:review") && task.RequestUserID != uid && task.ApproverUserID != uid {
		return nil, ErrPermissionDenied
	}
	return &task, nil
}

// ReviewApprovalTask 审核审批任务。
// 批准时会自动执行工具，拒绝时记录原因。
//
// 参数:
//   - ctx: 上下文。
//   - uid: 审批人用户 ID。
//   - id: 审批任务 ID。
//   - approve: 是否批准。
//   - reason: 拒绝原因（拒绝时使用）。
//
// 返回:
//   - *model.AIApprovalTask: 更新后的审批任务。
//   - *logic.ExecutionRecord: 执行记录（批准时）。
//   - error: 操作错误。
func (g *GatewayRuntime) ReviewApprovalTask(ctx context.Context, uid uint64, id string, approve bool, reason string) (*model.AIApprovalTask, *logic.ExecutionRecord, error) {
	task, err := g.GetApprovalTask(ctx, uid, id)
	if err != nil {
		return nil, nil, err
	}
	if !g.control.HasPermission(uid, "ai:approval:review") && task.ApproverUserID != uid {
		return nil, nil, ErrPermissionDenied
	}
	if !task.ExpiresAt.IsZero() && task.ExpiresAt.Before(time.Now()) {
		return nil, nil, ErrApprovalExpired
	}
	if task.Status != "pending" {
		return task, nil, nil
	}

	now := time.Now()
	task.ApproverUserID = uid
	task.UpdatedAt = now
	task.RejectReason = strings.TrimSpace(reason)
	if approve {
		task.Status = "approved"
		task.ApprovedAt = &now
		task.RejectedAt = nil
	} else {
		task.Status = "rejected"
		task.RejectedAt = &now
		task.ApprovedAt = nil
	}
	if g.db != nil {
		if err := g.db.WithContext(ctx).Save(task).Error; err != nil {
			return nil, nil, err
		}
	}
	g.runtime.UpdateApproval(task.ID, task.Status)
	publishApprovalUpdate(task, nil)

	if !approve || g.executor == nil {
		return task, nil, nil
	}
	outcome, execErr := g.executor.Execute(ctx, task)
	if outcome != nil && outcome.Record != nil {
		g.runtime.SaveExecution(outcome.Record)
	}
	if execErr != nil {
		return task, outcome.Record, execErr
	}
	return task, outcome.Record, nil
}

// ConfirmConfirmation 确认确认请求。
// 用于处理需要用户确认的操作。
//
// 参数:
//   - ctx: 上下文。
//   - uid: 用户 ID。
//   - id: 确认请求 ID。
//   - approve: 是否确认。
//
// 返回:
//   - *model.ConfirmationRequest: 更新后的确认请求。
//   - error: 操作错误。
func (g *GatewayRuntime) ConfirmConfirmation(ctx context.Context, uid uint64, id string, approve bool) (*model.ConfirmationRequest, error) {
	if g.db == nil {
		return nil, logic.ErrConfirmationNotFound
	}
	var req model.ConfirmationRequest
	if err := g.db.WithContext(ctx).Where("id = ?", id).First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, logic.ErrConfirmationNotFound
		}
		return nil, err
	}
	if !g.control.HasPermission(uid, "ai:approval:review") && req.RequestUserID != uid {
		return nil, ErrPermissionDenied
	}
	if !req.ExpiresAt.IsZero() && req.ExpiresAt.Before(time.Now()) {
		return nil, ErrApprovalExpired
	}
	if approve {
		req.Status = "confirmed"
	} else {
		req.Status = "cancelled"
	}
	if err := g.db.WithContext(ctx).Save(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

// approvalTicketFromTask 将数据库审批任务转换为业务层审批票证。
func approvalTicketFromTask(task *model.AIApprovalTask) *logic.ApprovalTicket {
	if task == nil {
		return nil
	}
	params, _ := decodeApprovalParams(task.ParamsJSON)
	return &logic.ApprovalTicket{
		ID:        task.ID,
		UserID:    task.RequestUserID,
		Tool:      task.ToolName,
		Status:    task.Status,
		Params:    params,
		CreatedAt: task.CreatedAt,
	}
}

// decodeApprovalParams 解码审批参数 JSON 字符串。
func decodeApprovalParams(raw string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

// inferApprovalTarget 从参数中推断审批目标资源。
// 根据参数中的 ID 字段推断资源类型和名称。
func inferApprovalTarget(params map[string]any) (string, string, string) {
	targetType := strings.TrimSpace(logic.ToString(params["resource_type"]))
	targetID := strings.TrimSpace(logic.ToString(params["resource_id"]))
	targetName := strings.TrimSpace(logic.ToString(params["resource_name"]))

	typeCandidates := []struct {
		key string
		typ string
	}{
		{"service_id", "service"},
		{"project_id", "project"},
		{"cluster_id", "cluster"},
		{"deployment_id", "deployment"},
	}
	for _, item := range typeCandidates {
		if targetID == "" {
			if v := strings.TrimSpace(logic.ToString(params[item.key])); v != "" {
				targetID = v
				if targetType == "" {
					targetType = item.typ
				}
			}
		}
	}

	if targetName == "" {
		for _, key := range []string{"service_name", "project_name", "cluster_name", "name"} {
			if v := strings.TrimSpace(logic.ToString(params[key])); v != "" {
				targetName = v
				break
			}
		}
	}
	return targetType, targetID, targetName
}

// publishApprovalUpdate 发布审批更新事件。
// 通过 SSE Hub 推送给相关用户。
func publishApprovalUpdate(task *model.AIApprovalTask, execution *logic.ExecutionRecord) {
	if task == nil {
		return
	}
	payload := aievents.ApprovalUpdate{
		ID:             task.ID,
		ApprovalToken:  task.ApprovalToken,
		ToolName:       task.ToolName,
		Status:         task.Status,
		RequestUserID:  task.RequestUserID,
		ApproverUserID: task.ApproverUserID,
		UpdatedAt:      task.UpdatedAt,
	}
	if execution != nil {
		payload.Execution = map[string]any{
			"id":         execution.ID,
			"tool":       execution.Tool,
			"status":     execution.Status,
			"result":     execution.Result,
			"created_at": execution.CreatedAt,
		}
	}
	aievents.DefaultApprovalHub().Publish(payload, task.RequestUserID, task.ApproverUserID)
}

func parseUint64Loose(v string) uint64 {
	n, _ := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
	return n
}

func formatAny(v any) string {
	if s := strings.TrimSpace(logic.ToString(v)); s != "" {
		return s
	}
	return fmt.Sprintf("%v", v)
}
