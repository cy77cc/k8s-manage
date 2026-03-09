package aspect

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/callbacks"
	cbtool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	"github.com/cy77cc/k8s-manage/internal/ai/tools/core"
)

type DecisionAction string

const (
	DecisionExecute        DecisionAction = "execute"
	DecisionInterrupt      DecisionAction = "interrupt"
	DecisionCreateApproval DecisionAction = "create_approval"
)

type ToolCallDecision struct {
	Action  DecisionAction
	Reason  string
	Preview map[string]any
}

type PermissionChecker interface {
	HasPermission(ctx context.Context, permission string) (bool, error)
}

type InterruptHandler interface {
	BuildInterrupt(ctx context.Context, meta core.ToolMeta, args map[string]any) (*tools.ApprovalInfo, error)
}

type AuditLogger interface {
	Log(ctx context.Context, record AuditRecord) error
}

type AuditRecord struct {
	ToolName   string
	Action     string
	Permission string
	Allowed    bool
	Reason     string
	Arguments  string
	CreatedAt  time.Time
}

type SecurityAspect struct {
	registry          *tools.Registry
	permissionChecker PermissionChecker
	interruptHandler  InterruptHandler
	auditLogger       AuditLogger
}

func NewSecurityAspect(registered []core.RegisteredTool, checker PermissionChecker, handler InterruptHandler, logger AuditLogger) *SecurityAspect {
	return &SecurityAspect{
		registry:          tools.NewRegistry(registered),
		permissionChecker: checker,
		interruptHandler:  handler,
		auditLogger:       logger,
	}
}

func (a *SecurityAspect) EvaluateToolCall(ctx context.Context, name, argumentsInJSON string) (*ToolCallDecision, error) {
	meta, ok := a.lookupMeta(name)
	if !ok {
		return &ToolCallDecision{Action: DecisionExecute}, nil
	}

	args, err := decodeArguments(argumentsInJSON)
	if err != nil {
		return nil, err
	}

	allowed, err := a.checkPermission(ctx, meta)
	if err != nil {
		return nil, err
	}
	if !allowed {
		decision := &ToolCallDecision{
			Action: DecisionCreateApproval,
			Reason: "user lacks required permission",
		}
		a.log(ctx, meta, false, decision.Reason, argumentsInJSON)
		return decision, nil
	}

	if meta.Risk == tools.ToolRiskMedium || meta.Risk == tools.ToolRiskHigh {
		reason := "tool execution requires user confirmation"
		decision := &ToolCallDecision{Action: DecisionInterrupt, Reason: reason}
		if a.interruptHandler != nil {
			info, err := a.interruptHandler.BuildInterrupt(ctx, meta, args)
			if err != nil {
				return nil, err
			}
			if info != nil {
				decision.Preview = info.Preview
				if strings.TrimSpace(info.ToolName) != "" {
					reason = info.ToolName + ": " + reason
					decision.Reason = reason
				}
			}
		}
		a.log(ctx, meta, true, decision.Reason, argumentsInJSON)
		return decision, nil
	}

	a.log(ctx, meta, true, "", argumentsInJSON)
	return &ToolCallDecision{Action: DecisionExecute}, nil
}

func (a *SecurityAspect) Middleware() compose.ToolMiddleware {
	return compose.ToolMiddleware{
		Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
			return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
				decision, err := a.EvaluateToolCall(ctx, input.Name, input.Arguments)
				if err != nil {
					return nil, err
				}
				switch decision.Action {
				case DecisionCreateApproval:
					return nil, &tools.ApprovalRequiredError{
						Tool:    input.Name,
						Message: decision.Reason,
					}
				case DecisionInterrupt:
					return nil, cbtool.StatefulInterrupt(ctx, &tools.ApprovalInfo{
						ToolName:        input.Name,
						ArgumentsInJSON: input.Arguments,
						Preview:         decision.Preview,
					}, input.Arguments)
				default:
					return next(ctx, input)
				}
			}
		},
	}
}

func (a *SecurityAspect) CallbackHandler() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, _ *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			cb := cbtool.ConvCallbackInput(input)
			if cb == nil {
				return ctx
			}
			toolName, _ := cb.Extra["tool_name"].(string)
			a.log(ctx, core.ToolMeta{Name: toolName}, true, "callback_start", cb.ArgumentsInJSON)
			return ctx
		}).
		OnEndFn(func(ctx context.Context, _ *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			cb := cbtool.ConvCallbackOutput(output)
			if cb == nil {
				return ctx
			}
			toolName, _ := cb.Extra["tool_name"].(string)
			a.log(ctx, core.ToolMeta{Name: toolName}, true, "callback_end", cb.Response)
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, _ *callbacks.RunInfo, err error) context.Context {
			a.log(ctx, core.ToolMeta{}, false, err.Error(), "")
			return ctx
		}).
		Build()
}

func (a *SecurityAspect) lookupMeta(name string) (core.ToolMeta, bool) {
	if a == nil || a.registry == nil {
		return core.ToolMeta{}, false
	}
	item, ok := a.registry.Get(name)
	if !ok {
		return core.ToolMeta{}, false
	}
	return item.Meta, true
}

func (a *SecurityAspect) checkPermission(ctx context.Context, meta core.ToolMeta) (bool, error) {
	if a == nil || a.permissionChecker == nil || strings.TrimSpace(meta.Permission) == "" {
		return true, nil
	}
	return a.permissionChecker.HasPermission(ctx, meta.Permission)
}

func (a *SecurityAspect) log(ctx context.Context, meta core.ToolMeta, allowed bool, reason, arguments string) {
	if a == nil || a.auditLogger == nil {
		return
	}
	_ = a.auditLogger.Log(ctx, AuditRecord{
		ToolName:   meta.Name,
		Action:     string(meta.Risk),
		Permission: meta.Permission,
		Allowed:    allowed,
		Reason:     reason,
		Arguments:  arguments,
		CreatedAt:  time.Now(),
	})
}

func decodeArguments(argumentsInJSON string) (map[string]any, error) {
	args := map[string]any{}
	if strings.TrimSpace(argumentsInJSON) == "" {
		return args, nil
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return nil, fmt.Errorf("decode tool arguments: %w", err)
	}
	return args, nil
}
