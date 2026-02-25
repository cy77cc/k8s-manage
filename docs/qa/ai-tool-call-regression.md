# AI Tool Calling Regression Plan (QA)

## Core Cases
1. Empty arguments from model:
- Expected: resolver fills defaults/context and tool succeeds when possible.
2. Required field still missing after resolution:
- Expected: one retry only, then structured `missing_param` error.
3. Mutating tool without approval token:
- Expected: `approval_required`, no execution.
4. Mutating tool with approved token:
- Expected: execution success/failure with full trace.
5. SSE trace fields:
- Expected: `param_resolution` and `retry` present for tool events.

## Smoke Commands
- `go test ./...`
- `npm run build`

## Release Gate
- No panic on tool calling.
- No endless retry loop.
- Tool result always emitted after tool call (success or structured failure).
