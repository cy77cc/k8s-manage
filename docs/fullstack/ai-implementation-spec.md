# AI Implementation Spec (Fullstack)

## Backend

- New APIs:
  - `GET /ai/capabilities`
  - `POST /ai/tools/preview`
  - `POST /ai/tools/execute`
  - `GET /ai/executions/:id`
  - `POST /ai/approvals`
  - `POST /ai/approvals/:id/confirm`
- Chat SSE adds:
  - `tool_call`
  - `tool_result`
  - `approval_required`

## Frontend

- `web/src/api/modules/ai.ts` adds:
  - `AICapability`
  - `ToolExecution`
  - `ApprovalTicket`
  - tool APIs and SSE handlers
