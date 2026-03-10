package observability

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	expertspec "github.com/cy77cc/OpsPilot/internal/ai/experts/spec"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
	governancetools "github.com/cy77cc/OpsPilot/internal/ai/tools/governance"
	monitortools "github.com/cy77cc/OpsPilot/internal/ai/tools/monitor"
)

type Expert struct {
	deps common.PlatformDeps
}

func New(deps common.PlatformDeps) *Expert {
	return &Expert{deps: deps}
}

func (e *Expert) Name() string { return "observability" }

func (e *Expert) Description() string {
	return "Observability expert for alerts, metrics, topology, and audit evidence."
}

func (e *Expert) Tools(ctx context.Context) []tool.InvokableTool {
	out := make([]tool.InvokableTool, 0, 8)
	out = append(out, monitortools.NewMonitorTools(ctx, e.deps)...)
	out = append(out, governancetools.TopologyGet(ctx, e.deps))
	out = append(out, governancetools.AuditLogSearch(ctx, e.deps))
	return out
}

func (e *Expert) Capabilities() []expertspec.ToolCapability {
	return []expertspec.ToolCapability{
		{Name: "monitor_alert_rule_list", Mode: "readonly", Risk: "low", Description: "List alerting rules."},
		{Name: "monitor_alert", Mode: "readonly", Risk: "low", Description: "Inspect firing alerts."},
		{Name: "monitor_alert_active", Mode: "readonly", Risk: "low", Description: "Inspect active alerts."},
		{Name: "monitor_metric", Mode: "readonly", Risk: "low", Description: "Query metric time series."},
		{Name: "monitor_metric_query", Mode: "readonly", Risk: "low", Description: "Query metric expressions."},
		{Name: "topology_get", Mode: "readonly", Risk: "low", Description: "Inspect service topology."},
		{Name: "audit_log_search", Mode: "readonly", Risk: "low", Description: "Search audit evidence."},
	}
}

func (e *Expert) AsTool() expertspec.ToolExport {
	return expertspec.ToolExport{
		Name:         "observability_expert",
		Description:  e.Description(),
		Capabilities: e.Capabilities(),
	}
}
