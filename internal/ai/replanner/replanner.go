package replanner

import "github.com/cy77cc/k8s-manage/internal/ai/executor"

type Replanner struct{}

func NewReplanner() *Replanner { return &Replanner{} }

func (r *Replanner) Decide(result executor.ExecutionResult) ReplanDecision {
	if !HasFailures(result) {
		return ReplanDecision{NeedReplan: false}
	}
	return ReplanDecision{
		NeedReplan:  true,
		Reason:      "one or more execution steps failed",
		Suggestions: []string{"review failed step outputs", "re-run orchestrator planning with updated context"},
	}
}
