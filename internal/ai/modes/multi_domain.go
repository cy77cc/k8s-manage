package modes

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cy77cc/k8s-manage/internal/ai/graph"
	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	plannerpkg "github.com/cy77cc/k8s-manage/internal/ai/planner"
	aitools "github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type MultiDomainMode struct {
	planner  *orchestrator.OrchestratorPlanner
	registry *plannerpkg.Registry
}

func NewMultiDomainMode(chatModel model.ToolCallingChatModel, _ aitools.PlatformDeps) *MultiDomainMode {
	registry := plannerpkg.NewRegistry()
	registry.Register(plannerpkg.NewInfrastructurePlanner())
	registry.Register(plannerpkg.NewServicePlanner())
	registry.Register(plannerpkg.NewCICDPlanner())
	registry.Register(plannerpkg.NewMonitorPlanner())
	registry.Register(plannerpkg.NewConfigPlanner())
	registry.Register(plannerpkg.NewUserPlanner())
	return &MultiDomainMode{planner: orchestrator.NewPlanner(chatModel), registry: registry}
}

func (m *MultiDomainMode) Execute(ctx context.Context, _ string, message string, gen *adk.AsyncGenerator[*types.AgentResult]) {
	if gen == nil {
		return
	}
	if m == nil || m.planner == nil || m.registry == nil {
		gen.Send(&types.AgentResult{Type: "error", Content: "multi-domain mode not initialized"})
		return
	}
	requests, err := m.planner.Plan(ctx, message)
	if err != nil {
		gen.Send(&types.AgentResult{Type: "error", Content: err.Error()})
		return
	}
	plans, err := graph.PlanDomains(ctx, m.registry, requests)
	if err != nil {
		gen.Send(&types.AgentResult{Type: "error", Content: err.Error()})
		return
	}
	gen.Send(&types.AgentResult{Type: "text", Content: summarizePlans(plans)})
}

func summarizePlans(plans []types.DomainPlan) string {
	parts := make([]string, 0, len(plans))
	for _, plan := range plans {
		parts = append(parts, fmt.Sprintf("%s(%d steps)", plan.Domain, len(plan.Steps)))
	}
	if len(parts) == 0 {
		return "multi-domain planner produced no work"
	}
	return "multi-domain plan ready: " + strings.Join(parts, ", ")
}
