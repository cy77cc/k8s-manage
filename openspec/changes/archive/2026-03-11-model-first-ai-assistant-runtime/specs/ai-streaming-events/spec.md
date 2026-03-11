## MODIFIED Requirements

### Requirement: AI Chat Stream Events

The system SHALL emit Server-Sent Events (SSE) that preserve model-stage output as faithfully as possible during AI chat streaming.

Required event categories:

- `meta`: session metadata at stream start
- `stage_delta`: incremental model-stage output for rewrite, plan, execute, and summary stages
- `delta`: incremental final answer content
- `thinking_delta`: model reasoning/thinking content when available
- `tool_call`: tool invocation with name and arguments
- `tool_result`: tool execution completion with result
- `approval_required`: high-risk tool requiring user approval
- `heartbeat`: periodic keep-alive event
- `done`: stream completion
- `error`: error notification

The system MUST treat `stage_delta` and `delta` as model-facing content channels rather than code-generated template placeholders.

#### Scenario: successful chat stream preserves model-stage output
- **WHEN** user sends a message that triggers rewrite, planning, execution, and summarization
- **THEN** the system MUST emit `stage_delta` events as stage content becomes available
- **AND** the system MUST emit `delta` events for the final answer before `done`
- **AND** the system MUST NOT wait until the entire request finishes before exposing all model-stage content

#### Scenario: unavailable stage is reported explicitly
- **WHEN** a model-backed stage becomes unavailable during streaming
- **THEN** the system MUST emit an explicit user-visible failure or unavailability signal for that stage
- **AND** the stream MUST NOT substitute code-generated semantic content pretending the stage completed successfully

#### Scenario: high-risk tool requires approval
- **WHEN** AI attempts to execute a tool marked as high-risk
- **THEN** system emits `approval_required` event with tool name, arguments, and preview
- **AND** stream pauses with runtime step identity for later resume

#### Scenario: heartbeat maintains connection
- **WHEN** streaming session lasts longer than 10 seconds
- **THEN** system emits `heartbeat` events every 10 seconds with timestamp

## MODIFIED Requirements

### Requirement: Checkpoint-based Resume

The system SHALL support resuming interrupted sessions by stable runtime identity rather than model-specific checkpoint semantics.

Resume inputs MUST be based on:

- `session_id`
- `plan_id`
- `step_id`
- approval decision

Compatibility aliases MAY exist for legacy clients, but runtime semantics MUST be plan-step based.

#### Scenario: resume after approval
- **WHEN** user submits approval for an interrupted session
- **THEN** system resumes execution for the specific `session_id + plan_id + step_id`
- **AND** continues emitting SSE events from that runtime point

#### Scenario: resume after rejection
- **WHEN** user rejects approval for an interrupted session
- **THEN** system returns a rejection message
- **AND** does not execute the interrupted step

#### Scenario: legacy resume compatibility does not redefine runtime identity
- **WHEN** a legacy client sends a checkpoint-style resume request
- **THEN** the system MAY translate that request into current runtime identifiers
- **AND** the canonical runtime identity MUST remain `session_id + plan_id + step_id`
