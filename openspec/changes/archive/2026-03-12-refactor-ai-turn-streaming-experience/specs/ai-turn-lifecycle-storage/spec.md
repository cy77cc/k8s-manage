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

### Requirement: session replay MUST reconstruct the same user-visible turn structure

The system MUST be able to reconstruct session history from persisted turn data so historical playback matches the user-visible structure of the original live interaction.

#### Scenario: session detail replays turn blocks
- **WHEN** a client requests AI session history
- **THEN** the system MUST return enough data to reconstruct ordered turns and their blocks
- **AND** completed, waiting-user, partial, and failed turn states MUST remain distinguishable in the replayed response

### Requirement: compatibility message projection MUST remain available during rollout

The system MUST support a rollout period where structured turn persistence coexists with legacy message-oriented history consumers.

#### Scenario: legacy message consumer remains supported
- **WHEN** a compatibility client still expects message-oriented session history
- **THEN** the system MUST project persisted turn data or dual-written state into the legacy response shape
- **AND** the canonical persisted state for new behavior MUST remain turn-and-block based
