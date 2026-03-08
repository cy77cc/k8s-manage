package tools

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/cicd"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/deployment"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/governance"
	hostimpl "github.com/cy77cc/k8s-manage/internal/ai/tools/impl/host"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/infrastructure"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/kubernetes"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/monitor"
	serviceimpl "github.com/cy77cc/k8s-manage/internal/ai/tools/impl/service"
)

func osGetCPUMem(ctx context.Context, deps PlatformDeps, input OSCPUMemInput) (ToolResult, error) {
	return hostimpl.OSGetCPUMem(ctx, deps, input)
}

func osGetDiskFS(ctx context.Context, deps PlatformDeps, input OSDiskInput) (ToolResult, error) {
	return hostimpl.OSGetDiskFS(ctx, deps, input)
}

func osGetNetStat(ctx context.Context, deps PlatformDeps, input OSNetInput) (ToolResult, error) {
	return hostimpl.OSGetNetStat(ctx, deps, input)
}

func osGetProcessTop(ctx context.Context, deps PlatformDeps, input OSProcessTopInput) (ToolResult, error) {
	return hostimpl.OSGetProcessTop(ctx, deps, input)
}

func osGetJournalTail(ctx context.Context, deps PlatformDeps, input OSJournalInput) (ToolResult, error) {
	return hostimpl.OSGetJournalTail(ctx, deps, input)
}

func osGetContainerRuntime(ctx context.Context, deps PlatformDeps, input OSContainerRuntimeInput) (ToolResult, error) {
	return hostimpl.OSGetContainerRuntime(ctx, deps, input)
}

func hostSSHReadonly(ctx context.Context, deps PlatformDeps, input HostSSHReadonlyInput) (ToolResult, error) {
	return hostimpl.HostSSHReadonly(ctx, deps, input)
}

func hostExec(ctx context.Context, deps PlatformDeps, input HostExecInput) (ToolResult, error) {
	return hostimpl.HostExec(ctx, deps, input)
}

func hostListInventory(ctx context.Context, deps PlatformDeps, input HostInventoryInput) (ToolResult, error) {
	return hostimpl.HostListInventory(ctx, deps, input)
}

func hostBatch(ctx context.Context, deps PlatformDeps, input HostBatchInput) (ToolResult, error) {
	return hostimpl.HostBatch(ctx, deps, input)
}

func hostBatchExecPreview(ctx context.Context, deps PlatformDeps, input HostBatchExecPreviewInput) (ToolResult, error) {
	return hostimpl.HostBatchExecPreview(ctx, deps, input)
}

func hostBatchExecApply(ctx context.Context, deps PlatformDeps, input HostBatchExecApplyInput) (ToolResult, error) {
	return hostimpl.HostBatchExecApply(ctx, deps, input)
}

func hostBatchStatusUpdate(ctx context.Context, deps PlatformDeps, input HostBatchStatusInput) (ToolResult, error) {
	return hostimpl.HostBatchStatusUpdate(ctx, deps, input)
}

func k8sQuery(ctx context.Context, deps PlatformDeps, input K8sQueryInput) (ToolResult, error) {
	return kubernetes.K8sQuery(ctx, deps, input)
}

func k8sListResources(ctx context.Context, deps PlatformDeps, input K8sListInput) (ToolResult, error) {
	return kubernetes.K8sListResources(ctx, deps, input)
}

func k8sEvents(ctx context.Context, deps PlatformDeps, input K8sEventsQueryInput) (ToolResult, error) {
	return kubernetes.K8sEvents(ctx, deps, input)
}

func k8sGetEvents(ctx context.Context, deps PlatformDeps, input K8sEventsInput) (ToolResult, error) {
	return kubernetes.K8sGetEvents(ctx, deps, input)
}

func k8sLogs(ctx context.Context, deps PlatformDeps, input K8sLogsInput) (ToolResult, error) {
	return kubernetes.K8sLogs(ctx, deps, input)
}

func k8sGetPodLogs(ctx context.Context, deps PlatformDeps, input K8sPodLogsInput) (ToolResult, error) {
	return kubernetes.K8sGetPodLogs(ctx, deps, input)
}

func serviceGetDetail(ctx context.Context, deps PlatformDeps, input ServiceDetailInput) (ToolResult, error) {
	return serviceimpl.ServiceGetDetail(ctx, deps, input)
}

func serviceStatus(ctx context.Context, deps PlatformDeps, input ServiceStatusInput) (ToolResult, error) {
	return serviceimpl.ServiceStatus(ctx, deps, input)
}

func serviceDeployPreview(ctx context.Context, deps PlatformDeps, input ServiceDeployPreviewInput) (ToolResult, error) {
	return serviceimpl.ServiceDeployPreview(ctx, deps, input)
}

func serviceDeployApply(ctx context.Context, deps PlatformDeps, input ServiceDeployApplyInput) (ToolResult, error) {
	return serviceimpl.ServiceDeployApply(ctx, deps, input)
}

func serviceDeploy(ctx context.Context, deps PlatformDeps, input ServiceDeployInput) (ToolResult, error) {
	return serviceimpl.ServiceDeploy(ctx, deps, input)
}

func serviceCatalogList(ctx context.Context, deps PlatformDeps, input ServiceCatalogListInput) (ToolResult, error) {
	return serviceimpl.ServiceCatalogList(ctx, deps, input)
}

func serviceVisibilityCheck(ctx context.Context, deps PlatformDeps, input ServiceVisibilityCheckInput) (ToolResult, error) {
	return serviceimpl.ServiceVisibilityCheck(ctx, deps, input)
}

func serviceCategoryTree(ctx context.Context, deps PlatformDeps) (ToolResult, error) {
	return serviceimpl.ServiceCategoryTree(ctx, deps)
}

func monitorAlertRuleList(ctx context.Context, deps PlatformDeps, input MonitorAlertRuleListInput) (ToolResult, error) {
	return monitor.MonitorAlertRuleList(ctx, deps, input)
}

func monitorAlert(ctx context.Context, deps PlatformDeps, input MonitorAlertInput) (ToolResult, error) {
	return monitor.MonitorAlert(ctx, deps, input)
}

func monitorAlertActive(ctx context.Context, deps PlatformDeps, input MonitorAlertActiveInput) (ToolResult, error) {
	return monitor.MonitorAlertActive(ctx, deps, input)
}

func monitorMetric(ctx context.Context, deps PlatformDeps, input MonitorMetricInput) (ToolResult, error) {
	return monitor.MonitorMetric(ctx, deps, input)
}

func monitorMetricQuery(ctx context.Context, deps PlatformDeps, input MonitorMetricQueryInput) (ToolResult, error) {
	return monitor.MonitorMetricQuery(ctx, deps, input)
}

func cicdPipelineList(ctx context.Context, deps PlatformDeps, input CICDPipelineListInput) (ToolResult, error) {
	return cicd.CICDPipelineList(ctx, deps, input)
}

func cicdPipelineStatus(ctx context.Context, deps PlatformDeps, input CICDPipelineStatusInput) (ToolResult, error) {
	return cicd.CICDPipelineStatus(ctx, deps, input)
}

func cicdPipelineTrigger(ctx context.Context, deps PlatformDeps, input CICDPipelineTriggerInput) (ToolResult, error) {
	return cicd.CICDPipelineTrigger(ctx, deps, input)
}

func jobList(ctx context.Context, deps PlatformDeps, input JobListInput) (ToolResult, error) {
	return cicd.JobList(ctx, deps, input)
}

func jobExecutionStatus(ctx context.Context, deps PlatformDeps, input JobExecutionStatusInput) (ToolResult, error) {
	return cicd.JobExecutionStatus(ctx, deps, input)
}

func jobRun(ctx context.Context, deps PlatformDeps, input JobRunInput) (ToolResult, error) {
	return cicd.JobRun(ctx, deps, input)
}

func deploymentTargetList(ctx context.Context, deps PlatformDeps, input DeploymentTargetListInput) (ToolResult, error) {
	return deployment.DeploymentTargetList(ctx, deps, input)
}

func deploymentTargetDetail(ctx context.Context, deps PlatformDeps, input DeploymentTargetDetailInput) (ToolResult, error) {
	return deployment.DeploymentTargetDetail(ctx, deps, input)
}

func deploymentBootstrapStatus(ctx context.Context, deps PlatformDeps, input DeploymentBootstrapStatusInput) (ToolResult, error) {
	return deployment.DeploymentBootstrapStatus(ctx, deps, input)
}

func configAppList(ctx context.Context, deps PlatformDeps, input ConfigAppListInput) (ToolResult, error) {
	return deployment.ConfigAppList(ctx, deps, input)
}

func configItemGet(ctx context.Context, deps PlatformDeps, input ConfigItemGetInput) (ToolResult, error) {
	return deployment.ConfigItemGet(ctx, deps, input)
}

func configDiff(ctx context.Context, deps PlatformDeps, input ConfigDiffInput) (ToolResult, error) {
	return deployment.ConfigDiff(ctx, deps, input)
}

func clusterListInventory(ctx context.Context, deps PlatformDeps, input ClusterInventoryInput) (ToolResult, error) {
	return deployment.ClusterListInventory(ctx, deps, input)
}

func serviceListInventory(ctx context.Context, deps PlatformDeps, input ServiceInventoryInput) (ToolResult, error) {
	return deployment.ServiceListInventory(ctx, deps, input)
}

func userList(ctx context.Context, deps PlatformDeps, input UserListInput) (ToolResult, error) {
	return governance.UserList(ctx, deps, input)
}

func roleList(ctx context.Context, deps PlatformDeps, input RoleListInput) (ToolResult, error) {
	return governance.RoleList(ctx, deps, input)
}

func permissionCheck(ctx context.Context, deps PlatformDeps, input PermissionCheckInput) (ToolResult, error) {
	return governance.PermissionCheck(ctx, deps, input)
}

func topologyGet(ctx context.Context, deps PlatformDeps, input TopologyGetInput) (ToolResult, error) {
	return governance.TopologyGet(ctx, deps, input)
}

func auditLogSearch(ctx context.Context, deps PlatformDeps, input AuditLogSearchInput) (ToolResult, error) {
	return governance.AuditLogSearch(ctx, deps, input)
}

func credentialList(ctx context.Context, deps PlatformDeps, input CredentialListInput) (ToolResult, error) {
	return infrastructure.CredentialList(ctx, deps, input)
}

func credentialTest(ctx context.Context, deps PlatformDeps, input CredentialTestInput) (ToolResult, error) {
	return infrastructure.CredentialTest(ctx, deps, input)
}
