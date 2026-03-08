package types

import (
	"fmt"
	"strings"
)

type Reference struct {
	Domain Domain `json:"domain"`
	StepID string `json:"step_id"`
	Field  string `json:"field"`
}

func ParseReferenceValue(value any, currentDomain Domain) (Reference, bool, error) {
	obj, ok := value.(map[string]any)
	if !ok {
		return Reference{}, false, nil
	}
	raw, ok := obj["$ref"]
	if !ok {
		return Reference{}, false, nil
	}
	ref, ok := raw.(string)
	if !ok {
		return Reference{}, false, fmt.Errorf("$ref must be a string")
	}
	parsed, err := ParseReference(ref, currentDomain)
	if err != nil {
		return Reference{}, true, err
	}
	return parsed, true, nil
}

func ParseReference(ref string, currentDomain Domain) (Reference, error) {
	parts := strings.Split(strings.TrimSpace(ref), ".")
	switch len(parts) {
	case 2:
		if currentDomain == "" {
			return Reference{}, fmt.Errorf("domain is required for local reference: %s", ref)
		}
		if parts[0] == "" || parts[1] == "" {
			return Reference{}, fmt.Errorf("invalid reference: %s", ref)
		}
		return Reference{Domain: currentDomain, StepID: parts[0], Field: parts[1]}, nil
	case 3:
		for _, part := range parts {
			if strings.TrimSpace(part) == "" {
				return Reference{}, fmt.Errorf("invalid reference: %s", ref)
			}
		}
		return Reference{Domain: Domain(parts[0]), StepID: parts[1], Field: parts[2]}, nil
	default:
		return Reference{}, fmt.Errorf("invalid reference: %s", ref)
	}
}
