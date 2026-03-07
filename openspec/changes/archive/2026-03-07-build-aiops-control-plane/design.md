## Context

The current AI backend already hosts orchestration-related code in `internal/ai`, including `PlatformRunner`, `ControlPlane`, and an `Orchestrator` that owns chat and resume semantics. That is a useful transition state, but the runtime is still effectively organized around a single platform agent and a chat-first event flow. The product direction has moved beyond that. The next backend iteration must behave like an AIOps control plane rather than a chat assistant with tools.

Official CloudWeGo Eino ADK examples and documentation now provide a clearer target shape for this system: explicit `plan-execute-replan` orchestration, interrupt-aware human-in-the-loop flows, checkpointed resume, and supervisor/sub-agent composition. The current codebase already uses ADK runners, checkpoint stores, review/approval wrappers, and a tool registry, so this change should build on those assets instead of introducing a parallel orchestration substrate.

Constraints:
- Existing `/api/v1/ai` gateway routes and SSE transport should remain the compatibility surface during this phase.
- Existing approval, preview, confirmation, execution, and session capabilities must continue to work.
- The change is large and cross-cutting, so the architecture must create a stable backbone for later UI and protocol evolution without requiring another orchestration rewrite.

## Goals / Non-Goals

**Goals:**
- Introduce explicit planner, executor, and replanner roles under `internal/ai`.
- Shift the core runtime model from chat turns to task lifecycle: objective, plan, execution, evidence, interrupt, replan, outcome.
- Formalize domain executor routing for Host, K8s, Service, and Monitor domains.
- Define a platform event model aimed at application-card streams.
- Keep `internal/service/ai` as the gateway surface while moving semantics into the AI control plane.
- Reuse the current control-plane and tooling assets instead of replacing them wholesale.

**Non-Goals:**
- Implement every future domain workflow or automation path.
- Fully redesign the frontend drawer in the same change.
- Replace all existing AI routes or remove current SSE compatibility immediately.
- Introduce the final long-term multi-agent graph platform in one step.

## Decisions

### Decision: `internal/ai` hosts an explicit AIOps control-plane runtime

The system SHALL introduce an explicit orchestration runtime inside `internal/ai` with clear Planner, Executor, and Replanner responsibilities.

Rationale:
- The current `PlatformRunner` and single platform agent do not provide an explicit task-state backbone.
- The official ADK `plan-execute-replan` pattern is a better fit for operational tasks than a general single-agent loop.
- Future domain routing and card-event projection require stable orchestration contracts.

Alternatives considered:
- Continue evolving the current single platform agent.
  - Rejected because it keeps planning, execution, and replanning implicit and makes AIOps state hard to surface.
- Push orchestration back toward gateway-side services.
  - Rejected because ownership has already been correctly moved into `internal/ai`.

### Decision: planner outputs domain steps, not tool calls

The Planner SHALL emit domain-level steps such as host identification, K8s diagnosis, service deployment preview, or monitor investigation, rather than direct tool call lists.

Rationale:
- Domain steps decouple planning from the tool registry.
- Tool-level plans would tightly couple the planner to implementation details and make executor routing brittle.
- Domain steps map cleanly to card UI and operational review.

Alternatives considered:
- Planner emits concrete tool names and arguments.
  - Rejected because it collapses planning and execution into one layer and reduces flexibility.

### Decision: executor outputs execution records and evidence

Executors SHALL normalize work into structured execution records, evidence items, artifacts, and issues.

Rationale:
- Replanning requires structured facts rather than prose.
- Card-oriented frontend rendering needs evidence and status objects, not only `delta` text.
- Execution auditing and future automation require durable operational records.

Alternatives considered:
- Preserve text-first execution summaries and derive evidence in the frontend.
  - Rejected because it hides semantics in narrative output and makes the platform difficult to trust.

### Decision: replanner produces finite state decisions

The Replanner SHALL output limited control-plane decisions such as `continue`, `revise`, `ask_user`, `finish`, or `abort`.

Rationale:
- AIOps workflows need deterministic state transitions.
- Free-form replanning text is difficult to route, test, and render.
- Interrupt and approval flow needs machine-usable decisions.

Alternatives considered:
- Replanner outputs narrative reasoning only.
  - Rejected because it leaves state transitions ambiguous.

### Decision: domain executors become first-class modules

The control plane SHALL add first-class domain executors for Host, K8s, Service, and Monitor domains.

Rationale:
- The current single platform agent treats every domain as one tool pool.
- Different domains require different routing, evidence extraction, and operational constraints.
- Domain executors are the natural unit for future supervisor composition and deeper workflow specialization.

Alternatives considered:
- Keep one executor with conditional logic for all domains.
  - Rejected because it recreates a monolith under a new name.

### Decision: application-card event semantics originate in `internal/ai`

The backend SHALL define a platform event model in `internal/ai` for plan creation, step status, evidence, approvals, replan decisions, summaries, and next actions. Gateway code SHALL only project these semantics through SSE transport.

Rationale:
- Event semantics are business concepts, not transport details.
- Application-card UI depends on structured event meaning.
- Current `delta`-oriented behavior can remain as compatibility, but it should no longer define the system.

Alternatives considered:
- Continue treating SSE as primarily text and tool events.
  - Rejected because it does not support the desired AIOps product shape.

### Decision: gateway compatibility is preserved while semantics migrate

`internal/service/ai` SHALL remain the `/api/v1/ai` compatibility surface while the internal control plane evolves.

Rationale:
- This reduces rollout risk for a large backend refactor.
- It allows phased frontend adoption of richer events.
- It keeps auth, transport, and HTTP binding isolated from control-plane growth.

Alternatives considered:
- Redesign the API surface in the same change.
  - Rejected because it would mix architecture change with full protocol migration.

## Risks / Trade-offs

- [Risk] This is a broad change touching orchestration, domain routing, and streaming semantics. → Mitigation: implement against explicit contracts and phase tasks internally, even within one change.
- [Risk] The first control-plane iteration could still become a monolith if planner/executor/replanner are only cosmetic wrappers. → Mitigation: define separate contracts and modules for each role, with dedicated tests.
- [Risk] Card-native event projection could diverge from current frontend assumptions. → Mitigation: preserve SSE compatibility and introduce the richer event family as the semantic source of truth.
- [Risk] Domain executors may develop unevenly, with Host/K8s maturing faster than Service/Monitor. → Mitigation: allow asymmetric implementation depth, but require all four first-class module boundaries in this change.
- [Risk] Approval and confirmation flows could break if control-plane state is not consistently reused. → Mitigation: keep existing approval, preview, session, and confirmation services as orchestration dependencies and verify resume flows explicitly.

## Migration Plan

1. Define orchestration contracts for Plan, ExecutionRecord, Evidence, ReplanDecision, and NextAction under `internal/ai`.
2. Introduce planner, executor, and replanner modules, initially integrating them with the current runner and tooling substrate.
3. Add domain executor routing for Host, K8s, Service, and Monitor domains.
4. Add platform event types and event projection logic in `internal/ai`, while keeping SSE projection in the gateway.
5. Rewire current gateway handlers to consume the new control-plane runtime and preserve existing `/api/v1/ai` behavior.
6. Verify compatibility and leave follow-up room for deeper domain workflows and richer frontend cards.

Rollback strategy:
- Because the gateway surface remains stable, rollback can revert internal control-plane routing while preserving public routes and transport behavior.

## Open Questions

- Which current tool groups should become the initial step-kind taxonomy for planner output?
- Should the first executor router dispatch directly to domain modules or through a supervisor abstraction from day one?
- How much of the current `delta/thinking_delta` stream should remain visible once application-card events exist as the semantic source of truth?
