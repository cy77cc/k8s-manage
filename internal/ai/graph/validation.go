package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

type Validator interface {
	Validate(ctx context.Context, calls []schema.ToolCall) []ValidationError
}

type OpenAPIValidator struct {
	registry *tools.Registry
}

func NewOpenAPIValidator(registry *tools.Registry) *OpenAPIValidator {
	return &OpenAPIValidator{registry: registry}
}

func (v *OpenAPIValidator) Validate(_ context.Context, calls []schema.ToolCall) []ValidationError {
	errs := make([]ValidationError, 0)
	for _, call := range calls {
		errs = append(errs, v.validateToolCall(call)...)
	}
	return errs
}

func (v *OpenAPIValidator) validateToolCall(call schema.ToolCall) []ValidationError {
	var payload map[string]any
	if strings.TrimSpace(call.Function.Arguments) == "" {
		payload = map[string]any{}
	} else if err := json.Unmarshal([]byte(call.Function.Arguments), &payload); err != nil {
		return []ValidationError{{
			ToolName: call.Function.Name,
			Message:  fmt.Sprintf("invalid tool arguments JSON: %v", err),
		}}
	}

	errs := make([]ValidationError, 0)
	if v.registry != nil {
		if registered, ok := v.registry.Get(call.Function.Name); ok {
			for _, field := range registered.Meta.Required {
				if _, exists := payload[field]; !exists {
					errs = append(errs, ValidationError{
						ToolName: call.Function.Name,
						Field:    field,
						Message:  "missing required field",
					})
				}
			}
		}
	}

	if isK8sTool(call.Function.Name) {
		errs = append(errs, validateK8sOpenAPI(payload, call.Function.Name)...)
	}

	return errs
}

func isK8sTool(name string) bool {
	name = tools.NormalizeToolName(name)
	return strings.HasPrefix(name, "k8s_") || strings.HasPrefix(name, "cluster_")
}

// validateK8sOpenAPI performs lightweight resource-shape validation before execution.
func validateK8sOpenAPI(payload map[string]any, toolName string) []ValidationError {
	manifest, ok := payload["manifest"].(map[string]any)
	if !ok {
		resource, hasResource := payload["resource"].(string)
		if hasResource && strings.TrimSpace(resource) != "" {
			return nil
		}
		return nil
	}

	errs := make([]ValidationError, 0)
	if strings.TrimSpace(asString(manifest["apiVersion"])) == "" {
		errs = append(errs, ValidationError{ToolName: toolName, Field: "manifest.apiVersion", Message: "missing apiVersion"})
	}
	if strings.TrimSpace(asString(manifest["kind"])) == "" {
		errs = append(errs, ValidationError{ToolName: toolName, Field: "manifest.kind", Message: "missing kind"})
	}
	meta, _ := manifest["metadata"].(map[string]any)
	if strings.TrimSpace(asString(meta["name"])) == "" {
		errs = append(errs, ValidationError{ToolName: toolName, Field: "manifest.metadata.name", Message: "missing metadata.name"})
	}
	return errs
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
