# AI API

## Chat

- `POST /api/v1/ai/chat`
  Streams AI responses over SSE.
- `POST /api/v1/ai/approval/respond`
  Resumes an interrupt-based approval checkpoint.
- `POST /api/v1/ai/adk/resume`
  Resumes a checkpoint by explicit target payload.

## Tools

- `GET /api/v1/ai/capabilities`
- `POST /api/v1/ai/tools/preview`
- `POST /api/v1/ai/tools/execute`
- `GET /api/v1/ai/executions/:id`

## Approval Tasks

- `POST /api/v1/ai/approvals`
  Create a persisted approval task.
- `GET /api/v1/ai/approvals`
  List approval tasks visible to the current user.
- `GET /api/v1/ai/approvals/:id`
  Fetch one approval task.
- `POST /api/v1/ai/approvals/:id/approve`
  Approve and execute the task.
- `POST /api/v1/ai/approvals/:id/reject`
  Reject the task.
- `POST /api/v1/ai/approvals/:id/confirm`
  Compatibility endpoint for existing approval clients.
- `GET /api/v1/ai/approvals/stream`
  SSE stream for approval state changes and execution results.

## Feedback

- `POST /api/v1/ai/feedback`
  Ingest positive feedback into the AI knowledge index. The payload can provide `question` and `answer` directly or rely on the stored session transcript when Redis-backed session state is enabled.
