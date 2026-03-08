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

type GatewayRuntime struct {
	db       *gorm.DB
	ai       *AIAgent
	control  *ControlPlane
	sessions *logic.SessionStore
	runtime  *logic.RuntimeStore
	router   aiapproval.ApprovalRouter
	gen      *aiapproval.TaskGenerator
	executor *aiapproval.ApprovalExecutor
}

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

func (g *GatewayRuntime) ListSessions(uid uint64, scene string) []*logic.AISession {
	return g.sessions.List(uid, scene)
}

func (g *GatewayRuntime) CurrentSession(uid uint64, scene string) (*logic.AISession, bool) {
	items := g.sessions.List(uid, scene)
	if len(items) == 0 {
		return nil, false
	}
	return items[0], true
}

func (g *GatewayRuntime) GetSession(uid uint64, id string) (*logic.AISession, bool) {
	return g.sessions.Get(uid, id)
}

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

func (g *GatewayRuntime) DeleteSession(uid uint64, id string) {
	g.sessions.Delete(uid, id)
}

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

func (g *GatewayRuntime) GetExecution(id string) (*logic.ExecutionRecord, bool) {
	return g.runtime.GetExecution(id)
}

func (g *GatewayRuntime) CreateApproval(uid uint64, tool string, params map[string]any) (*logic.ApprovalTicket, error) {
	task, err := g.CreateApprovalTask(context.Background(), uid, tool, params)
	if err != nil {
		return nil, err
	}
	return approvalTicketFromTask(task), nil
}

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

func (g *GatewayRuntime) ConfirmApproval(uid uint64, id string, approve bool) (*logic.ApprovalTicket, error) {
	task, _, err := g.ReviewApprovalTask(context.Background(), uid, id, approve, "")
	if err != nil {
		return nil, err
	}
	return approvalTicketFromTask(task), nil
}

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
