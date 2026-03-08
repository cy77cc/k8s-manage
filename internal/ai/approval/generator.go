package approval

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/model"
)

type TaskGenerator struct {
	model einomodel.ToolCallingChatModel
}

func NewTaskGenerator(model einomodel.ToolCallingChatModel) *TaskGenerator {
	return &TaskGenerator{model: model}
}

func (g *TaskGenerator) Generate(ctx context.Context, task *model.AIApprovalTask) (model.TaskDetail, error) {
	if task == nil {
		return model.TaskDetail{}, fmt.Errorf("approval task is nil")
	}
	if g.model != nil {
		if detail, ok, err := g.generateWithLLM(ctx, task); err == nil && ok {
			return detail, nil
		}
	}
	return fallbackTaskDetail(task), nil
}

func (g *TaskGenerator) generateWithLLM(ctx context.Context, task *model.AIApprovalTask) (model.TaskDetail, bool, error) {
	prompt := buildPrompt(task)
	msg, err := g.model.Generate(ctx, []*schema.Message{
		schema.SystemMessage("Return compact JSON for an approval task detail."),
		schema.UserMessage(prompt),
	})
	if err != nil || msg == nil {
		return model.TaskDetail{}, false, err
	}
	var detail model.TaskDetail
	if err := json.Unmarshal([]byte(strings.TrimSpace(msg.Content)), &detail); err != nil {
		return model.TaskDetail{}, false, err
	}
	if strings.TrimSpace(detail.Summary) == "" {
		return model.TaskDetail{}, false, nil
	}
	return detail, true, nil
}

func buildPrompt(task *model.AIApprovalTask) string {
	return fmt.Sprintf(`tool=%s
resource_type=%s
resource_id=%s
resource_name=%s
risk=%s
params=%s

Return JSON:
{"summary":"","steps":[{"title":"","description":""}],"risk_assessment":{"level":"","summary":"","items":[]},"rollback_plan":""}`,
		task.ToolName,
		task.TargetResourceType,
		task.TargetResourceID,
		task.TargetResourceName,
		task.RiskLevel,
		task.ParamsJSON,
	)
}

func fallbackTaskDetail(task *model.AIApprovalTask) model.TaskDetail {
	resourceName := strings.TrimSpace(task.TargetResourceName)
	if resourceName == "" {
		resourceName = task.TargetResourceID
	}
	summary := strings.TrimSpace(task.ToolName)
	if resourceName != "" {
		summary += " on " + resourceName
	}
	return model.TaskDetail{
		Summary: summary,
		Steps: []model.ExecutionStep{
			{Title: "Review request", Description: "Check the target resource and requested parameters."},
			{Title: "Assess risk", Description: "Confirm the operation matches the declared risk and change window."},
			{Title: "Approve or reject", Description: "Approve when the request is safe, otherwise reject with a reason."},
		},
		RiskAssessment: model.RiskAssessment{
			Level:   strings.TrimSpace(task.RiskLevel),
			Summary: "Review impact before execution.",
			Items:   []string{"Validate target scope", "Confirm rollback path", "Check requester intent"},
		},
		RollbackPlan: "Revert the change through the corresponding tool or restore the previous deployment state.",
	}
}
