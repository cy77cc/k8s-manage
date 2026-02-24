# AI Release Test Plan (QA)

## Critical Paths

1. Login -> `GET /ai/capabilities` -> chat stream
2. readonly tool preview/execute success
3. mutating tool blocked without approval
4. mutating tool execute after approval
5. forbidden user denied on tool execute

## Security Tests

- command whitelist bypass attempts
- invalid approval token / expired token
- execution replay with mismatched tool name

## Regression

- `go test ./...`
- `npm run build`
