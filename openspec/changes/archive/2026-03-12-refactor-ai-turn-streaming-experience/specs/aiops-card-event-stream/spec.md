## MODIFIED Requirements

### Requirement: AI control plane MUST emit application-card-oriented platform events

The system MUST emit structured platform events that represent AIOps task progress, evidence, approvals, replanning, and next actions for application-card style rendering, and those platform events MUST be projectable into persisted turn blocks.

#### Scenario: backend emits card-native event semantics
- **WHEN** an AIOps task runs through the control plane
- **THEN** the backend MUST emit structured platform events for orchestration progress and operational state
- **AND** those events MUST be suitable for projection into application-card UI components
- **AND** each event MUST be attributable to a stable `turn_id` and an owning render block when user-visible

## MODIFIED Requirements

### Requirement: card event stream MUST cover plan, execution, and outcome phases

The platform event family MUST represent planning, step execution, evidence, user interaction, and final outcome states, and MUST distinguish lightweight status updates from durable user-visible content blocks.

#### Scenario: event family spans the task lifecycle
- **WHEN** reviewers inspect the event model
- **THEN** the event family MUST cover plan creation, step status, tool activity, evidence, ask-user interactions, approval-required states, replanning, summaries, next actions, completion, and errors
- **AND** the event family MUST define which events are projected as status, plan, tool, approval, evidence, text, or error blocks

## MODIFIED Requirements

### Requirement: gateway MUST project platform events through SSE compatibility

The gateway MUST project AI control-plane events through SSE without becoming the owner of their semantics, and MUST preserve both compatibility events and richer turn/block lifecycle events during rollout.

#### Scenario: SSE transport remains compatible while semantics evolve
- **WHEN** the frontend consumes AI stream responses
- **THEN** SSE transport MUST remain available as the compatibility mechanism
- **AND** the semantic source of truth MUST remain the platform event model inside `internal/ai`
- **AND** the gateway MUST be able to deliver both compatibility text-oriented events and turn/block lifecycle events in the same stream

## MODIFIED Requirements

### Requirement: event stream MUST support mixed compatibility during rollout

The platform event stream MUST support a rollout period where card-native events coexist with compatibility text-oriented events, and the rollout MUST allow newer clients to prefer turn/block rendering without breaking older clients.

#### Scenario: rollout preserves compatibility
- **WHEN** richer platform events are introduced
- **THEN** the backend MUST allow compatibility with existing consumers during rollout
- **AND** the new platform event semantics MUST remain available for newer application-card consumers
- **AND** disabling new rendering paths on the frontend MUST still leave the compatibility event family usable
