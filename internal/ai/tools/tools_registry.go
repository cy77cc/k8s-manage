package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
)

type Registry struct {
	byName     map[string]RegisteredTool
	byDomain   map[ToolDomain][]RegisteredTool
	byCategory map[ToolCategory][]RegisteredTool
}

func NewRegistry(registered []RegisteredTool) *Registry {
	r := &Registry{
		byName:     make(map[string]RegisteredTool, len(registered)),
		byDomain:   make(map[ToolDomain][]RegisteredTool),
		byCategory: make(map[ToolCategory][]RegisteredTool),
	}
	for _, item := range registered {
		item.Meta = normalizeToolMeta(item.Meta)
		name := NormalizeToolName(item.Meta.Name)
		if name == "" {
			continue
		}
		r.byName[name] = item
		r.byDomain[item.Meta.Domain] = append(r.byDomain[item.Meta.Domain], item)
		r.byCategory[item.Meta.Category] = append(r.byCategory[item.Meta.Category], item)
	}
	return r
}

func (r *Registry) Get(name string) (RegisteredTool, bool) {
	if r == nil {
		return RegisteredTool{}, false
	}
	item, ok := r.byName[NormalizeToolName(name)]
	return item, ok
}

func (r *Registry) ByDomain(domain ToolDomain) []RegisteredTool {
	if r == nil {
		return nil
	}
	items := r.byDomain[domain]
	return append([]RegisteredTool(nil), items...)
}

func (r *Registry) ByCategory(category ToolCategory) []RegisteredTool {
	if r == nil {
		return nil
	}
	items := r.byCategory[category]
	return append([]RegisteredTool(nil), items...)
}

func registerTool(tools *[]RegisteredTool, meta ToolMeta, t tool.InvokableTool) {
	meta = normalizeToolMeta(meta)
	*tools = append(*tools, RegisteredTool{Meta: meta, Tool: t})
}

func BuildLocalTools(deps PlatformDeps) ([]RegisteredTool, error) {
	tools := make([]RegisteredTool, 0, 60)
	ctx := context.Background()

	// OS and Host tools
	registerTool(&tools, ToolMeta{
		Name:        "os_get_cpu_mem",
		Description: "Get CPU, memory and load average information from a target host.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, osGetCPUMem(ctx, deps, OSCPUMemInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "os_get_disk_fs",
		Description: "Get disk and filesystem usage information using 'df -h' command.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, osGetDiskFS(ctx, deps, OSDiskInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "os_get_net_stat",
		Description: "Get network statistics including device traffic and listening ports.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, osGetNetStat(ctx, deps, OSNetInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "os_get_process_top",
		Description: "Get top processes sorted by CPU usage.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, osGetProcessTop(ctx, deps, OSProcessTopInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "os_get_journal_tail",
		Description: "Get systemd journal logs for a specific service.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskMedium,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, osGetJournalTail(ctx, deps, OSJournalInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "os_get_container_runtime",
		Description: "Get container runtime information and running containers.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, osGetContainerRuntime(ctx, deps, OSContainerRuntimeInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "host_ssh_exec_readonly",
		Description: "Execute a readonly SSH command on a host.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskMedium,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, hostSSHReadonly(ctx, deps, HostSSHReadonlyInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "host_exec",
		Description: "Execute a readonly command on a single host via SSH.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskMedium,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, hostExec(ctx, deps, HostExecInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "host_list_inventory",
		Description: "Query host inventory list with detailed information.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, hostListInventory(ctx, deps, HostInventoryInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "host_batch",
		Description: "Execute a command on multiple hosts in batch.",
		Mode:        ToolModeMutating,
		Risk:        ToolRiskHigh,
		Provider:    "local",
		Permission:  "ai:tool:execute",
	}, hostBatch(ctx, deps, HostBatchInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "host_batch_exec_preview",
		Description: "Preview batch command execution before running.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskMedium,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, hostBatchExecPreview(ctx, deps, HostBatchExecPreviewInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "host_batch_exec_apply",
		Description: "Execute a command on multiple hosts after preview.",
		Mode:        ToolModeMutating,
		Risk:        ToolRiskHigh,
		Provider:    "local",
		Permission:  "ai:tool:execute",
	}, hostBatchExecApply(ctx, deps, HostBatchExecApplyInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "host_batch_status_update",
		Description: "Batch update host status to online/offline/maintenance.",
		Mode:        ToolModeMutating,
		Risk:        ToolRiskMedium,
		Provider:    "local",
		Permission:  "ai:tool:execute",
	}, hostBatchStatusUpdate(ctx, deps, HostBatchStatusInput{}))

	// Kubernetes tools
	registerTool(&tools, ToolMeta{
		Name:        "k8s_query",
		Description: "Query Kubernetes resources with filtering options.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, k8sQuery(ctx, deps, K8sQueryInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "k8s_list_resources",
		Description: "List Kubernetes resources of a specific type.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, k8sListResources(ctx, deps, K8sListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "k8s_events",
		Description: "Query Kubernetes events with optional filtering.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, k8sEvents(ctx, deps, K8sEventsQueryInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "k8s_get_events",
		Description: "Get Kubernetes events from a namespace.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, k8sGetEvents(ctx, deps, K8sEventsInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "k8s_logs",
		Description: "Get logs from a Kubernetes pod.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskMedium,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, k8sLogs(ctx, deps, K8sLogsInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "k8s_get_pod_logs",
		Description: "Get logs from a specific Kubernetes pod.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskMedium,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, k8sGetPodLogs(ctx, deps, K8sPodLogsInput{}))

	// Service tools
	registerTool(&tools, ToolMeta{
		Name:        "service_get_detail",
		Description: "Get detailed information about a specific service.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, serviceGetDetail(ctx, deps, ServiceDetailInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "service_status",
		Description: "Get current status and basic runtime information of a service.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, serviceStatus(ctx, deps, ServiceStatusInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "service_deploy_preview",
		Description: "Preview a service deployment without applying changes.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskMedium,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, serviceDeployPreview(ctx, deps, ServiceDeployPreviewInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "service_deploy_apply",
		Description: "Execute a service deployment to a target cluster.",
		Mode:        ToolModeMutating,
		Risk:        ToolRiskHigh,
		Provider:    "local",
		Permission:  "ai:tool:execute",
	}, serviceDeployApply(ctx, deps, ServiceDeployApplyInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "service_deploy",
		Description: "Unified service deployment tool supporting both preview and apply modes.",
		Mode:        ToolModeMutating,
		Risk:        ToolRiskHigh,
		Provider:    "local",
		Permission:  "ai:tool:execute",
	}, serviceDeploy(ctx, deps, ServiceDeployInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "service_catalog_list",
		Description: "Query the service catalog with filtering options.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, serviceCatalogList(ctx, deps, ServiceCatalogListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "service_category_tree",
		Description: "Get the service category tree structure.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, serviceCategoryTree(ctx, deps))

	registerTool(&tools, ToolMeta{
		Name:        "service_visibility_check",
		Description: "Check the visibility configuration of a service.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, serviceVisibilityCheck(ctx, deps, ServiceVisibilityCheckInput{}))

	// Monitor tools
	registerTool(&tools, ToolMeta{
		Name:        "monitor_alert_rule_list",
		Description: "Query the list of alert rules configured in the monitoring system.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, monitorAlertRuleList(ctx, deps, MonitorAlertRuleListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "monitor_alert",
		Description: "Query active/firing alert events from the monitoring system.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, monitorAlert(ctx, deps, MonitorAlertInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "monitor_alert_active",
		Description: "Query all active/firing alerts currently affecting the system.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, monitorAlertActive(ctx, deps, MonitorAlertActiveInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "monitor_metric",
		Description: "Query time-series metric data from the monitoring system.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, monitorMetric(ctx, deps, MonitorMetricInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "monitor_metric_query",
		Description: "Query metric data points over a time range for analysis.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, monitorMetricQuery(ctx, deps, MonitorMetricQueryInput{}))

	// CICD tools
	registerTool(&tools, ToolMeta{
		Name:        "cicd_pipeline_list",
		Description: "Query CI pipeline list with filtering options.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, cicdPipelineList(ctx, deps, CICDPipelineListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "cicd_pipeline_status",
		Description: "Query pipeline status with recent runs.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, cicdPipelineStatus(ctx, deps, CICDPipelineStatusInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "cicd_pipeline_trigger",
		Description: "Trigger pipeline build.",
		Mode:        ToolModeMutating,
		Risk:        ToolRiskHigh,
		Provider:    "local",
		Permission:  "ai:tool:execute",
	}, cicdPipelineTrigger(ctx, deps, CICDPipelineTriggerInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "job_list",
		Description: "Query job list with filtering options.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, jobList(ctx, deps, JobListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "job_execution_status",
		Description: "Query job execution status.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, jobExecutionStatus(ctx, deps, JobExecutionStatusInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "job_run",
		Description: "Manually trigger job execution.",
		Mode:        ToolModeMutating,
		Risk:        ToolRiskMedium,
		Provider:    "local",
		Permission:  "ai:tool:execute",
	}, jobRun(ctx, deps, JobRunInput{}))

	// Deployment tools
	registerTool(&tools, ToolMeta{
		Name:        "deployment_target_list",
		Description: "Query deployment target list with filtering options.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, deploymentTargetList(ctx, deps, DeploymentTargetListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "deployment_target_detail",
		Description: "Query deployment target detail.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, deploymentTargetDetail(ctx, deps, DeploymentTargetDetailInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "deployment_bootstrap_status",
		Description: "Query deployment target bootstrap status.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, deploymentBootstrapStatus(ctx, deps, DeploymentBootstrapStatusInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "config_app_list",
		Description: "Query config app list with filtering options.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, configAppList(ctx, deps, ConfigAppListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "config_item_get",
		Description: "Query config item value.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, configItemGet(ctx, deps, ConfigItemGetInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "config_diff",
		Description: "Compare config difference between environments.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, configDiff(ctx, deps, ConfigDiffInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "cluster_list_inventory",
		Description: "Query cluster inventory list with filtering options.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, clusterListInventory(ctx, deps, ClusterInventoryInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "service_list_inventory",
		Description: "Query service inventory list with filtering options.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, serviceListInventory(ctx, deps, ServiceInventoryInput{}))

	// Governance tools
	registerTool(&tools, ToolMeta{
		Name:        "user_list",
		Description: "Query the list of users in the platform.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, userList(ctx, deps, UserListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "role_list",
		Description: "Query the list of roles in the platform.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, roleList(ctx, deps, RoleListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "permission_check",
		Description: "Check if a user has a specific permission.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, permissionCheck(ctx, deps, PermissionCheckInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "topology_get",
		Description: "Query service topology showing relationships.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, topologyGet(ctx, deps, TopologyGetInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "audit_log_search",
		Description: "Search audit logs for platform activities.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, auditLogSearch(ctx, deps, AuditLogSearchInput{}))

	// Infrastructure tools
	registerTool(&tools, ToolMeta{
		Name:        "credential_list",
		Description: "Query cluster credential list for accessing Kubernetes clusters.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, credentialList(ctx, deps, CredentialListInput{}))

	registerTool(&tools, ToolMeta{
		Name:        "credential_test",
		Description: "Get credential connectivity test result.",
		Mode:        ToolModeReadonly,
		Risk:        ToolRiskLow,
		Provider:    "local",
		Permission:  "ai:tool:read",
	}, credentialTest(ctx, deps, CredentialTestInput{}))

	return tools, nil
}

func normalizeToolMeta(meta ToolMeta) ToolMeta {
	if meta.Domain == "" {
		meta.Domain = classifyToolDomain(meta.Name)
	}
	if meta.Category == "" {
		meta.Category = classifyToolCategory(meta)
	}
	return meta
}
