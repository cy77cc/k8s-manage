package graph

import "github.com/cloudwego/eino/schema"

type ActionInput struct {
	SessionID string
	Message   string
	UserID    uint64
	Context   map[string]any
}

type ToolCallResult struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

type InterruptInfo struct {
	Type     string         `json:"type"`
	ToolName string         `json:"tool_name,omitempty"`
	Reason   string         `json:"reason,omitempty"`
	Preview  map[string]any `json:"preview,omitempty"`
}

type ActionOutput struct {
	Response  string           `json:"response"`
	ToolCalls []ToolCallResult `json:"tool_calls,omitempty"`
	Interrupt *InterruptInfo   `json:"interrupt,omitempty"`
}

type ValidationError struct {
	ToolName string `json:"tool_name,omitempty"`
	Field    string `json:"field,omitempty"`
	Message  string `json:"message"`
}

type ToolResult struct {
	CallID   string `json:"call_id"`
	ToolName string `json:"tool_name"`
	Content  string `json:"content,omitempty"`
}

type GraphState struct {
	Messages         []*schema.Message
	PendingToolCalls []schema.ToolCall
	ToolResults      map[string]ToolResult
	ValidationErrors []ValidationError
}

func NewGraphState() *GraphState {
	return &GraphState{
		Messages:         make([]*schema.Message, 0, 4),
		PendingToolCalls: make([]schema.ToolCall, 0, 2),
		ToolResults:      make(map[string]ToolResult),
		ValidationErrors: make([]ValidationError, 0, 2),
	}
}

func (s *GraphState) AddMessage(msg *schema.Message) {
	if s == nil || msg == nil {
		return
	}
	s.Messages = append(s.Messages, msg)
}

func (s *GraphState) SetPendingToolCalls(calls []schema.ToolCall) {
	if s == nil {
		return
	}
	s.PendingToolCalls = append([]schema.ToolCall(nil), calls...)
}

func (s *GraphState) AddValidationError(err ValidationError) {
	if s == nil {
		return
	}
	s.ValidationErrors = append(s.ValidationErrors, err)
}

func (s *GraphState) AddToolResult(result ToolResult) {
	if s == nil {
		return
	}
	s.ToolResults[result.CallID] = result
}

func (s *GraphState) HasValidationErrors() bool {
	return s != nil && len(s.ValidationErrors) > 0
}
