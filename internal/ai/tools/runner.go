package tools

import (
	"context"

	core "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
	"k8s.io/client-go/kubernetes"
)

func runWithPolicyAndEvent[T any](ctx context.Context, meta ToolMeta, input T, do func(T) (any, string, error)) (ToolResult, error) {
	return core.RunWithPolicyAndEvent(ctx, meta, input, do)
}

func emitPolicyRequiredEvent(ctx context.Context, meta ToolMeta, err error) {
	core.EmitPolicyRequiredEvent(ctx, meta, err)
}

func resolveK8sClient(deps PlatformDeps, params map[string]any) (*kubernetes.Clientset, string, error) {
	return core.ResolveK8sClient(deps, params)
}

func structToMap(v any) map[string]any {
	return core.StructToMap(v)
}

func nextToolCallID() string {
	return core.NextToolCallID()
}
