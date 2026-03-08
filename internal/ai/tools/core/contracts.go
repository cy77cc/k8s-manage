package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
)

type ToolMode string

const (
	ToolModeReadonly ToolMode = "readonly"
	ToolModeMutating ToolMode = "mutating"
)

type ToolRisk string

const (
	ToolRiskLow    ToolRisk = "low"
	ToolRiskMedium ToolRisk = "medium"
	ToolRiskHigh   ToolRisk = "high"
)

type ToolResult struct {
	OK        bool   `json:"ok"`
	ErrorCode string `json:"error_code,omitempty"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
	Source    string `json:"source"`
	LatencyMS int64  `json:"latency_ms"`
}

type ToolExecutionError struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Recoverable bool     `json:"recoverable"`
	Suggestions []string `json:"suggestions,omitempty"`
	HintAction  string   `json:"hint_action,omitempty"`
}

type ToolMeta struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Mode         ToolMode          `json:"mode"`
	Risk         ToolRisk          `json:"risk"`
	Provider     string            `json:"provider"`
	Permission   string            `json:"permission"`
	Schema       map[string]any    `json:"schema,omitempty"`
	Required     []string          `json:"required,omitempty"`
	DefaultHint  map[string]any    `json:"default_hint,omitempty"`
	Examples     []string          `json:"examples,omitempty"`
	EnumSources  map[string]string `json:"enum_sources,omitempty"`
	ParamHints   map[string]string `json:"param_hints,omitempty"`
	RelatedTools []string          `json:"related_tools,omitempty"`
	SceneScope   []string          `json:"scene_scope,omitempty"`
}

type ApprovalRequiredError struct {
	Token     string
	Tool      string
	ExpiresAt time.Time
	Message   string
}

func (e *ApprovalRequiredError) Error() string {
	if e == nil {
		return "approval required"
	}
	if e.Message != "" {
		return e.Message
	}
	return "approval required"
}

func IsApprovalRequired(err error) (*ApprovalRequiredError, bool) {
	var apErr *ApprovalRequiredError
	if errors.As(err, &apErr) {
		return apErr, true
	}
	return nil, false
}

type ConfirmationRequiredError struct {
	Token     string
	Tool      string
	Preview   map[string]any
	ExpiresAt time.Time
	Message   string
}

func (e *ConfirmationRequiredError) Error() string {
	if e == nil {
		return "confirmation required"
	}
	if e.Message != "" {
		return e.Message
	}
	return "confirmation required"
}

func IsConfirmationRequired(err error) (*ConfirmationRequiredError, bool) {
	var cfErr *ConfirmationRequiredError
	if errors.As(err, &cfErr) {
		return cfErr, true
	}
	return nil, false
}

type ToolPolicyChecker func(ctx context.Context, meta ToolMeta, params map[string]any) error
type ToolEventEmitter func(event string, payload any)
type ToolMemoryAccessor interface {
	GetLastToolParams(toolName string) map[string]any
	SetLastToolParams(toolName string, params map[string]any)
}

type ctxKey string

const (
	policyCheckerCtxKey  ctxKey = "ai_tool_policy_checker"
	eventEmitterCtxKey   ctxKey = "ai_tool_event_emitter"
	userIDCtxKey         ctxKey = "ai_user_id"
	approvalCtxKey       ctxKey = "ai_approval_token"
	runtimeCtxKey        ctxKey = "ai_runtime_context"
	memoryAccessorCtxKey ctxKey = "ai_tool_memory_accessor"
)

func WithToolPolicyChecker(ctx context.Context, checker ToolPolicyChecker) context.Context {
	return context.WithValue(ctx, policyCheckerCtxKey, checker)
}

func WithToolEventEmitter(ctx context.Context, emitter ToolEventEmitter) context.Context {
	return context.WithValue(ctx, eventEmitterCtxKey, emitter)
}

func WithToolUser(ctx context.Context, userID uint64, approvalToken string) context.Context {
	ctx = context.WithValue(ctx, userIDCtxKey, userID)
	return context.WithValue(ctx, approvalCtxKey, approvalToken)
}

func WithToolRuntimeContext(ctx context.Context, runtime map[string]any) context.Context {
	return context.WithValue(ctx, runtimeCtxKey, runtime)
}

func ToolRuntimeContextFromContext(ctx context.Context) map[string]any {
	v, _ := ctx.Value(runtimeCtxKey).(map[string]any)
	if v == nil {
		return map[string]any{}
	}
	return v
}

func WithToolMemoryAccessor(ctx context.Context, accessor ToolMemoryAccessor) context.Context {
	return context.WithValue(ctx, memoryAccessorCtxKey, accessor)
}

func ToolMemoryAccessorFromContext(ctx context.Context) ToolMemoryAccessor {
	v := ctx.Value(memoryAccessorCtxKey)
	if v == nil {
		return nil
	}
	acc, _ := v.(ToolMemoryAccessor)
	return acc
}

func ToolUserFromContext(ctx context.Context) (uint64, string) {
	uid, _ := ctx.Value(userIDCtxKey).(uint64)
	token, _ := ctx.Value(approvalCtxKey).(string)
	return uid, token
}

func CheckToolPolicy(ctx context.Context, meta ToolMeta, params map[string]any) error {
	v := ctx.Value(policyCheckerCtxKey)
	if v == nil {
		return nil
	}
	checker, ok := v.(ToolPolicyChecker)
	if !ok || checker == nil {
		return nil
	}
	return checker(ctx, meta, params)
}

func EmitToolEvent(ctx context.Context, event string, payload any) {
	v := ctx.Value(eventEmitterCtxKey)
	if v == nil {
		return
	}
	emitter, ok := v.(ToolEventEmitter)
	if !ok || emitter == nil {
		return
	}
	emitter(event, payload)
}

func MarshalToolResult(result ToolResult) (string, error) {
	raw, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

type ToolInputError struct {
	Code    string
	Field   string
	Message string
}

func (e *ToolInputError) Error() string {
	if e == nil {
		return "invalid tool input"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Field)
	}
	return e.Code
}

func NewMissingParam(field, message string) error {
	return &ToolInputError{Code: "missing_param", Field: field, Message: message}
}

func NewInvalidParam(field, message string) error {
	return &ToolInputError{Code: "invalid_param", Field: field, Message: message}
}

func NewParamConflict(field, message string) error {
	return &ToolInputError{Code: "param_conflict", Field: field, Message: message}
}

func AsToolInputError(err error) (*ToolInputError, bool) {
	var ie *ToolInputError
	if errors.As(err, &ie) {
		return ie, true
	}
	return nil, false
}

var toolCallSeq uint64

func nextToolCallID() string {
	n := atomic.AddUint64(&toolCallSeq, 1)
	return fmt.Sprintf("tc-%d-%d", time.Now().UnixNano(), n)
}

func NextToolCallID() string {
	return nextToolCallID()
}

// NormalizeToolName keeps tool naming compatible with OpenAI-style function names.
// Legacy dotted names are converted to underscore style.
func NormalizeToolName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	return strings.ReplaceAll(trimmed, ".", "_")
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

type PlatformDeps struct {
	DB        *gorm.DB
	Clientset *kubernetes.Clientset
}

type RegisteredTool struct {
	Meta ToolMeta
	Tool tool.InvokableTool
}

type OSCPUMemInput struct {
	Target string `json:"target,omitempty" jsonschema:"description=target host id/ip/hostname,default=localhost"`
}

type OSDiskInput struct {
	Target string `json:"target,omitempty" jsonschema:"description=target host id/ip/hostname,default=localhost"`
}

type OSNetInput struct {
	Target string `json:"target,omitempty" jsonschema:"description=target host id/ip/hostname,default=localhost"`
}

type OSProcessTopInput struct {
	Target string `json:"target,omitempty" jsonschema:"description=target host id/ip/hostname,default=localhost"`
	Limit  int    `json:"limit,omitempty" jsonschema:"description=top process count,default=10"`
}

type OSJournalInput struct {
	Target  string `json:"target,omitempty" jsonschema:"description=target host id/ip/hostname,default=localhost"`
	Service string `json:"service" jsonschema:"required,description=systemd service unit"`
	Lines   int    `json:"lines,omitempty" jsonschema:"description=log lines,default=200"`
}

type OSContainerRuntimeInput struct {
	Target string `json:"target,omitempty" jsonschema:"description=target host id/ip/hostname,default=localhost"`
}

type K8sListInput struct {
	ClusterID int    `json:"cluster_id,omitempty" jsonschema:"description=cluster id in database"`
	Namespace string `json:"namespace,omitempty" jsonschema:"description=kubernetes namespace,default=default"`
	Resource  string `json:"resource" jsonschema:"required,description=resource type,enum=pods,enum=services,enum=deployments,enum=nodes"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max items,default=50"`
}

type K8sQueryInput struct {
	ClusterID int    `json:"cluster_id,omitempty" jsonschema:"description=cluster id in database"`
	Namespace string `json:"namespace,omitempty" jsonschema:"description=kubernetes namespace,default=default"`
	Resource  string `json:"resource" jsonschema:"required,description=resource type,enum=pods,enum=services,enum=deployments,enum=nodes"`
	Name      string `json:"name,omitempty" jsonschema:"description=resource name for exact lookup"`
	Label     string `json:"label,omitempty" jsonschema:"description=label selector"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max items,default=50"`
}

type K8sEventsInput struct {
	ClusterID int    `json:"cluster_id,omitempty" jsonschema:"description=cluster id in database"`
	Namespace string `json:"namespace,omitempty" jsonschema:"description=kubernetes namespace,default=default"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max events,default=50"`
}

type K8sEventsQueryInput struct {
	ClusterID int    `json:"cluster_id,omitempty" jsonschema:"description=cluster id in database"`
	Namespace string `json:"namespace,omitempty" jsonschema:"description=kubernetes namespace,default=default"`
	Kind      string `json:"kind,omitempty" jsonschema:"description=involved object kind,enum=Pod,enum=Deployment,enum=Service,enum=Node"`
	Name      string `json:"name,omitempty" jsonschema:"description=involved object name"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max events,default=50"`
}

type K8sPodLogsInput struct {
	ClusterID int    `json:"cluster_id,omitempty" jsonschema:"description=cluster id in database"`
	Namespace string `json:"namespace,omitempty" jsonschema:"description=kubernetes namespace,default=default"`
	Pod       string `json:"pod" jsonschema:"required,description=pod name"`
	Container string `json:"container,omitempty" jsonschema:"description=container name"`
	TailLines int    `json:"tail_lines,omitempty" jsonschema:"description=tail lines,default=200"`
}

type K8sLogsInput struct {
	ClusterID int    `json:"cluster_id,omitempty" jsonschema:"description=cluster id in database"`
	Namespace string `json:"namespace,omitempty" jsonschema:"description=kubernetes namespace,default=default"`
	Pod       string `json:"pod" jsonschema:"required,description=pod name"`
	Container string `json:"container,omitempty" jsonschema:"description=container name"`
	TailLines int    `json:"tail_lines,omitempty" jsonschema:"description=tail lines,default=200"`
}

type ServiceDetailInput struct {
	ServiceID int `json:"service_id" jsonschema:"required,description=service id"`
}

type ServiceDeployPreviewInput struct {
	ServiceID int `json:"service_id" jsonschema:"required,description=service id"`
	ClusterID int `json:"cluster_id" jsonschema:"required,description=cluster id"`
}

type ServiceDeployApplyInput struct {
	ServiceID int `json:"service_id" jsonschema:"required,description=service id"`
	ClusterID int `json:"cluster_id" jsonschema:"required,description=cluster id"`
}

type ServiceDeployInput struct {
	ServiceID int  `json:"service_id" jsonschema:"required,description=service id"`
	ClusterID int  `json:"cluster_id" jsonschema:"required,description=cluster id"`
	Preview   bool `json:"preview,omitempty" jsonschema:"description=preview deploy without apply"`
	Apply     bool `json:"apply,omitempty" jsonschema:"description=apply deploy after approval"`
}

type ServiceStatusInput struct {
	ServiceID int `json:"service_id" jsonschema:"required,description=service id"`
}

type HostSSHReadonlyInput struct {
	HostID  int    `json:"host_id" jsonschema:"required,description=host id"`
	Command string `json:"command" jsonschema:"required,description=readonly command"`
}

type HostExecInput struct {
	HostID  int    `json:"host_id" jsonschema:"required,description=host id"`
	Command string `json:"command" jsonschema:"required,description=readonly command"`
}

type HostInventoryInput struct {
	Status  string `json:"status,omitempty" jsonschema:"description=optional host status filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on name/ip/hostname"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max hosts,default=50"`
}

type ClusterInventoryInput struct {
	Status  string `json:"status,omitempty" jsonschema:"description=optional cluster status filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on name/endpoint"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max clusters,default=50"`
}

type ServiceInventoryInput struct {
	Keyword     string `json:"keyword,omitempty" jsonschema:"description=optional keyword on service name/owner"`
	RuntimeType string `json:"runtime_type,omitempty" jsonschema:"description=optional runtime type filter,k8s/compose/helm"`
	Env         string `json:"env,omitempty" jsonschema:"description=optional environment filter"`
	Status      string `json:"status,omitempty" jsonschema:"description=optional service status filter"`
	Limit       int    `json:"limit,omitempty" jsonschema:"description=max services,default=50"`
}

type HostBatchExecPreviewInput struct {
	HostIDs []int  `json:"host_ids" jsonschema:"required,description=target host ids"`
	Command string `json:"command" jsonschema:"required,description=shell command to run"`
	Reason  string `json:"reason,omitempty" jsonschema:"description=execution reason for audit context"`
}

type HostBatchExecApplyInput struct {
	HostIDs []int  `json:"host_ids" jsonschema:"required,description=target host ids"`
	Command string `json:"command" jsonschema:"required,description=shell command to run"`
	Reason  string `json:"reason,omitempty" jsonschema:"description=execution reason for audit context"`
}

type HostBatchInput struct {
	HostIDs []int  `json:"host_ids" jsonschema:"required,description=target host ids"`
	Command string `json:"command" jsonschema:"required,description=shell command to run"`
	Reason  string `json:"reason,omitempty" jsonschema:"description=execution reason for audit context"`
}

type HostBatchStatusInput struct {
	HostIDs []int  `json:"host_ids" jsonschema:"required,description=target host ids"`
	Action  string `json:"action" jsonschema:"required,description=status action: online/offline/maintenance"`
	Reason  string `json:"reason,omitempty" jsonschema:"description=change reason for audit context"`
}

type ServiceCatalogListInput struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"description=optional keyword on service name/owner"`
	CategoryID int    `json:"category_id,omitempty" jsonschema:"description=optional category id: 1 middleware, 2 business"`
	Limit      int    `json:"limit,omitempty" jsonschema:"description=max services,default=50"`
}

type ServiceVisibilityCheckInput struct {
	ServiceID int `json:"service_id" jsonschema:"required,description=service id"`
}

type DeploymentTargetListInput struct {
	Env     string `json:"env,omitempty" jsonschema:"description=optional environment filter"`
	Status  string `json:"status,omitempty" jsonschema:"description=optional target status filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional target keyword filter"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max targets,default=50"`
}

type DeploymentTargetDetailInput struct {
	TargetID int `json:"target_id" jsonschema:"required,description=deployment target id"`
}

type DeploymentBootstrapStatusInput struct {
	TargetID int `json:"target_id" jsonschema:"required,description=deployment target id"`
}

type CredentialListInput struct {
	Type    string `json:"type,omitempty" jsonschema:"description=credential type or runtime type"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on name/endpoint"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max credentials,default=50"`
}

type CredentialTestInput struct {
	CredentialID int `json:"credential_id" jsonschema:"required,description=credential id"`
}

type CICDPipelineListInput struct {
	Status  string `json:"status,omitempty" jsonschema:"description=optional status filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on repo/branch"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max pipelines,default=50"`
}

type CICDPipelineStatusInput struct {
	PipelineID int `json:"pipeline_id" jsonschema:"required,description=pipeline config id"`
}

type CICDPipelineTriggerInput struct {
	PipelineID int               `json:"pipeline_id" jsonschema:"required,description=pipeline config id"`
	Branch     string            `json:"branch" jsonschema:"required,description=branch to build"`
	Params     map[string]string `json:"params,omitempty" jsonschema:"description=optional trigger params"`
}

type JobListInput struct {
	Status  string `json:"status,omitempty" jsonschema:"description=optional status filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on name/type"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max jobs,default=50"`
}

type JobExecutionStatusInput struct {
	JobID       int `json:"job_id" jsonschema:"required,description=job id"`
	ExecutionID int `json:"execution_id,omitempty" jsonschema:"description=optional execution id"`
}

type JobRunInput struct {
	JobID  int            `json:"job_id" jsonschema:"required,description=job id"`
	Params map[string]any `json:"params,omitempty" jsonschema:"description=optional run params"`
}

type ConfigAppListInput struct {
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on service name"`
	Env     string `json:"env,omitempty" jsonschema:"description=optional env filter"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max apps,default=50"`
}

type ConfigItemGetInput struct {
	AppID int    `json:"app_id" jsonschema:"required,description=service id as config app id"`
	Key   string `json:"key" jsonschema:"required,description=config key"`
	Env   string `json:"env,omitempty" jsonschema:"description=optional env"`
}

type ConfigDiffInput struct {
	AppID int    `json:"app_id" jsonschema:"required,description=service id as config app id"`
	EnvA  string `json:"env_a" jsonschema:"required,description=compare env a"`
	EnvB  string `json:"env_b" jsonschema:"required,description=compare env b"`
}

type MonitorAlertRuleListInput struct {
	Status  string `json:"status,omitempty" jsonschema:"description=optional rule state filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on name/metric"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max rules,default=50"`
}

type MonitorAlertActiveInput struct {
	Severity  string `json:"severity,omitempty" jsonschema:"description=optional severity filter"`
	ServiceID int    `json:"service_id,omitempty" jsonschema:"description=optional service id filter"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max alerts,default=50"`
}

type MonitorAlertInput struct {
	Severity  string `json:"severity,omitempty" jsonschema:"description=optional severity filter"`
	ServiceID int    `json:"service_id,omitempty" jsonschema:"description=optional service id filter"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max alerts,default=50"`
}

type MonitorMetricQueryInput struct {
	Query     string `json:"query" jsonschema:"required,description=metric query or metric name"`
	TimeRange string `json:"time_range,omitempty" jsonschema:"description=time range,default=1h"`
	Step      int    `json:"step,omitempty" jsonschema:"description=step seconds,default=60"`
}

type MonitorMetricInput struct {
	Query     string `json:"query" jsonschema:"required,description=metric query or metric name"`
	TimeRange string `json:"time_range,omitempty" jsonschema:"description=time range,default=1h"`
	Step      int    `json:"step,omitempty" jsonschema:"description=step seconds,default=60"`
}

type TopologyGetInput struct {
	ServiceID int `json:"service_id,omitempty" jsonschema:"description=optional service id"`
	Depth     int `json:"depth,omitempty" jsonschema:"description=max depth,default=2"`
}

type AuditLogSearchInput struct {
	TimeRange    string `json:"time_range,omitempty" jsonschema:"description=time range,default=24h"`
	ResourceType string `json:"resource_type,omitempty" jsonschema:"description=optional resource type"`
	Action       string `json:"action,omitempty" jsonschema:"description=optional action type"`
	UserID       int    `json:"user_id,omitempty" jsonschema:"description=optional actor user id"`
	Limit        int    `json:"limit,omitempty" jsonschema:"description=max logs,default=50"`
}

type UserListInput struct {
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional username/email keyword"`
	Status  int    `json:"status,omitempty" jsonschema:"description=optional status filter"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max users,default=50"`
}

type RoleListInput struct {
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional role keyword"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max roles,default=50"`
}

type PermissionCheckInput struct {
	UserID   int    `json:"user_id" jsonschema:"required,description=user id"`
	Resource string `json:"resource" jsonschema:"required,description=resource name"`
	Action   string `json:"action" jsonschema:"required,description=action name"`
}
