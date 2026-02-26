package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type Logic struct {
	svcCtx *svc.ServiceContext
}

func NewLogic(svcCtx *svc.ServiceContext) *Logic {
	return &Logic{svcCtx: svcCtx}
}

func (l *Logic) listInventories(ctx context.Context) ([]model.AutomationInventory, error) {
	rows := make([]model.AutomationInventory, 0, 32)
	err := l.svcCtx.DB.WithContext(ctx).Order("id desc").Find(&rows).Error
	return rows, err
}

func (l *Logic) createInventory(ctx context.Context, actor uint, req createInventoryReq) (*model.AutomationInventory, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	row := model.AutomationInventory{
		Name:      name,
		HostsJSON: strings.TrimSpace(req.HostsJSON),
		CreatedBy: actor,
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (l *Logic) listPlaybooks(ctx context.Context) ([]model.AutomationPlaybook, error) {
	rows := make([]model.AutomationPlaybook, 0, 32)
	err := l.svcCtx.DB.WithContext(ctx).Order("id desc").Find(&rows).Error
	return rows, err
}

func (l *Logic) createPlaybook(ctx context.Context, actor uint, req createPlaybookReq) (*model.AutomationPlaybook, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	risk := strings.ToLower(strings.TrimSpace(req.RiskLevel))
	if risk == "" {
		risk = "medium"
	}
	row := model.AutomationPlaybook{
		Name:       name,
		ContentYML: strings.TrimSpace(req.ContentYML),
		RiskLevel:  risk,
		CreatedBy:  actor,
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (l *Logic) previewRun(ctx context.Context, req previewRunReq) (map[string]any, error) {
	action := strings.TrimSpace(req.Action)
	if action == "" {
		return nil, fmt.Errorf("action is required")
	}
	return map[string]any{
		"preview_token": fmt.Sprintf("preview-%d", time.Now().UnixNano()),
		"action":        action,
		"risk_level":    "medium",
		"params":        req.Params,
		"status":        "ready",
	}, nil
}

func (l *Logic) executeRun(ctx context.Context, actor uint, req executeRunReq) (*model.AutomationRun, error) {
	if strings.TrimSpace(req.ApprovalToken) == "" {
		return nil, fmt.Errorf("approval_token is required")
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		action = "generic"
	}
	buf, _ := json.Marshal(req.Params)
	run := model.AutomationRun{
		ID:         fmt.Sprintf("run-%d", time.Now().UnixNano()),
		Action:     action,
		Status:     "running",
		ParamsJSON: string(buf),
		OperatorID: actor,
		StartedAt:  time.Now(),
	}
	if err := l.svcCtx.DB.WithContext(ctx).Create(&run).Error; err != nil {
		return nil, err
	}
	_ = l.svcCtx.DB.WithContext(ctx).Create(&model.AutomationRunLog{
		RunID:   run.ID,
		Level:   "info",
		Message: "run queued and started",
	}).Error

	run.Status = "succeeded"
	run.ResultJSON = `{"summary":"skeleton execution completed"}`
	run.FinishedAt = time.Now()
	_ = l.svcCtx.DB.WithContext(ctx).Model(&model.AutomationRun{}).
		Where("id = ?", run.ID).
		Updates(map[string]any{
			"status":      run.Status,
			"result_json": run.ResultJSON,
			"finished_at": run.FinishedAt,
		}).Error
	_ = l.svcCtx.DB.WithContext(ctx).Create(&model.AutomationRunLog{
		RunID:   run.ID,
		Level:   "info",
		Message: "run finished",
	}).Error

	detail, _ := json.Marshal(map[string]any{
		"approval_token": strings.TrimSpace(req.ApprovalToken),
		"params":         req.Params,
	})
	_ = l.svcCtx.DB.WithContext(ctx).Create(&model.AutomationExecutionAudit{
		RunID:      run.ID,
		Action:     run.Action,
		Status:     run.Status,
		ActorID:    actor,
		DetailJSON: string(detail),
	}).Error
	return &run, nil
}

func (l *Logic) getRun(ctx context.Context, id string) (*model.AutomationRun, error) {
	var row model.AutomationRun
	err := l.svcCtx.DB.WithContext(ctx).Where("id = ?", strings.TrimSpace(id)).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (l *Logic) listRunLogs(ctx context.Context, id string) ([]model.AutomationRunLog, error) {
	rows := make([]model.AutomationRunLog, 0, 32)
	err := l.svcCtx.DB.WithContext(ctx).Where("run_id = ?", strings.TrimSpace(id)).Order("id asc").Find(&rows).Error
	return rows, err
}
