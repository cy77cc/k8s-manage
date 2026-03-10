package ai

import "github.com/cy77cc/OpsPilot/internal/ai/events"

type RunRequest struct {
	SessionID      string         `json:"session_id,omitempty"`
	Message        string         `json:"message"`
	RuntimeContext RuntimeContext `json:"runtime_context,omitempty"`
}

type ResumeRequest struct {
	SessionID string `json:"session_id,omitempty"`
	PlanID    string `json:"plan_id,omitempty"`
	StepID    string `json:"step_id,omitempty"`
	Target    string `json:"target,omitempty"`
	Approved  bool   `json:"approved"`
	Reason    string `json:"reason,omitempty"`
}

type RuntimeContext struct {
	Scene             string                 `json:"scene,omitempty"`
	Route             string                 `json:"route,omitempty"`
	ProjectID         string                 `json:"project_id,omitempty"`
	CurrentPage       string                 `json:"current_page,omitempty"`
	SelectedResources []SelectedResource     `json:"selected_resources,omitempty"`
	UserContext       map[string]any         `json:"user_context,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type SelectedResource struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type StreamEvent struct {
	Type     events.Name      `json:"type"`
	Audience events.Audience  `json:"audience"`
	Meta     events.EventMeta `json:"meta"`
	Data     map[string]any   `json:"data,omitempty"`
}

type ResumeResult struct {
	Resumed     bool   `json:"resumed"`
	Interrupted bool   `json:"interrupted,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	PlanID      string `json:"plan_id,omitempty"`
	StepID      string `json:"step_id,omitempty"`
	Status      string `json:"status,omitempty"`
	Message     string `json:"message,omitempty"`
}
