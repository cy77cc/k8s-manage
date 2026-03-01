package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
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

type ToolMeta struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Mode        ToolMode       `json:"mode"`
	Risk        ToolRisk       `json:"risk"`
	Provider    string         `json:"provider"`
	Permission  string         `json:"permission"`
	Schema      map[string]any `json:"schema,omitempty"`
	Required    []string       `json:"required,omitempty"`
	DefaultHint map[string]any `json:"default_hint,omitempty"`
	Examples    []string       `json:"examples,omitempty"`
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

type K8sEventsInput struct {
	ClusterID int    `json:"cluster_id,omitempty" jsonschema:"description=cluster id in database"`
	Namespace string `json:"namespace,omitempty" jsonschema:"description=kubernetes namespace,default=default"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max events,default=50"`
}

type K8sPodLogsInput struct {
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

type HostSSHReadonlyInput struct {
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

type HostBatchStatusInput struct {
	HostIDs []int  `json:"host_ids" jsonschema:"required,description=target host ids"`
	Action  string `json:"action" jsonschema:"required,description=status action: online/offline/maintenance"`
	Reason  string `json:"reason,omitempty" jsonschema:"description=change reason for audit context"`
}
