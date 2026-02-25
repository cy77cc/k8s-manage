package ai

import (
	"context"
	"encoding/json"
	"fmt"

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
	info, err := utils.GoStruct2ToolInfo[I](meta.Name, meta.Description)
	if err != nil {
		return err
	}
	schemaMap, required := toolInfoSchema(info)
	meta.Schema = schemaMap
	if len(meta.Required) == 0 {
		meta.Required = required
	}
	t := utils.NewTool[I, ToolResult](info, fn)
	if t == nil {
		return fmt.Errorf("create tool %s failed", meta.Name)
	}
	*tools = append(*tools, RegisteredTool{Meta: meta, Tool: t})
	return nil
}

func BuildLocalTools(deps PlatformDeps) ([]RegisteredTool, error) {
	tools := make([]RegisteredTool, 0, 24)

	if err := addLocalTool(&tools, ToolMeta{Name: "os_get_cpu_mem", Description: "读取 CPU/内存/负载概览。默认 target=localhost。示例: {\"target\":\"10.0.0.8\"}", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"target": "localhost"}}, func(ctx context.Context, input OSCPUMemInput) (ToolResult, error) {
		return osGetCPUMem(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os_get_disk_fs", Description: "读取磁盘与文件系统占用。默认 target=localhost。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"target": "localhost"}}, func(ctx context.Context, input OSDiskInput) (ToolResult, error) {
		return osGetDiskFS(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os_get_net_stat", Description: "读取网络连接与监听端口摘要。默认 target=localhost。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"target": "localhost"}}, func(ctx context.Context, input OSNetInput) (ToolResult, error) {
		return osGetNetStat(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os_get_process_top", Description: "读取高占用进程列表。limit 默认 10。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"target": "localhost", "limit": 10}}, func(ctx context.Context, input OSProcessTopInput) (ToolResult, error) {
		return osGetProcessTop(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os_get_journal_tail", Description: "按服务名读取系统日志窗口。service 必填。", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read", Required: []string{"service"}, DefaultHint: map[string]any{"target": "localhost", "lines": 200}}, func(ctx context.Context, input OSJournalInput) (ToolResult, error) {
		return osGetJournalTail(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os_get_container_runtime", Description: "读取容器运行时摘要。默认 target=localhost。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"target": "localhost"}}, func(ctx context.Context, input OSContainerRuntimeInput) (ToolResult, error) {
		return osGetContainerRuntime(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "k8s_list_resources", Description: "列出 K8s 资源。resource 必填，可选 pods/services/deployments/nodes。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", Required: []string{"resource"}, DefaultHint: map[string]any{"namespace": "default", "limit": 50}}, func(ctx context.Context, input K8sListInput) (ToolResult, error) {
		return k8sListResources(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "k8s_get_events", Description: "读取 K8s 事件，namespace 默认 default。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"namespace": "default", "limit": 50}}, func(ctx context.Context, input K8sEventsInput) (ToolResult, error) {
		return k8sGetEvents(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "k8s_get_pod_logs", Description: "读取 Pod 日志，pod 必填。", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read", Required: []string{"pod"}, DefaultHint: map[string]any{"namespace": "default", "tail_lines": 200}}, func(ctx context.Context, input K8sPodLogsInput) (ToolResult, error) {
		return k8sGetPodLogs(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "service_get_detail", Description: "查询服务详情，service_id 必填。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", Required: []string{"service_id"}}, func(ctx context.Context, input ServiceDetailInput) (ToolResult, error) {
		return serviceGetDetail(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "service_deploy_preview", Description: "预览服务部署动作，service_id/cluster_id 必填。", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read", Required: []string{"service_id", "cluster_id"}}, func(ctx context.Context, input ServiceDeployPreviewInput) (ToolResult, error) {
		return serviceDeployPreview(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "service_deploy_apply", Description: "执行服务部署（需审批），service_id/cluster_id 必填。", Mode: ToolModeMutating, Risk: ToolRiskHigh, Provider: "local", Permission: "ai:tool:execute", Required: []string{"service_id", "cluster_id"}}, func(ctx context.Context, input ServiceDeployApplyInput) (ToolResult, error) {
		return serviceDeployApply(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "host_ssh_exec_readonly", Description: "远程只读命令执行，host_id/command 必填。", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read", Required: []string{"host_id", "command"}}, func(ctx context.Context, input HostSSHReadonlyInput) (ToolResult, error) {
		return hostSSHReadonly(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "host_list_inventory", Description: "查询主机资产清单，可按 status/keyword 过滤。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"limit": 50}}, func(ctx context.Context, input HostInventoryInput) (ToolResult, error) {
		return hostListInventory(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "cluster_list_inventory", Description: "查询集群资产清单，可按 status/keyword 过滤。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"limit": 50}}, func(ctx context.Context, input ClusterInventoryInput) (ToolResult, error) {
		return clusterListInventory(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "service_list_inventory", Description: "查询服务资产清单，可按 runtime_type/env/status/keyword 过滤。", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read", DefaultHint: map[string]any{"limit": 50}}, func(ctx context.Context, input ServiceInventoryInput) (ToolResult, error) {
		return serviceListInventory(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "host_batch_exec_preview", Description: "批量命令执行预检查，返回目标主机与风险评估。", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read", Required: []string{"host_ids", "command"}}, func(ctx context.Context, input HostBatchExecPreviewInput) (ToolResult, error) {
		return hostBatchExecPreview(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "host_batch_exec_apply", Description: "批量执行主机命令（需审批），禁止危险命令。", Mode: ToolModeMutating, Risk: ToolRiskHigh, Provider: "local", Permission: "ai:tool:execute", Required: []string{"host_ids", "command"}}, func(ctx context.Context, input HostBatchExecApplyInput) (ToolResult, error) {
		return hostBatchExecApply(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "host_batch_status_update", Description: "批量更新主机状态（online/offline/maintenance，需审批）。", Mode: ToolModeMutating, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:execute", Required: []string{"host_ids", "action"}}, func(ctx context.Context, input HostBatchStatusInput) (ToolResult, error) {
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
