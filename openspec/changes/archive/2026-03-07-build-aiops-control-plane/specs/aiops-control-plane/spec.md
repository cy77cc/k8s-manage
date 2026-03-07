## ADDED Requirements

### Requirement: AI core MUST host explicit plan-execute-replan orchestration

The system MUST host explicit Planner, Executor, and Replanner responsibilities inside `internal/ai` as the primary runtime for AIOps task orchestration.

#### Scenario: orchestration roles are defined in the AI core
- **WHEN** maintainers inspect the AIOps backend architecture
- **THEN** `internal/ai` MUST define explicit planning, execution, and replanning roles
- **AND** the control-plane runtime MUST no longer depend on a single implicit platform-agent loop as the only orchestration abstraction

### Requirement: planner MUST output domain-level operational steps

The Planner MUST emit domain-level operational steps rather than direct tool call lists.

#### Scenario: plan output is domain-oriented
- **WHEN** a user objective is converted into a plan
- **THEN** the resulting plan MUST contain operational step types such as host, k8s, service, or monitor actions
- **AND** the plan MUST NOT require direct tool names as the primary planning abstraction

### Requirement: executor MUST produce structured execution records and evidence

Executors MUST normalize execution into structured records that include execution status, actions, evidence, and issues.

#### Scenario: executor reports operational facts
- **WHEN** a plan step is executed
- **THEN** the control plane MUST produce a structured execution record
- **AND** the record MUST include execution state and operational evidence that can be consumed by replanning and UI layers

### Requirement: replanner MUST emit finite control-plane decisions

The Replanner MUST emit bounded control-plane decisions that can drive workflow continuation or revision.

#### Scenario: replanner decides the next control-plane state
- **WHEN** execution results are evaluated
- **THEN** the replanner MUST produce a decision such as continue, revise, ask_user, finish, or abort
- **AND** the result MUST be machine-usable without requiring frontend parsing of narrative text

### Requirement: control plane MUST route work through domain executors

The AI control plane MUST route operational steps through formal domain executors for Host, K8s, Service, and Monitor domains.

#### Scenario: domain routing is explicit
- **WHEN** an execution step is dispatched
- **THEN** the control plane MUST route the step through a domain executor boundary
- **AND** Host, K8s, Service, and Monitor MUST each have a first-class executor host in the architecture

### Requirement: existing approval and resume flows MUST remain part of the control plane

The control plane MUST keep approval, confirmation, preview, and resume flows as formal orchestration dependencies.

#### Scenario: mutating execution pauses for control-plane review
- **WHEN** a step reaches a mutating or interrupting path
- **THEN** the control plane MUST continue to use approval, confirmation, preview, and resume flows as orchestration dependencies
- **AND** the task lifecycle MUST remain resumable after interruption
