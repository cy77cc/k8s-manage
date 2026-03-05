package graph

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cy77cc/k8s-manage/internal/ai/experts"
)

func helperStrategyBranch() *compose.GraphBranch {
	return compose.NewGraphBranch(func(_ context.Context, in *GraphInput) (string, error) {
		if in == nil || len(in.HelperRequests) == 0 {
			return nodeAggregate, nil
		}
		switch in.Strategy {
		case experts.StrategySequential:
			return nodeHelpersSeq, nil
		case experts.StrategyParallel:
			return nodeHelpersParallel, nil
		default:
			return nodeAggregate, nil
		}
	}, map[string]bool{
		nodeHelpersSeq:      true,
		nodeHelpersParallel: true,
		nodeAggregate:       true,
	})
}
