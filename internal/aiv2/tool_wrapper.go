package aiv2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type approvalAwareTool struct {
	inner          tool.InvokableTool
	policy         ToolPolicy
	sessionID      string
	turnID         string
	runtimeContext map[string]any
}

func wrapApprovalTool(inner tool.InvokableTool, policy ToolPolicy, sessionID, turnID string, runtimeContext map[string]any) tool.InvokableTool {
	if inner == nil {
		return nil
	}
	return &approvalAwareTool{
		inner:          inner,
		policy:         policy,
		sessionID:      strings.TrimSpace(sessionID),
		turnID:         strings.TrimSpace(turnID),
		runtimeContext: cloneMap(runtimeContext),
	}
}

func (t *approvalAwareTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return t.inner.Info(ctx)
}

func (t *approvalAwareTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if !t.policy.ApprovalRequired {
		return t.inner.InvokableRun(ctx, argumentsInJSON, opts...)
	}

	wasInterrupted, _, storedArguments := tool.GetInterruptState[string](ctx)
	if !wasInterrupted {
		return "", tool.StatefulInterrupt(ctx, t.interruptInfo(ctx, argumentsInJSON), argumentsInJSON)
	}

	isResumeTarget, hasData, data := tool.GetResumeContext[*ApprovalDecision](ctx)
	if !isResumeTarget {
		return "", tool.StatefulInterrupt(ctx, t.interruptInfo(ctx, storedArguments), storedArguments)
	}
	if !hasData {
		return "", fmt.Errorf("approval decision missing for tool %s", t.policy.Name)
	}
	if !data.Approved {
		if strings.TrimSpace(data.Reason) != "" {
			return fmt.Sprintf("Operation cancelled: %s", strings.TrimSpace(data.Reason)), nil
		}
		return "Operation cancelled by user approval.", nil
	}
	return t.inner.InvokableRun(ctx, storedArguments, opts...)
}

func (t *approvalAwareTool) interruptInfo(ctx context.Context, argumentsInJSON string) *ApprovalInterruptInfo {
	return &ApprovalInterruptInfo{
		ToolName:        t.policy.Name,
		Expert:          t.policy.Expert,
		ArgumentsInJSON: argumentsInJSON,
		ToolCallID:      compose.GetToolCallID(ctx),
		Summary:         approvalSummary(t.policy, argumentsInJSON),
		Risk:            t.policy.Risk,
		Mode:            t.policy.Mode,
		SessionID:       t.sessionID,
		TurnID:          t.turnID,
		RuntimeContext:  cloneMap(t.runtimeContext),
	}
}

func approvalSummary(policy ToolPolicy, argumentsInJSON string) string {
	var payload map[string]any
	_ = json.Unmarshal([]byte(argumentsInJSON), &payload)
	if len(payload) == 0 {
		return fmt.Sprintf("%s requires approval before execution.", policy.Name)
	}
	target := ""
	for _, key := range []string{"host_id", "host_ids", "cluster_id", "resource_id", "name", "target"} {
		if value, ok := payload[key]; ok && fmt.Sprint(value) != "" {
			target = fmt.Sprintf("%s=%v", key, value)
			break
		}
	}
	if target != "" {
		return fmt.Sprintf("Approve %s on %s.", policy.Name, target)
	}
	return fmt.Sprintf("Approve %s with requested arguments.", policy.Name)
}
