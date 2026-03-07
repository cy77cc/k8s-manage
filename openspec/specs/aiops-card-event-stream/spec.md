## ADDED Requirements

### Requirement: AI control plane MUST emit application-card-oriented platform events

The system MUST emit structured platform events that represent AIOps task progress, evidence, approvals, replanning, and next actions for application-card style rendering.

#### Scenario: backend emits card-native event semantics
- **WHEN** an AIOps task runs through the control plane
- **THEN** the backend MUST emit structured platform events for orchestration progress and operational state
- **AND** those events MUST be suitable for projection into application-card UI components

### Requirement: card event stream MUST cover plan, execution, and outcome phases

The platform event family MUST represent planning, step execution, evidence, user interaction, and final outcome states.

#### Scenario: event family spans the task lifecycle
- **WHEN** reviewers inspect the event model
- **THEN** the event family MUST cover plan creation, step status, tool activity, evidence, ask-user interactions, approval-required states, replanning, summaries, next actions, completion, and errors

### Requirement: gateway MUST project platform events through SSE compatibility

The gateway MUST project AI control-plane events through SSE without becoming the owner of their semantics.

#### Scenario: SSE transport remains compatible while semantics evolve
- **WHEN** the frontend consumes AI stream responses
- **THEN** SSE transport MUST remain available as the compatibility mechanism
- **AND** the semantic source of truth MUST remain the platform event model inside `internal/ai`

### Requirement: event stream MUST support mixed compatibility during rollout

The platform event stream MUST support a rollout period where card-native events coexist with compatibility text-oriented events.

#### Scenario: rollout preserves compatibility
- **WHEN** richer platform events are introduced
- **THEN** the backend MUST allow compatibility with existing consumers during rollout
- **AND** the new platform event semantics MUST remain available for newer application-card consumers
