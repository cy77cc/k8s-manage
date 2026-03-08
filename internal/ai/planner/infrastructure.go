package planner

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type InfrastructurePlanner struct{}

func NewInfrastructurePlanner() *InfrastructurePlanner { return &InfrastructurePlanner{} }

func (p *InfrastructurePlanner) Domain() types.Domain { return types.DomainInfrastructure }

func (p *InfrastructurePlanner) Plan(_ context.Context, req orchestrator.DomainRequest) (types.DomainPlan, error) {
	return types.DomainPlan{
		Domain: p.Domain(),
		Steps: []types.PlanStep{
			{
				ID:       "get_hosts",
				Tool:     "host_list_inventory",
				Params:   map[string]any{"keyword": req.UserIntent},
				Produces: []string{"host_id"},
			},
			{
				ID:        "inspect_host",
				Tool:      "host_ssh_exec_readonly",
				Params:    map[string]any{"host_id": map[string]any{"$ref": "get_hosts.host_id"}, "command": "status"},
				DependsOn: []string{"get_hosts"},
				Produces:  []string{"inspection_result"},
			},
		},
	}, nil
}
