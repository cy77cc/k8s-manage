# ThoughtChain Frontend Integration

## Purpose

The frontend AI panel uses ThoughtChain to show stage progress without mixing stage process text into the final assistant answer.

The model is:

- ThoughtChain: process visibility
- `delta`: final answer body
- `summary`: structured conclusion for stage state and program logic

## Stage Model

Supported stage keys:

- `rewrite`
- `plan`
- `execute`
- `user_action`
- `summary`

Current title mapping:

- `rewrite`: `理解你的问题`
- `plan`: `整理排查计划`
- `execute`: `调用专家执行`
- `user_action`: `等待你确认` or `等待你补充信息`
- `summary`: `生成结论`

## Event Mapping

### Rewrite

- `rewrite_result` marks the stage as created/success
- `stage_delta(stage=rewrite)` appends incremental visible text

### Plan

- `planner_state` creates or updates the stage as loading
- `plan_created` marks the stage as success
- `clarify_required` may terminate planning and move attention to `user_action`
- `replan_started` reopens the plan stage as loading
- `stage_delta(stage=plan)` appends planning text

### Execute

- `step_update` updates the aggregate execute stage
- `tool_call` and `tool_result` are detail items, not primary stages
- `stage_delta(stage=execute)` appends execution text

### User Action

- `approval_required` maps to `user_action`
- `clarify_required` maps to `user_action`
- this stage is user-facing and should not be merged into `execute`

### Summary

- `summary` marks the structured conclusion as ready
- `stage_delta(stage=summary)` streams stage text
- `delta` is still the final answer body and must remain separate

## Persistence

ThoughtChain is persisted per assistant message, not as a page-global buffer.

Persisted message metadata includes:

- `thoughtChain`
- `traceId`
- `recommendations`
- `status`

On restore:

- restore the exact `thoughtChain` snapshot
- reset `blink` to `false`
- if the message is already finished, loading states should be normalized to final statuses

## Rendering Rules

- default expanded keys should be empty
- ThoughtChain should render as soon as a message has either `thoughtChain`, `thinking`, or `content`
- the assistant bubble must not stay hidden until final `delta`
- `summary.output.narrative` and `summary.output.conclusion` should not be dumped directly into ThoughtChain content

## Event Completeness

Frontend metrics should track:

- expected stage keys inferred from high-level events
- delivered stage keys inferred from `stage_delta`
- rendered stage keys inferred from persisted/restored `thoughtChain`

Current helper:

- [thoughtChainMetrics.ts](/root/project/k8s-manage/web/src/components/AI/thoughtChainMetrics.ts)

## Implementation References

- [Copilot.tsx](/root/project/k8s-manage/web/src/components/AI/Copilot.tsx)
- [types.ts](/root/project/k8s-manage/web/src/components/AI/types.ts)
- [useConversationRestore.ts](/root/project/k8s-manage/web/src/components/AI/hooks/useConversationRestore.ts)
- [session_recorder.go](/root/project/k8s-manage/internal/service/ai/session_recorder.go)
