# AI Runtime Migration Guide

## Scope

This document describes the migration from the old checkpoint-oriented AI chain to the current orchestrated multi-expert runtime.

It also covers the model-first rollout toggle and rollback posture.

## Architecture Migration

Old shape:

```text
handler -> monolithic agent -> tool events
```

Current shape:

```text
handler -> orchestrator -> rewrite -> planner -> executor runtime -> experts -> summarizer
```

Key changes:

- orchestration is now owned by `internal/ai`
- execution state is explicit and durable
- experts are isolated by domain
- frontend consumes high-level orchestration events
- semantic authority belongs to Rewrite / Planner / Experts / Summarizer
- runtime code no longer impersonates missing model reasoning with deterministic semantic fallbacks

## Rollout Migration

Primary flags:

- `ai.use_multi_domain_arch`
- `feature_flags.ai_model_first_runtime`
- `feature_flags.ai_legacy_semantic_fallback`

Recommended target:

```yaml
ai:
  use_multi_domain_arch: true

feature_flags:
  ai_model_first_runtime: true
  ai_legacy_semantic_fallback: false
```

Rollback posture:

- keep `ai.use_multi_domain_arch=true` for the stage-based runtime
- set `feature_flags.ai_model_first_runtime=false` only for operational rollback
- if rollback is required, treat the deployment as compatibility mode and instruct users to fall back to manual operations or compatibility endpoints as documented
- do not reintroduce hidden code semantic fallbacks as a silent runtime behavior

## Resume Migration

Old mental model:

- resume by `checkpoint_id`

Current model:

- resume by `session_id + plan_id + step_id`

Migration rules:

- move clients to `POST /api/v1/ai/resume/step`
- keep `/api/v1/ai/approval/respond` only as a compatibility alias
- treat `/api/v1/ai/adk/resume` as deprecated compatibility only
- if a client still sends `checkpoint_id`, map it into `step_id` during migration only

## Event Model Migration

Old assumptions:

- body stream and tool traces were mixed
- some clients expected low-level or compatibility-only events
- approval/clarify semantics were not sharply separated

Current model:

- ThoughtChain consumes stage-oriented events
- final answer body consumes `delta`
- `tool_call` / `tool_result` are supplementary only
- `summary` is structured only
- `error` carries explicit `error_code` and `stage`
- `heartbeat` keeps long streams alive

Removed or deprecated assumptions:

- `confirmation_required` as a chat-runtime primary path
- `tool_intent_unresolved`
- `expert_progress`
- page-global ThoughtChain state
- checkpoint-driven approval UX

## Session Migration

Old history source:

- Redis session snapshots and runtime state

Current history source:

- database-backed `AIChatSession / AIChatMessage`

Migration target:

- persist assistant message snapshots, not just raw text
- restore `thoughtChain`, `traceId`, `recommendations`, and `status`
- isolate history by `scene`

## Frontend Migration

Move from:

- waiting for final body before showing assistant state
- global thought buffer
- compatibility event branches

Move to:

- show assistant bubble as soon as ThoughtChain exists
- bind ThoughtChain to each assistant message
- consume `stage_delta` for stage process text
- keep final answer body driven by `delta`

## Compatibility Policy

Still supported:

- `/api/v1/ai/approval/respond`
- `/api/v1/ai/adk/resume`
- `message` as a compatibility alias for final answer chunk handling in some clients

Operational compatibility:

- response headers expose `X-AI-Runtime-Mode`
- SSE `meta` exposes `runtime_mode`, `model_first_enabled`, and `compatibility_enabled`

Not recommended for new clients:

- `checkpoint_id`
- compatibility-only SSE branches
- old monolithic chat-runtime assumptions

## Validation Checklist

- chat works through `/api/v1/ai/chat`
- approval resumes through `/api/v1/ai/resume/step`
- ThoughtChain persists across refresh
- session restore is filtered by scene
- final answer body is streamed through `delta`
