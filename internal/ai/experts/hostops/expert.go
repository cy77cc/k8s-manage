package hostops

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	expertspec "github.com/cy77cc/OpsPilot/internal/ai/experts/spec"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
	hosttools "github.com/cy77cc/OpsPilot/internal/ai/tools/host"
)

type Expert struct {
	deps common.PlatformDeps
}

func New(deps common.PlatformDeps) *Expert {
	return &Expert{deps: deps}
}

func (e *Expert) Name() string { return "hostops" }

func (e *Expert) Description() string {
	return "Host operations expert for host inventory, readonly diagnostics, and governed batch execution."
}

func (e *Expert) Tools(ctx context.Context) []tool.InvokableTool {
	return expertspec.FilterToolsByName(ctx, hosttools.NewHostTools(ctx, e.deps),
		"host_list_inventory",
		"host_batch",
	)
}

func (e *Expert) Capabilities() []expertspec.ToolCapability {
	return []expertspec.ToolCapability{
		{Name: "host_exec", Mode: "readonly", Risk: "low", Description: "Run readonly host commands."},
		{Name: "host_batch_exec_preview", Mode: "readonly", Risk: "medium", Description: "Preview batch host execution."},
		{Name: "host_batch_exec_apply", Mode: "mutating", Risk: "high", Description: "Apply batch host execution."},
		{Name: "host_batch_status_update", Mode: "mutating", Risk: "high", Description: "Change host status in batch."},
		{Name: "os_get_cpu_mem", Mode: "readonly", Risk: "low", Description: "Collect CPU and memory diagnostics."},
		{Name: "os_get_disk_fs", Mode: "readonly", Risk: "low", Description: "Collect disk filesystem diagnostics."},
	}
}

func (e *Expert) AsTool() expertspec.ToolExport {
	return expertspec.ToolExport{
		Name:         "hostops_expert",
		Description:  e.Description(),
		Capabilities: e.Capabilities(),
	}
}
