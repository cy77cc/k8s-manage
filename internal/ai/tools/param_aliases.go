package tools

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/tools/param"
)

type ParamHintValue = param.ParamHintValue
type ParamHintItem = param.ParamHintItem
type ToolParamHintsResponse = param.ToolParamHintsResponse

func ResolveToolParamHints(ctx context.Context, deps PlatformDeps, meta ToolMeta) ToolParamHintsResponse {
	return param.ResolveToolParamHints(ctx, deps, meta)
}

func resolveToolParams(ctx context.Context, meta ToolMeta, params map[string]any, missingField string) (map[string]any, map[string]any) {
	return param.ResolveToolParams(ctx, meta, params, missingField)
}

func validateResolvedParams(meta ToolMeta, params map[string]any) error {
	return param.ValidateResolvedParams(meta, params)
}
