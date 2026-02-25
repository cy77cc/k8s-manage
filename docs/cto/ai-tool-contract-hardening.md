# AI Tool Contract Hardening (CTO)

## Goal
- Eliminate empty `{}` tool arguments by enforcing typed tool contracts and deterministic param resolution.

## Architecture
- Local tools use `InferTool[T,D]` with typed input structs and `jsonschema` tags.
- `ToolMeta` carries `schema`, `required`, `default_hint`, and examples for model alignment.
- Runtime flow:
  1. model arguments
  2. param resolver (runtime context > tool memory > safety defaults)
  3. policy check
  4. tool execution
  5. optional one-time retry on `missing_param`

## Reliability Rules
- Retry only once.
- Never override explicit model/user parameters.
- Mutating tools still require approval token.

## Observability
- SSE `tool_call/tool_result` payload includes:
  - `param_resolution`
  - `retry`
  - final params snapshot

## Risks
- Over-aggressive defaulting can hide user intent mistakes.
- Mitigation: white-list only, trace every applied default source.
