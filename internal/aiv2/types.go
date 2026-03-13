package aiv2

import (
	"time"

	"github.com/cloudwego/eino/schema"
	legacyai "github.com/cy77cc/OpsPilot/internal/ai"
)

const runtimeMode = "aiv2"

type emitFunc = legacyai.StreamEmitter

type streamEnvelope struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}

type ToolPolicy struct {
	Name             string `json:"name"`
	Expert           string `json:"expert,omitempty"`
	Mode             string `json:"mode"`
	Risk             string `json:"risk"`
	ApprovalRequired bool   `json:"approval_required"`
}

type ApprovalInterruptInfo struct {
	ToolName        string         `json:"tool_name"`
	Expert          string         `json:"expert,omitempty"`
	ArgumentsInJSON string         `json:"arguments_in_json"`
	ToolCallID      string         `json:"tool_call_id,omitempty"`
	Summary         string         `json:"summary,omitempty"`
	Risk            string         `json:"risk,omitempty"`
	Mode            string         `json:"mode,omitempty"`
	SessionID       string         `json:"session_id,omitempty"`
	TurnID          string         `json:"turn_id,omitempty"`
	RuntimeContext  map[string]any `json:"runtime_context,omitempty"`
}

type ApprovalDecision struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason,omitempty"`
}

type PendingApproval struct {
	CheckPointID string    `json:"checkpoint_id"`
	InterruptID  string    `json:"interrupt_id"`
	SessionID    string    `json:"session_id"`
	TurnID       string    `json:"turn_id"`
	TraceID      string    `json:"trace_id"`
	ToolName     string    `json:"tool_name"`
	Expert       string    `json:"expert,omitempty"`
	ToolArgs     string    `json:"tool_args"`
	Summary      string    `json:"summary"`
	Risk         string    `json:"risk,omitempty"`
	Mode         string    `json:"mode,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type runtimeContextKey struct{}

func init() {
	schema.Register[*ApprovalInterruptInfo]()
	schema.Register[*ApprovalDecision]()
}
