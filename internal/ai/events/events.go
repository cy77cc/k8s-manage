package events

import "time"

type Audience string

const (
	AudienceUser  Audience = "user"
	AudienceDebug Audience = "debug"
)

type Name string

const (
	Meta             Name = "meta"
	RewriteResult    Name = "rewrite_result"
	PlannerState     Name = "planner_state"
	PlanCreated      Name = "plan_created"
	StageDelta       Name = "stage_delta"
	StepUpdate       Name = "step_update"
	ApprovalRequired Name = "approval_required"
	ClarifyRequired  Name = "clarify_required"
	ReplanStarted    Name = "replan_started"
	Delta            Name = "delta"
	Summary          Name = "summary"
	Done             Name = "done"
	Error            Name = "error"
	ToolCall         Name = "tool_call"
	ToolResult       Name = "tool_result"
	Heartbeat        Name = "heartbeat"
)

type EventMeta struct {
	SessionID string    `json:"session_id,omitempty"`
	TraceID   string    `json:"trace_id,omitempty"`
	PlanID    string    `json:"plan_id,omitempty"`
	StepID    string    `json:"step_id,omitempty"`
	Iteration int       `json:"iteration,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func (m EventMeta) WithDefaults() EventMeta {
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now().UTC()
	}
	if m.Iteration == 0 {
		m.Iteration = 1
	}
	return m
}
