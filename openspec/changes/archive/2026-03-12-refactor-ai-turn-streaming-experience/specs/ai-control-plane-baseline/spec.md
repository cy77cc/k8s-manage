## MODIFIED Requirements

### Requirement: AI Chat SHALL Support SSE Streaming Contract
AI chat capability SHALL be specified with an SSE event contract that includes message streaming and completion signaling, while preserving a separation between gateway transport responsibilities and AI core event semantics. The baseline SHALL also recognize that AI streaming is evolving toward platform events and turn/block lifecycle events that represent planning, execution, evidence, approvals, replanning, next actions, and final answer rendering.

#### Scenario: SSE event family is defined
- **WHEN** reviewers inspect the AI baseline
- **THEN** the spec SHALL include `meta`, `delta`, `thinking_delta`, `tool_call`, `tool_result`, `approval_required`, `done`, and `error` as baseline stream events
- **AND** the gateway SHALL remain responsible for transport framing and delivery compatibility
- **AND** the AI core SHALL remain responsible for the semantic meaning of streamed execution and interrupt events
- **AND** the baseline SHALL allow richer platform events and turn/block lifecycle events for planning, evidence, replanning, and next actions

## MODIFIED Requirements

### Requirement: AI gateway and AI core MUST have separate ownership boundaries
The baseline MUST define `internal/service/ai` as the gateway surface for AI APIs and streaming, and `internal/ai` as the owner of AI orchestration, turn projection, and control-plane semantics. The AI core MUST host the AIOps control-plane runtime rather than only a chat-oriented orchestration shim.

#### Scenario: ownership boundary is documented
- **WHEN** maintainers inspect the AI control-plane baseline
- **THEN** the baseline MUST describe `internal/service/ai` as handling routes, auth-aware request mapping, and transport delivery
- **AND** the baseline MUST describe `internal/ai` as handling orchestration, execution semantics, interrupt-aware flow, turn/block projection, and AI platform behavior
- **AND** the baseline MUST recognize planner, executor, replanner, and domain executor routing as AI-core responsibilities

## ADDED Requirements

### Requirement: AI control plane MUST anchor lifecycle state to persisted turns
The baseline MUST require the AI control plane to bind streaming, approval, resume, and outcome behavior to a persisted assistant turn identity in addition to execution state.

#### Scenario: execution lifecycle remains attached to one turn
- **WHEN** an agentic chat request begins producing assistant-visible output
- **THEN** the control plane MUST assign a stable assistant `turn_id`
- **AND** approvals, evidence, and resumed execution MUST continue against that same turn until it reaches a terminal state
