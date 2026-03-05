package graph

import "github.com/cy77cc/k8s-manage/internal/ai/experts"

// GraphInput is the shared state flowing through expert graph nodes.
type GraphInput struct {
	Message        string
	Request        *experts.ExecuteRequest
	PrimaryOutput  string
	HelperRequests []experts.HelperRequest
	HelperResults  []experts.ExpertResult
	Strategy       experts.ExecutionStrategy
	PrimaryStream  string
	HelperStreams  []string
}

// GraphOutput is the final graph result projected to caller.
type GraphOutput struct {
	Response string
	Results  []experts.ExpertResult
	Metadata map[string]any
}
