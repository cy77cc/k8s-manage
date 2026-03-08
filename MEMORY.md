# AI Module Rewrite Memory

## Architecture

The AI runtime now lives under `internal/ai` and is split into focused packages:

- `router`: intent classification and domain routing
- `graph`: action workflow for sanitize, reasoning, validation, and execution
- `aspect`: security callbacks for permission checks, interrupt handling, and audit hooks
- `state`: Redis-backed session and checkpoint storage
- `approval`: persisted approval tasks, approver routing, execution, and SSE updates
- `rag`: namespace-aware knowledge indexing, retrieval, and feedback ingestion

## Runtime Flow

1. `internal/service/ai/handler` accepts chat or approval API calls.
2. `internal/ai/orchestrator.go` creates or resumes a session and hands requests to `AIAgent`.
3. `AIAgent` routes the request by domain and executes the matching `ActionGraph`.
4. `ActionGraph` sanitizes input, augments the prompt with RAG context, validates tool calls, and executes tools.
5. Medium/high-risk actions can create `AIApprovalTask` rows and stream updates over SSE.
6. Positive feedback can be converted into namespace-scoped knowledge entries for later retrieval.

## Persistence

- `AIApprovalTask` is stored in `ai_approval_tickets` with task detail JSON, tool call JSON, approval token, and execution timestamps.
- Session/checkpoint state uses Redis when configured.
- RAG knowledge in the new AI-facing package is namespace-aware and currently uses an in-process index abstraction with a Milvus backend interface.
