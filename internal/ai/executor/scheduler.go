package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/runtime"
)

func advanceScheduler(ctx context.Context, store *runtime.ExecutionStore, state *runtime.ExecutionState) (*Result, error) {
	if state == nil {
		return nil, fmt.Errorf("execution state is required")
	}
	results := make([]StepResult, 0, len(state.Steps))
	progress := true
	for progress {
		progress = false
		for stepID, step := range state.Steps {
			switch step.Status {
			case runtime.StepPending:
				if depsSatisfied(state, stepID) {
					if err := transitionStep(state, stepID, runtime.StepReady, "dependencies satisfied"); err != nil {
						return nil, err
					}
					results = append(results, snapshotResult(state.Steps[stepID]))
					progress = true
				} else if depsBlocked(state, stepID) {
					if err := transitionStep(state, stepID, runtime.StepBlocked, "upstream dependency failed"); err != nil {
						return nil, err
					}
					results = append(results, snapshotResult(state.Steps[stepID]))
					progress = true
				}
			case runtime.StepReady:
				if err := transitionStep(state, stepID, runtime.StepRunning, "step entered executor runtime"); err != nil {
					return nil, err
				}
				results = append(results, snapshotResult(state.Steps[stepID]))
				progress = true

				step = state.Steps[stepID]
				if needsApproval(step) {
					policy := riskPolicy(step.Mode, step.Risk)
					pending := &runtime.PendingApproval{
						PlanID:      state.PlanID,
						StepID:      stepID,
						ApprovalKey: fmt.Sprintf("%s:%s", state.PlanID, stepID),
						Status:      "pending",
						Title:       step.Title,
						Mode:        step.Mode,
						Risk:        step.Risk,
						Summary:     step.UserVisibleSummary,
						RequestedAt: time.Now().UTC(),
					}
					state.PendingApproval = pending
					state.Status = runtime.ExecutionStatusWaitingApproval
					state.Phase = fmt.Sprintf("waiting_approval:%s", approvalStageName(policy))
					if err := transitionStep(state, stepID, runtime.StepWaitingApproval, "step requires approval before execution"); err != nil {
						return nil, err
					}
					results = append(results, snapshotResult(state.Steps[stepID]))
					if store != nil {
						if err := store.Save(ctx, *state); err != nil {
							return nil, err
						}
					}
					return &Result{State: *state, Steps: results}, nil
				}

				if err := transitionStep(state, stepID, runtime.StepCompleted, "step scheduled by executor runtime"); err != nil {
					return nil, err
				}
				results = append(results, snapshotResult(state.Steps[stepID]))
				progress = true
			}
		}
	}

	if hasFailedOrBlocked(state) {
		state.Status = runtime.ExecutionStatusFailed
		state.Phase = "executor_blocked"
	} else {
		state.Status = runtime.ExecutionStatusCompleted
		state.Phase = "executor_completed"
	}
	if store != nil {
		if err := store.Save(ctx, *state); err != nil {
			return nil, err
		}
	}
	return &Result{State: *state, Steps: results}, nil
}

func advanceAfterApproval(ctx context.Context, store *runtime.ExecutionStore, state *runtime.ExecutionState, req ResumeRequest) (*Result, error) {
	if state == nil {
		return nil, fmt.Errorf("execution state is required")
	}
	if state.PendingApproval == nil {
		return &Result{State: *state}, nil
	}
	stepID := strings.TrimSpace(req.StepID)
	if stepID == "" {
		stepID = state.PendingApproval.StepID
	}
	if stepID == "" {
		return nil, fmt.Errorf("step_id is required")
	}
	if strings.TrimSpace(req.PlanID) != "" && strings.TrimSpace(req.PlanID) != state.PlanID {
		return nil, fmt.Errorf("plan_id mismatch")
	}

	decision := ApprovalDecision{
		PlanID:      state.PlanID,
		StepID:      stepID,
		Approved:    req.Approved,
		Reason:      strings.TrimSpace(req.Reason),
		Idempotency: runtime.ApprovalDecisionHash(state.PlanID, stepID, req.Approved),
	}
	if state.PendingApproval.DecisionHash == decision.Idempotency {
		return &Result{State: *state}, nil
	}
	now := time.Now().UTC()
	decision.Status = "rejected"
	if req.Approved {
		decision.Status = "approved"
	}
	state.PendingApproval.DecisionHash = decision.Idempotency
	state.PendingApproval.Approved = &decision.Approved
	state.PendingApproval.Reason = decision.Reason
	state.PendingApproval.ResolvedAt = now
	state.PendingApproval.Status = decision.Status

	if req.Approved {
		step := state.Steps[stepID]
		step.ApprovalSatisfied = true
		state.Steps[stepID] = step
		if err := transitionStep(state, stepID, runtime.StepReady, "approval granted, step returned to ready"); err != nil {
			return nil, err
		}
		state.Status = runtime.ExecutionStatusRunning
		state.Phase = "approval_granted"
	} else {
		if err := transitionStep(state, stepID, runtime.StepCancelled, "approval rejected by user"); err != nil {
			return nil, err
		}
		state.Status = runtime.ExecutionStatusFailed
		state.Phase = "cancelled"
		markDependentsBlocked(state, stepID)
	}
	return advanceScheduler(ctx, store, state)
}

func depsSatisfied(state *runtime.ExecutionState, stepID string) bool {
	deps := stepDependencies(state, stepID)
	if len(deps) == 0 {
		return true
	}
	for _, dep := range deps {
		if state.Steps[dep].Status != runtime.StepCompleted {
			return false
		}
	}
	return true
}

func depsBlocked(state *runtime.ExecutionState, stepID string) bool {
	for _, dep := range stepDependencies(state, stepID) {
		status := state.Steps[dep].Status
		if status == runtime.StepFailed || status == runtime.StepBlocked || status == runtime.StepCancelled {
			return true
		}
	}
	return false
}

func stepDependencies(state *runtime.ExecutionState, stepID string) []string {
	if state == nil || state.Steps == nil {
		return nil
	}
	raw, ok := state.Steps[stepID]
	if !ok {
		return nil
	}
	if raw.DependsOn != nil {
		return raw.DependsOn
	}
	return nil
}

func needsApproval(step runtime.StepState) bool {
	if step.ApprovalSatisfied {
		return false
	}
	return riskPolicy(step.Mode, step.Risk).RequiresApproval
}

func shouldAutoRetry(step runtime.StepState) bool {
	policy := riskPolicy(step.Mode, step.Risk)
	if !policy.AutoRetry {
		return false
	}
	return step.Attempts < step.MaxAttempts
}

func approvalStageName(policy RiskPolicy) string {
	if !policy.RequiresApproval {
		return "none"
	}
	if policy.MaxAttempts <= 1 {
		return "strict"
	}
	return "guarded"
}

func hasFailedOrBlocked(state *runtime.ExecutionState) bool {
	for _, step := range state.Steps {
		if step.Status == runtime.StepFailed || step.Status == runtime.StepBlocked || step.Status == runtime.StepCancelled {
			return true
		}
	}
	return false
}

func markDependentsBlocked(state *runtime.ExecutionState, failedStepID string) {
	for stepID := range state.Steps {
		for _, dep := range stepDependencies(state, stepID) {
			if dep == failedStepID {
				_ = transitionStep(state, stepID, runtime.StepBlocked, "upstream dependency was cancelled")
				break
			}
		}
	}
}

func transitionStep(state *runtime.ExecutionState, stepID string, target runtime.StepStatus, summary string) error {
	if state == nil {
		return fmt.Errorf("execution state is required")
	}
	step, ok := state.Steps[stepID]
	if !ok {
		return fmt.Errorf("step %s not found", stepID)
	}
	if !validTransition(step.Status, target) {
		return fmt.Errorf("invalid step transition: %s -> %s", step.Status, target)
	}
	step.Status = target
	if target == runtime.StepRunning {
		step.Attempts++
		if step.MaxAttempts == 0 {
			step.MaxAttempts = riskPolicy(step.Mode, step.Risk).MaxAttempts
		}
	}
	step.UserVisibleSummary = strings.TrimSpace(summary)
	step.UpdatedAt = time.Now().UTC()
	state.Steps[stepID] = step
	state.UpdatedAt = step.UpdatedAt
	return nil
}

func validTransition(from, to runtime.StepStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case runtime.StepPending:
		return to == runtime.StepReady || to == runtime.StepBlocked
	case runtime.StepReady:
		return to == runtime.StepRunning || to == runtime.StepCancelled
	case runtime.StepRunning:
		return to == runtime.StepCompleted || to == runtime.StepFailed || to == runtime.StepWaitingApproval || to == runtime.StepCancelled
	case runtime.StepWaitingApproval:
		return to == runtime.StepReady || to == runtime.StepCancelled
	case runtime.StepCompleted, runtime.StepFailed, runtime.StepBlocked, runtime.StepCancelled:
		return false
	default:
		return false
	}
}

func snapshotResult(step runtime.StepState) StepResult {
	var errInfo *StepError
	if step.ErrorCode != "" || step.ErrorMessage != "" {
		errInfo = &StepError{Code: step.ErrorCode, Message: step.ErrorMessage}
	}
	return StepResult{
		StepID:    step.StepID,
		Status:    step.Status,
		Summary:   step.UserVisibleSummary,
		Error:     errInfo,
		UpdatedAt: step.UpdatedAt,
	}
}

func describePreparedStep(step planner.PlanStep, status runtime.StepStatus) string {
	switch status {
	case runtime.StepReady:
		return fmt.Sprintf("步骤 %q 已满足依赖，等待 Executor Runtime 调度。", step.Title)
	default:
		return fmt.Sprintf("步骤 %q 已记录，等待上游步骤完成。", step.Title)
	}
}
