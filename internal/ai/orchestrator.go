package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/OpsPilot/internal/ai/adkstage"
	"github.com/cy77cc/OpsPilot/internal/ai/events"
	"github.com/cy77cc/OpsPilot/internal/ai/executor"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/rewrite"
	"github.com/cy77cc/OpsPilot/internal/ai/runtime"
	"github.com/cy77cc/OpsPilot/internal/ai/state"
	"github.com/google/uuid"
)

type StreamEmitter func(StreamEvent) bool

type Orchestrator struct {
	sessions   *state.SessionState
	executions *runtime.ExecutionStore
	rewriter   *rewrite.Rewriter
	planner    *planner.Planner
	executor   *executor.Executor
}

func NewOrchestrator(sessions *state.SessionState, executions *runtime.ExecutionStore) *Orchestrator {
	return &Orchestrator{
		sessions:   sessions,
		executions: executions,
		rewriter: rewrite.New(adkstage.New(adkstage.Config{
			Name:         "rewrite-stage",
			Description:  "Rewrite user requests into a stable semi-structured task draft.",
			Instruction:  rewrite.SystemPrompt(),
			ModelFactory: NewBaseChatModel,
		})),
		planner: planner.New(adkstage.New(adkstage.Config{
			Name:        "planner-stage",
			Description: "Plan clarified AI operations tasks into structured decisions and steps.",
			Instruction: planner.SystemPrompt(),
			ToolsConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{},
			},
			ModelFactory: NewBaseChatModel,
		})),
		executor: executor.New(executions),
	}
}

func (o *Orchestrator) Run(ctx context.Context, req RunRequest, emit StreamEmitter) error {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return fmt.Errorf("message is required")
	}

	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = uuid.NewString()
	}
	traceID := uuid.NewString()
	meta := events.EventMeta{
		SessionID: sessionID,
		TraceID:   traceID,
		Iteration: 1,
		Timestamp: time.Now().UTC(),
	}

	if o.sessions != nil {
		if err := o.sessions.AppendMessage(ctx, sessionID, schema.UserMessage(message)); err != nil {
			return err
		}
		if err := o.sessions.EnsureTitle(ctx, sessionID, deriveSessionTitle(message)); err != nil {
			return err
		}
	}

	if o.executions != nil {
		resourceIDs := make([]string, 0, len(req.RuntimeContext.SelectedResources))
		for _, item := range req.RuntimeContext.SelectedResources {
			if item.ID != "" {
				resourceIDs = append(resourceIDs, item.ID)
			}
		}
		if err := o.executions.Save(ctx, runtime.ExecutionState{
			TraceID:   traceID,
			SessionID: sessionID,
			Message:   message,
			Status:    runtime.ExecutionStatusRunning,
			Phase:     "gateway_entry",
			RuntimeContext: runtime.ContextSnapshot{
				Scene:       req.RuntimeContext.Scene,
				Route:       req.RuntimeContext.Route,
				ProjectID:   req.RuntimeContext.ProjectID,
				CurrentPage: req.RuntimeContext.CurrentPage,
				ResourceIDs: resourceIDs,
			},
		}); err != nil {
			return err
		}
	}

	emitEvent(emit, events.Meta, meta, map[string]any{
		"session_id": sessionID,
		"sessionId":  sessionID,
		"trace_id":   traceID,
		"traceId":    traceID,
		"createdAt":  meta.Timestamp.Format(time.RFC3339),
	})

	var rewriteOut rewrite.Output
	if o.rewriter != nil {
		var err error
		rewriteOut, err = o.rewriter.Rewrite(ctx, rewrite.Input{
			Message:           message,
			Scene:             req.RuntimeContext.Scene,
			CurrentPage:       req.RuntimeContext.CurrentPage,
			SelectedResources: toRewriteResources(req.RuntimeContext.SelectedResources),
		})
		if err == nil {
			emitEvent(emit, events.RewriteResult, meta, map[string]any{
				"rewrite":              rewriteOut,
				"user_visible_summary": rewriteOut.Narrative,
			})
		}
	}

	reply, genErr := o.planAndReply(ctx, message, rewriteOut, req.RuntimeContext, meta, emit, sessionID)
	if genErr != nil {
		reply = fmt.Sprintf("AI 编排入口已经切换到新的宿主边界，但当前模型暂不可用：%s", genErr.Error())
	}

	if o.sessions != nil {
		if err := o.sessions.AppendMessage(ctx, sessionID, schema.AssistantMessage(reply, nil)); err != nil {
			return err
		}
	}

	emitEvent(emit, events.Delta, meta, map[string]any{
		"content_chunk": reply,
		"contentChunk":  reply,
	})

	sessionPayload := map[string]any{
		"id":        sessionID,
		"title":     deriveSessionTitle(message),
		"messages":  o.sessionMessages(ctx, sessionID),
		"createdAt": meta.Timestamp.Format(time.RFC3339),
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
	}
	emitEvent(emit, events.Done, meta, map[string]any{
		"session": sessionPayload,
	})

	if o.executions != nil {
		st, err := o.executions.Load(ctx, sessionID)
		if err == nil && st != nil {
			st.Status = runtime.ExecutionStatusCompleted
			st.Phase = "completed"
			_ = o.executions.Save(ctx, *st)
		}
	}
	return nil
}

func (o *Orchestrator) Resume(ctx context.Context, req ResumeRequest) (*ResumeResult, error) {
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if o.executions == nil {
		return &ResumeResult{
			Resumed:   false,
			SessionID: sessionID,
			Status:    "unavailable",
			Message:   "execution store is not configured",
		}, nil
	}

	st, err := o.executions.Load(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return &ResumeResult{
			Resumed:   false,
			SessionID: sessionID,
			Status:    "missing",
			Message:   "execution state not found",
		}, nil
	}

	if st.PendingApproval == nil {
		return &ResumeResult{
			Resumed:   false,
			SessionID: sessionID,
			PlanID:    st.PlanID,
			StepID:    firstNonEmpty(req.StepID, req.Target),
			Status:    "noop",
			Message:   "no pending approval for this session",
		}, nil
	}
	stepID := firstNonEmpty(req.StepID, req.Target, st.PendingApproval.StepID)
	planID := firstNonEmpty(req.PlanID, st.PlanID, st.PendingApproval.PlanID)

	decisionHash := runtime.ApprovalDecisionHash(planID, stepID, req.Approved)
	if st.PendingApproval.DecisionHash == decisionHash {
		return &ResumeResult{
			Resumed:   true,
			SessionID: sessionID,
			PlanID:    planID,
			StepID:    stepID,
			Status:    "idempotent",
			Message:   "duplicate approval request ignored",
		}, nil
	}

	now := time.Now().UTC()
	st.PendingApproval.Approved = &req.Approved
	st.PendingApproval.Reason = strings.TrimSpace(req.Reason)
	st.PendingApproval.ResolvedAt = now
	st.PendingApproval.DecisionHash = decisionHash
	if req.Approved {
		st.PendingApproval.Status = "approved"
		st.Status = runtime.ExecutionStatusRunning
		st.Phase = "resume_requested"
	} else {
		st.PendingApproval.Status = "rejected"
		st.Status = runtime.ExecutionStatusFailed
		st.Phase = "cancelled"
	}
	if err := o.executions.Save(ctx, *st); err != nil {
		return nil, err
	}

	return &ResumeResult{
		Resumed:     req.Approved,
		Interrupted: !req.Approved,
		SessionID:   sessionID,
		PlanID:      planID,
		StepID:      stepID,
		Status:      st.PendingApproval.Status,
		Message:     approvalMessage(req.Approved),
	}, nil
}

func (o *Orchestrator) planAndReply(ctx context.Context, message string, rewritten rewrite.Output, runtimeCtx RuntimeContext, meta events.EventMeta, emit StreamEmitter, sessionID string) (string, error) {
	if o.planner != nil {
		emitEvent(emit, events.PlannerState, meta, map[string]any{
			"status":               "planning",
			"user_visible_summary": "正在根据 Rewrite 结果整理执行计划。",
		})
		decision, err := o.planner.Plan(ctx, planner.Input{
			Message: message,
			Rewrite: rewritten,
		})
		if err == nil {
			switch decision.Type {
			case planner.DecisionClarify:
				emitEvent(emit, events.ClarifyRequired, meta, map[string]any{
					"title":      "需要你补充信息",
					"message":    decision.Message,
					"candidates": decision.Candidates,
					"kind":       "clarify",
				})
				return decision.Message, nil
			case planner.DecisionDirectReply:
				return decision.Message, nil
			case planner.DecisionPlan:
				if decision.Plan != nil {
					meta.PlanID = decision.Plan.PlanID
					emitEvent(emit, events.PlanCreated, meta, map[string]any{
						"plan":                 decision.Plan,
						"user_visible_summary": decision.Narrative,
					})
					if o.executor != nil {
						prepared, prepErr := o.executor.SavePreparedState(ctx, executor.Request{
							TraceID:   meta.TraceID,
							SessionID: sessionID,
							Message:   message,
							Plan:      *decision.Plan,
							RuntimeContext: runtime.ContextSnapshot{
								Scene:       runtimeCtx.Scene,
								Route:       runtimeCtx.Route,
								ProjectID:   runtimeCtx.ProjectID,
								CurrentPage: runtimeCtx.CurrentPage,
								ResourceIDs: selectedResourceIDs(runtimeCtx.SelectedResources),
							},
						})
						if prepErr == nil && prepared != nil {
							for _, step := range prepared.Steps {
								stepMeta := meta
								stepMeta.StepID = step.StepID
								emitEvent(emit, events.StepUpdate, stepMeta, map[string]any{
									"plan_id":              prepared.State.PlanID,
									"step_id":              step.StepID,
									"status":               step.Status,
									"title":                prepared.State.Steps[step.StepID].Title,
									"expert":               prepared.State.Steps[step.StepID].Expert,
									"user_visible_summary": step.Summary,
								})
							}
						}
					}
				}
			}
		}
	}
	return o.generateReply(ctx, message)
}

func (o *Orchestrator) generateReply(ctx context.Context, message string) (string, error) {
	model, err := NewToolCallingChatModel(ctx)
	if err != nil {
		return "", err
	}
	resp, err := model.Generate(ctx, []*schema.Message{
		schema.SystemMessage("You are the OpsPilot AI assistant. Keep answers concise, factual, and action-oriented."),
		schema.UserMessage(message),
	})
	if err != nil {
		return "", err
	}
	content := strings.TrimSpace(resp.Content)
	if content == "" {
		return "我已经接收到你的请求，但当前没有生成可展示的回答。", nil
	}
	return content, nil
}

func (o *Orchestrator) sessionMessages(ctx context.Context, sessionID string) []map[string]any {
	if o.sessions == nil {
		return []map[string]any{}
	}
	snapshot, err := o.sessions.Load(ctx, sessionID)
	if err != nil || snapshot == nil {
		return []map[string]any{}
	}
	out := make([]map[string]any, 0, len(snapshot.Messages))
	for i, msg := range snapshot.Messages {
		out = append(out, map[string]any{
			"id":        fmt.Sprintf("%s-%d", sessionID, i+1),
			"role":      msg.Role,
			"content":   msg.Content,
			"timestamp": msg.Timestamp.Format(time.RFC3339),
		})
	}
	return out
}

func emitEvent(emit StreamEmitter, name events.Name, meta events.EventMeta, data map[string]any) {
	if emit == nil {
		return
	}
	emit(StreamEvent{
		Type:     name,
		Audience: events.AudienceUser,
		Meta:     meta.WithDefaults(),
		Data:     data,
	})
}

func deriveSessionTitle(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return "新对话"
	}
	runes := []rune(message)
	if len(runes) > 24 {
		return string(runes[:24])
	}
	return message
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func approvalMessage(approved bool) string {
	if approved {
		return "审批已记录，执行链路可继续恢复"
	}
	return "审批已拒绝，当前执行已保持中断"
}

func toRewriteResources(items []SelectedResource) []rewrite.SelectedResource {
	out := make([]rewrite.SelectedResource, 0, len(items))
	for _, item := range items {
		out = append(out, rewrite.SelectedResource{
			Type: item.Type,
			ID:   item.ID,
			Name: item.Name,
		})
	}
	return out
}

func selectedResourceIDs(items []SelectedResource) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		out = append(out, item.ID)
	}
	return out
}
