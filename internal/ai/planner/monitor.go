package planner

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type MonitorPlanner struct{}

func NewMonitorPlanner() *MonitorPlanner { return &MonitorPlanner{} }

func (p *MonitorPlanner) Domain() types.Domain { return types.DomainMonitor }

func (p *MonitorPlanner) Plan(_ context.Context, req orchestrator.DomainRequest) (types.DomainPlan, error) {
	return types.DomainPlan{
		Domain: p.Domain(),
		Steps: []types.PlanStep{
			{ID: "get_alerts", Tool: "monitor_alert_active", Params: map[string]any{"keyword": req.UserIntent}, Produces: []string{"alert_id"}},
			{ID: "get_topology", Tool: "topology_get", Params: map[string]any{"keyword": req.UserIntent}, Produces: []string{"topology_id"}},
		},
	}, nil
}
