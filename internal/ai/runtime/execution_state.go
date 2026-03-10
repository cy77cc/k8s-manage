package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultExecutionTTL = 24 * time.Hour

type ExecutionStatus string

const (
	ExecutionStatusRunning         ExecutionStatus = "running"
	ExecutionStatusWaitingApproval ExecutionStatus = "waiting_approval"
	ExecutionStatusCompleted       ExecutionStatus = "completed"
	ExecutionStatusFailed          ExecutionStatus = "failed"
)

type StepStatus string

const (
	StepPending         StepStatus = "pending"
	StepReady           StepStatus = "ready"
	StepRunning         StepStatus = "running"
	StepWaitingApproval StepStatus = "waiting_approval"
	StepCompleted       StepStatus = "completed"
	StepFailed          StepStatus = "failed"
	StepBlocked         StepStatus = "blocked"
	StepCancelled       StepStatus = "cancelled"
)

type ContextSnapshot struct {
	Scene       string   `json:"scene,omitempty"`
	Route       string   `json:"route,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	CurrentPage string   `json:"current_page,omitempty"`
	ResourceIDs []string `json:"resource_ids,omitempty"`
}

type StepState struct {
	StepID             string         `json:"step_id"`
	Title              string         `json:"title,omitempty"`
	Expert             string         `json:"expert,omitempty"`
	Intent             string         `json:"intent,omitempty"`
	Task               string         `json:"task,omitempty"`
	Input              map[string]any `json:"input,omitempty"`
	DependsOn          []string       `json:"depends_on,omitempty"`
	Status             StepStatus     `json:"status"`
	Mode               string         `json:"mode,omitempty"`
	Risk               string         `json:"risk,omitempty"`
	Attempts           int            `json:"attempts,omitempty"`
	MaxAttempts        int            `json:"max_attempts,omitempty"`
	IdempotencyKey     string         `json:"idempotency_key,omitempty"`
	ApprovalSatisfied  bool           `json:"approval_satisfied,omitempty"`
	UserVisibleSummary string         `json:"user_visible_summary,omitempty"`
	ErrorCode          string         `json:"error_code,omitempty"`
	ErrorMessage       string         `json:"error_message,omitempty"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type PendingApproval struct {
	PlanID       string    `json:"plan_id,omitempty"`
	StepID       string    `json:"step_id,omitempty"`
	ApprovalKey  string    `json:"approval_key,omitempty"`
	Status       string    `json:"status,omitempty"`
	Title        string    `json:"title,omitempty"`
	Mode         string    `json:"mode,omitempty"`
	Risk         string    `json:"risk,omitempty"`
	Summary      string    `json:"summary,omitempty"`
	Approved     *bool     `json:"approved,omitempty"`
	Reason       string    `json:"reason,omitempty"`
	RequestedAt  time.Time `json:"requested_at,omitempty"`
	ResolvedAt   time.Time `json:"resolved_at,omitempty"`
	DecisionHash string    `json:"decision_hash,omitempty"`
}

type ExecutionState struct {
	TraceID         string               `json:"trace_id"`
	SessionID       string               `json:"session_id"`
	PlanID          string               `json:"plan_id,omitempty"`
	Message         string               `json:"message,omitempty"`
	Status          ExecutionStatus      `json:"status"`
	Phase           string               `json:"phase,omitempty"`
	RuntimeContext  ContextSnapshot      `json:"runtime_context,omitempty"`
	Steps           map[string]StepState `json:"steps,omitempty"`
	PendingApproval *PendingApproval     `json:"pending_approval,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

type ExecutionStore struct {
	client redis.UniversalClient
	prefix string
	ttl    time.Duration
}

func NewExecutionStore(client redis.UniversalClient, prefix string) *ExecutionStore {
	if prefix == "" {
		prefix = "ai:execution:"
	}
	return &ExecutionStore{client: client, prefix: prefix, ttl: defaultExecutionTTL}
}

func (s *ExecutionStore) Save(ctx context.Context, st ExecutionState) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("execution store is not initialized")
	}
	now := time.Now().UTC()
	if st.CreatedAt.IsZero() {
		st.CreatedAt = now
	}
	st.UpdatedAt = now
	data, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("marshal execution state: %w", err)
	}
	return s.client.Set(ctx, s.key(st.SessionID), data, s.ttl).Err()
}

func (s *ExecutionStore) Load(ctx context.Context, sessionID string) (*ExecutionState, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("execution store is not initialized")
	}
	raw, err := s.client.Get(ctx, s.key(sessionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var out ExecutionState
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("unmarshal execution state: %w", err)
	}
	return &out, nil
}

func (s *ExecutionStore) key(sessionID string) string {
	return s.prefix + sessionID
}

func ApprovalDecisionHash(planID, stepID string, approved bool) string {
	return fmt.Sprintf("%s:%s:%t", planID, stepID, approved)
}
