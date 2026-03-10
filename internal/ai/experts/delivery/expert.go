package delivery

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	expertspec "github.com/cy77cc/OpsPilot/internal/ai/experts/spec"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/cicd"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
)

type Expert struct {
	deps common.PlatformDeps
}

func New(deps common.PlatformDeps) *Expert {
	return &Expert{deps: deps}
}

func (e *Expert) Name() string { return "delivery" }

func (e *Expert) Description() string {
	return "Delivery expert for CI/CD pipelines, release jobs, and automation run status."
}

func (e *Expert) Tools(ctx context.Context) []tool.InvokableTool {
	return cicd.NewCICDTools(ctx, e.deps)
}

func (e *Expert) Capabilities() []expertspec.ToolCapability {
	return []expertspec.ToolCapability{
		{Name: "cicd_pipeline_list", Mode: "readonly", Risk: "low", Description: "List pipeline configs."},
		{Name: "cicd_pipeline_status", Mode: "readonly", Risk: "low", Description: "Inspect pipeline status and runs."},
		{Name: "cicd_pipeline_trigger", Mode: "mutating", Risk: "high", Description: "Trigger a pipeline run."},
		{Name: "job_list", Mode: "readonly", Risk: "low", Description: "List platform jobs."},
		{Name: "job_execution_status", Mode: "readonly", Risk: "low", Description: "Inspect job execution status."},
		{Name: "job_run", Mode: "mutating", Risk: "high", Description: "Trigger a job run."},
	}
}

func (e *Expert) AsTool() expertspec.ToolExport {
	return expertspec.ToolExport{
		Name:         "delivery_expert",
		Description:  e.Description(),
		Capabilities: e.Capabilities(),
	}
}
