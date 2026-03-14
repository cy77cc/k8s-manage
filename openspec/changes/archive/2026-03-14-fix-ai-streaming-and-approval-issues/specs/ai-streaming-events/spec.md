## MODIFIED Requirements

### Requirement: AI Chat Stream Events

The system SHALL emit Server-Sent Events (SSE) that preserve assistant output semantics across streaming, approval gating, resume, and compatibility clients.

Required compatibility event categories:

- `meta`: session metadata at stream start
- `stage_delta`: incremental model-stage output for rewrite, plan, approval-gate messaging, execute, and summary stages
- `delta`: incremental final answer content
- `thinking_delta`: model reasoning/thinking content when available
- `tool_call`: tool invocation with name and arguments
- `tool_result`: tool execution completion with result
- `approval_required`: approval gate requiring user confirmation before execution starts
- `heartbeat`: periodic keep-alive event
- `done`: stream completion
- `error`: error notification

Required turn/block-native event categories:

- `turn_started`: assistant turn lifecycle start
- `block_open`: a renderable block is created
- `block_delta`: incremental content or state update for a block
- `block_replace`: block payload replacement for structured cards
- `block_close`: block is finalized
- `turn_state`: assistant turn status or phase transition
- `turn_done`: assistant turn completion

The system MUST treat `delta` as append-only assistant text chunks rather than cumulative snapshots, and `stage_delta` MUST carry semantic fields required for UI and persistence consumers.

#### Scenario: assistant text delta remains append-only
- **WHEN** the runtime receives cumulative provider content snapshots during one assistant turn
- **THEN** the emitted `delta` events MUST contain only the newly appended assistant-visible text
- **AND** concatenating all emitted `delta` chunks in order MUST reconstruct the final assistant answer
- **AND** the system MUST NOT resend the full accumulated text as a later `delta`

#### Scenario: stage delta carries semantic planning metadata
- **WHEN** the runtime enters planning or updates a generated plan
- **THEN** each emitted `stage_delta` event for the plan stage MUST include a stable stage identifier and status
- **AND** planning milestones MUST include user-visible semantic fields such as `title`, `description`, and `steps` when that data exists
- **AND** consumers MUST NOT be forced to synthesize core plan semantics from hard-coded frontend templates alone

#### Scenario: internal structured payloads do not leak into assistant text
- **WHEN** the model or runtime produces internal structured payloads such as tool arguments or plan JSON
- **THEN** that payload MUST be routed through structured lifecycle events or filtered from the user-visible text stream
- **AND** the assistant `delta` stream MUST NOT expose raw internal payloads like `{\"steps\": [...]}` as normal prose

#### Scenario: approval gate appears before execution
- **WHEN** AI identifies a planned step that requires approval
- **THEN** the stream MUST emit `approval_required` before the gated step produces `tool_call` or `tool_result`
- **AND** the stream MUST include resumable runtime identity for that gate
- **AND** the user-visible lifecycle MUST show the turn waiting for approval rather than already executing the step

#### Scenario: heartbeat maintains connection
- **WHEN** streaming session lasts longer than 10 seconds
- **THEN** system emits `heartbeat` events every 10 seconds with timestamp
