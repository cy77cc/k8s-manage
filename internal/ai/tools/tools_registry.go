package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
)

type PlatformDeps struct {
	DB        *gorm.DB
	Clientset *kubernetes.Clientset
}

type RegisteredTool struct {
	Meta ToolMeta
	Tool tool.InvokableTool
}

func addLocalTool[I any](tools *[]RegisteredTool, meta ToolMeta, fn func(ctx context.Context, input I) (ToolResult, error)) error {
	meta = normalizeToolMeta(meta)
	// 从函数参数中推断 ToolInfo
	info, err := utils.GoStruct2ToolInfo[I](meta.Name, meta.Description)
	if err != nil {
		return err
	}
	schemaMap, required := toolInfoSchema(info)
	meta.Schema = schemaMap
	if len(meta.Required) == 0 {
		meta.Required = required
	}
	meta = normalizeToolMeta(meta)
	t := utils.NewTool(info, fn)
	if t == nil {
		return fmt.Errorf("create tool %s failed", meta.Name)
	}
	*tools = append(*tools, RegisteredTool{Meta: meta, Tool: t})
	return nil
}

var defaultEnumSourceByField = map[string]string{
	"host_id":       "host_list_inventory",
	"cluster_id":    "cluster_list_inventory",
	"service_id":    "service_list_inventory",
	"target_id":     "deployment_target_list",
	"credential_id": "credential_list",
	"pipeline_id":   "cicd_pipeline_list",
	"job_id":        "job_list",
	"user_id":       "user_list",
	"app_id":        "config_app_list",
}

func normalizeToolMeta(meta ToolMeta) ToolMeta {
	if meta.EnumSources == nil {
		meta.EnumSources = map[string]string{}
	}
	if meta.ParamHints == nil {
		meta.ParamHints = map[string]string{}
	}
	for _, field := range meta.Required {
		source, exists := meta.EnumSources[field]
		if exists && strings.TrimSpace(source) != "" {
			continue
		}
		if mapped := strings.TrimSpace(defaultEnumSourceByField[field]); mapped != "" {
			meta.EnumSources[field] = mapped
		}
	}
	meta.Description = normalizeToolDescription(meta)
	return meta
}

func normalizeToolDescription(meta ToolMeta) string {
	base := strings.TrimSpace(meta.Description)
	if base == "" {
		base = "执行平台工具操作。"
	}
	if !strings.Contains(base, "。") {
		base += "。"
	}
	parts := []string{base}
	if len(meta.Required) > 0 {
		req := append([]string{}, meta.Required...)
		sort.Strings(req)
		parts = append(parts, fmt.Sprintf("必填参数: %s。", strings.Join(req, ", ")))
	}
	if len(meta.DefaultHint) > 0 {
		keys := make([]string, 0, len(meta.DefaultHint))
		for k := range meta.DefaultHint {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		defaultItems := make([]string, 0, len(keys))
		for _, k := range keys {
			defaultItems = append(defaultItems, fmt.Sprintf("%s=%v", k, meta.DefaultHint[k]))
		}
		parts = append(parts, fmt.Sprintf("默认值: %s。", strings.Join(defaultItems, ", ")))
	}
	if len(meta.EnumSources) > 0 {
		keys := make([]string, 0, len(meta.EnumSources))
		for k := range meta.EnumSources {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sourceItems := make([]string, 0, len(keys))
		for _, k := range keys {
			source := strings.TrimSpace(meta.EnumSources[k])
			if source == "" {
				continue
			}
			sourceItems = append(sourceItems, fmt.Sprintf("%s 可从 %s 获取", k, source))
		}
		if len(sourceItems) > 0 {
			parts = append(parts, fmt.Sprintf("参数来源: %s。", strings.Join(sourceItems, "；")))
		}
	}
	if len(meta.Examples) > 0 {
		parts = append(parts, fmt.Sprintf("示例: %s。", strings.TrimSpace(meta.Examples[0])))
	}
	return strings.Join(parts, " ")
}

func BuildLocalTools(deps PlatformDeps) ([]RegisteredTool, error) {
	tools := make([]RegisteredTool, 0, 24)

	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "os_get_cpu_mem",
			Description: "读取 CPU/内存/负载概览。默认 target=localhost。示例: {\"target\":\"10.0.0.8\"}",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"target": "localhost"},
		},
		func(ctx context.Context, input OSCPUMemInput) (ToolResult, error) {
			return osGetCPUMem(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "os_get_disk_fs",
			Description: "读取磁盘与文件系统占用。默认 target=localhost；target 支持主机 ID/IP/主机名（如 香港云服务器）。", Mode: ToolModeReadonly,
			Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"target": "localhost"}},
		func(ctx context.Context, input OSDiskInput) (ToolResult, error) {
			return osGetDiskFS(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "os_get_net_stat",
			Description: "读取网络连接与监听端口摘要。默认 target=localhost。", Mode: ToolModeReadonly,
			Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"target": "localhost"}},
		func(ctx context.Context, input OSNetInput) (ToolResult, error) {
			return osGetNetStat(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "os_get_process_top",
			Description: "读取高占用进程列表。limit 默认 10。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"target": "localhost", "limit": 10},
		},
		func(ctx context.Context, input OSProcessTopInput) (ToolResult, error) {
			return osGetProcessTop(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "os_get_journal_tail",
			Description: "按服务名读取系统日志窗口。service 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskMedium,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"service"},
			DefaultHint: map[string]any{"target": "localhost", "lines": 200},
		},
		func(ctx context.Context, input OSJournalInput) (ToolResult, error) {
			return osGetJournalTail(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "os_get_container_runtime",
			Description: "读取容器运行时摘要。默认 target=localhost。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"target": "localhost"},
		},
		func(ctx context.Context, input OSContainerRuntimeInput) (ToolResult, error) {
			return osGetContainerRuntime(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "k8s_list_resources",
			Description: "列出 K8s 资源。resource 必填，可选 pods/services/deployments/nodes。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"resource"},
			DefaultHint: map[string]any{"namespace": "default", "limit": 50},
		},
		func(ctx context.Context, input K8sListInput) (ToolResult, error) {
			return k8sListResources(ctx, deps, input)
		}); err != nil {
		return nil, err
	}

	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "k8s_get_events",
			Description: "读取 K8s 事件，namespace 默认 default。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"namespace": "default", "limit": 50},
		},
		func(ctx context.Context, input K8sEventsInput) (ToolResult, error) {
			return k8sGetEvents(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "k8s_get_pod_logs",
			Description: "读取 Pod 日志，pod 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskMedium,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"pod"},
			DefaultHint: map[string]any{"namespace": "default", "tail_lines": 200},
		},
		func(ctx context.Context, input K8sPodLogsInput) (ToolResult, error) {
			return k8sGetPodLogs(ctx, deps, input)
		}); err != nil {
		return nil, err
	}

	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "service_get_detail",
			Description: "查询服务详情，service_id 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"service_id"},
		},
		func(ctx context.Context, input ServiceDetailInput) (ToolResult, error) {
			return serviceGetDetail(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "service_deploy_preview",
			Description: "预览服务部署动作，service_id/cluster_id 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskMedium,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"service_id", "cluster_id"},
		},
		func(ctx context.Context, input ServiceDeployPreviewInput) (ToolResult, error) {
			return serviceDeployPreview(ctx, deps, input)
		}); err != nil {
		return nil, err
	}

	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "service_deploy_apply",
			Description: "执行服务部署（需审批），service_id/cluster_id 必填。",
			Mode:        ToolModeMutating,
			Risk:        ToolRiskHigh,
			Provider:    "local",
			Permission:  "ai:tool:execute",
			Required:    []string{"service_id", "cluster_id"},
		},
		func(ctx context.Context, input ServiceDeployApplyInput) (ToolResult, error) {
			return serviceDeployApply(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "service_catalog_list",
			Description: "查询服务目录，支持 category_id/keyword 过滤。示例: {\"category_id\":2,\"keyword\":\"payment\"}。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input ServiceCatalogListInput) (ToolResult, error) {
			return serviceCatalogList(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "service_category_tree",
			Description: "查询服务分类树。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, _ struct{}) (ToolResult, error) {
			return serviceCategoryTree(ctx, deps)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "service_visibility_check",
			Description: "查询服务可见性配置，service_id 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"service_id"},
		},
		func(ctx context.Context, input ServiceVisibilityCheckInput) (ToolResult, error) {
			return serviceVisibilityCheck(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "deployment_target_list",
			Description: "查询部署目标列表，支持 env/status/keyword 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input DeploymentTargetListInput) (ToolResult, error) {
			return deploymentTargetList(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "deployment_target_detail",
			Description: "查询部署目标详情，target_id 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"target_id"},
		},
		func(ctx context.Context, input DeploymentTargetDetailInput) (ToolResult, error) {
			return deploymentTargetDetail(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "deployment_bootstrap_status",
			Description: "查询部署目标引导状态，target_id 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"target_id"},
		},
		func(ctx context.Context, input DeploymentBootstrapStatusInput) (ToolResult, error) {
			return deploymentBootstrapStatus(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "credential_list",
			Description: "查询凭证列表，支持 type/keyword 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input CredentialListInput) (ToolResult, error) {
			return credentialList(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "credential_test",
			Description: "查询凭证连通性测试结果，credential_id 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"credential_id"},
		},
		func(ctx context.Context, input CredentialTestInput) (ToolResult, error) {
			return credentialTest(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "cicd_pipeline_list",
			Description: "查询流水线列表，支持 status/keyword 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input CICDPipelineListInput) (ToolResult, error) {
			return cicdPipelineList(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "cicd_pipeline_status",
			Description: "查询流水线状态，pipeline_id 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"pipeline_id"},
		},
		func(ctx context.Context, input CICDPipelineStatusInput) (ToolResult, error) {
			return cicdPipelineStatus(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "cicd_pipeline_trigger",
			Description: "触发流水线构建，pipeline_id/branch 必填。",
			Mode:        ToolModeMutating,
			Risk:        ToolRiskHigh,
			Provider:    "local",
			Permission:  "ai:tool:execute",
			Required:    []string{"pipeline_id", "branch"},
		},
		func(ctx context.Context, input CICDPipelineTriggerInput) (ToolResult, error) {
			return cicdPipelineTrigger(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "job_list",
			Description: "查询任务列表，支持 status/keyword 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input JobListInput) (ToolResult, error) {
			return jobList(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "job_execution_status",
			Description: "查询任务执行状态，job_id 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"job_id"},
		},
		func(ctx context.Context, input JobExecutionStatusInput) (ToolResult, error) {
			return jobExecutionStatus(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "job_run",
			Description: "手动触发任务，job_id 必填。",
			Mode:        ToolModeMutating,
			Risk:        ToolRiskMedium,
			Provider:    "local",
			Permission:  "ai:tool:execute",
			Required:    []string{"job_id"},
		},
		func(ctx context.Context, input JobRunInput) (ToolResult, error) {
			return jobRun(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "config_app_list",
			Description: "查询配置应用列表，支持 keyword/env 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input ConfigAppListInput) (ToolResult, error) {
			return configAppList(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "config_item_get",
			Description: "查询配置项，app_id/key 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"app_id", "key"},
		},
		func(ctx context.Context, input ConfigItemGetInput) (ToolResult, error) {
			return configItemGet(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "config_diff",
			Description: "对比配置差异，app_id/env_a/env_b 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"app_id", "env_a", "env_b"},
		},
		func(ctx context.Context, input ConfigDiffInput) (ToolResult, error) {
			return configDiff(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "monitor_alert_rule_list",
			Description: "查询告警规则列表，支持 status/keyword 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input MonitorAlertRuleListInput) (ToolResult, error) {
			return monitorAlertRuleList(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "monitor_alert_active",
			Description: "查询活跃告警，支持 severity/service_id 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input MonitorAlertActiveInput) (ToolResult, error) {
			return monitorAlertActive(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "monitor_metric_query",
			Description: "查询指标数据，query 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"query"},
		},
		func(ctx context.Context, input MonitorMetricQueryInput) (ToolResult, error) {
			return monitorMetricQuery(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "topology_get",
			Description: "查询服务拓扑，支持 service_id/depth。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input TopologyGetInput) (ToolResult, error) {
			return topologyGet(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "audit_log_search",
			Description: "查询审计日志，支持 time_range/resource_type/action/user_id。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input AuditLogSearchInput) (ToolResult, error) {
			return auditLogSearch(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "user_list",
			Description: "查询用户列表，支持 keyword/status 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input UserListInput) (ToolResult, error) {
			return userList(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "role_list",
			Description: "查询角色列表，支持 keyword 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
		},
		func(ctx context.Context, input RoleListInput) (ToolResult, error) {
			return roleList(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "permission_check",
			Description: "检查权限，user_id/resource/action 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"user_id", "resource", "action"},
		},
		func(ctx context.Context, input PermissionCheckInput) (ToolResult, error) {
			return permissionCheck(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "host_ssh_exec_readonly",
			Description: "远程只读命令执行，host_id/command 必填。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskMedium,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"host_id", "command"},
		},
		func(ctx context.Context, input HostSSHReadonlyInput) (ToolResult, error) {
			return hostSSHReadonly(ctx, deps, input)
		}); err != nil {
		return nil, err
	}

	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "host_list_inventory",
			Description: "查询主机资产清单，可按 status/keyword 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
		},
		func(ctx context.Context, input HostInventoryInput) (ToolResult, error) {
			return hostListInventory(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "cluster_list_inventory",
			Description: "查询集群资产清单，可按 status/keyword 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
		},
		func(ctx context.Context, input ClusterInventoryInput) (ToolResult, error) {
			return clusterListInventory(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "service_list_inventory",
			Description: "查询服务资产清单，可按 runtime_type/env/status/keyword 过滤。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskLow,
			Provider:    "local",
			Permission:  "ai:tool:read",
			DefaultHint: map[string]any{"limit": 50},
		},
		func(ctx context.Context, input ServiceInventoryInput) (ToolResult, error) {
			return serviceListInventory(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "host_batch_exec_preview",
			Description: "批量命令执行预检查，返回目标主机与风险评估。",
			Mode:        ToolModeReadonly,
			Risk:        ToolRiskMedium,
			Provider:    "local",
			Permission:  "ai:tool:read",
			Required:    []string{"host_ids", "command"},
		},
		func(ctx context.Context, input HostBatchExecPreviewInput) (ToolResult, error) {
			return hostBatchExecPreview(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "host_batch_exec_apply",
			Description: "批量执行主机命令（需审批），禁止危险命令。",
			Mode:        ToolModeMutating,
			Risk:        ToolRiskHigh,
			Provider:    "local",
			Permission:  "ai:tool:execute",
			Required:    []string{"host_ids", "command"},
		},
		func(ctx context.Context, input HostBatchExecApplyInput) (ToolResult, error) {
			return hostBatchExecApply(ctx, deps, input)
		}); err != nil {
		return nil, err
	}
	if err := addLocalTool(
		&tools,
		ToolMeta{
			Name:        "host_batch_status_update",
			Description: "批量更新主机状态（online/offline/maintenance，需审批）。",
			Mode:        ToolModeMutating,
			Risk:        ToolRiskMedium,
			Provider:    "local",
			Permission:  "ai:tool:execute",
			Required:    []string{"host_ids", "action"},
		},
		func(ctx context.Context, input HostBatchStatusInput) (ToolResult, error) {
			return hostBatchStatusUpdate(ctx, deps, input)
		}); err != nil {
		return nil, err
	}

	return tools, nil
}

func toolInfoSchema(info *schema.ToolInfo) (map[string]any, []string) {
	if info == nil || info.ParamsOneOf == nil {
		return nil, nil
	}
	js, err := info.ParamsOneOf.ToJSONSchema()
	if err != nil || js == nil {
		return nil, nil
	}
	raw, err := json.Marshal(js)
	if err != nil {
		return nil, nil
	}
	m := map[string]any{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, nil
	}
	req := make([]string, 0)
	if list, ok := m["required"].([]any); ok {
		for _, item := range list {
			s := fmt.Sprintf("%v", item)
			if s != "" {
				req = append(req, s)
			}
		}
	}
	return m, req
}
