package rewrite

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cloudwego/eino/adk"
)

type Output struct {
	RawUserInput      string            `json:"raw_user_input,omitempty"`
	NormalizedRequest NormalizedRequest `json:"normalized_request,omitempty"`
	Ambiguities       []string          `json:"ambiguities,omitempty"`
	Assumptions       []string          `json:"assumptions,omitempty"`
	NormalizedGoal    string            `json:"normalized_goal"`
	OperationMode     string            `json:"operation_mode"`
	ResourceHints     ResourceHints     `json:"resource_hints,omitempty"`
	DomainHints       []string          `json:"domain_hints,omitempty"`
	AmbiguityFlags    []string          `json:"ambiguity_flags,omitempty"`
	Narrative         string            `json:"narrative"`
}

type ResourceHints struct {
	ServiceName string `json:"service_name,omitempty"`
	ServiceID   int    `json:"service_id,omitempty"`
	ClusterName string `json:"cluster_name,omitempty"`
	ClusterID   int    `json:"cluster_id,omitempty"`
	HostName    string `json:"host_name,omitempty"`
	HostID      int    `json:"host_id,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

type NormalizedRequest struct {
	Intent         string           `json:"intent,omitempty"`
	Targets        []RequestTarget  `json:"targets,omitempty"`
	Symptoms       []RequestSymptom `json:"symptoms,omitempty"`
	Context        RequestContext   `json:"context,omitempty"`
	UserHypotheses []string         `json:"user_hypotheses,omitempty"`
	Priority       string           `json:"priority,omitempty"`
}

type RequestTarget struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type RequestSymptom struct {
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

type RequestContext struct {
	TimeHint     string `json:"time_hint,omitempty"`
	TriggerEvent string `json:"trigger_event,omitempty"`
	Environment  string `json:"environment,omitempty"`
}

type Input struct {
	Message           string
	Scene             string
	CurrentPage       string
	SelectedResources []SelectedResource
}

type SelectedResource struct {
	Type string
	ID   string
	Name string
}

type Rewriter struct {
	runner *adk.Runner
}

func New(runner *adk.Runner) *Rewriter {
	return &Rewriter{runner: runner}
}

func (r *Rewriter) Rewrite(ctx context.Context, in Input) (Output, error) {
	base := buildBaseOutput(in)

	if r == nil || r.runner == nil {
		return buildFallbackOutput(base, "rewrite_runner_unavailable"), nil
	}
	raw, err := runADKRewrite(ctx, r.runner, buildPromptInput(in))
	if err != nil {
		return buildFallbackOutput(base, "rewrite_model_unavailable"), nil
	}

	var parsed Output
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err != nil {
		return buildFallbackOutput(base, "rewrite_invalid_json"), nil
	}
	return normalizeOutput(base, parsed), nil
}

func buildBaseOutput(in Input) Output {
	message := strings.TrimSpace(in.Message)
	hints := detectResourceHints(in.SelectedResources)
	return Output{
		RawUserInput:   message,
		NormalizedGoal: message,
		OperationMode:  "query",
		ResourceHints:  hints,
		NormalizedRequest: NormalizedRequest{
			Intent:  "user_request",
			Targets: buildTargets(in.SelectedResources),
		},
		Narrative: buildNarrative(message, "query", hints, nil, nil),
	}
}

func buildPromptInput(in Input) string {
	var b strings.Builder
	b.WriteString("message: ")
	b.WriteString(strings.TrimSpace(in.Message))
	if strings.TrimSpace(in.Scene) != "" {
		b.WriteString("\nscene: ")
		b.WriteString(strings.TrimSpace(in.Scene))
	}
	if strings.TrimSpace(in.CurrentPage) != "" {
		b.WriteString("\ncurrent_page: ")
		b.WriteString(strings.TrimSpace(in.CurrentPage))
	}
	if len(in.SelectedResources) > 0 {
		b.WriteString("\nselected_resources:")
		for _, item := range in.SelectedResources {
			b.WriteString("\n- type=")
			b.WriteString(item.Type)
			if item.ID != "" {
				b.WriteString(", id=")
				b.WriteString(item.ID)
			}
			if item.Name != "" {
				b.WriteString(", name=")
				b.WriteString(item.Name)
			}
		}
	}
	return b.String()
}

func mergeNormalizedRequest(parsed, base NormalizedRequest) NormalizedRequest {
	if strings.TrimSpace(parsed.Intent) == "" {
		parsed.Intent = base.Intent
	}
	if len(parsed.Targets) == 0 {
		parsed.Targets = base.Targets
	}
	if len(parsed.Symptoms) == 0 {
		parsed.Symptoms = base.Symptoms
	}
	if parsed.Context == (RequestContext{}) {
		parsed.Context = base.Context
	}
	if len(parsed.UserHypotheses) == 0 {
		parsed.UserHypotheses = base.UserHypotheses
	}
	if strings.TrimSpace(parsed.Priority) == "" {
		parsed.Priority = base.Priority
	}
	return parsed
}

func normalizeOutput(base, parsed Output) Output {
	if strings.TrimSpace(parsed.NormalizedGoal) == "" {
		parsed.NormalizedGoal = base.NormalizedGoal
	}
	parsed.RawUserInput = firstNonEmpty(parsed.RawUserInput, base.RawUserInput)
	parsed.OperationMode = normalizeMode(parsed.OperationMode, base.OperationMode)
	if parsed.ResourceHints == (ResourceHints{}) {
		parsed.ResourceHints = base.ResourceHints
	}
	parsed.NormalizedRequest = mergeNormalizedRequest(parsed.NormalizedRequest, base.NormalizedRequest)
	parsed.Assumptions = dedupe(parsed.Assumptions)
	parsed.Ambiguities = dedupe(parsed.Ambiguities)
	parsed.DomainHints = dedupe(parsed.DomainHints)
	parsed.AmbiguityFlags = dedupe(parsed.AmbiguityFlags)
	parsed.Narrative = buildNarrative(parsed.NormalizedGoal, parsed.OperationMode, parsed.ResourceHints, parsed.DomainHints, parsed.AmbiguityFlags)
	return parsed
}

func buildFallbackOutput(base Output, reason string) Output {
	base.Assumptions = dedupe(append(base.Assumptions, reason))
	base.Narrative = buildFallbackNarrative(base.NormalizedGoal, reason)
	return base
}

func buildTargets(resources []SelectedResource) []RequestTarget {
	targets := make([]RequestTarget, 0, len(resources))
	for _, item := range resources {
		targetType := strings.ToLower(strings.TrimSpace(item.Type))
		targetName := firstNonEmpty(item.Name, item.ID)
		if targetType == "" || targetName == "" {
			continue
		}
		targets = append(targets, RequestTarget{Type: targetType, Name: targetName})
	}
	return targets
}

func detectResourceHints(resources []SelectedResource) ResourceHints {
	hints := ResourceHints{}
	for _, item := range resources {
		switch strings.ToLower(strings.TrimSpace(item.Type)) {
		case "service":
			hints.ServiceName = firstNonEmpty(item.Name, item.ID)
			hints.ServiceID = parseResourceID(item.ID)
		case "cluster":
			hints.ClusterName = firstNonEmpty(item.Name, item.ID)
			hints.ClusterID = parseResourceID(item.ID)
		case "host":
			hints.HostName = firstNonEmpty(item.Name, item.ID)
			hints.HostID = parseResourceID(item.ID)
		case "namespace":
			hints.Namespace = firstNonEmpty(item.Name, item.ID)
		}
	}
	return hints
}

func buildNarrative(goal, mode string, hints ResourceHints, domains, ambiguity []string) string {
	parts := []string{"用户请求已被规整为可执行任务。", "目标：" + goal + "。", "模式：" + mode + "。"}
	if hints.ServiceName != "" {
		parts = append(parts, "服务线索："+hints.ServiceName+"。")
	}
	if hints.ClusterName != "" {
		parts = append(parts, "集群线索："+hints.ClusterName+"。")
	}
	if hints.ClusterID > 0 {
		parts = append(parts, "集群ID线索："+itoa(hints.ClusterID)+"。")
	}
	if len(domains) > 0 {
		parts = append(parts, "涉及领域："+strings.Join(domains, " / ")+"。")
	}
	if len(ambiguity) > 0 {
		parts = append(parts, "当前仍存在歧义："+strings.Join(ambiguity, ", ")+"。")
	}
	return strings.Join(parts, " ")
}

func buildFallbackNarrative(goal, reason string) string {
	parts := []string{"Rewrite 模型不可用，已将用户原始请求标准化透传给 Planner。", "目标：" + goal + "。"}
	if reason != "" {
		parts = append(parts, "原因："+reason+"。")
	}
	return strings.Join(parts, " ")
}

func normalizeMode(value, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "query", "investigate", "mutate":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return fallback
	}
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

func parseResourceID(raw string) int {
	for _, r := range strings.TrimSpace(raw) {
		if r < '0' || r > '9' {
			return 0
		}
	}
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	value := 0
	for _, r := range raw {
		value = value*10 + int(r-'0')
	}
	return value
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	buf := make([]byte, 0, 12)
	for value > 0 {
		buf = append([]byte{byte('0' + value%10)}, buf...)
		value /= 10
	}
	return string(buf)
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
