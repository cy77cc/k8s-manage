package tools

import (
	"context"
	"strings"
)

func resolveToolParams(ctx context.Context, meta ToolMeta, params map[string]any, missingField string) (map[string]any, map[string]any) {
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

	// 1) runtime context from chat request
	runtime := ToolRuntimeContextFromContext(ctx)
	setIfMissing("target", runtime["target"], "runtime")
	setIfMissing("host_id", runtime["host_id"], "runtime")
	setIfMissing("cluster_id", runtime["cluster_id"], "runtime")
	setIfMissing("namespace", runtime["namespace"], "runtime")
	setIfMissing("service_id", runtime["service_id"], "runtime")
	setIfMissing("env", runtime["env"], "runtime")
	setIfMissing("runtime_type", runtime["runtime_type"], "runtime")

	// 2) session memory by tool
	if accessor := ToolMemoryAccessorFromContext(ctx); accessor != nil {
		for k, v := range accessor.GetLastToolParams(meta.Name) {
			setIfMissing(k, v, "memory")
		}
	}

	// 3) configured defaults + safety defaults
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
