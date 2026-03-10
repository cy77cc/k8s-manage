package events

import (
	"github.com/cy77cc/OpsPilot/internal/ai/executor"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/rewrite"
	"github.com/cy77cc/OpsPilot/internal/ai/runtime"
	"github.com/cy77cc/OpsPilot/internal/ai/summarizer"
)

type RewriteResultPayload struct {
	Rewrite            rewrite.Output `json:"rewrite"`
	UserVisibleSummary string         `json:"user_visible_summary,omitempty"`
}

type PlannerStatePayload struct {
	Status             string `json:"status"`
	UserVisibleSummary string `json:"user_visible_summary,omitempty"`
}

type PlanCreatedPayload struct {
	Plan               *planner.ExecutionPlan `json:"plan,omitempty"`
	UserVisibleSummary string                 `json:"user_visible_summary,omitempty"`
}

type StageDeltaPayload struct {
	Stage        string `json:"stage"`
	ContentChunk string `json:"content_chunk,omitempty"`
	Status       string `json:"status,omitempty"`
	StepID       string `json:"step_id,omitempty"`
	Expert       string `json:"expert,omitempty"`
	Replace      bool   `json:"replace,omitempty"`
}

type StepUpdatePayload struct {
	PlanID             string             `json:"plan_id,omitempty"`
	StepID             string             `json:"step_id,omitempty"`
	Status             runtime.StepStatus `json:"status"`
	Title              string             `json:"title,omitempty"`
	Expert             string             `json:"expert,omitempty"`
	UserVisibleSummary string             `json:"user_visible_summary,omitempty"`
}

type ApprovalRequiredPayload struct {
	SessionID          string                 `json:"session_id,omitempty"`
	PlanID             string                 `json:"plan_id,omitempty"`
	StepID             string                 `json:"step_id,omitempty"`
	Title              string                 `json:"title,omitempty"`
	Risk               string                 `json:"risk,omitempty"`
	Mode               string                 `json:"mode,omitempty"`
	Status             string                 `json:"status,omitempty"`
	UserVisibleSummary string                 `json:"user_visible_summary,omitempty"`
	Resume             executor.ResumeRequest `json:"resume"`
}

type ClarifyRequiredPayload struct {
	Kind       string           `json:"kind,omitempty"`
	Title      string           `json:"title,omitempty"`
	Message    string           `json:"message,omitempty"`
	Candidates []map[string]any `json:"candidates,omitempty"`
}

type ReplanStartedPayload struct {
	Reason         string `json:"reason,omitempty"`
	PreviousPlanID string `json:"previous_plan_id,omitempty"`
}

type DeltaPayload struct {
	ContentChunk string `json:"content_chunk,omitempty"`
}

type SummaryPayload struct {
	Output summarizer.SummaryOutput `json:"output"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}
