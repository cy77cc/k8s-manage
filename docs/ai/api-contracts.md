# AI API Contracts

## Overview

The AI assistant now exposes a single `/api/v1/ai` surface. The frontend entry is `/ai`, backed by the new `AIChatPage` and the SSE chat flow.

Canonical endpoints:

- `POST /api/v1/ai/chat`
- `POST /api/v1/ai/chat/respond`
- `GET /api/v1/ai/sessions`
- `GET /api/v1/ai/sessions/:id`
- `DELETE /api/v1/ai/sessions/:id`
- `GET /api/v1/ai/tools`

Auxiliary endpoints still used by the current V2 UI and approval flow:

- `GET /api/v1/ai/sessions/current`
- `POST /api/v1/ai/approval/respond`
- `POST /api/v1/ai/confirmations/:id/confirm`
- `GET /api/v1/ai/capabilities`

Compatibility and command-center endpoints remain available under the same `/api/v1/ai/*` namespace until the command workflow is retired.

## Chat

- `POST /api/v1/ai/chat`
- Response type: `text/event-stream`

SSE events:

- `meta`
- `delta`
- `thinking_delta`
- `tool_call`
- `tool_result`
- `approval_required`
- `review_required`
- `interrupt_required`
- `heartbeat`
- `done`
- `error`

### Request

```json
{
  "sessionId": "sess-optional",
  "message": "帮我查看服务状态",
  "context": {
    "scene": "services",
    "pageData": {
      "service_id": 10
    }
  }
}
```

### Context Rules

`context` supports:

- `scene`
- `pageData.cluster_id`
- `pageData.service_id`
- `pageData.host_id`
- `pageData.target_id`
- `pageData.namespace`
- `pageData.env`
- `selectedItems`
- `approval_token`
- `confirmation_token`

Backend flattens `pageData` into runtime context for tool parameter resolution.

### Done Event

`done` contains:

- `session`
- `stream_state`
- `tool_summary`
- `turn_recommendations`

Example:

```json
{
  "session": {
    "id": "sess-1",
    "title": "AI Session",
    "messages": []
  },
  "stream_state": "ok",
  "tool_summary": {
    "calls": 1,
    "results": 1
  },
  "turn_recommendations": [
    {
      "id": "rec-1",
      "title": "继续检查",
      "content": "查看最近一次执行记录"
    }
  ]
}
```

## Sessions

- `GET /api/v1/ai/sessions`
- `GET /api/v1/ai/sessions/current`
- `GET /api/v1/ai/sessions/:id`
- `DELETE /api/v1/ai/sessions/:id`

Session payload includes:

- `id`
- `scene`
- `title`
- `messages`
- `createdAt`
- `updatedAt`

## Tool Catalog

- `GET /api/v1/ai/tools`
- `GET /api/v1/ai/capabilities`

Both return permission-filtered tool metadata. `GET /ai/tools` is the canonical route for the V2 page; `GET /ai/capabilities` is kept as a compatibility alias.

Tool fields may include:

- `name`
- `description`
- `mode`
- `risk`
- `provider`
- `permission`
- `scene_scope`
- `param_hints`
- `enum_sources`

## Interactive Approval

### Resume From Chat Ask

- `POST /api/v1/ai/chat/respond`

Request:

```json
{
  "checkpoint_id": "sess-1",
  "target": "tool:approval-1",
  "approved": true
}
```

Response:

- `resumed: true` when execution continues normally
- `interrupted: true` when another approval/review interrupt is raised

### Compatibility Approval Resume

- `POST /api/v1/ai/approval/respond`

Used by the current frontend hook layer to resume interrupted tool execution. Payload and behavior mirror the chat respond flow.

### Confirmation

- `POST /api/v1/ai/confirmations/:id/confirm`

Used for explicit confirmation tickets outside the ADK interrupt resume path.

## Tool Execution

The following endpoints remain available for direct tool UX and regression coverage:

- `GET /api/v1/ai/tools/:name/params/hints`
- `POST /api/v1/ai/tools/preview`
- `POST /api/v1/ai/tools/execute`
- `GET /api/v1/ai/executions/:id`

Unified tool result payload:

```json
{
  "ok": true,
  "data": {},
  "error": "",
  "error_code": "",
  "source": "local|remote_ssh|db|deploy|preview|cluster_kubeconfig",
  "latency_ms": 12
}
```

## Notes

- Old `internal/service/ai/v2` route wrapper has been removed.
- Old floating AI widget and legacy `ChatInterface` page are no longer part of the frontend.
- `/ai` now always points to the new `AIChatPage`.
