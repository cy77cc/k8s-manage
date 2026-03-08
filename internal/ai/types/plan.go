package types

import (
	"fmt"
	"regexp"
)

type Domain string

const (
	DomainGeneral        Domain = "general"
	DomainInfrastructure Domain = "infrastructure"
	DomainService        Domain = "service"
	DomainCICD           Domain = "cicd"
	DomainMonitor        Domain = "monitor"
	DomainConfig         Domain = "config"
	DomainUser           Domain = "user"
)

var stepIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type DomainPlan struct {
	Domain Domain     `json:"domain"`
	Steps  []PlanStep `json:"steps"`
}

type PlanStep struct {
	ID        string         `json:"id"`
	Tool      string         `json:"tool"`
	Params    map[string]any `json:"params,omitempty"`
	DependsOn []string       `json:"depends_on,omitempty"`
	Produces  []string       `json:"produces,omitempty"`
	Requires  []string       `json:"requires,omitempty"`
}

type StepResult struct {
	StepID string         `json:"step_id"`
	Output map[string]any `json:"output,omitempty"`
	Error  string         `json:"error,omitempty"`
}

func (p DomainPlan) Validate() error {
	if p.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if len(p.Steps) == 0 {
		return fmt.Errorf("steps must not be empty")
	}
	seen := make(map[string]struct{}, len(p.Steps))
	for _, step := range p.Steps {
		if !stepIDPattern.MatchString(step.ID) {
			return fmt.Errorf("step id must be snake_case: %s", step.ID)
		}
		if _, ok := seen[step.ID]; ok {
			return fmt.Errorf("duplicate step id: %s", step.ID)
		}
		seen[step.ID] = struct{}{}
		if step.Tool == "" {
			return fmt.Errorf("tool is required for step %s", step.ID)
		}
	}
	return nil
}
