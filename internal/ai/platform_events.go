package ai

import "time"

const (
	EventMeta             = "meta"
	EventPlanCreated      = "plan_created"
	EventStepStatus       = "step_status"
	EventToolCall         = "tool_call"
	EventToolResult       = "tool_result"
	EventEvidence         = "evidence"
	EventAskUser          = "ask_user"
	EventApprovalRequired = "approval_required"
	EventReplanDecision   = "replan_decision"
	EventSummary          = "summary"
	EventNextActions      = "next_actions"
)

type PlatformEvent struct {
	Type      string         `json:"type"`
	SessionID string         `json:"session_id,omitempty"`
	PlanID    string         `json:"plan_id,omitempty"`
	StepID    string         `json:"step_id,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type PlanStepView struct {
	StepID string `json:"step_id"`
	Title  string `json:"title"`
	Kind   string `json:"kind"`
	Domain string `json:"domain"`
	Status string `json:"status"`
}

type PlatformEventProjector struct{}

func NewPlatformEventProjector() *PlatformEventProjector { return &PlatformEventProjector{} }

func (p *PlatformEventProjector) PlanCreated(plan Plan) PlatformEvent {
	steps := make([]PlanStepView, 0, len(plan.Steps))
	for _, step := range plan.Steps {
		steps = append(steps, PlanStepView{
			StepID: step.StepID,
			Title:  step.Title,
			Kind:   string(step.Kind),
			Domain: string(step.Domain),
			Status: string(step.Status),
		})
	}
	return PlatformEvent{
		Type:      EventPlanCreated,
		SessionID: plan.SessionID,
		PlanID:    plan.PlanID,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"objective": plan.Objective.Summary,
			"steps":     steps,
		},
	}
}

func (p *PlatformEventProjector) StepStatus(plan Plan, step PlanStep, status StepStatus) PlatformEvent {
	return PlatformEvent{
		Type:      EventStepStatus,
		SessionID: plan.SessionID,
		PlanID:    plan.PlanID,
		StepID:    step.StepID,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"title":  step.Title,
			"kind":   string(step.Kind),
			"domain": string(step.Domain),
			"status": string(status),
		},
	}
}

func (p *PlatformEventProjector) Evidence(plan Plan, step PlanStep, item EvidenceItem) PlatformEvent {
	return PlatformEvent{
		Type:      EventEvidence,
		SessionID: plan.SessionID,
		PlanID:    plan.PlanID,
		StepID:    step.StepID,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"evidence_id": item.EvidenceID,
			"type":        string(item.Type),
			"title":       item.Title,
			"summary":     item.Summary,
			"severity":    string(item.Severity),
			"data":        item.Data,
		},
	}
}

func (p *PlatformEventProjector) Replan(plan Plan, decision ReplanDecision) PlatformEvent {
	return PlatformEvent{
		Type:      EventReplanDecision,
		SessionID: plan.SessionID,
		PlanID:    plan.PlanID,
		StepID:    decision.BasedOnStepID,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"outcome":   string(decision.Outcome),
			"rationale": decision.Rationale,
		},
	}
}

func (p *PlatformEventProjector) Summary(plan Plan, outcome FinalOutcome) PlatformEvent {
	return PlatformEvent{
		Type:      EventSummary,
		SessionID: plan.SessionID,
		PlanID:    plan.PlanID,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"status":       outcome.Status,
			"summary":      outcome.Summary,
			"key_findings": outcome.KeyFindings,
		},
	}
}

func (p *PlatformEventProjector) NextActions(plan Plan, actions []NextAction) PlatformEvent {
	return PlatformEvent{
		Type:      EventNextActions,
		SessionID: plan.SessionID,
		PlanID:    plan.PlanID,
		Timestamp: time.Now(),
		Payload: map[string]any{
			"actions": actions,
		},
	}
}
