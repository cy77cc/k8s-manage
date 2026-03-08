package orchestrator

import "github.com/cy77cc/k8s-manage/internal/ai/types"

type DomainRequest struct {
	Domain     types.Domain   `json:"domain"`
	UserIntent string         `json:"user_intent,omitempty"`
	Context    map[string]any `json:"context,omitempty"`
}

type DomainDescriptor struct {
	Domain      types.Domain `json:"domain"`
	Description string       `json:"description"`
}
