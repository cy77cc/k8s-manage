package planner

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type ServicePlanner struct{}

func NewServicePlanner() *ServicePlanner { return &ServicePlanner{} }

func (p *ServicePlanner) Domain() types.Domain { return types.DomainService }

func (p *ServicePlanner) Plan(_ context.Context, req orchestrator.DomainRequest) (types.DomainPlan, error) {
	return types.DomainPlan{
		Domain: p.Domain(),
		Steps: []types.PlanStep{
			{ID: "get_service", Tool: "service_list_inventory", Params: map[string]any{"keyword": req.UserIntent}, Produces: []string{"service_id"}},
			{ID: "get_target", Tool: "deployment_target_list", Params: map[string]any{"keyword": req.UserIntent}, Produces: []string{"target_id"}},
			{ID: "deploy_service", Tool: "service_deploy_apply", Params: map[string]any{"service_id": map[string]any{"$ref": "get_service.service_id"}, "target_id": map[string]any{"$ref": "get_target.target_id"}}, DependsOn: []string{"get_service", "get_target"}, Produces: []string{"deployment_id"}},
		},
	}, nil
}
