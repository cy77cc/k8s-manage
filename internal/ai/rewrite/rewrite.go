package rewrite

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
)

type Output struct {
	NormalizedGoal string        `json:"normalized_goal"`
	OperationMode  string        `json:"operation_mode"`
	ResourceHints  ResourceHints `json:"resource_hints,omitempty"`
	DomainHints    []string      `json:"domain_hints,omitempty"`
	AmbiguityFlags []string      `json:"ambiguity_flags,omitempty"`
	Narrative      string        `json:"narrative"`
}

type ResourceHints struct {
	ServiceName string `json:"service_name,omitempty"`
	ClusterName string `json:"cluster_name,omitempty"`
	HostName    string `json:"host_name,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
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

type StageRunner interface {
	Run(ctx context.Context, input string) (string, error)
}

type Rewriter struct {
	runner StageRunner
}

func New(runner StageRunner) *Rewriter {
	return &Rewriter{runner: runner}
}

func (r *Rewriter) Rewrite(ctx context.Context, in Input) (Output, error) {
	out := heuristicRewrite(in)

	if r == nil || r.runner == nil {
		return out, nil
	}
	raw, err := r.runner.Run(ctx, buildPromptInput(in))
	if err != nil {
		return out, nil
	}

	var parsed Output
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err != nil {
		return out, nil
	}
	if strings.TrimSpace(parsed.NormalizedGoal) == "" {
		return out, nil
	}
	parsed.OperationMode = normalizeMode(parsed.OperationMode, out.OperationMode)
	if strings.TrimSpace(parsed.Narrative) == "" {
		parsed.Narrative = out.Narrative
	}
	if len(parsed.DomainHints) == 0 {
		parsed.DomainHints = out.DomainHints
	}
	if len(parsed.AmbiguityFlags) == 0 {
		parsed.AmbiguityFlags = out.AmbiguityFlags
	}
	if parsed.ResourceHints == (ResourceHints{}) {
		parsed.ResourceHints = out.ResourceHints
	}
	return parsed, nil
}

func heuristicRewrite(in Input) Output {
	message := strings.TrimSpace(in.Message)
	mode := detectMode(message)
	hints := detectResourceHints(message, in.SelectedResources)
	domains := detectDomains(message, in.Scene)
	ambiguity := make([]string, 0, 2)
	if hints.ServiceName == "" && len(in.SelectedResources) == 0 {
		ambiguity = append(ambiguity, "resource_target_not_explicit")
	}

	goal := message
	if hints.ServiceName != "" && !strings.Contains(goal, hints.ServiceName) {
		goal = goal + "，目标服务：" + hints.ServiceName
	}

	return Output{
		NormalizedGoal: goal,
		OperationMode:  mode,
		ResourceHints:  hints,
		DomainHints:    domains,
		AmbiguityFlags: ambiguity,
		Narrative:      buildNarrative(goal, mode, hints, domains, ambiguity),
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

func detectMode(message string) string {
	lower := strings.ToLower(message)
	switch {
	case containsAny(lower, "发布", "部署", "回滚", "重启", "删除", "apply", "rollback", "restart"):
		return "mutate"
	case containsAny(lower, "排查", "分析", "诊断", "看看", "查下", "check", "investigate"):
		return "investigate"
	default:
		return "query"
	}
}

func detectDomains(message, scene string) []string {
	lower := strings.ToLower(message + " " + scene)
	out := make([]string, 0, 3)
	appendIf := func(domain string, keywords ...string) {
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				out = append(out, domain)
				return
			}
		}
	}
	appendIf("service", "service", "服务", "api")
	appendIf("k8s", "k8s", "pod", "deployment", "cluster", "namespace")
	appendIf("hostops", "host", "主机", "机器", "节点")
	appendIf("delivery", "deploy", "发布", "rollout", "pipeline", "cicd")
	appendIf("observability", "监控", "latency", "error", "告警", "metrics", "日志")
	if len(out) == 0 {
		out = append(out, "service")
	}
	return dedupe(out)
}

func detectResourceHints(message string, resources []SelectedResource) ResourceHints {
	hints := ResourceHints{}
	for _, item := range resources {
		switch strings.ToLower(strings.TrimSpace(item.Type)) {
		case "service":
			hints.ServiceName = firstNonEmpty(item.Name, item.ID)
		case "cluster":
			hints.ClusterName = firstNonEmpty(item.Name, item.ID)
		case "host":
			hints.HostName = firstNonEmpty(item.Name, item.ID)
		case "namespace":
			hints.Namespace = firstNonEmpty(item.Name, item.ID)
		}
	}

	servicePattern := regexp.MustCompile(`([a-z0-9-]+(?:api|service|svc))`)
	if hints.ServiceName == "" {
		if match := servicePattern.FindString(strings.ToLower(message)); match != "" {
			hints.ServiceName = match
		}
	}
	namespacePattern := regexp.MustCompile(`namespace[:= ]([a-z0-9-]+)`)
	if hints.Namespace == "" {
		if groups := namespacePattern.FindStringSubmatch(strings.ToLower(message)); len(groups) > 1 {
			hints.Namespace = groups[1]
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
	if len(domains) > 0 {
		parts = append(parts, "涉及领域："+strings.Join(domains, " / ")+"。")
	}
	if len(ambiguity) > 0 {
		parts = append(parts, "当前仍存在歧义："+strings.Join(ambiguity, ", ")+"。")
	}
	return strings.Join(parts, " ")
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
