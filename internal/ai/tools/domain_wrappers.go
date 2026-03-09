package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/cicd"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/deployment"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/governance"
	hostimpl "github.com/cy77cc/k8s-manage/internal/ai/tools/impl/host"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/infrastructure"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/kubernetes"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/impl/monitor"
	serviceimpl "github.com/cy77cc/k8s-manage/internal/ai/tools/impl/service"
)

func osGetCPUMem(ctx context.Context, deps PlatformDeps, input OSCPUMemInput) tool.InvokableTool {
	return hostimpl.OSGetCPUMem(ctx, deps, input)
}

func osGetDiskFS(ctx context.Context, deps PlatformDeps, input OSDiskInput) tool.InvokableTool {
	return hostimpl.OSGetDiskFS(ctx, deps, input)
}

func osGetNetStat(ctx context.Context, deps PlatformDeps, input OSNetInput) tool.InvokableTool {
	return hostimpl.OSGetNetStat(ctx, deps, input)
}

func osGetProcessTop(ctx context.Context, deps PlatformDeps, input OSProcessTopInput) tool.InvokableTool {
	return hostimpl.OSGetProcessTop(ctx, deps, input)
}

func osGetJournalTail(ctx context.Context, deps PlatformDeps, input OSJournalInput) tool.InvokableTool {
	return hostimpl.OSGetJournalTail(ctx, deps, input)
}

func osGetContainerRuntime(ctx context.Context, deps PlatformDeps, input OSContainerRuntimeInput) tool.InvokableTool {
	return hostimpl.OSGetContainerRuntime(ctx, deps, input)
}

func hostSSHReadonly(ctx context.Context, deps PlatformDeps, input HostSSHReadonlyInput) tool.InvokableTool {
	return hostimpl.HostSSHReadonly(ctx, deps, input)
}

func hostExec(ctx context.Context, deps PlatformDeps, input HostExecInput) tool.InvokableTool {
	return hostimpl.HostExec(ctx, deps, input)
}

func hostListInventory(ctx context.Context, deps PlatformDeps, input HostInventoryInput) tool.InvokableTool {
	return hostimpl.HostListInventory(ctx, deps, input)
}

func hostBatch(ctx context.Context, deps PlatformDeps, input HostBatchInput) tool.InvokableTool {
	return hostimpl.HostBatch(ctx, deps, input)
}

func hostBatchExecPreview(ctx context.Context, deps PlatformDeps, input HostBatchExecPreviewInput) tool.InvokableTool {
	return hostimpl.HostBatchExecPreview(ctx, deps, input)
}

func hostBatchExecApply(ctx context.Context, deps PlatformDeps, input HostBatchExecApplyInput) tool.InvokableTool {
	return hostimpl.HostBatchExecApply(ctx, deps, input)
}

func hostBatchStatusUpdate(ctx context.Context, deps PlatformDeps, input HostBatchStatusInput) tool.InvokableTool {
	return hostimpl.HostBatchStatusUpdate(ctx, deps, input)
}

func k8sQuery(ctx context.Context, deps PlatformDeps, input K8sQueryInput) tool.InvokableTool {
	return kubernetes.K8sQuery(ctx, deps, input)
}

func k8sListResources(ctx context.Context, deps PlatformDeps, input K8sListInput) tool.InvokableTool {
	return kubernetes.K8sListResources(ctx, deps, input)
}

func k8sEvents(ctx context.Context, deps PlatformDeps, input K8sEventsQueryInput) tool.InvokableTool {
	return kubernetes.K8sEvents(ctx, deps, input)
}

func k8sGetEvents(ctx context.Context, deps PlatformDeps, input K8sEventsInput) tool.InvokableTool {
	return kubernetes.K8sGetEvents(ctx, deps, input)
}

func k8sLogs(ctx context.Context, deps PlatformDeps, input K8sLogsInput) tool.InvokableTool {
	return kubernetes.K8sLogs(ctx, deps, input)
}

func k8sGetPodLogs(ctx context.Context, deps PlatformDeps, input K8sPodLogsInput) tool.InvokableTool {
	return kubernetes.K8sGetPodLogs(ctx, deps, input)
}

func serviceGetDetail(ctx context.Context, deps PlatformDeps, input ServiceDetailInput) tool.InvokableTool {
	return serviceimpl.ServiceGetDetail(ctx, deps, input)
}

func serviceStatus(ctx context.Context, deps PlatformDeps, input ServiceStatusInput) tool.InvokableTool {
	return serviceimpl.ServiceStatus(ctx, deps, input)
}

func serviceDeployPreview(ctx context.Context, deps PlatformDeps, input ServiceDeployPreviewInput) tool.InvokableTool {
	return serviceimpl.ServiceDeployPreview(ctx, deps, input)
}

func serviceDeployApply(ctx context.Context, deps PlatformDeps, input ServiceDeployApplyInput) tool.InvokableTool {
	return serviceimpl.ServiceDeployApply(ctx, deps, input)
}

func serviceDeploy(ctx context.Context, deps PlatformDeps, input ServiceDeployInput) tool.InvokableTool {
	return serviceimpl.ServiceDeploy(ctx, deps, input)
}

func serviceCatalogList(ctx context.Context, deps PlatformDeps, input ServiceCatalogListInput) tool.InvokableTool {
	return serviceimpl.ServiceCatalogList(ctx, deps, input)
}

func serviceVisibilityCheck(ctx context.Context, deps PlatformDeps, input ServiceVisibilityCheckInput) tool.InvokableTool {
	return serviceimpl.ServiceVisibilityCheck(ctx, deps, input)
}

func serviceCategoryTree(ctx context.Context, deps PlatformDeps) tool.InvokableTool {
	return serviceimpl.ServiceCategoryTree(ctx, deps)
}

func monitorAlertRuleList(ctx context.Context, deps PlatformDeps, input MonitorAlertRuleListInput) tool.InvokableTool {
	return monitor.MonitorAlertRuleList(ctx, deps, input)
}

func monitorAlert(ctx context.Context, deps PlatformDeps, input MonitorAlertInput) tool.InvokableTool {
	return monitor.MonitorAlert(ctx, deps, input)
}

func monitorAlertActive(ctx context.Context, deps PlatformDeps, input MonitorAlertActiveInput) tool.InvokableTool {
	return monitor.MonitorAlertActive(ctx, deps, input)
}

func monitorMetric(ctx context.Context, deps PlatformDeps, input MonitorMetricInput) tool.InvokableTool {
	return monitor.MonitorMetric(ctx, deps, input)
}

func monitorMetricQuery(ctx context.Context, deps PlatformDeps, input MonitorMetricQueryInput) tool.InvokableTool {
	return monitor.MonitorMetricQuery(ctx, deps, input)
}

func cicdPipelineList(ctx context.Context, deps PlatformDeps, input CICDPipelineListInput) tool.InvokableTool {
	return cicd.CICDPipelineList(ctx, deps, input)
}

func cicdPipelineStatus(ctx context.Context, deps PlatformDeps, input CICDPipelineStatusInput) tool.InvokableTool {
	return cicd.CICDPipelineStatus(ctx, deps, input)
}

func cicdPipelineTrigger(ctx context.Context, deps PlatformDeps, input CICDPipelineTriggerInput) tool.InvokableTool {
	return cicd.CICDPipelineTrigger(ctx, deps, input)
}

func jobList(ctx context.Context, deps PlatformDeps, input JobListInput) tool.InvokableTool {
	return cicd.JobList(ctx, deps, input)
}

func jobExecutionStatus(ctx context.Context, deps PlatformDeps, input JobExecutionStatusInput) tool.InvokableTool {
	return cicd.JobExecutionStatus(ctx, deps, input)
}

func jobRun(ctx context.Context, deps PlatformDeps, input JobRunInput) tool.InvokableTool {
	return cicd.JobRun(ctx, deps, input)
}

func deploymentTargetList(ctx context.Context, deps PlatformDeps, input DeploymentTargetListInput) tool.InvokableTool {
	return deployment.DeploymentTargetList(ctx, deps, input)
}

func deploymentTargetDetail(ctx context.Context, deps PlatformDeps, input DeploymentTargetDetailInput) tool.InvokableTool {
	return deployment.DeploymentTargetDetail(ctx, deps, input)
}

func deploymentBootstrapStatus(ctx context.Context, deps PlatformDeps, input DeploymentBootstrapStatusInput) tool.InvokableTool {
	return deployment.DeploymentBootstrapStatus(ctx, deps, input)
}

func configAppList(ctx context.Context, deps PlatformDeps, input ConfigAppListInput) tool.InvokableTool {
	return deployment.ConfigAppList(ctx, deps, input)
}

func configItemGet(ctx context.Context, deps PlatformDeps, input ConfigItemGetInput) tool.InvokableTool {
	return deployment.ConfigItemGet(ctx, deps, input)
}

func configDiff(ctx context.Context, deps PlatformDeps, input ConfigDiffInput) tool.InvokableTool {
	return deployment.ConfigDiff(ctx, deps, input)
}

func clusterListInventory(ctx context.Context, deps PlatformDeps, input ClusterInventoryInput) tool.InvokableTool {
	return deployment.ClusterListInventory(ctx, deps, input)
}

func serviceListInventory(ctx context.Context, deps PlatformDeps, input ServiceInventoryInput) tool.InvokableTool {
	return deployment.ServiceListInventory(ctx, deps, input)
}

func userList(ctx context.Context, deps PlatformDeps, input UserListInput) tool.InvokableTool {
	return governance.UserList(ctx, deps, input)
}

func roleList(ctx context.Context, deps PlatformDeps, input RoleListInput) tool.InvokableTool {
	return governance.RoleList(ctx, deps, input)
}

func permissionCheck(ctx context.Context, deps PlatformDeps, input PermissionCheckInput) tool.InvokableTool {
	return governance.PermissionCheck(ctx, deps, input)
}

func topologyGet(ctx context.Context, deps PlatformDeps, input TopologyGetInput) tool.InvokableTool {
	return governance.TopologyGet(ctx, deps, input)
}

func auditLogSearch(ctx context.Context, deps PlatformDeps, input AuditLogSearchInput) tool.InvokableTool {
	return governance.AuditLogSearch(ctx, deps, input)
}

func credentialList(ctx context.Context, deps PlatformDeps, input CredentialListInput) tool.InvokableTool {
	return infrastructure.CredentialList(ctx, deps, input)
}

func credentialTest(ctx context.Context, deps PlatformDeps, input CredentialTestInput) tool.InvokableTool {
	return infrastructure.CredentialTest(ctx, deps, input)
}
