// Package tools 提供工具注册、构建和执行的核心功能。
// 包含工具元数据定义、注册表、工具包装器（审批/审核）以及各领域工具的实现。
package tools

import (
	"context"
	"strings"

	core "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
)

// ToolMode 是工具模式类型别名。
type ToolMode = core.ToolMode

// 工具模式常量。
const (
	// ToolModeReadonly 只读模式，不修改系统状态。
	ToolModeReadonly = core.ToolModeReadonly
	// ToolModeMutating 变更模式，会修改系统状态。
	ToolModeMutating = core.ToolModeMutating
)

// ToolRisk 是工具风险级别类型别名。
type ToolRisk = core.ToolRisk

// ToolDomain 是工具领域类型别名。
type ToolDomain = core.ToolDomain

// ToolCategory 是工具类别类型别名。
type ToolCategory = core.ToolCategory

// 风险级别常量。
const (
	// ToolRiskLow 低风险，无需审批。
	ToolRiskLow = core.ToolRiskLow
	// ToolRiskMedium 中风险，需要确认。
	ToolRiskMedium = core.ToolRiskMedium
	// ToolRiskHigh 高风险，需要审批。
	ToolRiskHigh = core.ToolRiskHigh
)

// 工具领域常量。
const (
	DomainGeneral        = core.DomainGeneral        // 通用领域
	DomainInfrastructure = core.DomainInfrastructure // 基础设施领域
	DomainService        = core.DomainService        // 服务领域
	DomainCICD           = core.DomainCICD           // CI/CD 领域
	DomainMonitor        = core.DomainMonitor        // 监控领域
	DomainConfig         = core.DomainConfig         // 配置领域
	DomainUser           = core.DomainUser           // 用户管理领域
)

// 工具类别常量。
const (
	CategoryDiscovery = core.CategoryDiscovery // 发现类（只读查询）
	CategoryAction    = core.CategoryAction    // 操作类（变更操作）
)

// 核心类型别名。
type ToolResult = core.ToolResult                     // 工具执行结果
type ToolExecutionError = core.ToolExecutionError     // 工具执行错误
type ToolMeta = core.ToolMeta                         // 工具元信息
type ApprovalRequiredError = core.ApprovalRequiredError // 需要审批错误
type ConfirmationRequiredError = core.ConfirmationRequiredError // 需要确认错误
type ToolPolicyChecker = core.ToolPolicyChecker       // 工具策略检查器
type ToolEventEmitter = core.ToolEventEmitter         // 工具事件发射器
type ToolMemoryAccessor = core.ToolMemoryAccessor     // 工具内存访问器
type PlatformDeps = core.PlatformDeps                 // 平台依赖项
type RegisteredTool = core.RegisteredTool             // 已注册工具

type ToolInputError = core.ToolInputError // 工具输入错误

// 工具输入类型别名。
type OSCPUMemInput = core.OSCPUMemInput                     // OS CPU/内存输入
type OSDiskInput = core.OSDiskInput                         // OS 磁盘输入
type OSNetInput = core.OSNetInput                           // OS 网络输入
type OSProcessTopInput = core.OSProcessTopInput             // OS 进程 Top 输入
type OSJournalInput = core.OSJournalInput                   // OS 日志输入
type OSContainerRuntimeInput = core.OSContainerRuntimeInput // 容器运行时输入
type K8sListInput = core.K8sListInput                       // K8s 列表输入
type K8sQueryInput = core.K8sQueryInput                     // K8s 查询输入
type K8sEventsInput = core.K8sEventsInput                   // K8s 事件输入
type K8sEventsQueryInput = core.K8sEventsQueryInput         // K8s 事件查询输入
type K8sPodLogsInput = core.K8sPodLogsInput                 // K8s Pod 日志输入
type K8sLogsInput = core.K8sLogsInput                       // K8s 日志输入
type ServiceDetailInput = core.ServiceDetailInput           // 服务详情输入
type ServiceDeployPreviewInput = core.ServiceDeployPreviewInput // 服务部署预览输入
type ServiceDeployApplyInput = core.ServiceDeployApplyInput     // 服务部署应用输入
type ServiceDeployInput = core.ServiceDeployInput           // 服务部署输入
type ServiceStatusInput = core.ServiceStatusInput           // 服务状态输入
type HostSSHReadonlyInput = core.HostSSHReadonlyInput       // 主机 SSH 只读输入
type HostExecInput = core.HostExecInput                     // 主机执行输入
type HostInventoryInput = core.HostInventoryInput           // 主机资产输入
type ClusterInventoryInput = core.ClusterInventoryInput     // 集群资产输入
type ServiceInventoryInput = core.ServiceInventoryInput     // 服务资产输入
type HostBatchExecPreviewInput = core.HostBatchExecPreviewInput // 主机批量执行预览输入
type HostBatchExecApplyInput = core.HostBatchExecApplyInput     // 主机批量执行应用输入
type HostBatchInput = core.HostBatchInput                   // 主机批量输入
type HostBatchStatusInput = core.HostBatchStatusInput       // 主机批量状态输入
type ServiceCatalogListInput = core.ServiceCatalogListInput // 服务目录列表输入
type ServiceVisibilityCheckInput = core.ServiceVisibilityCheckInput // 服务可见性检查输入
type DeploymentTargetListInput = core.DeploymentTargetListInput     // 部署目标列表输入
type DeploymentTargetDetailInput = core.DeploymentTargetDetailInput // 部署目标详情输入
type DeploymentBootstrapStatusInput = core.DeploymentBootstrapStatusInput // 部署引导状态输入
type CredentialListInput = core.CredentialListInput         // 凭证列表输入
type CredentialTestInput = core.CredentialTestInput         // 凭证测试输入
type CICDPipelineListInput = core.CICDPipelineListInput     // CI/CD 流水线列表输入
type CICDPipelineStatusInput = core.CICDPipelineStatusInput // CI/CD 流水线状态输入
type CICDPipelineTriggerInput = core.CICDPipelineTriggerInput // CI/CD 流水线触发输入
type JobListInput = core.JobListInput                       // 任务列表输入
type JobExecutionStatusInput = core.JobExecutionStatusInput // 任务执行状态输入
type JobRunInput = core.JobRunInput                         // 任务运行输入
type ConfigAppListInput = core.ConfigAppListInput           // 配置应用列表输入
type ConfigItemGetInput = core.ConfigItemGetInput           // 配置项获取输入
type ConfigDiffInput = core.ConfigDiffInput                 // 配置差异输入
type MonitorAlertRuleListInput = core.MonitorAlertRuleListInput // 监控告警规则列表输入
type MonitorAlertActiveInput = core.MonitorAlertActiveInput // 活跃告警输入
type MonitorAlertInput = core.MonitorAlertInput             // 监控告警输入
type MonitorMetricQueryInput = core.MonitorMetricQueryInput // 监控指标查询输入
type MonitorMetricInput = core.MonitorMetricInput           // 监控指标输入
type TopologyGetInput = core.TopologyGetInput               // 拓扑获取输入
type AuditLogSearchInput = core.AuditLogSearchInput         // 审计日志搜索输入
type UserListInput = core.UserListInput                     // 用户列表输入
type RoleListInput = core.RoleListInput                     // 角色列表输入
type PermissionCheckInput = core.PermissionCheckInput       // 权限检查输入

// IsApprovalRequired 检查错误是否为需要审批错误。
func IsApprovalRequired(err error) (*ApprovalRequiredError, bool) {
	return core.IsApprovalRequired(err)
}

// IsConfirmationRequired 检查错误是否为需要确认错误。
func IsConfirmationRequired(err error) (*ConfirmationRequiredError, bool) {
	return core.IsConfirmationRequired(err)
}

// WithToolPolicyChecker 将工具策略检查器存入上下文。
func WithToolPolicyChecker(ctx context.Context, checker ToolPolicyChecker) context.Context {
	return core.WithToolPolicyChecker(ctx, checker)
}

// WithToolEventEmitter 将工具事件发射器存入上下文。
func WithToolEventEmitter(ctx context.Context, emitter ToolEventEmitter) context.Context {
	return core.WithToolEventEmitter(ctx, emitter)
}

// WithToolUser 将用户 ID 和审批令牌存入上下文。
func WithToolUser(ctx context.Context, userID uint64, approvalToken string) context.Context {
	return core.WithToolUser(ctx, userID, approvalToken)
}

// WithToolRuntimeContext 将运行时上下文存入上下文。
func WithToolRuntimeContext(ctx context.Context, runtime map[string]any) context.Context {
	return core.WithToolRuntimeContext(ctx, runtime)
}

// ToolRuntimeContextFromContext 从上下文获取运行时上下文。
func ToolRuntimeContextFromContext(ctx context.Context) map[string]any {
	return core.ToolRuntimeContextFromContext(ctx)
}

// WithToolMemoryAccessor 将工具内存访问器存入上下文。
func WithToolMemoryAccessor(ctx context.Context, accessor ToolMemoryAccessor) context.Context {
	return core.WithToolMemoryAccessor(ctx, accessor)
}

// ToolMemoryAccessorFromContext 从上下文获取工具内存访问器。
func ToolMemoryAccessorFromContext(ctx context.Context) ToolMemoryAccessor {
	return core.ToolMemoryAccessorFromContext(ctx)
}

// ToolUserFromContext 从上下文获取用户 ID 和审批令牌。
func ToolUserFromContext(ctx context.Context) (uint64, string) {
	return core.ToolUserFromContext(ctx)
}

// CheckToolPolicy 检查工具执行策略。
func CheckToolPolicy(ctx context.Context, meta ToolMeta, params map[string]any) error {
	return core.CheckToolPolicy(ctx, meta, params)
}

// EmitToolEvent 发射工具事件。
func EmitToolEvent(ctx context.Context, event string, payload any) {
	core.EmitToolEvent(ctx, event, payload)
}

// MarshalToolResult 将工具结果序列化为 JSON 字符串。
func MarshalToolResult(result ToolResult) (string, error) {
	return core.MarshalToolResult(result)
}

// NewMissingParam 创建参数缺失错误。
func NewMissingParam(field, message string) error {
	return core.NewMissingParam(field, message)
}

// NewInvalidParam 创建参数无效错误。
func NewInvalidParam(field, message string) error {
	return core.NewInvalidParam(field, message)
}

// NewParamConflict 创建参数冲突错误。
func NewParamConflict(field, message string) error {
	return core.NewParamConflict(field, message)
}

// AsToolInputError 检查错误是否为工具输入错误。
func AsToolInputError(err error) (*ToolInputError, bool) {
	return core.AsToolInputError(err)
}

// NormalizeToolName 规范化工具名称，将点号转换为下划线。
func NormalizeToolName(name string) string {
	return core.NormalizeToolName(name)
}

// filterToolsByPrefix 根据名称前缀过滤工具。
func filterToolsByPrefix(all []RegisteredTool, prefixes ...string) []RegisteredTool {
	out := make([]RegisteredTool, 0)
	for _, item := range all {
		for _, prefix := range prefixes {
			if strings.HasPrefix(item.Meta.Name, prefix) {
				out = append(out, item)
				break
			}
		}
	}
	return out
}

// filterToolsByName 根据名称精确匹配过滤工具。
func filterToolsByName(all []RegisteredTool, names ...string) []RegisteredTool {
	want := make(map[string]struct{}, len(names))
	for _, name := range names {
		want[name] = struct{}{}
	}
	out := make([]RegisteredTool, 0)
	for _, item := range all {
		if _, ok := want[item.Meta.Name]; ok {
			out = append(out, item)
		}
	}
	return out
}
