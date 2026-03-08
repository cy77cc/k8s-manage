package approval

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
	"gorm.io/gorm"
)

type ToolRunner interface {
	RunTool(ctx context.Context, toolName string, params map[string]any) (tools.ToolResult, error)
}

type ApprovalUpdatePublisher interface {
	Publish(update any, userIDs ...uint64)
}

type ApprovalExecutor struct {
	db        *gorm.DB
	runner    ToolRunner
	publisher ApprovalUpdatePublisher
	now       func() time.Time
}

type ExecutionOutcome struct {
	Record *logic.ExecutionRecord
	Task   *model.AIApprovalTask
}

func NewApprovalExecutor(db *gorm.DB, runner ToolRunner, publisher ApprovalUpdatePublisher) *ApprovalExecutor {
	return &ApprovalExecutor{
		db:        db,
		runner:    runner,
		publisher: publisher,
		now:       time.Now,
	}
}

func (e *ApprovalExecutor) Execute(ctx context.Context, task *model.AIApprovalTask) (*ExecutionOutcome, error) {
	if task == nil {
		return nil, fmt.Errorf("approval task is nil")
	}
	if e.runner == nil {
		return nil, fmt.Errorf("tool runner is nil")
	}

	params, err := decodeParams(task.ParamsJSON)
	if err != nil {
		return nil, fmt.Errorf("decode params: %w", err)
	}

	runCtx := tools.WithToolUser(ctx, task.RequestUserID, task.ApprovalToken)
	result, runErr := e.runner.RunTool(runCtx, task.ToolName, params)

	now := e.now()
	record := &logic.ExecutionRecord{
		ID:        "exe-" + task.ID,
		Tool:      task.ToolName,
		Status:    "succeeded",
		Result:    map[string]any{"ok": result.OK, "data": result.Data, "error": result.Error, "source": result.Source},
		CreatedAt: now,
	}
	if runErr != nil {
		record.Status = "failed"
		record.Result = map[string]any{"ok": false, "error": runErr.Error(), "source": "approval_executor"}
	}

	task.ExecutedAt = &now
	if runErr != nil {
		task.Status = "failed"
		task.RejectReason = runErr.Error()
	} else {
		task.Status = "executed"
	}

	if e.db != nil {
		if err := e.db.WithContext(ctx).Save(task).Error; err != nil {
			return nil, err
		}
	}

	e.publish(task, record)

	if runErr != nil {
		return &ExecutionOutcome{Record: record, Task: task}, runErr
	}
	return &ExecutionOutcome{Record: record, Task: task}, nil
}

func (e *ApprovalExecutor) publish(task *model.AIApprovalTask, record *logic.ExecutionRecord) {
	if e.publisher == nil || task == nil {
		return
	}
	payload := map[string]any{
		"id":               task.ID,
		"approval_token":   task.ApprovalToken,
		"tool_name":        task.ToolName,
		"status":           task.Status,
		"request_user_id":  task.RequestUserID,
		"approver_user_id": task.ApproverUserID,
		"execution":        record,
		"updated_at":       task.UpdatedAt,
	}
	e.publisher.Publish(payload, task.RequestUserID, task.ApproverUserID)
}

func decodeParams(raw string) (map[string]any, error) {
	if raw == "" {
		return map[string]any{}, nil
	}
	var params map[string]any
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return nil, err
	}
	if params == nil {
		params = map[string]any{}
	}
	return params, nil
}
