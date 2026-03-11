package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/OpsPilot/internal/ai/events"
	"github.com/cy77cc/OpsPilot/internal/ai/executor"
	"github.com/cy77cc/OpsPilot/internal/ai/experts"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/rewrite"
	"github.com/cy77cc/OpsPilot/internal/ai/runtime"
	"github.com/cy77cc/OpsPilot/internal/ai/state"
	"github.com/cy77cc/OpsPilot/internal/ai/summarizer"
	"github.com/cy77cc/OpsPilot/internal/ai/tools/common"
	"github.com/google/uuid"
)

type StreamEmitter func(StreamEvent) bool

type Orchestrator struct {
	sessions   *state.SessionState
	executions *runtime.ExecutionStore
	rewriter   *rewrite.Rewriter
	planner    *planner.Planner
	executor   *executor.Executor
	summarizer *summarizer.Summarizer
	renderer   *finalAnswerRenderer
	metrics    *AIMetrics
	maxIters   int
}

func NewOrchestrator(sessions *state.SessionState, executions *runtime.ExecutionStore, deps common.PlatformDeps) *Orchestrator {
	out := &Orchestrator{
		sessions:   sessions,
		executions: executions,
		executor:   executor.New(executions),
		renderer:   newFinalAnswerRenderer(),
		metrics:    NewAIMetrics(),
		maxIters:   2,
	}
	bootstrapCtx := context.Background()
	if rewriteModel, err := NewRewriteChatModel(bootstrapCtx); err == nil {
		if stage, stageErr := rewrite.NewWithADK(bootstrapCtx, rewriteModel); stageErr == nil {
			out.rewriter = stage
		}
	}
	if plannerModel, err := NewToolCallingChatModel(bootstrapCtx); err == nil {
		if stage, stageErr := planner.NewWithADK(bootstrapCtx, plannerModel, deps); stageErr == nil {
			out.planner = stage
		}
		if stepRunner, stepErr := executor.NewAgentStepRunner(bootstrapCtx, plannerModel, experts.DefaultRegistry(deps)); stepErr == nil {
			out.executor = executor.New(executions, executor.WithStepRunner(stepRunner))
		}
	}
	if summarizerModel, err := NewSummarizerChatModel(bootstrapCtx); err == nil {
		if stage, stageErr := summarizer.NewWithADK(bootstrapCtx, summarizerModel); stageErr == nil {
			out.summarizer = stage
		}
	}
	return out
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
	streamEmit := emit
	if o.metrics != nil {
		thoughtRun := o.metrics.StartThoughtChainRun()
		if thoughtRun != nil {
			defer thoughtRun.Finalize()
			streamEmit = func(evt StreamEvent) bool {
				thoughtRun.Observe(evt)
				if emit == nil {
					return true
				}
				return emit(evt)
			}
		}
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

	emitEvent(streamEmit, events.Meta, meta, map[string]any{
		"session_id": sessionID,
		"sessionId":  sessionID,
		"trace_id":   traceID,
		"traceId":    traceID,
		"createdAt":  meta.Timestamp.Format(time.RFC3339),
	})

	var rewriteOut rewrite.Output
	if o.rewriter != nil {
		emitStageDelta(streamEmit, meta, "rewrite", "loading", "开始理解你的问题并提取目标线索。", "", "")
		var err error
		rewriteOut, err = o.rewriter.RewriteStream(ctx, rewrite.Input{
			Message:           message,
			Scene:             req.RuntimeContext.Scene,
			CurrentPage:       req.RuntimeContext.CurrentPage,
			SelectedResources: toRewriteResources(req.RuntimeContext.SelectedResources),
		}, func(chunk string) {
			emitStageDelta(streamEmit, meta, "rewrite", "loading", chunk, "", "")
		})
		if err == nil {
			if o.metrics != nil {
				o.metrics.RecordRewrite(rewriteOut)
			}
			emitEvent(streamEmit, events.RewriteResult, meta, map[string]any{
				"rewrite":              rewriteOut,
				"user_visible_summary": rewriteOut.Narrative,
			})
			emitStageDelta(streamEmit, meta, "rewrite", "success", rewriteOut.Narrative, "", "")
		}
	}

	reply, genErr := o.planAndReply(ctx, message, rewriteOut, req.RuntimeContext, meta, streamEmit, sessionID)
	if genErr != nil {
		reply = fmt.Sprintf("AI 编排入口已经切换到新的宿主边界，但当前模型暂不可用：%s", genErr.Error())
		emitDeltaChunks(streamEmit, meta, reply)
	}

	if o.sessions != nil {
		if err := o.sessions.AppendMessage(ctx, sessionID, schema.AssistantMessage(reply, nil)); err != nil {
			return err
		}
	}

	sessionPayload := map[string]any{
		"id":        sessionID,
		"title":     deriveSessionTitle(message),
		"messages":  o.sessionMessages(ctx, sessionID),
		"createdAt": meta.Timestamp.Format(time.RFC3339),
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
	}
	emitEvent(streamEmit, events.Done, meta, map[string]any{
		"session": sessionPayload,
	})

	if o.executions != nil {
		st, err := o.executions.Load(ctx, sessionID)
		if err == nil && st != nil {
			if st.Status == runtime.ExecutionStatusRunning {
				st.Status = runtime.ExecutionStatusCompleted
				st.Phase = "completed"
				_ = o.executions.Save(ctx, *st)
			}
		}
	}
	return nil
}

func (o *Orchestrator) Resume(ctx context.Context, req ResumeRequest) (*ResumeResult, error) {
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if o.executor == nil {
		return &ResumeResult{
			Resumed:   false,
			SessionID: sessionID,
			Status:    "unavailable",
			Message:   "executor is not configured",
		}, nil
	}
	if o.executions != nil {
		if st, err := o.executions.Load(ctx, sessionID); err == nil && st != nil && st.PendingApproval != nil {
			stepID := firstNonEmpty(req.StepID, req.Target, st.PendingApproval.StepID)
			planID := firstNonEmpty(req.PlanID, st.PlanID, st.PendingApproval.PlanID)
			decisionHash := runtime.ApprovalDecisionHash(planID, stepID, req.Approved)
			if st.PendingApproval.DecisionHash == decisionHash {
				result := &ResumeResult{
					Resumed:   true,
					SessionID: sessionID,
					PlanID:    planID,
					StepID:    stepID,
					Status:    "idempotent",
					Message:   "duplicate approval request ignored",
				}
				if o.metrics != nil {
					o.metrics.RecordResume(result.Status, nil)
				}
				return result, nil
			}
		}
	}
	result, err := o.executor.Resume(ctx, executor.ResumeRequest{
		SessionID: sessionID,
		PlanID:    req.PlanID,
		StepID:    firstNonEmpty(req.StepID, req.Target),
		Approved:  req.Approved,
		Reason:    req.Reason,
	})
	if err != nil {
		if o.metrics != nil {
			o.metrics.RecordResume("", err)
		}
		return nil, err
	}
	if result == nil {
		res := &ResumeResult{
			Resumed:   false,
			SessionID: sessionID,
			Status:    "missing",
			Message:   "execution state not found",
		}
		if o.metrics != nil {
			o.metrics.RecordResume(res.Status, nil)
		}
		return res, nil
	}
	if result.State.SessionID == "" {
		res := &ResumeResult{
			Resumed:   false,
			SessionID: sessionID,
			PlanID:    req.PlanID,
			StepID:    firstNonEmpty(req.StepID, req.Target),
			Status:    "noop",
			Message:   "no pending approval for this session",
		}
		if o.metrics != nil {
			o.metrics.RecordResume(res.Status, nil)
		}
		return res, nil
	}
	pendingStepID := ""
	pendingPlanID := ""
	if result.State.PendingApproval != nil {
		pendingStepID = result.State.PendingApproval.StepID
		pendingPlanID = result.State.PendingApproval.PlanID
	}
	stepID := firstNonEmpty(req.StepID, req.Target, pendingStepID)
	planID := firstNonEmpty(req.PlanID, result.State.PlanID, pendingPlanID)
	status := firstNonEmpty(result.State.Phase)
	if approval := result.Approval(); approval != nil && approval.Status != "" {
		status = approval.Status
	}

	res := &ResumeResult{
		Resumed:     req.Approved,
		Interrupted: !req.Approved,
		SessionID:   sessionID,
		PlanID:      planID,
		StepID:      stepID,
		Status:      status,
		Message:     resumeStatusMessage(status, req.Approved),
	}
	if o.metrics != nil {
		o.metrics.RecordResume(res.Status, nil)
	}
	return res, nil
}

func (o *Orchestrator) MetricsSnapshot() AIMetricsSnapshot {
	if o == nil || o.metrics == nil {
		return AIMetricsSnapshot{}
	}
	return o.metrics.Snapshot()
}

func (o *Orchestrator) planAndReply(ctx context.Context, message string, rewritten rewrite.Output, runtimeCtx RuntimeContext, meta events.EventMeta, emit StreamEmitter, sessionID string) (string, error) {
	if o.planner != nil {
		emitEvent(emit, events.PlannerState, meta, map[string]any{
			"status":               "planning",
			"user_visible_summary": "正在根据 Rewrite 结果整理执行计划。",
		})
		emitStageDelta(emit, meta, "plan", "loading", "正在整理目标、资源和执行约束。", "", "")
		decision, err := o.planner.PlanStream(ctx, planner.Input{
			Message: message,
			Rewrite: rewritten,
		}, func(chunk string) {
			emitStageDelta(emit, meta, "plan", "loading", chunk, "", "")
		})
		if err == nil {
			if o.metrics != nil {
				o.metrics.RecordPlanner(decision)
			}
			switch decision.Type {
			case planner.DecisionClarify:
				emitEvent(emit, events.ClarifyRequired, meta, map[string]any{
					"title":      "需要你补充信息",
					"message":    decision.Message,
					"candidates": decision.Candidates,
					"kind":       "clarify",
				})
				emitStageDelta(emit, meta, "plan", "error", decision.Message, "", "")
				emitDeltaChunks(emit, meta, decision.Message)
				return decision.Message, nil
			case planner.DecisionDirectReply:
				emitStageDelta(emit, meta, "plan", "success", firstNonEmpty(decision.Message, decision.Narrative), "", "")
				emitDeltaChunks(emit, meta, decision.Message)
				return decision.Message, nil
			case planner.DecisionPlan:
				if decision.Plan != nil {
					meta.PlanID = decision.Plan.PlanID
					emitEvent(emit, events.PlanCreated, meta, map[string]any{
						"plan":                 decision.Plan,
						"user_visible_summary": decision.Narrative,
					})
					emitStageDelta(emit, meta, "plan", "success", firstNonEmpty(decision.Narrative, decision.Plan.Narrative), "", "")
					if o.executor != nil {
						executed, execErr := o.executor.Run(ctx, executor.Request{
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
							EventMeta: executor.EventMeta{
								SessionID: meta.SessionID,
								TraceID:   meta.TraceID,
								PlanID:    meta.PlanID,
								Iteration: meta.Iteration,
								Timestamp: meta.Timestamp,
							},
							EmitEvent: func(name string, eventMeta executor.EventMeta, data map[string]any) bool {
								emitEvent(emit, events.Name(name), events.EventMeta{
									SessionID: eventMeta.SessionID,
									TraceID:   eventMeta.TraceID,
									PlanID:    eventMeta.PlanID,
									StepID:    eventMeta.StepID,
									Iteration: eventMeta.Iteration,
									Timestamp: eventMeta.Timestamp,
								}, data)
								emitExecuteStageDelta(emit, events.EventMeta{
									SessionID: eventMeta.SessionID,
									TraceID:   eventMeta.TraceID,
									PlanID:    eventMeta.PlanID,
									StepID:    eventMeta.StepID,
									Iteration: eventMeta.Iteration,
									Timestamp: eventMeta.Timestamp,
								}, name, data)
								return true
							},
						})
						if execErr == nil && executed != nil {
							emitStageDelta(emit, meta, "summary", "loading", "正在汇总执行证据并生成结论。", "", "")
							summaryOut := o.summarizeExecution(ctx, message, decision.Plan, executed, func(chunk string) {
								emitStageDelta(emit, meta, "summary", "loading", chunk, "", "")
							})
							emitEvent(emit, events.Summary, meta, map[string]any{
								"output": summaryOut,
							})
							emitStageDelta(emit, meta, "summary", "success", firstNonEmpty(summaryOut.Summary, "已生成结构化结论。"), "", "")
							if summaryOut.NeedMoreInvestigation && meta.Iteration < o.maxIters {
								reason := ""
								if summaryOut.ReplanHint != nil {
									reason = summaryOut.ReplanHint.Reason
								}
								emitEvent(emit, events.ReplanStarted, meta, map[string]any{
									"reason":           reason,
									"previous_plan_id": executed.State.PlanID,
								})
							}
							if body := o.renderAndEmitFinalAnswer(decision.Plan, executed, summaryOut, emit, meta); body != "" {
								return body, nil
							}
						}
					}
				}
			}
		}
	}
	reply, err := o.generateReply(ctx, message)
	if err == nil {
		emitDeltaChunks(emit, meta, reply)
	}
	return reply, err
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

func (o *Orchestrator) renderAndEmitFinalAnswer(plan *planner.ExecutionPlan, result *executor.Result, summaryOut summarizer.SummaryOutput, emit StreamEmitter, meta events.EventMeta) string {
	if o == nil || o.renderer == nil {
		body := firstNonEmpty(summaryOut.Headline, summaryOut.Conclusion, summaryOut.Summary)
		emitDeltaChunks(emit, meta, body)
		return strings.TrimSpace(body)
	}
	paragraphs := o.renderer.Render("", plan, result, summaryOut)
	var builder strings.Builder
	for i, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(paragraph)
		chunk := paragraph
		if i < len(paragraphs)-1 {
			chunk += "\n\n"
		}
		emitEvent(emit, events.Delta, meta, map[string]any{
			"content_chunk": chunk,
			"contentChunk":  chunk,
		})
	}
	return strings.TrimSpace(builder.String())
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

func emitStageDelta(emit StreamEmitter, meta events.EventMeta, stage, status, chunk, stepID, expert string) {
	chunk = strings.TrimSpace(chunk)
	if emit == nil || stage == "" || chunk == "" {
		return
	}
	payload := map[string]any{
		"stage":         stage,
		"status":        status,
		"content_chunk": chunk,
	}
	if stepID != "" {
		payload["step_id"] = stepID
	}
	if expert != "" {
		payload["expert"] = expert
	}
	emitEvent(emit, events.StageDelta, meta, payload)
}

func emitExecuteStageDelta(emit StreamEmitter, meta events.EventMeta, name string, data map[string]any) {
	switch name {
	case string(events.StepUpdate):
		emitStageDelta(
			emit,
			meta,
			"execute",
			stageStatusFromValue(data["status"]),
			firstNonEmpty(stringValue(data["title"]), stringValue(data["user_visible_summary"])),
			stringValue(data["step_id"]),
			stringValue(data["expert"]),
		)
	case string(events.ToolCall):
		emitStageDelta(
			emit,
			meta,
			"execute",
			"loading",
			firstNonEmpty(stringValue(data["summary"]), stringValue(data["tool_name"])),
			stringValue(data["step_id"]),
			firstNonEmpty(stringValue(data["expert"]), stringValue(data["tool_name"])),
		)
	case string(events.ToolResult):
		status := "success"
		if strings.TrimSpace(stringValue(data["error"])) != "" || strings.EqualFold(stringValue(data["status"]), "error") {
			status = "error"
		}
		emitStageDelta(
			emit,
			meta,
			"execute",
			status,
			firstNonEmpty(stringValue(data["summary"]), stringValue(data["error"]), stringValue(data["tool_name"])),
			stringValue(data["step_id"]),
			firstNonEmpty(stringValue(data["expert"]), stringValue(data["tool_name"])),
		)
	}
}

func stageStatusFromValue(value any) string {
	switch strings.TrimSpace(stringValue(value)) {
	case "completed", "success":
		return "success"
	case "failed", "error", "blocked", "cancelled", "rejected":
		return "error"
	default:
		return "loading"
	}
}

func emitDeltaChunks(emit StreamEmitter, meta events.EventMeta, content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	runes := []rune(content)
	const chunkSize = 24
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[start:end])
		emitEvent(emit, events.Delta, meta, map[string]any{
			"content_chunk": chunk,
			"contentChunk":  chunk,
		})
	}
}

func eventsJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
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

func stringValue(value any) string {
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}

func resumeStatusMessage(status string, approved bool) string {
	switch strings.TrimSpace(status) {
	case "approved", "approval_granted":
		return "审批已通过，待审批步骤会继续执行。"
	case "rejected":
		return "审批已拒绝，待审批步骤不会执行，相关下游步骤已被取消或阻断。"
	case "cancelled":
		return "当前执行已取消，后续步骤不会继续执行。"
	case "idempotent":
		if approved {
			return "重复的审批通过请求已忽略，系统不会重复执行该步骤。"
		}
		return "重复的审批拒绝请求已忽略，系统不会再次取消该步骤。"
	case "noop":
		return "当前没有可恢复的待审批步骤。"
	case "missing":
		return "未找到对应的执行状态，无法恢复该步骤。"
	case "unavailable":
		return "当前恢复能力不可用。"
	default:
		if approved {
			return "审批已记录，执行链路可继续恢复。"
		}
		return "审批已拒绝，当前待审批步骤不会继续执行。"
	}
}

func (o *Orchestrator) summarizeExecution(ctx context.Context, message string, plan *planner.ExecutionPlan, result *executor.Result, onDelta func(string)) summarizer.SummaryOutput {
	if result == nil {
		return summarizer.SummaryOutput{
			Summary:   "当前执行结果不可用。",
			Narrative: "Summarizer 未拿到 executor result，因此回退到保守总结。",
		}
	}
	if o.summarizer == nil {
		fallback, _ := summarizer.New(nil).Summarize(context.Background(), summarizer.Input{
			Message: message,
			Plan:    plan,
			State:   result.State,
			Steps:   result.Steps,
		})
		return fallback
	}
	out, err := o.summarizer.SummarizeStream(ctx, summarizer.Input{
		Message: message,
		Plan:    plan,
		State:   result.State,
		Steps:   result.Steps,
	}, onDelta)
	if err != nil {
		fallback, _ := summarizer.New(nil).Summarize(context.Background(), summarizer.Input{
			Message: message,
			Plan:    plan,
			State:   result.State,
			Steps:   result.Steps,
		})
		return fallback
	}
	return out
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
