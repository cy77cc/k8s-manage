package tools

import (
	"context"
	"strings"

	core "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
)

type ToolMode = core.ToolMode

const (
	ToolModeReadonly = core.ToolModeReadonly
	ToolModeMutating = core.ToolModeMutating
)

type ToolRisk = core.ToolRisk
type ToolDomain = core.ToolDomain
type ToolCategory = core.ToolCategory

const (
	ToolRiskLow    = core.ToolRiskLow
	ToolRiskMedium = core.ToolRiskMedium
	ToolRiskHigh   = core.ToolRiskHigh
)

const (
	DomainGeneral        = core.DomainGeneral
	DomainInfrastructure = core.DomainInfrastructure
	DomainService        = core.DomainService
	DomainCICD           = core.DomainCICD
	DomainMonitor        = core.DomainMonitor
	DomainConfig         = core.DomainConfig
	DomainUser           = core.DomainUser
)

const (
	CategoryDiscovery = core.CategoryDiscovery
	CategoryAction    = core.CategoryAction
)

type ToolResult = core.ToolResult
type ToolExecutionError = core.ToolExecutionError
type ToolMeta = core.ToolMeta
type ApprovalRequiredError = core.ApprovalRequiredError
type ConfirmationRequiredError = core.ConfirmationRequiredError
type ToolPolicyChecker = core.ToolPolicyChecker
type ToolEventEmitter = core.ToolEventEmitter
type ToolMemoryAccessor = core.ToolMemoryAccessor
type PlatformDeps = core.PlatformDeps
type RegisteredTool = core.RegisteredTool

type ToolInputError = core.ToolInputError

type OSCPUMemInput = core.OSCPUMemInput
type OSDiskInput = core.OSDiskInput
type OSNetInput = core.OSNetInput
type OSProcessTopInput = core.OSProcessTopInput
type OSJournalInput = core.OSJournalInput
type OSContainerRuntimeInput = core.OSContainerRuntimeInput
type K8sListInput = core.K8sListInput
type K8sQueryInput = core.K8sQueryInput
type K8sEventsInput = core.K8sEventsInput
type K8sEventsQueryInput = core.K8sEventsQueryInput
type K8sPodLogsInput = core.K8sPodLogsInput
type K8sLogsInput = core.K8sLogsInput
type ServiceDetailInput = core.ServiceDetailInput
type ServiceDeployPreviewInput = core.ServiceDeployPreviewInput
type ServiceDeployApplyInput = core.ServiceDeployApplyInput
type ServiceDeployInput = core.ServiceDeployInput
type ServiceStatusInput = core.ServiceStatusInput
type HostSSHReadonlyInput = core.HostSSHReadonlyInput
type HostExecInput = core.HostExecInput
type HostInventoryInput = core.HostInventoryInput
type ClusterInventoryInput = core.ClusterInventoryInput
type ServiceInventoryInput = core.ServiceInventoryInput
type HostBatchExecPreviewInput = core.HostBatchExecPreviewInput
type HostBatchExecApplyInput = core.HostBatchExecApplyInput
type HostBatchInput = core.HostBatchInput
type HostBatchStatusInput = core.HostBatchStatusInput
type ServiceCatalogListInput = core.ServiceCatalogListInput
type ServiceVisibilityCheckInput = core.ServiceVisibilityCheckInput
type DeploymentTargetListInput = core.DeploymentTargetListInput
type DeploymentTargetDetailInput = core.DeploymentTargetDetailInput
type DeploymentBootstrapStatusInput = core.DeploymentBootstrapStatusInput
type CredentialListInput = core.CredentialListInput
type CredentialTestInput = core.CredentialTestInput
type CICDPipelineListInput = core.CICDPipelineListInput
type CICDPipelineStatusInput = core.CICDPipelineStatusInput
type CICDPipelineTriggerInput = core.CICDPipelineTriggerInput
type JobListInput = core.JobListInput
type JobExecutionStatusInput = core.JobExecutionStatusInput
type JobRunInput = core.JobRunInput
type ConfigAppListInput = core.ConfigAppListInput
type ConfigItemGetInput = core.ConfigItemGetInput
type ConfigDiffInput = core.ConfigDiffInput
type MonitorAlertRuleListInput = core.MonitorAlertRuleListInput
type MonitorAlertActiveInput = core.MonitorAlertActiveInput
type MonitorAlertInput = core.MonitorAlertInput
type MonitorMetricQueryInput = core.MonitorMetricQueryInput
type MonitorMetricInput = core.MonitorMetricInput
type TopologyGetInput = core.TopologyGetInput
type AuditLogSearchInput = core.AuditLogSearchInput
type UserListInput = core.UserListInput
type RoleListInput = core.RoleListInput
type PermissionCheckInput = core.PermissionCheckInput

func IsApprovalRequired(err error) (*ApprovalRequiredError, bool) {
	return core.IsApprovalRequired(err)
}

func IsConfirmationRequired(err error) (*ConfirmationRequiredError, bool) {
	return core.IsConfirmationRequired(err)
}

func WithToolPolicyChecker(ctx context.Context, checker ToolPolicyChecker) context.Context {
	return core.WithToolPolicyChecker(ctx, checker)
}

func WithToolEventEmitter(ctx context.Context, emitter ToolEventEmitter) context.Context {
	return core.WithToolEventEmitter(ctx, emitter)
}

func WithToolUser(ctx context.Context, userID uint64, approvalToken string) context.Context {
	return core.WithToolUser(ctx, userID, approvalToken)
}

func WithToolRuntimeContext(ctx context.Context, runtime map[string]any) context.Context {
	return core.WithToolRuntimeContext(ctx, runtime)
}

func ToolRuntimeContextFromContext(ctx context.Context) map[string]any {
	return core.ToolRuntimeContextFromContext(ctx)
}

func WithToolMemoryAccessor(ctx context.Context, accessor ToolMemoryAccessor) context.Context {
	return core.WithToolMemoryAccessor(ctx, accessor)
}

func ToolMemoryAccessorFromContext(ctx context.Context) ToolMemoryAccessor {
	return core.ToolMemoryAccessorFromContext(ctx)
}

func ToolUserFromContext(ctx context.Context) (uint64, string) {
	return core.ToolUserFromContext(ctx)
}

func CheckToolPolicy(ctx context.Context, meta ToolMeta, params map[string]any) error {
	return core.CheckToolPolicy(ctx, meta, params)
}

func EmitToolEvent(ctx context.Context, event string, payload any) {
	core.EmitToolEvent(ctx, event, payload)
}

func MarshalToolResult(result ToolResult) (string, error) {
	return core.MarshalToolResult(result)
}

func NewMissingParam(field, message string) error {
	return core.NewMissingParam(field, message)
}

func NewInvalidParam(field, message string) error {
	return core.NewInvalidParam(field, message)
}

func NewParamConflict(field, message string) error {
	return core.NewParamConflict(field, message)
}

func AsToolInputError(err error) (*ToolInputError, bool) {
	return core.AsToolInputError(err)
}

func NormalizeToolName(name string) string {
	return core.NormalizeToolName(name)
}

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
