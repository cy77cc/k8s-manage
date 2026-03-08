package planner

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type UserPlanner struct{}

func NewUserPlanner() *UserPlanner { return &UserPlanner{} }

func (p *UserPlanner) Domain() types.Domain { return types.DomainUser }

func (p *UserPlanner) Plan(_ context.Context, req orchestrator.DomainRequest) (types.DomainPlan, error) {
	return types.DomainPlan{
		Domain: p.Domain(),
		Steps: []types.PlanStep{
			{ID: "get_user", Tool: "user_list", Params: map[string]any{"keyword": req.UserIntent}, Produces: []string{"user_id"}},
			{ID: "check_permission", Tool: "permission_check", Params: map[string]any{"user_id": map[string]any{"$ref": "get_user.user_id"}}, DependsOn: []string{"get_user"}, Produces: []string{"permission_result"}},
		},
	}, nil
}
