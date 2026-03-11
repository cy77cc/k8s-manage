# AI Rewrite To RAG Bridge

## Purpose

The Rewrite stage is the semantic entrypoint for both planning and retrieval-augmented generation.

The system does not send raw colloquial user text directly into retrieval as the primary signal. Instead, Rewrite produces a normalized semantic contract that both Planner and RAG consume.

## Rewrite Contract For RAG

Key fields:

- `normalized_goal`
- `operation_mode`
- `normalized_request`
- `resource_hints`
- `ambiguities`
- `retrieval_intent`
- `retrieval_queries`
- `retrieval_keywords`
- `knowledge_scope`
- `requires_rag`

Location:

- [rewrite.go](/root/project/k8s-manage/internal/ai/rewrite/rewrite.go)
- [rewrite_bridge.go](/root/project/k8s-manage/internal/ai/rag/rewrite_bridge.go)

## Integration Rules

- Planner and RAG must consume the same Rewrite semantic contract.
- RAG must not reconstruct intent by reparsing the raw user message independently.
- If Rewrite is unavailable or returns invalid output, the system must report Rewrite unavailability explicitly instead of falling back to deterministic code understanding.

## Example Flow

```text
user message
  -> Rewrite
     -> semantic contract
        -> Planner
        -> RAG rewrite bridge
```

## Operational Notes

- `requires_rag=true` indicates retrieval should be considered before or during later stages.
- `knowledge_scope` should constrain retrieval namespaces or domains.
- `retrieval_queries` should be treated as primary retrieval candidates.
- `retrieval_keywords` are secondary expansion hints, not replacements for semantic queries.

## Failure Policy

- Rewrite unavailable: fail the turn explicitly with Rewrite-stage unavailability.
- Invalid Rewrite output: fail the turn explicitly with Rewrite invalid output.
- Do not substitute a code-generated semantic contract pretending Rewrite succeeded.
