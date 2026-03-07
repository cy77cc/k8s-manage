package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/k8s-manage/internal/service/ai/logic"
)

type PlanningRequest struct {
	SessionID string
	Message   string
	Context   map[string]any
}

type Planner struct{}

func NewPlanner() *Planner { return &Planner{} }

func (p *Planner) BuildPlan(_ context.Context, req PlanningRequest) (Plan, error) {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return Plan{}, fmt.Errorf("message is required")
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

type Replanner struct{}

func NewReplanner() *Replanner { return &Replanner{} }

func (r *Replanner) Decide(_ context.Context, req ReplanRequest) (ReplanDecision, error) {
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
