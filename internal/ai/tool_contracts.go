package ai

import (
	"context"
	"encoding/json"
	"errors"
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

type ctxKey string

const (
	policyCheckerCtxKey ctxKey = "ai_tool_policy_checker"
	eventEmitterCtxKey  ctxKey = "ai_tool_event_emitter"
	userIDCtxKey        ctxKey = "ai_user_id"
	approvalCtxKey      ctxKey = "ai_approval_token"
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
