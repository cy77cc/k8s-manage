## ADDED Requirements

### Requirement: AI chat persistence MUST model assistant output as turns and blocks

The system MUST persist AI chat history using a structured lifecycle model where each assistant response is stored as a turn containing ordered renderable blocks rather than only a single flattened message string.

#### Scenario: assistant response persists as structured turn
- **WHEN** an AI chat request starts producing assistant-visible output
- **THEN** the system MUST create or update an assistant turn record for that request
- **AND** the system MUST persist ordered blocks for status, text, tool, approval, evidence, thinking, or error content as they become available

### Requirement: persisted turn state MUST support interruption and resume

The persistence model MUST preserve enough identity and lifecycle state for interrupted execution to continue on the same assistant turn after approval or clarification.

#### Scenario: approval resume continues existing turn
- **WHEN** a running assistant turn enters approval-required state
- **THEN** the system MUST persist the relationship between session, turn, plan step, and pending approval state
- **AND** a later resume action MUST continue updating that same turn instead of creating an unrelated assistant response

#### Scenario: persistence failure remains diagnosable
- **WHEN** session, user-message, or assistant-message persistence fails during streaming
- **THEN** the system MUST emit operational diagnostics sufficient to identify the failed persistence stage
- **AND** the runtime MUST avoid silently dropping persisted user-visible history without traceable evidence

### Requirement: session replay MUST reconstruct the same user-visible turn structure

The system MUST be able to reconstruct session history from persisted turn data so historical playback matches the user-visible structure of the original live interaction.

#### Scenario: session detail replays turn blocks
- **WHEN** a client requests AI session history
- **THEN** the system MUST return enough data to reconstruct ordered turns and their blocks
- **AND** completed, waiting-user, partial, and failed turn states MUST remain distinguishable in the replayed response

### Requirement: compatibility message projection MUST remain available during rollout

The system MUST support a rollout period where structured turn persistence coexists with legacy message-oriented history consumers.

Compatibility projection MUST preserve both user prompts and assistant output even when runtime metadata arrives asynchronously or partially. Persisting the user prompt MUST NOT depend solely on a later streaming meta event containing a preexisting session identifier.

#### Scenario: user prompt persists before complete stream metadata arrives
- **WHEN** a chat request is accepted and the user prompt is known before all streaming metadata is available
- **THEN** the system MUST ensure a stable session identity exists or is created for persistence
- **AND** the user prompt MUST be persisted without waiting for a later meta event to make persistence possible
- **AND** later runtime metadata MUST enrich the same persisted session and assistant turn instead of creating a disconnected record

#### Scenario: compatibility client continues reading messages
- **WHEN** a legacy client requests session history
- **THEN** the system MUST still include message-compatible fields used by the current frontend
- **AND** the addition of structured turn replay fields MUST NOT remove required existing compatibility fields during rollout
