package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
)

type PlanningRequest struct {
	SessionID string
	Message   string
	Context   map[string]any
}

type textGenerator interface {
	Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
}

type Planner struct {
	model textGenerator
}

func NewPlanner(model textGenerator) *Planner { return &Planner{model: model} }

func (p *Planner) BuildPlan(ctx context.Context, req PlanningRequest) (Plan, error) {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return Plan{}, fmt.Errorf("message is required")
	}
	if p != nil && p.model != nil {
		if plan, err := p.buildPlanWithModel(ctx, req); err == nil && len(plan.Steps) > 0 {
			return plan, nil
		}
	}
	domain := detectDomain(message, req.Context)
	steps := buildDomainSteps(domain, message)
	return Plan{
		PlanID:    formatID("plan", time.Now()),
		SessionID: strings.TrimSpace(req.SessionID),
		Objective: Objective{
			Summary:    message,
			UserIntent: "operate",
			Urgency:    "medium",
			SuccessCriteria: []string{
				"形成明确计划",
				"完成至少一个领域步骤执行",
				"给出可执行下一步",
			},
		},
		Steps:     steps,
		Status:    "active",
		CreatedAt: time.Now(),
	}, nil
}

func (p *Planner) buildPlanWithModel(ctx context.Context, req PlanningRequest) (Plan, error) {
	prompt := strings.TrimSpace(`你是一个 AIOps Planner。请基于用户目标生成结构化执行计划。

输出必须是 JSON 对象，格式如下：
{
  "objective": "一句话目标",
  "steps": [
    {
      "title": "步骤标题",
      "kind": "host-identification|host-diagnosis|k8s-diagnosis|service-diagnosis|monitor-investigation|recommend-action",
      "domain": "platform|host|k8s|service|monitor",
      "goal": "该步骤要达成什么",
      "inputs": {"key":"value"}
    }
  ]
}

要求：
1. steps 至少 1 个，最多 4 个
2. domain 和 kind 必须使用给定枚举
3. inputs 只放执行该步骤所需的最小上下文
4. 如果只是咨询/建议类问题，使用 domain=platform, kind=recommend-action
5. 不要输出 markdown，不要输出解释，只输出 JSON`)
	userInput := req.Message
	if len(req.Context) > 0 {
		userInput += "\n\n上下文:\n" + mustJSON(req.Context)
	}
	msg, err := p.model.Generate(ctx, []*schema.Message{
		schema.SystemMessage(prompt),
		schema.UserMessage(userInput),
	})
	if err != nil {
		return Plan{}, err
	}
	if msg == nil || strings.TrimSpace(msg.Content) == "" {
		return Plan{}, fmt.Errorf("empty planner response")
	}
	return decodePlanResponse(req, msg.Content)
}

func detectDomain(message string, runtime map[string]any) Domain {
	scene := strings.ToLower(strings.TrimSpace(logic.ToString(runtime["scene"])))
	msg := strings.ToLower(message)
	switch {
	case strings.Contains(scene, "host"), strings.Contains(msg, "主机"), strings.Contains(msg, "服务器"), strings.Contains(msg, "磁盘"), strings.Contains(msg, "ssh"):
		return DomainHost
	case strings.Contains(scene, "k8s"), strings.Contains(msg, "pod"), strings.Contains(msg, "deployment"), strings.Contains(msg, "namespace"), strings.Contains(msg, "k8s"):
		return DomainK8s
	case strings.Contains(scene, "service"), strings.Contains(msg, "服务"), strings.Contains(msg, "deploy"), strings.Contains(msg, "发布"):
		return DomainService
	case strings.Contains(scene, "monitor"), strings.Contains(msg, "告警"), strings.Contains(msg, "监控"), strings.Contains(msg, "指标"):
		return DomainMonitor
	default:
		return DomainPlatform
	}
}

func buildDomainSteps(domain Domain, message string) []PlanStep {
	switch domain {
	case DomainHost:
		return []PlanStep{
			{StepID: formatID("step", time.Now()), Title: "确认目标主机", Kind: StepKindHostIdentification, Domain: DomainHost, Goal: "识别与请求相关的目标主机", Status: StepStatusReady, Inputs: map[string]any{"message": message}},
			{StepID: formatID("step", time.Now().Add(time.Nanosecond)), Title: "执行主机诊断", Kind: StepKindHostDiagnosis, Domain: DomainHost, Goal: "采集主机诊断证据并形成结论", Status: StepStatusPending, Dependencies: []string{}, Inputs: map[string]any{"message": message}},
		}
	case DomainK8s:
		return []PlanStep{
			{StepID: formatID("step", time.Now()), Title: "执行 K8s 诊断", Kind: StepKindK8sDiagnosis, Domain: DomainK8s, Goal: "采集集群和工作负载证据", Status: StepStatusReady, Inputs: map[string]any{"message": message}},
		}
	case DomainService:
		return []PlanStep{
			{StepID: formatID("step", time.Now()), Title: "执行服务诊断", Kind: StepKindServiceDiagnosis, Domain: DomainService, Goal: "采集服务相关证据", Status: StepStatusReady, Inputs: map[string]any{"message": message}},
		}
	case DomainMonitor:
		return []PlanStep{
			{StepID: formatID("step", time.Now()), Title: "执行监控调查", Kind: StepKindMonitorInvest, Domain: DomainMonitor, Goal: "采集监控与告警证据", Status: StepStatusReady, Inputs: map[string]any{"message": message}},
		}
	default:
		return []PlanStep{
			{StepID: formatID("step", time.Now()), Title: "生成处置建议", Kind: StepKindRecommendAction, Domain: DomainPlatform, Goal: "给出下一步动作建议", Status: StepStatusReady, Inputs: map[string]any{"message": message}},
		}
	}
}

type ExecutionRequest struct {
	Plan    Plan
	Step    PlanStep
	Message string
	Context map[string]any
	Emit    func(string, map[string]any) bool
}

type DomainExecutor interface {
	Execute(ctx context.Context, req ExecutionRequest) (ExecutionRecord, error)
}

type DomainExecutorFunc func(ctx context.Context, req ExecutionRequest) (ExecutionRecord, error)

func (f DomainExecutorFunc) Execute(ctx context.Context, req ExecutionRequest) (ExecutionRecord, error) {
	return f(ctx, req)
}

type ExecutorRouter struct {
	executors map[Domain]DomainExecutor
}

func NewExecutorRouter() *ExecutorRouter {
	return &ExecutorRouter{executors: map[Domain]DomainExecutor{}}
}

func (r *ExecutorRouter) Register(domain Domain, executor DomainExecutor) {
	if r == nil || executor == nil {
		return
	}
	r.executors[domain] = executor
}

func (r *ExecutorRouter) Execute(ctx context.Context, req ExecutionRequest) (ExecutionRecord, error) {
	if r == nil {
		return ExecutionRecord{}, fmt.Errorf("executor router not initialized")
	}
	executor, ok := r.executors[req.Step.Domain]
	if !ok {
		return ExecutionRecord{}, fmt.Errorf("no executor registered for domain %s", req.Step.Domain)
	}
	return executor.Execute(ctx, req)
}

type ReplanRequest struct {
	Plan      Plan
	Execution ExecutionRecord
}

type Replanner struct {
	model textGenerator
}

func NewReplanner(model textGenerator) *Replanner { return &Replanner{model: model} }

func (r *Replanner) Decide(ctx context.Context, req ReplanRequest) (ReplanDecision, error) {
	if r != nil && r.model != nil {
		if decision, err := r.decideWithModel(ctx, req); err == nil && decision.Outcome != "" {
			return decision, nil
		}
	}
	decision := ReplanDecision{
		DecisionID:    formatID("replan", time.Now()),
		PlanID:        req.Plan.PlanID,
		BasedOnStepID: req.Execution.StepID,
	}
	if req.Execution.Status == ExecutionStatusFailed || req.Execution.Status == ExecutionStatusBlocked {
		decision.Outcome = ReplanOutcomeAskUser
		decision.Rationale = "执行未完成，需要用户确认后续动作。"
		return decision, nil
	}
	findings := make([]string, 0, len(req.Execution.Evidence))
	actions := []NextAction{{
		ID:    formatID("na", time.Now()),
		Type:  "follow-up",
		Label: "继续执行下一步诊断或预检查",
		Risk:  "medium",
	}}
	for _, item := range req.Execution.Evidence {
		if strings.TrimSpace(item.Summary) != "" {
			findings = append(findings, item.Summary)
		} else if strings.TrimSpace(item.Title) != "" {
			findings = append(findings, item.Title)
		}
	}
	decision.Outcome = ReplanOutcomeFinish
	decision.Rationale = "已获得足够证据并生成下一步动作。"
	decision.FinalOutcome = FinalOutcome{
		Status:      "success",
		Summary:     strings.TrimSpace(req.Execution.Summary),
		KeyFindings: findings,
		NextActions: actions,
	}
	if decision.FinalOutcome.Summary == "" {
		decision.FinalOutcome.Summary = "本轮任务已完成。"
	}
	return decision, nil
}

func (r *Replanner) decideWithModel(ctx context.Context, req ReplanRequest) (ReplanDecision, error) {
	prompt := strings.TrimSpace(`你是 AIOps Replanner。请根据目标、计划和执行结果，判断下一步控制面决策。

输出必须是 JSON 对象，格式如下：
{
  "outcome": "continue|revise|ask_user|finish|abort",
  "rationale": "简洁说明原因",
  "final_outcome": {
    "status": "success|failed|partial",
    "summary": "给用户的结论",
    "key_findings": ["发现1"],
    "next_actions": [
      {
        "type": "follow-up|approval|diagnosis|apply",
        "label": "下一步建议",
        "risk": "low|medium|high"
      }
    ]
  }
}

规则：
1. 如果执行被中断、失败或缺少关键信息，优先使用 ask_user
2. 如果目标已完成，使用 finish
3. 除非 outcome=finish，否则 final_outcome 可以为空对象
4. 只输出 JSON`)
	msg, err := r.model.Generate(ctx, []*schema.Message{
		schema.SystemMessage(prompt),
		schema.UserMessage("计划:\n" + mustJSON(req.Plan) + "\n\n执行记录:\n" + mustJSON(req.Execution)),
	})
	if err != nil {
		return ReplanDecision{}, err
	}
	if msg == nil || strings.TrimSpace(msg.Content) == "" {
		return ReplanDecision{}, fmt.Errorf("empty replanner response")
	}
	return decodeReplanDecision(req, msg.Content)
}

type plannerResponse struct {
	Objective string `json:"objective"`
	Steps     []struct {
		Title  string         `json:"title"`
		Kind   StepKind       `json:"kind"`
		Domain Domain         `json:"domain"`
		Goal   string         `json:"goal"`
		Inputs map[string]any `json:"inputs"`
	} `json:"steps"`
}

func decodePlanResponse(req PlanningRequest, content string) (Plan, error) {
	raw := extractJSONObject(content)
	var payload plannerResponse
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return Plan{}, err
	}
	steps := make([]PlanStep, 0, len(payload.Steps))
	for idx, item := range payload.Steps {
		domain := normalizeDomain(item.Domain)
		kind := normalizeStepKind(item.Kind, domain)
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = fmt.Sprintf("执行步骤 %d", idx+1)
		}
		goal := strings.TrimSpace(item.Goal)
		if goal == "" {
			goal = title
		}
		inputs := item.Inputs
		if inputs == nil {
			inputs = map[string]any{}
		}
		inputs["message"] = req.Message
		steps = append(steps, PlanStep{
			StepID: formatID("step", time.Now().Add(time.Duration(idx)*time.Nanosecond)),
			Title:  title,
			Kind:   kind,
			Domain: domain,
			Goal:   goal,
			Inputs: inputs,
			Status: StepStatusReady,
		})
	}
	if len(steps) == 0 {
		return Plan{}, fmt.Errorf("planner returned no steps")
	}
	objective := strings.TrimSpace(payload.Objective)
	if objective == "" {
		objective = strings.TrimSpace(req.Message)
	}
	return Plan{
		PlanID:    formatID("plan", time.Now()),
		SessionID: strings.TrimSpace(req.SessionID),
		Objective: Objective{
			Summary:    objective,
			UserIntent: "operate",
			Urgency:    "medium",
			SuccessCriteria: []string{
				"形成明确计划",
				"完成至少一个领域步骤执行",
				"给出可执行下一步",
			},
		},
		Steps:     steps,
		Status:    "active",
		CreatedAt: time.Now(),
	}, nil
}

func decodeReplanDecision(req ReplanRequest, content string) (ReplanDecision, error) {
	raw := extractJSONObject(content)
	var payload struct {
		Outcome      ReplanOutcome `json:"outcome"`
		Rationale    string        `json:"rationale"`
		FinalOutcome FinalOutcome  `json:"final_outcome"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return ReplanDecision{}, err
	}
	decision := ReplanDecision{
		DecisionID:    formatID("replan", time.Now()),
		PlanID:        req.Plan.PlanID,
		BasedOnStepID: req.Execution.StepID,
		Outcome:       normalizeReplanOutcome(payload.Outcome),
		Rationale:     strings.TrimSpace(payload.Rationale),
		FinalOutcome:  payload.FinalOutcome,
	}
	if decision.Outcome == "" {
		return ReplanDecision{}, fmt.Errorf("replanner returned empty outcome")
	}
	return decision, nil
}

func extractJSONObject(content string) string {
	raw := strings.TrimSpace(content)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return raw[start : end+1]
	}
	return raw
}

func normalizeDomain(domain Domain) Domain {
	switch domain {
	case DomainPlatform, DomainHost, DomainK8s, DomainService, DomainMonitor:
		return domain
	default:
		return DomainPlatform
	}
}

func normalizeStepKind(kind StepKind, domain Domain) StepKind {
	switch kind {
	case StepKindHostIdentification, StepKindHostDiagnosis, StepKindK8sDiagnosis, StepKindServiceDiagnosis, StepKindMonitorInvest, StepKindRecommendAction:
		return kind
	}
	switch domain {
	case DomainHost:
		return StepKindHostDiagnosis
	case DomainK8s:
		return StepKindK8sDiagnosis
	case DomainService:
		return StepKindServiceDiagnosis
	case DomainMonitor:
		return StepKindMonitorInvest
	default:
		return StepKindRecommendAction
	}
}

func normalizeReplanOutcome(outcome ReplanOutcome) ReplanOutcome {
	switch outcome {
	case ReplanOutcomeContinue, ReplanOutcomeRevise, ReplanOutcomeAskUser, ReplanOutcomeFinish, ReplanOutcomeAbort:
		return outcome
	default:
		return ReplanOutcomeAskUser
	}
}
