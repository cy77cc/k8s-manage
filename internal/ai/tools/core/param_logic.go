package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func ResolveToolParams(ctx context.Context, meta ToolMeta, params map[string]any, missingField string) (map[string]any, map[string]any) {
	resolved := cloneMap(params)
	resolution := map[string]any{
		"original": cloneMap(params),
		"applied":  map[string]any{},
		"source":   map[string]string{},
	}
	applied := resolution["applied"].(map[string]any)
	source := resolution["source"].(map[string]string)

	setIfMissing := func(key string, value any, src string) {
		if isEmptyValue(value) {
			return
		}
		if !isEmptyValue(resolved[key]) {
			return
		}
		resolved[key] = value
		applied[key] = value
		source[key] = src
	}

	runtime := ToolRuntimeContextFromContext(ctx)
	setIfMissing("target", runtime["target"], "runtime")
	setIfMissing("host_id", runtime["host_id"], "runtime")
	setIfMissing("cluster_id", runtime["cluster_id"], "runtime")
	setIfMissing("namespace", runtime["namespace"], "runtime")
	setIfMissing("service_id", runtime["service_id"], "runtime")
	setIfMissing("env", runtime["env"], "runtime")
	setIfMissing("runtime_type", runtime["runtime_type"], "runtime")

	if accessor := ToolMemoryAccessorFromContext(ctx); accessor != nil {
		for k, v := range accessor.GetLastToolParams(meta.Name) {
			setIfMissing(k, v, "memory")
		}
	}

	for k, v := range meta.DefaultHint {
		setIfMissing(k, v, "meta_default")
	}
	setIfMissing("target", "localhost", "safety_default")
	setIfMissing("namespace", "default", "safety_default")
	setIfMissing("limit", 50, "safety_default")
	setIfMissing("tail_lines", 200, "safety_default")
	setIfMissing("lines", 200, "safety_default")

	if missingField != "" {
		resolution["missing_field"] = missingField
	}
	resolution["final"] = cloneMap(resolved)
	return resolved, resolution
}

func ValidateResolvedParams(meta ToolMeta, params map[string]any) error {
	for _, field := range meta.Required {
		if isEmptyValue(params[field]) {
			return NewMissingParam(field, missingParamMessage(meta, field))
		}
	}
	properties := schemaProperties(meta.Schema)
	for field, raw := range params {
		prop, ok := properties[field]
		if !ok {
			continue
		}
		if err := validateType(field, raw, prop); err != nil {
			return err
		}
		if err := validateEnum(meta, field, raw, prop); err != nil {
			return err
		}
	}
	return nil
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func isEmptyValue(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	return false
}

func missingParamMessage(meta ToolMeta, field string) string {
	parts := []string{fmt.Sprintf("missing required parameter `%s`", field)}
	if source := strings.TrimSpace(meta.EnumSources[field]); source != "" {
		parts = append(parts, fmt.Sprintf("you can call `%s` to get candidate values", source))
	}
	if hint := strings.TrimSpace(meta.ParamHints[field]); hint != "" {
		parts = append(parts, hint)
	}
	return strings.Join(parts, "; ")
}

func validateType(field string, value any, prop map[string]any) error {
	typ, _ := prop["type"].(string)
	typ = strings.TrimSpace(strings.ToLower(typ))
	if typ == "" {
		return nil
	}
	switch typ {
	case "integer":
		if _, ok := toInt64(value); !ok {
			return NewInvalidParam(field, fmt.Sprintf("`%s` expects integer, got %T", field, value))
		}
	case "number":
		if _, ok := toFloat64(value); !ok {
			return NewInvalidParam(field, fmt.Sprintf("`%s` expects number, got %T", field, value))
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			if s, ok := value.(string); ok {
				if _, err := strconv.ParseBool(strings.TrimSpace(s)); err == nil {
					return nil
				}
			}
			return NewInvalidParam(field, fmt.Sprintf("`%s` expects boolean(true/false)", field))
		}
	case "array":
		if _, ok := value.([]any); ok {
			return nil
		}
		if _, ok := value.([]string); ok {
			return nil
		}
		if _, ok := value.([]int); ok {
			return nil
		}
		return NewInvalidParam(field, fmt.Sprintf("`%s` expects array value", field))
	case "string":
		if _, ok := value.(string); !ok {
			return NewInvalidParam(field, fmt.Sprintf("`%s` expects string, got %T", field, value))
		}
	}
	return nil
}

func validateEnum(meta ToolMeta, field string, value any, prop map[string]any) error {
	rawEnum, ok := prop["enum"].([]any)
	if !ok || len(rawEnum) == 0 {
		return nil
	}
	current := strings.TrimSpace(fmt.Sprintf("%v", value))
	for _, item := range rawEnum {
		if strings.EqualFold(current, strings.TrimSpace(fmt.Sprintf("%v", item))) {
			return nil
		}
	}
	suggestions := make([]string, 0, len(rawEnum))
	for _, item := range rawEnum {
		suggestions = append(suggestions, fmt.Sprintf("%v", item))
	}
	msg := fmt.Sprintf("`%s` value `%s` is invalid, allowed values: %s", field, current, strings.Join(suggestions, ", "))
	if source := strings.TrimSpace(meta.EnumSources[field]); source != "" {
		msg += fmt.Sprintf("; or call `%s` to discover values", source)
	}
	return NewInvalidParam(field, msg)
}

func toInt64(v any) (int64, bool) {
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func toFloat64(v any) (float64, bool) {
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return 0, false
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}
