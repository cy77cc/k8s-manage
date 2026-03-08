# AI Troubleshooting

## Approval task stays pending

- Verify the current user has `ai:approval:review` or is the routed approver.
- Check `ai_approval_tickets.status` and `expires_at`.
- Subscribe to `/api/v1/ai/approvals/stream` to confirm update events are being published.

## Approved task did not execute

- Inspect the approval response payload for the attached execution record.
- Check whether the task moved to `failed` and whether `reject_reason` contains the execution error.
- Verify the approval token was passed into tool execution and the underlying tool still exists in the registry.

## Feedback endpoint returns collector not initialized

- Redis-backed session state is required for automatic session Q&A extraction.
- As a fallback, send `question` and `answer` explicitly in `POST /api/v1/ai/feedback`.

## No RAG context in responses

- Confirm knowledge entries exist for the request namespace.
- The graph only augments prompts when the namespace retriever returns matching entries.
- If the AI runtime was started without the current in-process knowledge index populated, restart-time state will be empty until knowledge is re-added or feedback is re-ingested.
