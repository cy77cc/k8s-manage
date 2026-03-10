package service

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	expertspec "github.com/cy77cc/OpsPilot/internal/ai/experts/spec"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
	deploymenttools "github.com/cy77cc/OpsPilot/internal/ai/tools/deployment"
	infratools "github.com/cy77cc/OpsPilot/internal/ai/tools/infrastructure"
	servicetools "github.com/cy77cc/OpsPilot/internal/ai/tools/service"
)

type Expert struct {
	deps common.PlatformDeps
}

func New(deps common.PlatformDeps) *Expert {
	return &Expert{deps: deps}
}

func (e *Expert) Name() string { return "service" }

func (e *Expert) Description() string {
	return "Service expert for service status, deployment targets, config inspection, and credential checks."
}

func (e *Expert) Tools(ctx context.Context) []tool.InvokableTool {
	out := make([]tool.InvokableTool, 0, 24)
	out = append(out, expertspec.FilterToolsByName(ctx, servicetools.NewServiceTools(ctx, e.deps),
		"service_deploy",
	)...)
	out = append(out, expertspec.FilterToolsByName(ctx, deploymenttools.NewDeploymentTools(ctx, e.deps),
		"cluster_list_inventory",
		"service_list_inventory",
	)...)
	out = append(out, infratools.NewInfrastructureTools(ctx, e.deps)...)
	return out
}

func (e *Expert) Capabilities() []expertspec.ToolCapability {
	return []expertspec.ToolCapability{
		{Name: "service_get_detail", Mode: "readonly", Risk: "low", Description: "Fetch full service detail."},
		{Name: "service_status", Mode: "readonly", Risk: "low", Description: "Fetch current service runtime status."},
		{Name: "service_deploy_preview", Mode: "readonly", Risk: "medium", Description: "Preview a service deployment."},
		{Name: "service_deploy_apply", Mode: "mutating", Risk: "high", Description: "Apply a service deployment."},
		{Name: "deployment_target_list", Mode: "readonly", Risk: "low", Description: "List deployment targets."},
		{Name: "config_item_get", Mode: "readonly", Risk: "low", Description: "Read service config items."},
		{Name: "config_diff", Mode: "readonly", Risk: "low", Description: "Compare service config across environments."},
		{Name: "credential_list", Mode: "readonly", Risk: "low", Description: "List infrastructure credentials."},
		{Name: "credential_test", Mode: "readonly", Risk: "medium", Description: "Check credential test status."},
	}
}

func (e *Expert) AsTool() expertspec.ToolExport {
	return expertspec.ToolExport{
		Name:         "service_expert",
		Description:  e.Description(),
		Capabilities: e.Capabilities(),
	}
}
