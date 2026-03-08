package router

import "github.com/cy77cc/k8s-manage/internal/ai/tools"

// DomainRouteConfig defines keyword-based hints for a tool domain.
type DomainRouteConfig struct {
	Domain   tools.ToolDomain
	Keywords []string
}

// Classification describes the router decision for a user input.
type Classification struct {
	Domain      tools.ToolDomain
	Normalized  string
	Confidence  float64
	MatchedRule string
}
