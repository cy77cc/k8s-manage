package executor

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type ToolExecutor func(ctx context.Context, params map[string]any) (map[string]any, error)

type ExecutionContext struct {
	Tools   map[string]ToolExecutor
	Results map[string]types.StepResult
}

type ExecutionResult struct {
	Order   []string
	Results map[string]types.StepResult
}

type DAG struct {
	Steps        map[string]types.PlanStep
	Domains      map[string]types.Domain
	Dependencies map[string][]string
	Order        []string
}
