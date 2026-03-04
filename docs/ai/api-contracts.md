# AI API Contracts

## Chat

- `POST /api/v1/ai/chat`
- Content-Type: `text/event-stream`
- Events:
  - `meta`
  - `delta`
  - `tool_call`
  - `tool_result`
  - `approval_required`
  - `done`
  - `error`

### Chat Context

Request `context` now supports:

- `scene`
- `pageData.cluster_id`
- `pageData.service_id`
- `pageData.host_id`
- `pageData.target_id`
- `pageData.namespace`
- `pageData.env`
- `selectedItems`

Backend will flatten `pageData` into runtime context for tool parameter resolution.

## Capabilities

- `GET /api/v1/ai/capabilities`
- Returns enabled tools based on permission.
- Tool fields include:
  - `enum_sources`
  - `param_hints`
  - `related_tools`
  - `scene_scope`

## Tool Parameter Hints

- `GET /api/v1/ai/tools/:name/params/hints`

Response example:

```json
{
  "tool": "service_get_detail",
  "params": {
    "service_id": {
      "type": "integer",
      "required": true,
      "hint": "可从 service_list_inventory 获取",
      "enum_source": "service_list_inventory",
      "values": [
        {"value": 1, "label": "svc-a"}
      ]
    }
  }
}
```

## Scene Recommendation

- `GET /api/v1/ai/scene/:scene/tools`
- Returns scene metadata and recommended tool list.

## Tool Execution

- `POST /api/v1/ai/tools/preview`
- `POST /api/v1/ai/tools/execute`
- `GET /api/v1/ai/executions/:id`

Unified result payload:

```json
{
  "ok": true,
  "data": {},
  "error": "",
  "source": "local|remote_ssh|db|deploy|preview|default_clientset|cluster_kubeconfig",
  "latency_ms": 12
}
```

Tool failure payload may include `execution_error`:

```json
{
  "code": "invalid_param",
  "message": "...",
  "recoverable": true,
  "suggestions": ["..."],
  "hint_action": "修正参数格式后重试"
}
```

## Commands

- `GET /api/v1/ai/commands/suggestions?scene=&q=`
- `POST /api/v1/ai/commands/preview`
- `POST /api/v1/ai/commands/execute`
- `GET /api/v1/ai/commands/history`
- `GET /api/v1/ai/commands/history/:id`

### Command Alias

- `GET /api/v1/ai/commands/aliases?scene=`
- `POST /api/v1/ai/commands/aliases`
- `DELETE /api/v1/ai/commands/aliases/:alias?scene=`

### Command Template

- `GET /api/v1/ai/commands/templates?scene=`
- `POST /api/v1/ai/commands/templates`

## Approval

- `POST /api/v1/ai/approvals`
- `POST /api/v1/ai/approvals/:id/confirm`
