# AI Troubleshooting

## Model-first runtime appears unavailable

- Check startup logs for layer-specific health checks: `rewrite`, `planner`, `expert`, `summarizer`.
- Inspect SSE `meta` and response headers:
  - `X-AI-Runtime-Mode`
  - `X-AI-Compatibility-Enabled`
  - `X-AI-Model-First-Enabled`
- If `runtime_mode=disabled`, verify:
  - `ai.use_multi_domain_arch`
  - `feature_flags.ai_model_first_runtime`
- If `runtime_mode=compatibility`, verify whether `feature_flags.ai_legacy_semantic_fallback` was enabled intentionally for rollback.

## Stage-specific failure appears in chat

- Inspect SSE `error.stage` and `error.error_code`.
- Expected layer-specific failures include:
  - `rewrite_*`
  - `planner_*`
  - `expert_*`
  - `summarizer_*`
- The system should report stage failure explicitly rather than silently fabricating semantic output. If you see fabricated stage content, treat that as a regression.

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
- Confirm Rewrite is producing retrieval fields:
  - `retrieval_intent`
  - `retrieval_queries`
  - `retrieval_keywords`
  - `knowledge_scope`
- The graph only augments prompts when the namespace retriever returns matching entries.
- If the AI runtime was started without the current in-process knowledge index populated, restart-time state will be empty until knowledge is re-added or feedback is re-ingested.
