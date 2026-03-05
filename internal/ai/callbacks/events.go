package callbacks

import "time"

// ToolCallEvent describes tool-call lifecycle events.
type ToolCallEvent struct {
	Tool      string         `json:"tool"`
	CallID    string         `json:"call_id"`
	Arguments map[string]any `json:"arguments,omitempty"`
	Result    any            `json:"result,omitempty"`
	Error     string         `json:"error,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Duration  time.Duration  `json:"duration,omitempty"`
}

// ExpertProgressEvent describes expert execution progress.
type ExpertProgressEvent struct {
	Expert     string `json:"expert"`
	Status     string `json:"status"` // running, done, failed
	Task       string `json:"task,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	Error      string `json:"error,omitempty"`
}

// StreamEvent describes token-level stream activity.
type StreamEvent struct {
	Type      string    `json:"type"` // delta, thinking_delta, done, error
	Content   string    `json:"content,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
