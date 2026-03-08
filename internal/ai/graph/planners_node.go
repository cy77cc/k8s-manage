package graph

import (
	"context"
	"fmt"

	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	plannerpkg "github.com/cy77cc/k8s-manage/internal/ai/planner"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
	"golang.org/x/sync/errgroup"
)

func PlanDomains(ctx context.Context, registry *plannerpkg.Registry, requests []orchestrator.DomainRequest) ([]types.DomainPlan, error) {
	if len(requests) == 0 {
		return nil, nil
	}
	if len(requests) == 1 {
		planner, ok := registry.Get(requests[0].Domain)
		if !ok {
			return nil, fmt.Errorf("planner not registered for domain %s", requests[0].Domain)
		}
		plan, err := planner.Plan(ctx, requests[0])
		if err != nil {
			return nil, err
		}
		return []types.DomainPlan{plan}, nil
	}
	plans := make([]types.DomainPlan, len(requests))
	group, groupCtx := errgroup.WithContext(ctx)
	for idx, req := range requests {
		idx, req := idx, req
		group.Go(func() error {
			planner, ok := registry.Get(req.Domain)
			if !ok {
				return fmt.Errorf("planner not registered for domain %s", req.Domain)
			}
			plan, err := planner.Plan(groupCtx, req)
			if err != nil {
				return err
			}
			plans[idx] = plan
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return plans, nil
}
