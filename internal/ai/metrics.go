package ai

import (
	"strings"
	"sync"

	"github.com/cy77cc/OpsPilot/internal/ai/events"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/rewrite"
)

type AIMetrics struct {
	mu           sync.Mutex
	rewrite      RewriteMetricsSnapshot
	planner      PlannerMetricsSnapshot
	resume       ResumeMetricsSnapshot
	thoughtChain ThoughtChainMetricsSnapshot
}

type AIMetricsSnapshot struct {
	Rewrite      RewriteMetricsSnapshot      `json:"rewrite"`
	Planner      PlannerMetricsSnapshot      `json:"planner"`
	Resume       ResumeMetricsSnapshot       `json:"resume"`
	ThoughtChain ThoughtChainMetricsSnapshot `json:"thought_chain"`
}

type RewriteMetricsSnapshot struct {
	Total             int     `json:"total"`
	StructuredOutputs int     `json:"structured_outputs"`
	Fallbacks         int     `json:"fallbacks"`
	AmbiguousOutputs  int     `json:"ambiguous_outputs"`
	QualityRate       float64 `json:"quality_rate"`
}

type PlannerMetricsSnapshot struct {
	Total              int     `json:"total"`
	Clarify            int     `json:"clarify"`
	Plans              int     `json:"plans"`
	ExecutablePlans    int     `json:"executable_plans"`
	DirectReplies      int     `json:"direct_replies"`
	Rejected           int     `json:"rejected"`
	ClarifyRate        float64 `json:"clarify_rate"`
	ExecutablePlanRate float64 `json:"executable_plan_rate"`
}

type ResumeMetricsSnapshot struct {
	Total                  int     `json:"total"`
	Successful             int     `json:"successful"`
	Failures               int     `json:"failures"`
	DuplicateIntercepted   int     `json:"duplicate_intercepted"`
	SuccessRate            float64 `json:"success_rate"`
	DuplicateInterceptRate float64 `json:"duplicate_intercept_rate"`
}

type ThoughtChainMetricsSnapshot struct {
	Runs                   int     `json:"runs"`
	ExpectedStageSignals   int     `json:"expected_stage_signals"`
	DeliveredStageSignals  int     `json:"delivered_stage_signals"`
	RunsWithMissingSignals int     `json:"runs_with_missing_signals"`
	EventCompletenessRate  float64 `json:"event_completeness_rate"`
}

type thoughtChainRunMetrics struct {
	parent          *AIMetrics
	requiredStages  map[string]struct{}
	deliveredStages map[string]struct{}
}

func NewAIMetrics() *AIMetrics {
	return &AIMetrics{}
}

func (m *AIMetrics) Snapshot() AIMetricsSnapshot {
	if m == nil {
		return AIMetricsSnapshot{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return AIMetricsSnapshot{
		Rewrite:      m.rewrite,
		Planner:      m.planner,
		Resume:       m.resume,
		ThoughtChain: m.thoughtChain,
	}
}

func (m *AIMetrics) RecordRewrite(out rewrite.Output) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rewrite.Total++
	if isStructuredRewrite(out) {
		m.rewrite.StructuredOutputs++
	}
	if isRewriteFallback(out) {
		m.rewrite.Fallbacks++
	}
	if len(out.AmbiguityFlags) > 0 || len(out.Ambiguities) > 0 {
		m.rewrite.AmbiguousOutputs++
	}
	m.rewrite.QualityRate = rate(m.rewrite.StructuredOutputs, m.rewrite.Total)
}

func (m *AIMetrics) RecordPlanner(decision planner.Decision) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.planner.Total++
	switch decision.Type {
	case planner.DecisionClarify:
		m.planner.Clarify++
	case planner.DecisionPlan:
		m.planner.Plans++
		if planIsExecutable(decision.Plan) {
			m.planner.ExecutablePlans++
		}
	case planner.DecisionDirectReply:
		m.planner.DirectReplies++
	case planner.DecisionReject:
		m.planner.Rejected++
	}
	m.planner.ClarifyRate = rate(m.planner.Clarify, m.planner.Total)
	m.planner.ExecutablePlanRate = rate(m.planner.ExecutablePlans, m.planner.Plans)
}

func (m *AIMetrics) RecordResume(status string, err error) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resume.Total++
	if err != nil {
		m.resume.Failures++
	} else {
		switch strings.TrimSpace(status) {
		case "idempotent":
			m.resume.Successful++
			m.resume.DuplicateIntercepted++
		case "approved", "approval_granted", "rejected", "cancelled", "noop":
			m.resume.Successful++
		default:
			m.resume.Failures++
		}
	}
	m.resume.SuccessRate = rate(m.resume.Successful, m.resume.Total)
	m.resume.DuplicateInterceptRate = rate(m.resume.DuplicateIntercepted, m.resume.Total)
}

func (m *AIMetrics) StartThoughtChainRun() *thoughtChainRunMetrics {
	if m == nil {
		return nil
	}
	return &thoughtChainRunMetrics{
		parent:          m,
		requiredStages:  make(map[string]struct{}),
		deliveredStages: make(map[string]struct{}),
	}
}

func (r *thoughtChainRunMetrics) Observe(evt StreamEvent) {
	if r == nil {
		return
	}
	switch evt.Type {
	case events.RewriteResult:
		r.requiredStages["rewrite"] = struct{}{}
	case events.PlannerState, events.PlanCreated:
		r.requiredStages["plan"] = struct{}{}
	case events.StepUpdate, events.ToolCall, events.ToolResult:
		r.requiredStages["execute"] = struct{}{}
	case events.ApprovalRequired, events.ClarifyRequired:
		r.requiredStages["user_action"] = struct{}{}
		r.deliveredStages["user_action"] = struct{}{}
	case events.Summary:
		r.requiredStages["summary"] = struct{}{}
	case events.StageDelta:
		stage := strings.TrimSpace(stringValue(evt.Data["stage"]))
		if stage != "" {
			r.deliveredStages[stage] = struct{}{}
		}
	}
}

func (r *thoughtChainRunMetrics) Finalize() {
	if r == nil || r.parent == nil {
		return
	}
	requiredCount := len(r.requiredStages)
	if requiredCount == 0 {
		return
	}
	deliveredCount := 0
	for stage := range r.requiredStages {
		if _, ok := r.deliveredStages[stage]; ok {
			deliveredCount++
		}
	}

	r.parent.mu.Lock()
	defer r.parent.mu.Unlock()
	r.parent.thoughtChain.Runs++
	r.parent.thoughtChain.ExpectedStageSignals += requiredCount
	r.parent.thoughtChain.DeliveredStageSignals += deliveredCount
	if deliveredCount < requiredCount {
		r.parent.thoughtChain.RunsWithMissingSignals++
	}
	r.parent.thoughtChain.EventCompletenessRate = rate(
		r.parent.thoughtChain.DeliveredStageSignals,
		r.parent.thoughtChain.ExpectedStageSignals,
	)
}

func isStructuredRewrite(out rewrite.Output) bool {
	return strings.TrimSpace(out.NormalizedGoal) != "" &&
		strings.TrimSpace(out.OperationMode) != "" &&
		strings.TrimSpace(out.Narrative) != ""
}

func isRewriteFallback(out rewrite.Output) bool {
	for _, assumption := range out.Assumptions {
		if strings.HasPrefix(strings.TrimSpace(assumption), "rewrite_") {
			return true
		}
	}
	return false
}

func planIsExecutable(plan *planner.ExecutionPlan) bool {
	if plan == nil || strings.TrimSpace(plan.PlanID) == "" || len(plan.Steps) == 0 {
		return false
	}
	for _, step := range plan.Steps {
		if strings.TrimSpace(step.StepID) == "" || strings.TrimSpace(step.Expert) == "" {
			return false
		}
	}
	return true
}

func rate(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
