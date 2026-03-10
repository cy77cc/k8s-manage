package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/cicd"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/deployment"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/governance"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/host"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/infrastructure"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/kubernetes"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/monitor"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/service"
)

func NewCommonTools(ctx context.Context, deps common.PlatformDeps) []tool.BaseTool {
	return []tool.BaseTool{
		cicd.CICDPipelineList(ctx, deps),
		deployment.ClusterListInventory(ctx, deps),
		governance.AuditLogSearch(ctx, deps),
		host.HostListInventory(ctx, deps),
		infrastructure.CredentialList(ctx, deps),
		kubernetes.K8sListResources(ctx, deps),
		monitor.MonitorAlertRuleList(ctx, deps),
		service.ServiceCatalogList(ctx, deps),
	}
}
