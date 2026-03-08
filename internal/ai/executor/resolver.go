package executor

import (
	"fmt"

	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

func ResolveParams(domain types.Domain, params map[string]any, results map[string]types.StepResult) (map[string]any, error) {
	if len(params) == 0 {
		return map[string]any{}, nil
	}
	resolved := make(map[string]any, len(params))
	for key, value := range params {
		item, err := resolveValue(domain, value, results)
		if err != nil {
			return nil, err
		}
		resolved[key] = item
	}
	return resolved, nil
}

func resolveValue(domain types.Domain, value any, results map[string]types.StepResult) (any, error) {
	ref, ok, err := types.ParseReferenceValue(value, domain)
	if err != nil {
		return nil, err
	}
	if ok {
		key := string(ref.Domain) + "." + ref.StepID
		result, exists := results[key]
		if !exists {
			return nil, fmt.Errorf("reference target not executed: %s", key)
		}
		field, exists := result.Output[ref.Field]
		if !exists {
			return nil, fmt.Errorf("reference field missing: %s.%s", key, ref.Field)
		}
		return field, nil
	}
	obj, ok := value.(map[string]any)
	if ok {
		resolved := make(map[string]any, len(obj))
		for key, item := range obj {
			resolvedItem, err := resolveValue(domain, item, results)
			if err != nil {
				return nil, err
			}
			resolved[key] = resolvedItem
		}
		return resolved, nil
	}
	items, ok := value.([]any)
	if ok {
		resolved := make([]any, 0, len(items))
		for _, item := range items {
			resolvedItem, err := resolveValue(domain, item, results)
			if err != nil {
				return nil, err
			}
			resolved = append(resolved, resolvedItem)
		}
		return resolved, nil
	}
	return value, nil
}
