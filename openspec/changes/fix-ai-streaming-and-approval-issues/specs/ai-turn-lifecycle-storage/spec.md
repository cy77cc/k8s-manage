## MODIFIED Requirements

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
