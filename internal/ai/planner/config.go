package planner

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type ConfigPlanner struct{}

func NewConfigPlanner() *ConfigPlanner { return &ConfigPlanner{} }

func (p *ConfigPlanner) Domain() types.Domain { return types.DomainConfig }

func (p *ConfigPlanner) Plan(_ context.Context, req orchestrator.DomainRequest) (types.DomainPlan, error) {
	return types.DomainPlan{
		Domain: p.Domain(),
		Steps: []types.PlanStep{
			{ID: "get_app", Tool: "config_app_list", Params: map[string]any{"keyword": req.UserIntent}, Produces: []string{"app_id"}},
			{ID: "diff_config", Tool: "config_diff", Params: map[string]any{"app_id": map[string]any{"$ref": "get_app.app_id"}}, DependsOn: []string{"get_app"}, Produces: []string{"diff_result"}},
		},
	}, nil
}
