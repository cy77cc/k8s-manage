package planner

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type CICDPlanner struct{}

func NewCICDPlanner() *CICDPlanner { return &CICDPlanner{} }

func (p *CICDPlanner) Domain() types.Domain { return types.DomainCICD }

func (p *CICDPlanner) Plan(_ context.Context, req orchestrator.DomainRequest) (types.DomainPlan, error) {
	return types.DomainPlan{
		Domain: p.Domain(),
		Steps: []types.PlanStep{
			{ID: "get_pipeline", Tool: "cicd_pipeline_list", Params: map[string]any{"keyword": req.UserIntent}, Produces: []string{"pipeline_id"}},
			{ID: "trigger_pipeline", Tool: "cicd_pipeline_trigger", Params: map[string]any{"pipeline_id": map[string]any{"$ref": "get_pipeline.pipeline_id"}}, DependsOn: []string{"get_pipeline"}, Produces: []string{"execution_id"}},
		},
	}, nil
}
