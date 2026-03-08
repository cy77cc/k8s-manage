package executor

import (
	"fmt"
	"sort"

	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

func BuildDAG(plans []types.DomainPlan) (*DAG, error) {
	dag := &DAG{
		Steps:        map[string]types.PlanStep{},
		Domains:      map[string]types.Domain{},
		Dependencies: map[string][]string{},
	}
	for _, plan := range plans {
		if err := plan.Validate(); err != nil {
			return nil, err
		}
		for _, step := range plan.Steps {
			globalID := string(plan.Domain) + "." + step.ID
			if _, exists := dag.Steps[globalID]; exists {
				return nil, fmt.Errorf("duplicate global step id: %s", globalID)
			}
			dag.Steps[globalID] = step
			dag.Domains[globalID] = plan.Domain
			deps := make([]string, 0, len(step.DependsOn)+1)
			for _, dep := range step.DependsOn {
				deps = append(deps, string(plan.Domain)+"."+dep)
			}
			if err := appendReferenceDeps(&deps, plan.Domain, step.Params); err != nil {
				return nil, err
			}
			dag.Dependencies[globalID] = uniqueStrings(deps)
		}
	}
	if err := validateDependencies(dag); err != nil {
		return nil, err
	}
	order, err := topoSort(dag)
	if err != nil {
		return nil, err
	}
	dag.Order = order
	return dag, nil
}

func appendReferenceDeps(deps *[]string, domain types.Domain, params map[string]any) error {
	for _, value := range params {
		if err := collectRefDeps(deps, domain, value); err != nil {
			return err
		}
	}
	return nil
}

func collectRefDeps(deps *[]string, domain types.Domain, value any) error {
	ref, ok, err := types.ParseReferenceValue(value, domain)
	if err != nil {
		return err
	}
	if ok {
		*deps = append(*deps, string(ref.Domain)+"."+ref.StepID)
		return nil
	}
	if obj, ok := value.(map[string]any); ok {
		for _, nested := range obj {
			if err := collectRefDeps(deps, domain, nested); err != nil {
				return err
			}
		}
	}
	if items, ok := value.([]any); ok {
		for _, nested := range items {
			if err := collectRefDeps(deps, domain, nested); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateDependencies(dag *DAG) error {
	for stepID, deps := range dag.Dependencies {
		for _, dep := range deps {
			if _, ok := dag.Steps[dep]; !ok {
				return fmt.Errorf("依赖步骤不存在: %s", dep)
			}
		}
		if _, ok := dag.Steps[stepID]; !ok {
			return fmt.Errorf("step missing: %s", stepID)
		}
	}
	return nil
}

func topoSort(dag *DAG) ([]string, error) {
	inDegree := make(map[string]int, len(dag.Steps))
	adj := make(map[string][]string, len(dag.Steps))
	for stepID := range dag.Steps {
		inDegree[stepID] = 0
	}
	for stepID, deps := range dag.Dependencies {
		inDegree[stepID] = len(deps)
		for _, dep := range deps {
			adj[dep] = append(adj[dep], stepID)
		}
	}
	queue := make([]string, 0)
	for stepID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, stepID)
		}
	}
	sort.Strings(queue)
	order := make([]string, 0, len(dag.Steps))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)
		for _, next := range adj[current] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
				sort.Strings(queue)
			}
		}
	}
	if len(order) != len(dag.Steps) {
		return nil, fmt.Errorf("检测到循环依赖")
	}
	return order, nil
}

func uniqueStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
