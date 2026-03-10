package experts

import (
	"sort"

	deliveryexpert "github.com/cy77cc/OpsPilot/internal/ai/experts/delivery"
	hostopsexpert "github.com/cy77cc/OpsPilot/internal/ai/experts/hostops"
	k8sexpert "github.com/cy77cc/OpsPilot/internal/ai/experts/k8s"
	observabilityexpert "github.com/cy77cc/OpsPilot/internal/ai/experts/observability"
	serviceexpert "github.com/cy77cc/OpsPilot/internal/ai/experts/service"
	expertspec "github.com/cy77cc/OpsPilot/internal/ai/experts/spec"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
)

type Registry struct {
	experts map[string]expertspec.Expert
}

func NewRegistry(experts ...expertspec.Expert) *Registry {
	items := make(map[string]expertspec.Expert, len(experts))
	for _, exp := range experts {
		if exp == nil {
			continue
		}
		items[exp.Name()] = exp
	}
	return &Registry{experts: items}
}

func DefaultRegistry(deps common.PlatformDeps) *Registry {
	return NewRegistry(
		hostopsexpert.New(deps),
		k8sexpert.New(deps),
		serviceexpert.New(deps),
		deliveryexpert.New(deps),
		observabilityexpert.New(deps),
	)
}

func (r *Registry) Get(name string) (expertspec.Expert, bool) {
	if r == nil {
		return nil, false
	}
	exp, ok := r.experts[name]
	return exp, ok
}

func (r *Registry) List() []expertspec.Expert {
	if r == nil {
		return nil
	}
	keys := make([]string, 0, len(r.experts))
	for key := range r.experts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]expertspec.Expert, 0, len(keys))
	for _, key := range keys {
		out = append(out, r.experts[key])
	}
	return out
}

func (r *Registry) ToolDirectory() []expertspec.ToolExport {
	exps := r.List()
	out := make([]expertspec.ToolExport, 0, len(exps))
	for _, exp := range exps {
		out = append(out, exp.AsTool())
	}
	return out
}
