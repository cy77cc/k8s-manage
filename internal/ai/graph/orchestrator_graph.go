package graph

import (
	"context"
	"fmt"

	"github.com/cy77cc/k8s-manage/internal/ai/executor"
	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	plannerpkg "github.com/cy77cc/k8s-manage/internal/ai/planner"
	"github.com/cy77cc/k8s-manage/internal/ai/replanner"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type domainDispatcher interface {
	Plan(ctx context.Context, message string) ([]orchestrator.DomainRequest, error)
}

type GraphResult struct {
	Requests  []orchestrator.DomainRequest
	Plans     []types.DomainPlan
	Execution executor.ExecutionResult
	Replan    replanner.ReplanDecision
}

type OrchestratorGraph struct {
	dispatcher domainDispatcher
	planners   *plannerpkg.Registry
	executor   *executor.Executor
	replanner  *replanner.Replanner
}

func NewOrchestratorGraph(dispatcher domainDispatcher, planners *plannerpkg.Registry, exec *executor.Executor, repl *replanner.Replanner) *OrchestratorGraph {
	return &OrchestratorGraph{dispatcher: dispatcher, planners: planners, executor: exec, replanner: repl}
}

func (g *OrchestratorGraph) Compile() error {
	if g == nil || g.dispatcher == nil {
		return fmt.Errorf("dispatcher is required")
	}
	if g.planners == nil {
		return fmt.Errorf("planner registry is required")
	}
	if g.executor == nil {
		return fmt.Errorf("executor is required")
	}
	if g.replanner == nil {
		return fmt.Errorf("replanner is required")
	}
	return nil
}

func (g *OrchestratorGraph) Execute(ctx context.Context, message string) (GraphResult, error) {
	if err := g.Compile(); err != nil {
		return GraphResult{}, err
	}
	requests, err := g.dispatcher.Plan(ctx, message)
	if err != nil {
		return GraphResult{}, err
	}
	plans, err := PlanDomains(ctx, g.planners, requests)
	if err != nil {
		return GraphResult{}, err
	}
	execution, err := g.executor.Execute(ctx, plans)
	if err != nil {
		return GraphResult{}, err
	}
	decision := g.replanner.Decide(execution)
	return GraphResult{Requests: requests, Plans: plans, Execution: execution, Replan: decision}, nil
}
