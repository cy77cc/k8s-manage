package planner

import (
	"sort"

	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type Registry struct {
	planners map[types.Domain]DomainPlanner
}

func NewRegistry() *Registry {
	return &Registry{planners: map[types.Domain]DomainPlanner{}}
}

func (r *Registry) Register(planner DomainPlanner) {
	if r == nil || planner == nil {
		return
	}
	r.planners[planner.Domain()] = planner
}

func (r *Registry) Get(domain types.Domain) (DomainPlanner, bool) {
	if r == nil {
		return nil, false
	}
	planner, ok := r.planners[domain]
	return planner, ok
}

func (r *Registry) List() []DomainPlanner {
	if r == nil {
		return nil
	}
	items := make([]DomainPlanner, 0, len(r.planners))
	for _, item := range r.planners {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Domain() < items[j].Domain() })
	return items
}
