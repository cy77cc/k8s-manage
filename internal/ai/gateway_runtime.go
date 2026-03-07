package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
	"gorm.io/gorm"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrToolNotFound     = errors.New("tool not found")
	ErrApprovalExpired  = errors.New("approval expired")
)

type gatewayToolRunner interface {
	ToolMetas() []aitools.ToolMeta
	RunTool(ctx context.Context, toolName string, params map[string]any) (aitools.ToolResult, error)
}

type GatewayRuntime struct {
	db       *gorm.DB
	runner   gatewayToolRunner
	control  *ControlPlane
	sessions *logic.SessionStore
	runtime  *logic.RuntimeStore
}

func NewGatewayRuntime(db *gorm.DB, runner gatewayToolRunner, control *ControlPlane, sessions *logic.SessionStore, runtime *logic.RuntimeStore) *GatewayRuntime {
	return &GatewayRuntime{db: db, runner: runner, control: control, sessions: sessions, runtime: runtime}
}

func (g *GatewayRuntime) ListSessions(uid uint64, scene string) []*logic.AISession {
	if g == nil || g.sessions == nil {
		return nil
	}
	return g.sessions.ListSessions(uid, scene)
}

func (g *GatewayRuntime) CurrentSession(uid uint64, scene string) (*logic.AISession, bool) {
	if g == nil || g.sessions == nil {
		return nil, false
	}
	return g.sessions.CurrentSession(uid, scene)
}

func (g *GatewayRuntime) GetSession(uid uint64, id string) (*logic.AISession, bool) {
	if g == nil || g.sessions == nil {
		return nil, false
	}
	return g.sessions.GetSession(uid, id)
}

func (g *GatewayRuntime) BranchSession(uid uint64, sourceSessionID, anchorMessageID, title string) (*logic.AISession, error) {
	if g == nil || g.sessions == nil {
		return nil, errors.New("session store unavailable")
	}
	return g.sessions.BranchSession(uid, sourceSessionID, anchorMessageID, title)
}

func (g *GatewayRuntime) DeleteSession(uid uint64, id string) {
	if g == nil || g.sessions == nil {
		return
	}
	g.sessions.DeleteSession(uid, id)
}

func (g *GatewayRuntime) UpdateSessionTitle(uid uint64, id, title string) (*logic.AISession, error) {
	if g == nil || g.sessions == nil {
		return nil, errors.New("session store unavailable")
	}
	return g.sessions.UpdateSessionTitle(uid, id, title)
}

func (g *GatewayRuntime) PreviewTool(uid uint64, tool string, params map[string]any) (map[string]any, error) {
	meta, err := g.requireTool(uid, tool)
	if err != nil {
		return nil, err
	}
	data := map[string]any{
		"tool":              meta.Name,
		"mode":              meta.Mode,
		"risk":              meta.Risk,
		"params":            params,
		"approval_required": meta.Mode == aitools.ToolModeMutating,
	}
	if meta.Mode == aitools.ToolModeMutating {
		t := g.runtime.NewApproval(uid, logic.ApprovalTicket{
			Tool:   meta.Name,
			Params: params,
			Risk:   meta.Risk,
			Mode:   meta.Mode,
			Meta:   meta,
		})
		data["approval_token"] = t.ID
		data["expiresAt"] = t.ExpiresAt
		data["previewDiff"] = "Mutating operation requires approval."
	}
	return data, nil
}

func (g *GatewayRuntime) ExecuteTool(ctx context.Context, uid uint64, tool string, params map[string]any, approvalToken string) (*logic.ExecutionRecord, error) {
	meta, err := g.requireTool(uid, tool)
	if err != nil {
		return nil, err
	}
	if g.runner == nil {
		return nil, errors.New("ai agent not initialized")
	}
	rec := &logic.ExecutionRecord{
		ID:         fmt.Sprintf("exe-%d", time.Now().UnixNano()),
		Tool:       meta.Name,
		Params:     params,
		Mode:       meta.Mode,
		Status:     "running",
		RequestUID: uid,
		CreatedAt:  time.Now(),
	}
	start := time.Now()
	runCtx := aitools.WithToolUser(ctx, uid, strings.TrimSpace(approvalToken))
	if g.control != nil {
		runCtx = aitools.WithToolPolicyChecker(runCtx, g.control.ToolPolicy)
	}
	result, runErr := g.runner.RunTool(runCtx, meta.Name, params)
	finished := time.Now()
	rec.FinishedAt = &finished
	rec.Result = &result
	if runErr != nil {
		rec.Status = "failed"
		rec.Error = runErr.Error()
	} else if result.OK {
		rec.Status = "succeeded"
	} else {
		rec.Status = "failed"
		rec.Error = result.Error
	}
	if apErr, ok := aitools.IsApprovalRequired(runErr); ok {
		rec.Status = "failed"
		rec.Error = apErr.Error()
	}
	if result.LatencyMS == 0 {
		rec.Result.LatencyMS = time.Since(start).Milliseconds()
	}
	if g.runtime != nil {
		g.runtime.SaveExecution(rec)
	}
	return rec, nil
}

func (g *GatewayRuntime) GetExecution(id string) (*logic.ExecutionRecord, bool) {
	if g == nil || g.runtime == nil {
		return nil, false
	}
	return g.runtime.GetExecution(id)
}

func (g *GatewayRuntime) CreateApproval(uid uint64, tool string, params map[string]any) (*logic.ApprovalTicket, error) {
	meta, err := g.requireTool(uid, tool)
	if err != nil {
		return nil, err
	}
	if meta.Mode == aitools.ToolModeReadonly {
		return nil, errors.New("readonly tool does not require approval")
	}
	t := g.runtime.NewApproval(uid, logic.ApprovalTicket{
		Tool:   meta.Name,
		Params: params,
		Risk:   meta.Risk,
		Mode:   meta.Mode,
		Meta:   meta,
	})
	return t, nil
}

func (g *GatewayRuntime) ConfirmApproval(uid uint64, id string, approve bool) (*logic.ApprovalTicket, error) {
	if g == nil || g.runtime == nil || g.control == nil {
		return nil, errors.New("approval runtime unavailable")
	}
	if !g.control.HasPermission(uid, "ai:approval:review") {
		return nil, ErrPermissionDenied
	}
	t, ok := g.runtime.GetApproval(id)
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	if time.Now().After(t.ExpiresAt) {
		_, _ = g.runtime.SetApprovalStatus(id, "expired", uid)
		return nil, ErrApprovalExpired
	}
	status := "rejected"
	if approve {
		status = "approved"
	}
	out, _ := g.runtime.SetApprovalStatus(id, status, uid)
	return out, nil
}

func (g *GatewayRuntime) ConfirmConfirmation(ctx context.Context, uid uint64, id string, approve bool) (*model.ConfirmationRequest, error) {
	if g == nil {
		return nil, errors.New("gateway runtime unavailable")
	}
	svc := logic.NewConfirmationService(g.db)
	item, err := svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.RequestUserID != uid && (g.control == nil || !g.control.IsAdmin(uid)) {
		return nil, ErrPermissionDenied
	}
	if time.Now().After(item.ExpiresAt) {
		_, _ = svc.ExpirePending(ctx, time.Now())
		return nil, ErrApprovalExpired
	}
	if approve {
		return svc.Confirm(ctx, id)
	}
	return svc.Cancel(ctx, id)
}

func (g *GatewayRuntime) requireTool(uid uint64, name string) (aitools.ToolMeta, error) {
	if g == nil || g.control == nil {
		return aitools.ToolMeta{}, errors.New("control plane unavailable")
	}
	meta, ok := g.control.FindMeta(name)
	if !ok {
		return aitools.ToolMeta{}, ErrToolNotFound
	}
	if !g.control.HasPermission(uid, meta.Permission) {
		return aitools.ToolMeta{}, ErrPermissionDenied
	}
	return meta, nil
}
