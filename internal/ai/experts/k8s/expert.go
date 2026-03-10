package k8s

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	expertspec "github.com/cy77cc/OpsPilot/internal/ai/experts/spec"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
	k8stools "github.com/cy77cc/OpsPilot/internal/ai/tools/kubernetes"
)

type Expert struct {
	deps common.PlatformDeps
}

func New(deps common.PlatformDeps) *Expert {
	return &Expert{deps: deps}
}

func (e *Expert) Name() string { return "k8s" }

func (e *Expert) Description() string {
	return "Kubernetes expert for cluster workload queries, events, and pod log inspection."
}

func (e *Expert) Tools(ctx context.Context) []tool.InvokableTool {
	return k8stools.NewKubernetesTools(ctx, e.deps)
}

func (e *Expert) Capabilities() []expertspec.ToolCapability {
	return []expertspec.ToolCapability{
		{Name: "k8s_query", Mode: "readonly", Risk: "low", Description: "Query Kubernetes resources."},
		{Name: "k8s_list_resources", Mode: "readonly", Risk: "low", Description: "List Kubernetes resources by type."},
		{Name: "k8s_events", Mode: "readonly", Risk: "low", Description: "Inspect Kubernetes events."},
		{Name: "k8s_get_events", Mode: "readonly", Risk: "low", Description: "Inspect filtered Kubernetes events."},
		{Name: "k8s_logs", Mode: "readonly", Risk: "low", Description: "Fetch pod/container logs."},
		{Name: "k8s_get_pod_logs", Mode: "readonly", Risk: "low", Description: "Fetch pod logs directly."},
	}
}

func (e *Expert) AsTool() expertspec.ToolExport {
	return expertspec.ToolExport{
		Name:         "k8s_expert",
		Description:  e.Description(),
		Capabilities: e.Capabilities(),
	}
}
