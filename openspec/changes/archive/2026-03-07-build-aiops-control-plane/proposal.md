## Why

The current AI backend has moved some orchestration ownership into `internal/ai`, but it is still fundamentally a chat-oriented assistant built around a single platform agent, compatibility SSE events, and gateway-driven interaction loops. That is not enough for the AIOps platform direction we have chosen. We now need a real AI control plane with explicit `plan-execute-replan` orchestration, domain executors, interrupt-aware operational workflows, and application-card event streams.

The opportunity is to establish the first complete AIOps control-plane iteration in one coherent backend change, instead of continuing to evolve the system through smaller chat-centric refactors that would keep the architecture fragmented.

## What Changes

- Introduce an explicit AIOps control-plane runtime in `internal/ai` with planner, executor, and replanner responsibilities.
- Replace the current single-agent execution shape with a task-lifecycle model centered on plans, execution records, evidence, interrupts, and replan decisions.
- Add a formal domain executor routing layer for Host, K8s, Service, and Monitor operations.
- Introduce a backend event model oriented around application-card streams rather than primarily around chat deltas.
- Preserve the existing `/api/v1/ai` gateway surface while shifting internal execution semantics to the AI control plane.
- Reuse and formalize existing approval, preview, execution, and session control-plane capabilities as orchestration dependencies.
- **BREAKING (internal architecture):** `internal/ai` will stop treating the single platform agent as the main orchestration abstraction and will instead host a multi-role orchestration runtime.

## Capabilities

### New Capabilities
- `aiops-control-plane`: planner/executor/replanner orchestration runtime, execution records, evidence, interrupts, and domain executor routing for AIOps tasks
- `aiops-card-event-stream`: backend platform event model for application-card style progress, evidence, approvals, replan decisions, and next actions

### Modified Capabilities
- `ai-assistant-adk-architecture`: ADK architecture requirements will expand from basic AI-core-owned orchestration to explicit multi-role `plan-execute-replan` control-plane orchestration
- `ai-control-plane-baseline`: baseline AI control-plane requirements will expand to include domain executors, structured execution state, and card-oriented platform events
- `ai-assistant-drawer`: the assistant drawer requirements will evolve from chat-first rendering toward consuming richer application-card-oriented AI events while preserving compatibility during rollout

## Impact

- Affected backend code:
  - `internal/ai/**`
  - `internal/service/ai/**`
  - selected wiring in `internal/svc/**`
- Affected architectural surfaces:
  - AI orchestration runtime
  - domain tool execution routing
  - interrupt/approval/resume coordination
  - backend streaming event semantics
- API and frontend impact:
  - `/api/v1/ai` routes remain the compatibility surface in this phase
  - SSE transport remains in place, but event semantics will expand toward application-card consumption
- Dependency and design impact:
  - stronger alignment with CloudWeGo Eino ADK `plan-execute-replan` and supervisor-style orchestration patterns
  - future AI features will build on the control-plane runtime instead of extending a single platform agent
