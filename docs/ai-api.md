# AI API

## Overview

The AI assistant now uses the agentic multi-expert orchestration path:

```text
/api/v1/ai/chat
  -> internal/service/ai transport shell
  -> internal/ai Orchestrator
  -> Rewrite
  -> Planner
  -> Executor Runtime
  -> Experts as Tools
  -> Summarizer
```

The transport contract is:

- `delta`: final assistant body stream
- `summary`: structured conclusion
- `stage_delta`: ThoughtChain stage stream
- `approval` / `clarify`: user action gates

## Chat

### `POST /api/v1/ai/chat`

Starts one assistant turn and streams SSE events.

Request:

```json
{
  "sessionId": "optional-session-id",
  "message": "查看 local 集群 kube-system 空间下的 cilium pod 日志",
  "context": {
    "scene": "deployment:k8s"
  }
}
```

Response:

- `Content-Type: text/event-stream`
- high-level events only

Primary SSE events:

- `meta`
- `rewrite_result`
- `planner_state`
- `plan_created`
- `stage_delta`
- `step_update`
- `tool_call`
- `tool_result`
- `approval_required`
- `clarify_required`
- `replan_started`
- `summary`
- `delta`
- `done`
- `error`

Example event order:

```text
meta
rewrite_result
planner_state
plan_created
step_update
tool_call
tool_result
summary
delta*
done
```

### `meta`

```json
{
  "sessionId": "3781107b-dd2b-462a-9467-1120293fb126",
  "session_id": "3781107b-dd2b-462a-9467-1120293fb126",
  "traceId": "c9d1c182-ef1d-4f68-9d20-922326be486a",
  "trace_id": "c9d1c182-ef1d-4f68-9d20-922326be486a",
  "createdAt": "2026-03-10T13:28:33Z"
}
```

### `rewrite_result`

```json
{
  "rewrite": {
    "normalized_goal": "Fetch pod logs and assess health",
    "operation_mode": "investigate",
    "resource_hints": {
      "cluster_name": "local",
      "namespace": "kube-system"
    },
    "narrative": "用户请求已被规整为可执行任务。"
  },
  "user_visible_summary": "用户请求已被规整为可执行任务。"
}
```

### `planner_state`

```json
{
  "status": "planning",
  "user_visible_summary": "正在根据 Rewrite 结果整理执行计划。"
}
```

### `plan_created`

```json
{
  "plan": {
    "plan_id": "plan-logs-cilium-87f2m",
    "goal": "Retrieve last 100 log lines and assess pod health.",
    "resolved": {
      "cluster_id": 1,
      "namespace": "kube-system",
      "pod_name": "cilium-87f2m"
    },
    "steps": [
      {
        "step_id": "step-1",
        "expert": "k8s",
        "mode": "readonly",
        "risk": "low"
      }
    ]
  },
  "user_visible_summary": "已生成结构化计划。"
}
```

### `stage_delta`

Used by the frontend ThoughtChain, not by the final answer body.

```json
{
  "stage": "rewrite",
  "status": "loading",
  "content_chunk": "开始理解你的问题并提取目标线索。"
}
```

Fields:

- `stage`: `rewrite | plan | execute | summary`
- `status`: `loading | success | error`
- `content_chunk`: incremental user-visible stage text
- `step_id`: optional for execute stage
- `expert`: optional for execute stage
- `replace`: optional overwrite mode

### `step_update`

```json
{
  "plan_id": "plan-logs-cilium-87f2m",
  "step_id": "step-1",
  "status": "running",
  "title": "Fetch Pod Logs",
  "expert": "k8s",
  "user_visible_summary": "正在获取目标 Pod 日志。"
}
```

### `approval_required`

```json
{
  "session_id": "session-1",
  "plan_id": "plan-1",
  "step_id": "step-2",
  "title": "等待你确认",
  "risk": "high",
  "mode": "mutating",
  "status": "waiting_approval",
  "user_visible_summary": "当前步骤需要审批后继续执行。",
  "resume": {
    "session_id": "session-1",
    "plan_id": "plan-1",
    "step_id": "step-2"
  }
}
```

### `clarify_required`

```json
{
  "kind": "clarify",
  "title": "需要你补充信息",
  "message": "当前没有可执行的集群上下文，需要先指定集群。",
  "candidates": []
}
```

### `summary`

Structured only. Do not render this as the final answer body.

```json
{
  "output": {
    "summary": "已生成结构化结论",
    "conclusion": "Cilium pod 运行正常，日志中未见错误。",
    "need_more_investigation": false,
    "narrative": "当前结论基于已执行步骤及其证据生成。"
  }
}
```

### `delta`

Final assistant answer stream.

```json
{
  "content_chunk": "Cilium pod cilium-87f2m 当前运行正常。"
}
```

### `done`

`done` closes the turn and returns the persisted session snapshot.

## Resume

### `POST /api/v1/ai/resume/step`

Canonical resume endpoint.

Request:

```json
{
  "session_id": "session-1",
  "plan_id": "plan-1",
  "step_id": "step-2",
  "approved": true,
  "reason": "approved by operator"
}
```

Response:

```json
{
  "resumed": true,
  "interrupted": false,
  "sessionId": "session-1",
  "session_id": "session-1",
  "plan_id": "plan-1",
  "step_id": "step-2",
  "status": "approved",
  "message": "审批已通过，待审批步骤会继续执行。"
}
```

### `POST /api/v1/ai/approval/respond`

Compatibility alias for `/api/v1/ai/resume/step`.

- same request handling
- same response shape
- intended for existing approval clients during migration

### `POST /api/v1/ai/adk/resume`

Legacy ADK compatibility endpoint.

Accepted compatibility fields:

- `checkpoint_id`
- `session_id`
- `plan_id`
- `step_id`

Behavior:

- `checkpoint_id` is mapped into `step_id`
- response includes `deprecated: true`
- response includes `compat_mode: "legacy_adk_resume"`
- clients should migrate to `session_id + plan_id + step_id`

## Sessions

### `GET /api/v1/ai/sessions?scene=<scene>`

Lists sessions for the current user and scene.

### `GET /api/v1/ai/sessions/current?scene=<scene>`

Returns the latest session for the current user and scene.

### `GET /api/v1/ai/sessions/:id?scene=<scene>`

Returns one session with full persisted messages.

Assistant messages may include:

- `thinking`
- `traceId`
- `thoughtChain`
- `recommendations`
- `status`

## Feedback

### `POST /api/v1/ai/feedback`

Stores user feedback for knowledge improvement.

## Notes

- `summary` is not a replacement for `delta`
- ThoughtChain should consume `rewrite_result / planner_state / plan_created / stage_delta / step_update / approval_required / clarify_required / summary`
- final user-visible answer body should consume `delta`
