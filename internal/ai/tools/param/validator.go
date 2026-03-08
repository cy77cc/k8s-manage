package param

import core "github.com/cy77cc/k8s-manage/internal/ai/tools/core"

func ValidateResolvedParams(meta core.ToolMeta, params map[string]any) error {
	return core.ValidateResolvedParams(meta, params)
}
