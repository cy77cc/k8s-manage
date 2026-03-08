package executor

import (
	"context"
	"fmt"

	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type Executor struct {
	tools map[string]ToolExecutor
}

func NewExecutor(tools map[string]ToolExecutor) *Executor {
	if tools == nil {
		tools = map[string]ToolExecutor{}
	}
	return &Executor{tools: tools}
}

func (e *Executor) Execute(ctx context.Context, plans []types.DomainPlan) (ExecutionResult, error) {
	dag, err := BuildDAG(plans)
	if err != nil {
		return ExecutionResult{}, err
	}
	results := make(map[string]types.StepResult, len(dag.Steps))
	order := make([]string, 0, len(dag.Order))
	for _, globalID := range dag.Order {
		step := dag.Steps[globalID]
		toolExec, ok := e.tools[step.Tool]
		if !ok {
			return ExecutionResult{}, fmt.Errorf("工具未注册: %s", step.Tool)
		}
		params, err := ResolveParams(dag.Domains[globalID], step.Params, results)
		if err != nil {
			return ExecutionResult{}, err
		}
		output, err := toolExec(ctx, params)
		if err != nil {
			results[globalID] = types.StepResult{StepID: globalID, Error: err.Error()}
			return ExecutionResult{Order: order, Results: results}, err
		}
		results[globalID] = types.StepResult{StepID: globalID, Output: output}
		order = append(order, globalID)
	}
	return ExecutionResult{Order: order, Results: results}, nil
}
