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

## Capabilities

- `GET /api/v1/ai/capabilities`
- Returns enabled tools based on permission.

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

## Approval

- `POST /api/v1/ai/approvals`
- `POST /api/v1/ai/approvals/:id/confirm`
