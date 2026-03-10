package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/runtime"
)

type Request struct {
	TraceID        string
	SessionID      string
	Message        string
	Plan           planner.ExecutionPlan
	RuntimeContext runtime.ContextSnapshot
}

type ResumeRequest struct {
	SessionID string `json:"session_id"`
	PlanID    string `json:"plan_id"`
	StepID    string `json:"step_id"`
	Approved  bool   `json:"approved"`
	Reason    string `json:"reason,omitempty"`
}

type ApprovalDecision struct {
	PlanID      string `json:"plan_id"`
	StepID      string `json:"step_id"`
	Approved    bool   `json:"approved"`
	Reason      string `json:"reason,omitempty"`
	Idempotency string `json:"idempotency"`
	Status      string `json:"status,omitempty"`
}

type Evidence struct {
	Kind   string         `json:"kind,omitempty"`
	Source string         `json:"source,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
}

type StepError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type StepResult struct {
	StepID    string             `json:"step_id"`
	Status    runtime.StepStatus `json:"status"`
	Summary   string             `json:"summary,omitempty"`
	Evidence  []Evidence         `json:"evidence,omitempty"`
	Error     *StepError         `json:"error,omitempty"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type Result struct {
	State           runtime.ExecutionState   `json:"state"`
	Steps           []StepResult             `json:"steps,omitempty"`
	PendingApproval *runtime.PendingApproval `json:"pending_approval,omitempty"`
}

type Executor struct {
	store *runtime.ExecutionStore
}

type RiskPolicy struct {
	RequiresApproval bool `json:"requires_approval"`
	MaxAttempts      int  `json:"max_attempts"`
	AutoRetry        bool `json:"auto_retry"`
}

func New(store *runtime.ExecutionStore) *Executor {
	return &Executor{store: store}
}

func riskPolicy(mode, risk string) RiskPolicy {
	mode = strings.ToLower(strings.TrimSpace(mode))
	risk = strings.ToLower(strings.TrimSpace(risk))
	switch {
	case mode == "mutating" || risk == "high":
		return RiskPolicy{RequiresApproval: true, MaxAttempts: 1, AutoRetry: false}
	case risk == "medium":
		return RiskPolicy{RequiresApproval: true, MaxAttempts: 1, AutoRetry: false}
	default:
		return RiskPolicy{RequiresApproval: false, MaxAttempts: 2, AutoRetry: true}
	}
}

func (e *Executor) PrepareState(_ context.Context, req Request) (runtime.ExecutionState, []StepResult, error) {
	planID := strings.TrimSpace(req.Plan.PlanID)
	if planID == "" {
		return runtime.ExecutionState{}, nil, fmt.Errorf("plan_id is required")
	}
	if strings.TrimSpace(req.SessionID) == "" {
		return runtime.ExecutionState{}, nil, fmt.Errorf("session_id is required")
	}

	now := time.Now().UTC()
	steps := make(map[string]runtime.StepState, len(req.Plan.Steps))
	results := make([]StepResult, 0, len(req.Plan.Steps))
	for _, step := range req.Plan.Steps {
		if strings.TrimSpace(step.StepID) == "" {
			return runtime.ExecutionState{}, nil, fmt.Errorf("plan step missing step_id")
		}
		policy := riskPolicy(step.Mode, step.Risk)
		status := runtime.StepPending
		if len(step.DependsOn) == 0 {
			status = runtime.StepReady
		}
		steps[step.StepID] = runtime.StepState{
			StepID:             step.StepID,
			Title:              step.Title,
			Expert:             step.Expert,
			DependsOn:          append([]string(nil), step.DependsOn...),
			Status:             status,
			Mode:               strings.TrimSpace(step.Mode),
			Risk:               strings.TrimSpace(step.Risk),
			MaxAttempts:        policy.MaxAttempts,
			IdempotencyKey:     runtime.ApprovalDecisionHash(planID, step.StepID, false),
			UserVisibleSummary: describePreparedStep(step, status),
			UpdatedAt:          now,
		}
		results = append(results, StepResult{
			StepID:    step.StepID,
			Status:    status,
			Summary:   describePreparedStep(step, status),
			UpdatedAt: now,
		})
	}

	state := runtime.ExecutionState{
		TraceID:        req.TraceID,
		SessionID:      req.SessionID,
		PlanID:         planID,
		Message:        strings.TrimSpace(req.Message),
		Status:         runtime.ExecutionStatusRunning,
		Phase:          "executor_prepared",
		RuntimeContext: req.RuntimeContext,
		Steps:          steps,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return state, results, nil
}

func (e *Executor) SavePreparedState(ctx context.Context, req Request) (*Result, error) {
	state, steps, err := e.PrepareState(ctx, req)
	if err != nil {
		return nil, err
	}
	if e != nil && e.store != nil {
		if err := e.store.Save(ctx, state); err != nil {
			return nil, err
		}
	}
	return &Result{State: state, Steps: steps}, nil
}

func (e *Executor) Resume(ctx context.Context, req ResumeRequest) (*Result, error) {
	if e == nil || e.store == nil {
		return nil, fmt.Errorf("execution store is not configured")
	}
	state, err := e.store.Load(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("execution state not found")
	}
	return advanceAfterApproval(ctx, e.store, state, req)
}

func (e *Executor) Run(ctx context.Context, req Request) (*Result, error) {
	prepared, err := e.SavePreparedState(ctx, req)
	if err != nil {
		return nil, err
	}
	if e == nil || e.store == nil {
		return prepared, nil
	}
	return advanceScheduler(ctx, e.store, &prepared.State)
}

func (e *Executor) RecordFailure(ctx context.Context, sessionID, stepID, code, message string) (*Result, error) {
	if e == nil || e.store == nil {
		return nil, fmt.Errorf("execution store is not configured")
	}
	state, err := e.store.Load(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("execution state not found")
	}
	step, ok := state.Steps[stepID]
	if !ok {
		return nil, fmt.Errorf("step %s not found", stepID)
	}
	step.ErrorCode = strings.TrimSpace(code)
	step.ErrorMessage = strings.TrimSpace(message)
	if shouldAutoRetry(step) {
		step.Status = runtime.StepReady
		step.UserVisibleSummary = "step failed once and will be retried automatically"
	} else {
		step.Status = runtime.StepFailed
		step.UserVisibleSummary = "step failed and requires manual follow-up"
		state.Status = runtime.ExecutionStatusFailed
		state.Phase = "executor_failed"
		markDependentsBlocked(state, stepID)
	}
	step.UpdatedAt = time.Now().UTC()
	state.Steps[stepID] = step
	if err := e.store.Save(ctx, *state); err != nil {
		return nil, err
	}
	return &Result{State: *state, Steps: []StepResult{snapshotResult(step)}}, nil
}
