package ai

import "time"

type Domain string

const (
	DomainPlatform Domain = "platform"
	DomainHost     Domain = "host"
	DomainK8s      Domain = "k8s"
	DomainService  Domain = "service"
	DomainMonitor  Domain = "monitor"
)

type StepKind string

const (
	StepKindHostIdentification StepKind = "host-identification"
	StepKindHostDiagnosis      StepKind = "host-diagnosis"
	StepKindK8sDiagnosis       StepKind = "k8s-diagnosis"
	StepKindServiceDiagnosis   StepKind = "service-diagnosis"
	StepKindMonitorInvest      StepKind = "monitor-investigation"
	StepKindRecommendAction    StepKind = "recommend-action"
)

type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusReady     StepStatus = "ready"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusBlocked   StepStatus = "blocked"
)

type ExecutionStatus string

const (
	ExecutionStatusRunning     ExecutionStatus = "running"
	ExecutionStatusCompleted   ExecutionStatus = "completed"
	ExecutionStatusFailed      ExecutionStatus = "failed"
	ExecutionStatusBlocked     ExecutionStatus = "blocked"
	ExecutionStatusInterrupted ExecutionStatus = "interrupted"
	ExecutionStatusPartial     ExecutionStatus = "partial"
)

type EvidenceType string

const (
	EvidenceTypeHostMatch     EvidenceType = "host-match"
	EvidenceTypeDiskUsage     EvidenceType = "disk-usage"
	EvidenceTypeDiagnosis     EvidenceType = "diagnosis"
	EvidenceTypeRecommendation EvidenceType = "recommendation"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityCritical Severity = "critical"
)

type ReplanOutcome string

const (
	ReplanOutcomeContinue ReplanOutcome = "continue"
	ReplanOutcomeRevise   ReplanOutcome = "revise"
	ReplanOutcomeAskUser  ReplanOutcome = "ask_user"
	ReplanOutcomeFinish   ReplanOutcome = "finish"
	ReplanOutcomeAbort    ReplanOutcome = "abort"
)

type OrchestrationPhase string

const (
	PhasePlanning    OrchestrationPhase = "planning"
	PhaseExecuting   OrchestrationPhase = "executing"
	PhaseInterrupted OrchestrationPhase = "interrupted"
	PhaseReplanning  OrchestrationPhase = "replanning"
	PhaseFinished    OrchestrationPhase = "finished"
	PhaseAborted     OrchestrationPhase = "aborted"
)

type Objective struct {
	Summary         string   `json:"summary"`
	SuccessCriteria []string `json:"success_criteria,omitempty"`
	UserIntent      string   `json:"user_intent,omitempty"`
	Urgency         string   `json:"urgency,omitempty"`
}

type NextAction struct {
	ID     string         `json:"id"`
	Type   string         `json:"type"`
	Label  string         `json:"label"`
	Risk   string         `json:"risk,omitempty"`
	Params map[string]any `json:"params,omitempty"`
}

type FinalOutcome struct {
	Status      string       `json:"status"`
	Summary     string       `json:"summary"`
	KeyFindings []string     `json:"key_findings,omitempty"`
	NextActions []NextAction `json:"next_actions,omitempty"`
}

type PlanStep struct {
	StepID       string         `json:"step_id"`
	Title        string         `json:"title"`
	Kind         StepKind       `json:"kind"`
	Domain       Domain         `json:"domain"`
	Goal         string         `json:"goal"`
	Dependencies []string       `json:"dependencies,omitempty"`
	Inputs       map[string]any `json:"inputs,omitempty"`
	Status       StepStatus     `json:"status"`
}

type Plan struct {
	PlanID     string      `json:"plan_id"`
	SessionID  string      `json:"session_id,omitempty"`
	Objective  Objective   `json:"objective"`
	Steps      []PlanStep  `json:"steps"`
	Status     string      `json:"status,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	FinishedAt *time.Time  `json:"finished_at,omitempty"`
}

type EvidenceItem struct {
	EvidenceID string         `json:"evidence_id"`
	Type       EvidenceType   `json:"type"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary,omitempty"`
	Severity   Severity       `json:"severity,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
}

type ExecutionIssue struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
}

type ExecutionRecord struct {
	ExecutionID string           `json:"execution_id,omitempty"`
	PlanID      string           `json:"plan_id,omitempty"`
	StepID      string           `json:"step_id"`
	Status      ExecutionStatus  `json:"status"`
	Summary     string           `json:"summary,omitempty"`
	Evidence    []EvidenceItem   `json:"evidence,omitempty"`
	Issues      []ExecutionIssue `json:"issues,omitempty"`
	StartedAt   time.Time        `json:"started_at,omitempty"`
	FinishedAt  *time.Time       `json:"finished_at,omitempty"`
}

type ReplanDecision struct {
	DecisionID   string        `json:"decision_id,omitempty"`
	PlanID       string        `json:"plan_id,omitempty"`
	BasedOnStepID string       `json:"based_on_step_id,omitempty"`
	Outcome      ReplanOutcome `json:"outcome"`
	Rationale    string        `json:"rationale,omitempty"`
	FinalOutcome FinalOutcome  `json:"final_outcome,omitempty"`
}

type OrchestrationState struct {
	Phase      OrchestrationPhase `json:"phase"`
	Objective  Objective          `json:"objective"`
	CurrentPlan *Plan             `json:"current_plan,omitempty"`
	LastRecord *ExecutionRecord   `json:"last_record,omitempty"`
	LastDecision *ReplanDecision  `json:"last_decision,omitempty"`
}
