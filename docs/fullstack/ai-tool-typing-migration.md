# AI Tool Typing Migration (Fullstack)

## Scope
- `internal/ai/tools_registry.go`
- `internal/ai/tool_contracts.go`
- `internal/ai/tools_os.go`
- `internal/ai/tools_k8s.go`
- `internal/ai/tools_service.go`
- `internal/ai/tools_host.go`
- `internal/ai/tool_param_resolver.go`
- `internal/service/ai/chat_handler.go`
- `internal/service/ai/store.go`

## Migration Steps
1. Replace `map[string]any` tool input with typed structs.
2. Use `GoStruct2ToolInfo` to generate tool schema and required list.
3. Add resolver before policy/execution.
4. Add single retry on `missing_param`.
5. Persist per-scene last successful params in memory store.

## Backward Compatibility
- External API routes unchanged.
- SSE event names unchanged; payload includes extra debug fields.

## Validation
- `go test ./...`
- `npm run build`
