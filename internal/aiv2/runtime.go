package aiv2

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	legacyai "github.com/cy77cc/OpsPilot/internal/ai"
	"github.com/cy77cc/OpsPilot/internal/ai/events"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
	"github.com/cy77cc/OpsPilot/internal/config"
	"github.com/cy77cc/OpsPilot/internal/svc"
	"github.com/google/uuid"
)

type Runtime struct {
	deps       common.PlatformDeps
	store      *PendingStore
	checkpoint *redisCheckpointStore
	metrics    *Metrics
	buildFn    func(ctx context.Context, sessionID, turnID string, runtimeCtx map[string]any) (adk.Agent, *adk.Runner, map[string]ToolPolicy, error)
}

func NewRuntime(svcCtx *svc.ServiceContext) *Runtime {
	return &Runtime{
		deps: common.PlatformDeps{
			DB:         svcCtx.DB,
			Prometheus: svcCtx.Prometheus,
		},
		store:      NewPendingStore(svcCtx.Rdb, "ai:v2:pending:", 24*time.Hour),
		checkpoint: newRedisCheckpointStore(svcCtx.Rdb, "ai:v2:checkpoint:", 24*time.Hour),
		metrics:    &Metrics{},
	}
}

func (r *Runtime) Run(ctx context.Context, req legacyai.RunRequest, emit emitFunc) error {
	return r.run(ctx, req, emit, nil)
}

func (r *Runtime) Resume(ctx context.Context, req legacyai.ResumeRequest) (*legacyai.ResumeResult, error) {
	_, res, err := r.resume(ctx, req, nil)
	return res, err
}

func (r *Runtime) ResumeStream(ctx context.Context, req legacyai.ResumeRequest, emit emitFunc) (*legacyai.ResumeResult, error) {
	_, res, err := r.resume(ctx, req, emit)
	return res, err
}

func (r *Runtime) run(ctx context.Context, req legacyai.RunRequest, emit emitFunc, turnIDHint *string) error {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return fmt.Errorf("message is required")
	}
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = uuid.NewString()
	}
	traceID := uuid.NewString()
	turnID := uuid.NewString()
	if turnIDHint != nil && strings.TrimSpace(*turnIDHint) != "" {
		turnID = strings.TrimSpace(*turnIDHint)
	}
	checkPointID := uuid.NewString()
	meta := events.EventMeta{
		SessionID: sessionID,
		TraceID:   traceID,
		TurnID:    turnID,
		Timestamp: time.Now().UTC(),
	}
	if emit != nil {
		emit(legacyai.StreamEvent{
			Type: events.Meta,
			Meta: meta,
			Data: map[string]any{
				"session_id":   sessionID,
				"sessionId":    sessionID,
				"trace_id":     traceID,
				"traceId":      traceID,
				"turn_id":      turnID,
				"runtime_mode": runtimeMode,
				"runtimeMode":  runtimeMode,
			},
		})
		emit(legacyai.StreamEvent{
			Type: events.PlannerState,
			Meta: meta,
			Data: map[string]any{"status": "loading", "user_visible_summary": "正在理解请求并选择工具"},
		})
	}

	runtimeCtx := runtimeContextMap(req.RuntimeContext)
	ctx = context.WithValue(ctx, runtimeContextKey{}, runtimeCtx)
	agent, runner, policies, err := r.buildRunner(ctx, sessionID, turnID, runtimeCtx)
	if err != nil {
		return err
	}
	_ = agent
	iter := runner.Run(ctx, []adk.Message{schema.UserMessage(message)}, adk.WithCheckPointID(checkPointID))
	answer, thinking, interrupted, interruptID, err := r.consume(ctx, iter, meta, emit, sessionID, turnID, traceID, checkPointID, policies)
	if err != nil {
		return err
	}
	if interrupted {
		if r.metrics != nil {
			r.metrics.Interrupts.Add(1)
		}
		if emit != nil {
			emit(legacyai.StreamEvent{
				Type: events.Done,
				Meta: meta,
				Data: map[string]any{
					"session_id":   sessionID,
					"sessionId":    sessionID,
					"turn_id":      turnID,
					"status":       "waiting_user",
					"runtime_mode": runtimeMode,
					"interrupt_id": interruptID,
				},
			})
		}
		return nil
	}
	if emit != nil {
		if strings.TrimSpace(answer) != "" {
			emit(legacyai.StreamEvent{Type: events.Summary, Meta: meta, Data: map[string]any{"summary": strings.TrimSpace(answer)}})
		}
		emit(legacyai.StreamEvent{
			Type: events.Done,
			Meta: meta,
			Data: map[string]any{
				"session_id":   sessionID,
				"sessionId":    sessionID,
				"turn_id":      turnID,
				"status":       "completed",
				"thinking":     thinking,
				"runtime_mode": runtimeMode,
			},
		})
	}
	if r.metrics != nil {
		r.metrics.Completions.Add(1)
	}
	return nil
}

func (r *Runtime) resume(ctx context.Context, req legacyai.ResumeRequest, emit emitFunc) (string, *legacyai.ResumeResult, error) {
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return "", nil, fmt.Errorf("session_id is required")
	}
	pending, err := r.store.Get(ctx, sessionID)
	if err != nil {
		return "", nil, err
	}
	if pending == nil {
		return "", &legacyai.ResumeResult{SessionID: sessionID, Status: "missing", Message: "no pending approval found"}, nil
	}
	if r.metrics != nil {
		r.metrics.Resumes.Add(1)
	}
	traceID := pending.TraceID
	if traceID == "" {
		traceID = uuid.NewString()
	}
	meta := events.EventMeta{
		SessionID: pending.SessionID,
		TraceID:   traceID,
		TurnID:    pending.TurnID,
		Timestamp: time.Now().UTC(),
	}
	if emit != nil {
		emit(legacyai.StreamEvent{Type: events.Meta, Meta: meta, Data: map[string]any{
			"session_id":   pending.SessionID,
			"sessionId":    pending.SessionID,
			"turn_id":      pending.TurnID,
			"trace_id":     traceID,
			"traceId":      traceID,
			"runtime_mode": runtimeMode,
		}})
		emit(legacyai.StreamEvent{Type: events.PlannerState, Meta: meta, Data: map[string]any{"status": "loading", "user_visible_summary": "正在恢复中断后的工具调用"}})
	}
	runtimeCtx := cloneMap(pendingToContext(pending))
	ctx = context.WithValue(ctx, runtimeContextKey{}, runtimeCtx)
	_, runner, policies, err := r.buildRunner(ctx, pending.SessionID, pending.TurnID, runtimeCtx)
	if err != nil {
		return "", nil, err
	}
	iter, err := runner.ResumeWithParams(ctx, pending.CheckPointID, &adk.ResumeParams{
		Targets: map[string]any{
			pending.InterruptID: &ApprovalDecision{Approved: req.Approved, Reason: strings.TrimSpace(req.Reason)},
		},
	})
	if err != nil {
		return "", nil, err
	}
	answer, thinking, interrupted, _, err := r.consume(ctx, iter, meta, emit, pending.SessionID, pending.TurnID, traceID, pending.CheckPointID, policies)
	if err != nil {
		return "", nil, err
	}
	if interrupted {
		return "", &legacyai.ResumeResult{
			Resumed:     false,
			Interrupted: true,
			SessionID:   pending.SessionID,
			TurnID:      pending.TurnID,
			Status:      "waiting_user",
			Message:     "approval is still required",
		}, nil
	}
	_ = r.store.Delete(ctx, sessionID)
	status := "approved"
	message := "approval granted"
	if !req.Approved {
		status = "rejected"
		message = "approval rejected"
	}
	if emit != nil {
		if strings.TrimSpace(answer) != "" {
			emit(legacyai.StreamEvent{Type: events.Summary, Meta: meta, Data: map[string]any{"summary": strings.TrimSpace(answer)}})
		}
		emit(legacyai.StreamEvent{Type: events.Done, Meta: meta, Data: map[string]any{
			"session_id":   pending.SessionID,
			"sessionId":    pending.SessionID,
			"turn_id":      pending.TurnID,
			"status":       status,
			"thinking":     thinking,
			"runtime_mode": runtimeMode,
		}})
	}
	return answer, &legacyai.ResumeResult{
		Resumed:   true,
		SessionID: pending.SessionID,
		TurnID:    pending.TurnID,
		Status:    status,
		Message:   message,
	}, nil
}

func (r *Runtime) buildRunner(ctx context.Context, sessionID, turnID string, runtimeCtx map[string]any) (adk.Agent, *adk.Runner, map[string]ToolPolicy, error) {
	if r.buildFn != nil {
		return r.buildFn(ctx, sessionID, turnID, runtimeCtx)
	}
	model, err := legacyai.NewToolCallingChatModel(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	registry, err := NewToolRegistry(ctx, r.deps, sessionID, turnID, runtimeCtx)
	if err != nil {
		return nil, nil, nil, err
	}
	handlers := []adk.ChatModelAgentMiddleware{
		NewContextInjectMiddleware(),
		NewStreamingProjectorMiddleware(registry.policies),
		NewObservabilityMiddleware(r.metrics),
	}
	agent, err := newSingleAgent(ctx, model, registry.Tools(), handlers)
	if err != nil {
		return nil, nil, nil, err
	}
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
		CheckPointStore: r.checkpoint,
	})
	return agent, runner, registry.policies, nil
}

func newSingleAgent(ctx context.Context, model einomodel.BaseChatModel, tools []tool.BaseTool, handlers []adk.ChatModelAgentMiddleware) (adk.Agent, error) {
	baseTools := make([]tool.BaseTool, 0, len(tools))
	baseTools = append(baseTools, tools...)
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "opspilot-aiv2",
		Description:   "Single agent runtime for OpsPilot with direct tool calling and HITL approval.",
		Instruction:   systemPrompt(),
		Model:         model,
		MaxIterations: 8,
		Handlers:      handlers,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: baseTools},
		},
	})
}

func (r *Runtime) consume(ctx context.Context, iter *adk.AsyncIterator[*adk.AgentEvent], meta events.EventMeta, emit emitFunc, sessionID, turnID, traceID, checkpointID string, policies map[string]ToolPolicy) (string, string, bool, string, error) {
	var answer, thinking string
	var summaryStarted bool
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event == nil {
			continue
		}
		if event.Err != nil {
			return answer, thinking, false, "", event.Err
		}
		if event.Action != nil && event.Action.CustomizedAction != nil {
			switch payload := event.Action.CustomizedAction.(type) {
			case *streamEnvelope:
				if emit != nil {
					emit(legacyai.StreamEvent{Type: events.Name(payload.Type), Meta: meta, Data: cloneMap(payload.Payload)})
				}
			case streamEnvelope:
				if emit != nil {
					emit(legacyai.StreamEvent{Type: events.Name(payload.Type), Meta: meta, Data: cloneMap(payload.Payload)})
				}
			}
		}
		if event.Action != nil && event.Action.Interrupted != nil && len(event.Action.Interrupted.InterruptContexts) > 0 {
			for _, interruptCtx := range event.Action.Interrupted.InterruptContexts {
				info := toApprovalInfo(interruptCtx.Info)
				if info == nil {
					continue
				}
				if strings.TrimSpace(info.SessionID) == "" {
					info.SessionID = sessionID
				}
				if strings.TrimSpace(info.TurnID) == "" {
					info.TurnID = turnID
				}
				policy := policies[info.ToolName]
				pending := PendingApproval{
					CheckPointID: checkpointID,
					InterruptID:  interruptCtx.ID,
					SessionID:    sessionID,
					TurnID:       turnID,
					TraceID:      traceID,
					ToolName:     info.ToolName,
					Expert:       firstNonEmpty(info.Expert, policy.Expert),
					ToolArgs:     info.ArgumentsInJSON,
					Summary:      firstNonEmpty(info.Summary, approvalSummary(policy, info.ArgumentsInJSON)),
					Risk:         firstNonEmpty(info.Risk, policy.Risk),
					Mode:         firstNonEmpty(info.Mode, policy.Mode),
					CreatedAt:    time.Now().UTC(),
				}
				if err := r.store.Save(ctx, pending); err != nil {
					return answer, thinking, false, "", err
				}
				if emit != nil {
					emit(legacyai.StreamEvent{
						Type: events.ApprovalRequired,
						Meta: meta,
						Data: map[string]any{
							"id":                   pending.InterruptID,
							"session_id":           sessionID,
							"turn_id":              turnID,
							"title":                firstNonEmpty(info.Summary, pending.Summary),
							"user_visible_summary": pending.Summary,
							"tool":                 pending.ToolName,
							"risk":                 pending.Risk,
							"mode":                 pending.Mode,
							"approval_required":    true,
							"resume": map[string]any{
								"session_id": sessionID,
								"step_id":    pending.InterruptID,
							},
						},
					})
				}
				return answer, thinking, true, pending.InterruptID, nil
			}
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		output := event.Output.MessageOutput
		if output.IsStreaming && output.MessageStream != nil {
			for {
				msg, err := output.MessageStream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					return answer, thinking, false, "", err
				}
				if msg == nil {
					continue
				}
				if len(msg.ToolCalls) > 0 && emit != nil {
					for _, call := range msg.ToolCalls {
						emit(legacyai.StreamEvent{
							Type: events.ToolCall,
							Meta: meta,
							Data: map[string]any{
								"call_id":   call.ID,
								"tool_name": call.Function.Name,
								"summary":   call.Function.Arguments,
							},
						})
					}
				}
				if msg.Role == schema.Tool && emit != nil {
					emit(legacyai.StreamEvent{
						Type: events.ToolResult,
						Meta: meta,
						Data: map[string]any{
							"tool_name": msg.ToolName,
							"summary":   truncate(msg.Content, 240),
							"status":    "success",
							"result":    map[string]any{"ok": true},
						},
					})
					continue
				}
				if msg.Role != schema.Assistant {
					continue
				}
				if emit != nil && !summaryStarted {
					summaryStarted = true
					emit(legacyai.StreamEvent{
						Type: events.StageDelta,
						Meta: meta,
						Data: map[string]any{"stage": "summary", "status": "loading", "content_chunk": "正在整理最终结论"},
					})
				}
				if msg.ReasoningContent != "" {
					delta := incrementalDelta(thinking, msg.ReasoningContent)
					if delta != "" && emit != nil {
						emit(legacyai.StreamEvent{Type: events.ThinkingDelta, Meta: meta, Data: map[string]any{"content_chunk": delta}})
					}
					thinking = mergeSnapshot(thinking, msg.ReasoningContent)
				}
				if msg.Content != "" {
					delta := incrementalDelta(answer, msg.Content)
					if delta != "" && emit != nil {
						emit(legacyai.StreamEvent{Type: events.Delta, Meta: meta, Data: map[string]any{"content_chunk": delta}})
					}
					answer = mergeSnapshot(answer, msg.Content)
				}
			}
			continue
		}
		if msg := output.Message; msg != nil {
			if len(msg.ToolCalls) > 0 && emit != nil {
				for _, call := range msg.ToolCalls {
					emit(legacyai.StreamEvent{
						Type: events.ToolCall,
						Meta: meta,
						Data: map[string]any{
							"call_id":   call.ID,
							"tool_name": call.Function.Name,
							"summary":   call.Function.Arguments,
						},
					})
				}
			}
			if msg.Role == schema.Tool {
				if emit != nil {
					emit(legacyai.StreamEvent{
						Type: events.ToolResult,
						Meta: meta,
						Data: map[string]any{
							"tool_name": msg.ToolName,
							"summary":   truncate(msg.Content, 240),
							"status":    "success",
							"result":    map[string]any{"ok": true},
						},
					})
				}
				continue
			}
			if msg.Role != schema.Assistant {
				continue
			}
			if emit != nil && !summaryStarted {
				summaryStarted = true
				emit(legacyai.StreamEvent{Type: events.StageDelta, Meta: meta, Data: map[string]any{"stage": "summary", "status": "loading", "content_chunk": "正在整理最终结论"}})
			}
			if msg.ReasoningContent != "" {
				delta := incrementalDelta(thinking, msg.ReasoningContent)
				if delta != "" && emit != nil {
					emit(legacyai.StreamEvent{Type: events.ThinkingDelta, Meta: meta, Data: map[string]any{"content_chunk": delta}})
				}
				thinking = mergeSnapshot(thinking, msg.ReasoningContent)
			}
			if msg.Content != "" {
				delta := incrementalDelta(answer, msg.Content)
				if delta != "" && emit != nil {
					emit(legacyai.StreamEvent{Type: events.Delta, Meta: meta, Data: map[string]any{"content_chunk": delta}})
				}
				answer = mergeSnapshot(answer, msg.Content)
			}
		}
	}
	return answer, thinking, false, "", nil
}

func incrementalDelta(previous, current string) string {
	if current == "" || current == previous {
		return ""
	}
	if previous != "" && strings.HasPrefix(current, previous) {
		return current[len(previous):]
	}
	return current
}

func mergeSnapshot(previous, current string) string {
	if current == "" {
		return previous
	}
	if previous == "" || strings.HasPrefix(current, previous) {
		return current
	}
	return previous + current
}

func runtimeContextMap(ctx legacyai.RuntimeContext) map[string]any {
	out := cloneMap(ctx.Metadata)
	if out == nil {
		out = map[string]any{}
	}
	if strings.TrimSpace(ctx.Scene) != "" {
		out["scene"] = strings.TrimSpace(ctx.Scene)
	}
	if strings.TrimSpace(ctx.Route) != "" {
		out["route"] = strings.TrimSpace(ctx.Route)
	}
	if strings.TrimSpace(ctx.ProjectID) != "" {
		out["project_id"] = strings.TrimSpace(ctx.ProjectID)
	}
	if strings.TrimSpace(ctx.CurrentPage) != "" {
		out["current_page"] = strings.TrimSpace(ctx.CurrentPage)
	}
	if len(ctx.SelectedResources) > 0 {
		items := make([]map[string]any, 0, len(ctx.SelectedResources))
		for _, item := range ctx.SelectedResources {
			items = append(items, map[string]any{"type": item.Type, "id": item.ID, "name": item.Name})
		}
		out["selected_resources"] = items
	}
	return out
}

func pendingToContext(pending *PendingApproval) map[string]any {
	if pending == nil {
		return nil
	}
	return map[string]any{
		"session_id": pending.SessionID,
		"turn_id":    pending.TurnID,
	}
}

func toApprovalInfo(raw any) *ApprovalInterruptInfo {
	switch value := raw.(type) {
	case *ApprovalInterruptInfo:
		return value
	case ApprovalInterruptInfo:
		v := value
		return &v
	case map[string]any:
		return &ApprovalInterruptInfo{
			ToolName:        stringify(value["tool_name"]),
			Expert:          stringify(value["expert"]),
			ArgumentsInJSON: stringify(value["arguments_in_json"]),
			ToolCallID:      stringify(value["tool_call_id"]),
			Summary:         stringify(value["summary"]),
			Risk:            stringify(value["risk"]),
			Mode:            stringify(value["mode"]),
			SessionID:       stringify(value["session_id"]),
			TurnID:          stringify(value["turn_id"]),
		}
	default:
		return nil
	}
}

func stringify(v any) string {
	return strings.TrimSpace(fmt.Sprint(v))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func cloneMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func Enabled() bool {
	return config.CFG.FeatureFlags.AIAssistantV2 != nil && *config.CFG.FeatureFlags.AIAssistantV2
}
