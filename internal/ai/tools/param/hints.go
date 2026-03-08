package param

import (
	"context"

	core "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
)

type ParamHintValue = core.ParamHintValue
type ParamHintItem = core.ParamHintItem
type ToolParamHintsResponse = core.ToolParamHintsResponse

func ResolveToolParamHints(ctx context.Context, deps core.PlatformDeps, meta core.ToolMeta) ToolParamHintsResponse {
	return core.ResolveToolParamHints(ctx, deps, meta)
}
