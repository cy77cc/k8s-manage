package planner

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type DomainPlanner interface {
	Domain() types.Domain
	Plan(ctx context.Context, req orchestrator.DomainRequest) (types.DomainPlan, error)
}
