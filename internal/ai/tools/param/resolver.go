package param

import (
	"context"

	core "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
)

func ResolveToolParams(ctx context.Context, meta core.ToolMeta, params map[string]any, missingField string) (map[string]any, map[string]any) {
	return core.ResolveToolParams(ctx, meta, params, missingField)
}
