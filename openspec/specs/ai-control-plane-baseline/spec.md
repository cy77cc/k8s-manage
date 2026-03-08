## MODIFIED Requirements

### Requirement: AI Chat SHALL Support SSE Streaming Contract
AI chat capability SHALL be specified with an SSE event contract that includes message streaming and completion signaling, while preserving a separation between gateway transport responsibilities and AI core event semantics. The baseline SHALL also recognize that AI streaming is evolving toward platform events that represent planning, execution, evidence, approvals, replanning, and next actions.

#### Scenario: SSE event family is defined
- **WHEN** reviewers inspect the AI baseline
- **THEN** the spec SHALL include `meta`, `delta`, `thinking_delta`, `tool_call`, `tool_result`, `approval_required`, `done`, and `error` as baseline stream events
- **AND** the gateway SHALL remain responsible for transport framing and delivery compatibility
- **AND** the AI core SHALL remain responsible for the semantic meaning of streamed execution and interrupt events
- **AND** the baseline SHALL allow richer platform events for planning, evidence, replanning, and next actions

### Requirement: AI gateway and AI core MUST have separate ownership boundaries
The baseline MUST define `internal/service/ai` as the gateway surface for AI APIs and streaming, and `internal/ai` as the owner of AI orchestration and control-plane semantics. The AI core MUST host the AIOps control-plane runtime rather than only a chat-oriented orchestration shim.

#### Scenario: ownership boundary is documented
- **WHEN** maintainers inspect the AI control-plane baseline
- **THEN** the baseline MUST describe `internal/service/ai` as handling routes, auth-aware request mapping, and transport delivery
- **AND** the baseline MUST describe `internal/ai` as handling orchestration, execution semantics, interrupt-aware flow, and AI platform behavior
- **AND** the baseline MUST recognize planner, executor, replanner, and domain executor routing as AI-core responsibilities

## ADDED Requirements

### Requirement: AI control plane MUST support structured task lifecycle state
The baseline MUST require the AI control plane to model objectives, plans, execution records, evidence, interrupts, and outcomes as structured task lifecycle state.

#### Scenario: task lifecycle state is part of the baseline
- **WHEN** reviewers inspect the AI control-plane baseline
- **THEN** the baseline MUST include structured task lifecycle concepts for planning, execution, evidence, interruption, replanning, and completion

### Requirement: AI control plane MUST include formal domain executor boundaries
The baseline MUST require formal domain executor boundaries for Host, K8s, Service, and Monitor task routing.

#### Scenario: domain executor boundaries are part of the baseline
- **WHEN** maintainers inspect the AI control-plane architecture
- **THEN** the baseline MUST require Host, K8s, Service, and Monitor executor boundaries as first-class routing targets

### Requirement: AI control plane MUST support rollout toggles for orchestration entry
The baseline MUST define a configuration toggle for selecting the multi-domain orchestration entrypoint during rollout. The toggle SHALL be exposed as `ai.use_multi_domain_arch`, default to `false`, and allow the existing agentic entrypoint to remain available as a fallback.

#### Scenario: multi-domain orchestration is gated by config
- **WHEN** operators enable `ai.use_multi_domain_arch`
- **THEN** agentic requests SHALL enter the multi-domain planning path
- **AND** simple-chat requests SHALL remain unaffected
- **AND** disabling the toggle SHALL preserve the legacy agentic path
