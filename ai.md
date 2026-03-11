# AI Architecture

## Overview

The current AI runtime is an orchestrated, stage-based system designed to keep model reasoning and execution control separate.

```text
Gateway / API
  -> internal/service/ai
  -> internal/ai Orchestrator
     -> Rewrite
     -> Planner
     -> Executor Runtime
        -> Expert Agent Tools
     -> Summarizer
```

Core rule:

- models decide and explain
- runtime code schedules, persists, resumes, and enforces safety

## Layer Boundaries

### Gateway / Transport

Location:

- [handler.go](/root/project/k8s-manage/internal/service/ai/handler.go)
- [routes.go](/root/project/k8s-manage/internal/service/ai/routes.go)

Responsibilities:

- HTTP/SSE framing
- auth
- request mapping
- session shell

Non-responsibilities:

- planning
- execution state transitions
- approval semantics
- stage policy

### Orchestrator Host

Location:

- [orchestrator.go](/root/project/k8s-manage/internal/ai/orchestrator.go)

Responsibilities:

- `RunRequest` / `ResumeRequest`
- trace and session lifecycle
- stage composition
- event emission
- metrics recording

### Rewrite

Location:

- [rewrite.go](/root/project/k8s-manage/internal/ai/rewrite/rewrite.go)

Responsibilities:

- normalize colloquial input
- extract resource hints
- preserve ambiguity flags

Must not:

- fabricate resource IDs
- claim permission outcomes
- claim execution conclusions

### Planner

Location:

- [planner.go](/root/project/k8s-manage/internal/ai/planner/planner.go)

Responsibilities:

- resolve resources
- decide `clarify / reject / direct_reply / plan`
- emit structured `ExecutionPlan`
- enforce prerequisite validation on structured fields only

Key properties:

- `ResolvedResources` supports single, multi, and scoped resources
- `step.input` is the execution truth source after normalization

### Executor Runtime

Location:

- [executor.go](/root/project/k8s-manage/internal/ai/executor/executor.go)
- [scheduler.go](/root/project/k8s-manage/internal/ai/executor/scheduler.go)

Responsibilities:

- deterministic DAG scheduling
- step state transitions
- approval waiting
- resume and idempotency
- retry policy
- expert dispatch

The executor is runtime code, not a freeform agent.

### Experts

Location:

- [registry.go](/root/project/k8s-manage/internal/ai/experts/registry.go)

Domains:

- hostops
- k8s
- service
- delivery
- observability

Experts are exposed as tools to the executor and cannot access planner-only support tools.

### Summarizer

Location:

- [summarizer.go](/root/project/k8s-manage/internal/ai/summarizer/summarizer.go)

Responsibilities:

- synthesize `StepResult` and `Evidence`
- produce structured summary
- decide `need_more_investigation`
- support replan signaling

## Persistence Model

### Runtime State

Execution runtime state is persisted separately from chat history.

Location:

- [execution_state.go](/root/project/k8s-manage/internal/ai/runtime/execution_state.go)

Stores:

- `trace_id`
- `session_id`
- `plan_id`
- step states
- pending approval
- runtime context snapshot

### Chat History

Chat history now uses database-backed message snapshots.

Locations:

- [chat_store.go](/root/project/k8s-manage/internal/ai/state/chat_store.go)
- [ai_chat.go](/root/project/k8s-manage/internal/model/ai_chat.go)
- [session_recorder.go](/root/project/k8s-manage/internal/service/ai/session_recorder.go)

Persisted assistant metadata includes:

- `thoughtChain`
- `traceId`
- `recommendations`
- `status`

## Event Model

Primary high-level events:

- `meta`
- `rewrite_result`
- `planner_state`
- `plan_created`
- `stage_delta`
- `step_update`
- `approval_required`
- `clarify_required`
- `replan_started`
- `summary`
- `delta`
- `done`
- `error`

Design intent:

- ThoughtChain consumes stage-oriented events
- final body consumes `delta`
- `tool_call` and `tool_result` are detail signals

## Frontend Integration

Primary UI path:

```text
AppLayout
  -> AICopilotButton
  -> AIAssistantDrawer
  -> CopilotSurface
  -> Copilot
```

Locations:

- [Copilot.tsx](/root/project/k8s-manage/web/src/components/AI/Copilot.tsx)
- [CopilotSurface.tsx](/root/project/k8s-manage/web/src/components/AI/CopilotSurface.tsx)

Rules:

- ThoughtChain is per assistant message
- default state is collapsed
- scene switching must isolate session restore
- final answer body must not be synthesized from `summary`

## Rollout And Compatibility

Configuration:

- `ai.use_multi_domain_arch`

Current compatibility endpoints:

- `/api/v1/ai/approval/respond`
- `/api/v1/ai/adk/resume`

Current canonical endpoint:

- `/api/v1/ai/resume/step`

## Observability

Lightweight runtime metrics are recorded in:

- [metrics.go](/root/project/k8s-manage/internal/ai/metrics.go)

Current metrics groups:

- rewrite quality
- planner clarify / executable plan rate
- resume success / duplicate intercept rate
- ThoughtChain completeness
