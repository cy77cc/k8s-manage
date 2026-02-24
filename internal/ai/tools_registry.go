package ai

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
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

func addLocalTool(tools *[]RegisteredTool, meta ToolMeta, fn func(ctx context.Context, input map[string]any) (ToolResult, error)) error {
	t, err := utils.InferTool[map[string]any, ToolResult](meta.Name, meta.Description, fn)
	if err != nil {
		return err
	}
	*tools = append(*tools, RegisteredTool{Meta: meta, Tool: t})
	return nil
}

func BuildLocalTools(deps PlatformDeps) ([]RegisteredTool, error) {
	tools := make([]RegisteredTool, 0, 13)

	if err := addLocalTool(&tools, ToolMeta{Name: "os.get_cpu_mem", Description: "读取 CPU/内存/负载概览", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return osGetCPUMem(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os.get_disk_fs", Description: "读取磁盘与文件系统占用", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return osGetDiskFS(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os.get_net_stat", Description: "读取网络连接与监听端口摘要", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return osGetNetStat(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os.get_process_top", Description: "读取高占用进程列表", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return osGetProcessTop(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os.get_journal_tail", Description: "按服务名读取系统日志窗口", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return osGetJournalTail(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "os.get_container_runtime", Description: "读取容器运行时摘要", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return osGetContainerRuntime(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "k8s.list_resources", Description: "列出 K8s 资源", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return k8sListResources(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "k8s.get_events", Description: "读取 K8s 事件", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return k8sGetEvents(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "k8s.get_pod_logs", Description: "读取 Pod 日志", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return k8sGetPodLogs(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "service.get_detail", Description: "查询服务详情", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return serviceGetDetail(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "service.deploy_preview", Description: "预览服务部署动作", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return serviceDeployPreview(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "service.deploy_apply", Description: "执行服务部署（需审批）", Mode: ToolModeMutating, Risk: ToolRiskHigh, Provider: "local", Permission: "ai:tool:execute"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return serviceDeployApply(ctx, deps, input)
	}); err != nil {
		return nil, err
	}
	if err := addLocalTool(&tools, ToolMeta{Name: "host.ssh_exec_readonly", Description: "远程只读命令执行", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read"}, func(ctx context.Context, input map[string]any) (ToolResult, error) {
		return hostSSHReadonly(ctx, deps, input)
	}); err != nil {
		return nil, err
	}

	return tools, nil
}
