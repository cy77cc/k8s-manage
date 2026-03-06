package ai

type AgentResult struct {
	Type     string
	Content  string
	ToolName string
	ToolData map[string]any
	Ask      *AskRequest
}

type AskRequest struct {
	ID          string
	Title       string
	Description string
	Risk        string
	Details     map[string]any
}
