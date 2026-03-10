package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/google/uuid"

	"github.com/cy77cc/OpsPilot/internal/ai/rewrite"
)

type Planner struct {
	runner *adk.Runner
}

type Input struct {
	Message string
	Rewrite rewrite.Output
}

type DecisionType string

const (
	DecisionClarify     DecisionType = "clarify"
	DecisionReject      DecisionType = "reject"
	DecisionDirectReply DecisionType = "direct_reply"
	DecisionPlan        DecisionType = "plan"
)

type Decision struct {
	Type       DecisionType     `json:"type"`
	Message    string           `json:"message,omitempty"`
	Reason     string           `json:"reason,omitempty"`
	Candidates []map[string]any `json:"candidates,omitempty"`
	Plan       *ExecutionPlan   `json:"plan,omitempty"`
	Narrative  string           `json:"narrative"`
}

type ExecutionPlan struct {
	PlanID    string            `json:"plan_id"`
	Goal      string            `json:"goal"`
	Resolved  ResolvedResources `json:"resolved"`
	Narrative string            `json:"narrative"`
	Steps     []PlanStep        `json:"steps"`
}

type ResolvedResources struct {
	ServiceName string   `json:"service_name,omitempty"`
	ClusterName string   `json:"cluster_name,omitempty"`
	HostNames   []string `json:"host_names,omitempty"`
	Namespace   string   `json:"namespace,omitempty"`
}

type PlanStep struct {
	StepID    string         `json:"step_id"`
	Title     string         `json:"title"`
	Expert    string         `json:"expert"`
	Intent    string         `json:"intent"`
	Task      string         `json:"task"`
	Input     map[string]any `json:"input,omitempty"`
	DependsOn []string       `json:"depends_on,omitempty"`
	Mode      string         `json:"mode"`
	Risk      string         `json:"risk"`
	Narrative string         `json:"narrative,omitempty"`
}

func New(runner *adk.Runner) *Planner {
	return &Planner{runner: runner}
}

func (p *Planner) Plan(ctx context.Context, in Input) (Decision, error) {
	base := buildBaseDecision(in)

	if p == nil || p.runner == nil {
		return base, nil
	}
	raw, err := runADKPlanner(ctx, p.runner, buildPromptInput(in))
	if err != nil {
		return base, nil
	}

	parsed, err := ParseDecision(strings.TrimSpace(raw))
	if err != nil {
		return base, nil
	}
	return normalizeDecision(base, parsed), nil
}

func ParseDecision(raw string) (Decision, error) {
	var out Decision
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return Decision{}, err
	}
	if out.Type == "" {
		return Decision{}, fmt.Errorf("planner decision missing type")
	}
	return out, nil
}

func buildBaseDecision(in Input) Decision {
	rewritten := in.Rewrite
	ambiguities := dedupe(append(append([]string(nil), rewritten.AmbiguityFlags...), rewritten.Ambiguities...))
	if len(ambiguities) > 0 {
		return Decision{
			Type:       DecisionClarify,
			Message:    "我需要先确认目标资源后再继续规划。",
			Narrative:  "Rewrite 输出仍有未消解歧义，Planner 先请求补充信息。",
			Candidates: buildClarifyCandidates(ambiguities),
		}
	}

	planID := uuid.NewString()
	goal := firstNonEmpty(rewritten.NormalizedGoal, strings.TrimSpace(in.Message))
	mode, risk := normalizeStepMode(rewritten.OperationMode)
	expert := pickPrimaryExpert(rewritten)
	return Decision{
		Type:      DecisionPlan,
		Narrative: "Planner 模型不可用时，使用最小结构化计划继续交给执行层处理。",
		Plan: &ExecutionPlan{
			PlanID: planID,
			Goal:   goal,
			Resolved: ResolvedResources{
				ServiceName: rewritten.ResourceHints.ServiceName,
				ClusterName: rewritten.ResourceHints.ClusterName,
				Namespace:   rewritten.ResourceHints.Namespace,
				HostNames:   collectHostNames(rewritten),
			},
			Narrative: "该计划是 Planner 失败时的最小兜底结构，保留用户目标与已知资源线索。",
			Steps: []PlanStep{
				{
					StepID:    "step-1",
					Title:     "处理用户请求",
					Expert:    expert,
					Intent:    "handle_request",
					Task:      goal,
					Mode:      mode,
					Risk:      risk,
					Narrative: goal,
					Input: map[string]any{
						"message":            strings.TrimSpace(in.Message),
						"normalized_request": rewritten.NormalizedRequest,
						"resource_hints":     rewritten.ResourceHints,
					},
				},
			},
		},
	}
}

func buildPromptInput(in Input) string {
	data, _ := json.Marshal(in.Rewrite)
	return "message: " + strings.TrimSpace(in.Message) + "\nrewrite: " + string(data)
}

func normalizeDecision(base, parsed Decision) Decision {
	if parsed.Type == "" {
		return base
	}
	if strings.TrimSpace(parsed.Narrative) == "" {
		parsed.Narrative = base.Narrative
	}
	if parsed.Type == DecisionPlan && parsed.Plan == nil {
		parsed.Plan = base.Plan
	}
	if parsed.Type == DecisionClarify && strings.TrimSpace(parsed.Message) == "" {
		parsed.Message = base.Message
	}
	if parsed.Type == DecisionDirectReply && strings.TrimSpace(parsed.Message) == "" {
		parsed.Message = base.Message
	}
	if parsed.Type == DecisionPlan {
		parsed.Plan = normalizePlan(base.Plan, parsed.Plan)
	}
	return parsed
}

func normalizePlan(base, parsed *ExecutionPlan) *ExecutionPlan {
	if parsed == nil {
		return base
	}
	if base == nil {
		return parsed
	}
	if strings.TrimSpace(parsed.PlanID) == "" {
		parsed.PlanID = base.PlanID
	}
	if strings.TrimSpace(parsed.Goal) == "" {
		parsed.Goal = base.Goal
	}
	parsed.Resolved.ServiceName = firstNonEmpty(parsed.Resolved.ServiceName, base.Resolved.ServiceName)
	parsed.Resolved.ClusterName = firstNonEmpty(parsed.Resolved.ClusterName, base.Resolved.ClusterName)
	parsed.Resolved.Namespace = firstNonEmpty(parsed.Resolved.Namespace, base.Resolved.Namespace)
	if len(parsed.Resolved.HostNames) == 0 {
		parsed.Resolved.HostNames = base.Resolved.HostNames
	}
	if strings.TrimSpace(parsed.Narrative) == "" {
		parsed.Narrative = base.Narrative
	}
	if len(parsed.Steps) == 0 {
		parsed.Steps = base.Steps
	}
	return parsed
}

func buildClarifyCandidates(ambiguities []string) []map[string]any {
	out := make([]map[string]any, 0, len(ambiguities))
	for _, item := range ambiguities {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, map[string]any{
			"kind":    "ambiguity",
			"message": item,
		})
	}
	return out
}

func normalizeStepMode(mode string) (string, string) {
	if strings.TrimSpace(mode) == "mutate" {
		return "mutating", "high"
	}
	return "readonly", "low"
}

func pickPrimaryExpert(r rewrite.Output) string {
	for _, domain := range r.DomainHints {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			return domain
		}
	}
	for _, target := range r.NormalizedRequest.Targets {
		switch strings.TrimSpace(target.Type) {
		case "host":
			return "hostops"
		case "cluster", "namespace", "pod", "deployment":
			return "k8s"
		case "pipeline":
			return "delivery"
		case "service":
			return "service"
		}
	}
	return "service"
}

func collectHostNames(r rewrite.Output) []string {
	if strings.TrimSpace(r.ResourceHints.HostName) != "" {
		return []string{strings.TrimSpace(r.ResourceHints.HostName)}
	}
	hosts := make([]string, 0, len(r.NormalizedRequest.Targets))
	for _, target := range r.NormalizedRequest.Targets {
		if strings.TrimSpace(target.Type) != "host" {
			continue
		}
		name := strings.TrimSpace(target.Name)
		if name == "" {
			continue
		}
		hosts = append(hosts, name)
	}
	return dedupe(hosts)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func dedupe(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
