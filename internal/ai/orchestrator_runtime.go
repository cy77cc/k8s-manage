package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
)

func (o *Orchestrator) newExecutorRouter() *ExecutorRouter {
	router := NewExecutorRouter()
	executor := DomainExecutorFunc(func(ctx context.Context, req ExecutionRequest) (ExecutionRecord, error) {
		return o.executeStep(ctx, req)
	})
	for _, domain := range []Domain{DomainPlatform, DomainHost, DomainK8s, DomainService, DomainMonitor} {
		router.Register(domain, executor)
	}
	return router
}

func (o *Orchestrator) emitPlatformEvent(emit func(string, map[string]any) bool, event PlatformEvent) bool {
	payload := event.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	if event.PlanID != "" {
		payload["plan_id"] = event.PlanID
	}
	if event.StepID != "" {
		payload["step_id"] = event.StepID
	}
	if !event.Timestamp.IsZero() {
		payload["timestamp"] = event.Timestamp.UTC().Format(time.RFC3339Nano)
	}
	return emit(event.Type, payload)
}

func (o *Orchestrator) executeStep(ctx context.Context, req ExecutionRequest) (ExecutionRecord, error) {
	if o == nil || o.runner == nil {
		return ExecutionRecord{}, fmt.Errorf("runner not initialized")
	}
	tracker := newToolEventTracker()
	var assistantContent strings.Builder
	var reasoningContent strings.Builder
	record := ExecutionRecord{
		ExecutionID: formatID("exec", time.Now()),
		PlanID:      req.Plan.PlanID,
		StepID:      req.Step.StepID,
		Status:      ExecutionStatusRunning,
		StartedAt:   time.Now(),
	}
	emit := req.Emit
	if emit == nil {
		emit = func(string, map[string]any) bool { return true }
	}
	streamCtx := o.buildToolContext(
		ctx,
		toolUserIDFromContext(req.Context),
		strings.TrimSpace(logic.ToString(req.Context["approval_token"])),
		logic.NormalizeScene(logic.ToString(req.Context["scene"])),
		req.Message,
		req.Context,
		emit,
		tracker,
	)
	iter := o.runner.Query(streamCtx, req.Plan.SessionID, o.buildExecutionPrompt(req))
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if err := o.processADKEvent(emit, tracker, event, &assistantContent, &reasoningContent); err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			record.Status = ExecutionStatusFailed
			record.Issues = append(record.Issues, ExecutionIssue{
				Code:        "stream_interrupted",
				Message:     err.Error(),
				Recoverable: true,
			})
			finished := time.Now()
			record.FinishedAt = &finished
			return record, nil
		}
	}
	record.Status = ExecutionStatusCompleted
	record.Summary = strings.TrimSpace(assistantContent.String())
	if record.Summary == "" {
		record.Summary = strings.TrimSpace(reasoningContent.String())
	}
	if record.Summary == "" {
		record.Summary = "执行完成。"
	}
	record.Evidence = evidenceFromExecution(req.Step, record.Summary, tracker.summary())
	finished := time.Now()
	record.FinishedAt = &finished
	return record, nil
}

func evidenceFromExecution(step PlanStep, summary string, toolSummary toolSummary) []EvidenceItem {
	items := []EvidenceItem{{
		EvidenceID: formatID("ev", time.Now()),
		Type:       EvidenceTypeDiagnosis,
		Title:      step.Title,
		Summary:    summary,
		Severity:   SeverityInfo,
		Data: map[string]any{
			"tool_calls":   toolSummary.Calls,
			"tool_results": toolSummary.Results,
		},
	}}
	if step.Domain == DomainHost && strings.Contains(strings.ToLower(summary), "磁盘") {
		items = append(items, EvidenceItem{
			EvidenceID: formatID("ev", time.Now().Add(time.Nanosecond)),
			Type:       EvidenceTypeDiskUsage,
			Title:      "主机磁盘诊断结果",
			Summary:    summary,
			Severity:   SeverityWarning,
		})
	}
	return items
}

func toolUserIDFromContext(runtime map[string]any) uint64 {
	switch v := runtime["user_id"].(type) {
	case uint64:
		return v
	case int:
		if v >= 0 {
			return uint64(v)
		}
	}
	return 0
}

func (o *Orchestrator) buildExecutionPrompt(req ExecutionRequest) string {
	scene := logic.NormalizeScene(logic.ToString(req.Context["scene"]))
	stepContext := map[string]any{
		"objective": req.Plan.Objective.Summary,
		"step": map[string]any{
			"id":     req.Step.StepID,
			"title":  req.Step.Title,
			"kind":   req.Step.Kind,
			"domain": req.Step.Domain,
			"goal":   req.Step.Goal,
			"inputs": req.Step.Inputs,
		},
		"runtime_context": req.Context,
	}
	task := strings.TrimSpace(fmt.Sprintf(
		"当前只执行这一个步骤，不要自己改写目标。\n领域: %s\n步骤: %s\n目标: %s\n用户原始请求: %s",
		req.Step.Domain,
		req.Step.Title,
		req.Step.Goal,
		req.Message,
	))
	return o.buildPrompt(task, scene, stepContext)
}
