package aiv2

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
)

type Metrics struct {
	ToolCalls   atomic.Int64
	Interrupts  atomic.Int64
	Resumes     atomic.Int64
	Completions atomic.Int64
}

type ContextInjectMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
}

func NewContextInjectMiddleware() adk.ChatModelAgentMiddleware {
	return &ContextInjectMiddleware{BaseChatModelAgentMiddleware: &adk.BaseChatModelAgentMiddleware{}}
}

func (m *ContextInjectMiddleware) BeforeAgent(ctx context.Context, runCtx *adk.ChatModelAgentContext) (context.Context, *adk.ChatModelAgentContext, error) {
	raw, _ := ctx.Value(runtimeContextKey{}).(map[string]any)
	if len(raw) == 0 {
		return ctx, runCtx, nil
	}
	var details []string
	for _, key := range []string{"scene", "route", "current_page", "project_id"} {
		if value, ok := raw[key]; ok && strings.TrimSpace(fmt.Sprint(value)) != "" {
			details = append(details, fmt.Sprintf("%s=%s", key, strings.TrimSpace(fmt.Sprint(value))))
		}
	}
	if len(details) == 0 {
		return ctx, runCtx, nil
	}
	runCtx.Instruction = strings.TrimSpace(runCtx.Instruction) + "\n\nRuntime context:\n- " + strings.Join(details, "\n- ")
	return ctx, runCtx, nil
}

type StreamingProjectorMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
	policies map[string]ToolPolicy
}

func NewStreamingProjectorMiddleware(policies map[string]ToolPolicy) adk.ChatModelAgentMiddleware {
	return &StreamingProjectorMiddleware{
		BaseChatModelAgentMiddleware: &adk.BaseChatModelAgentMiddleware{},
		policies:                     policies,
	}
}

func (m *StreamingProjectorMiddleware) WrapInvokableToolCall(ctx context.Context, endpoint adk.InvokableToolCallEndpoint, tCtx *adk.ToolContext) (adk.InvokableToolCallEndpoint, error) {
	return func(runCtx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
		policy := m.policies[strings.TrimSpace(tCtx.Name)]
		_ = adk.SendEvent(runCtx, &adk.AgentEvent{
			Action: &adk.AgentAction{CustomizedAction: &streamEnvelope{
				Type: "tool_call",
				Payload: map[string]any{
					"call_id":  tCtx.CallID,
					"tool_name": tCtx.Name,
					"expert":   policy.Expert,
					"summary":  approvalSummary(policy, argumentsInJSON),
					"risk":     policy.Risk,
					"mode":     policy.Mode,
					"ts":       time.Now().UTC().Format(time.RFC3339Nano),
				},
			}},
		})
		result, err := endpoint(runCtx, argumentsInJSON, opts...)
		if err == nil {
			_ = adk.SendEvent(runCtx, &adk.AgentEvent{
				Action: &adk.AgentAction{CustomizedAction: &streamEnvelope{
					Type: "tool_result",
					Payload: map[string]any{
						"call_id":  tCtx.CallID,
						"tool_name": tCtx.Name,
						"expert":   policy.Expert,
						"summary":  truncate(result, 240),
						"status":   "success",
						"result": map[string]any{
							"ok": true,
						},
						"ts": time.Now().UTC().Format(time.RFC3339Nano),
					},
				}},
			})
		}
		return result, err
	}, nil
}

type ObservabilityMiddleware struct {
	*adk.BaseChatModelAgentMiddleware
	metrics *Metrics
}

func NewObservabilityMiddleware(metrics *Metrics) adk.ChatModelAgentMiddleware {
	return &ObservabilityMiddleware{
		BaseChatModelAgentMiddleware: &adk.BaseChatModelAgentMiddleware{},
		metrics:                      metrics,
	}
}

func (m *ObservabilityMiddleware) WrapInvokableToolCall(ctx context.Context, endpoint adk.InvokableToolCallEndpoint, _ *adk.ToolContext) (adk.InvokableToolCallEndpoint, error) {
	return func(runCtx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
		if m.metrics != nil {
			m.metrics.ToolCalls.Add(1)
		}
		return endpoint(runCtx, argumentsInJSON, opts...)
	}, nil
}

func truncate(input string, max int) string {
	input = strings.TrimSpace(input)
	if len(input) <= max {
		return input
	}
	return input[:max] + "..."
}

