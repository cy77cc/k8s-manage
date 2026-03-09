## Why

The recent refactoring to use `react.Agent` simplified the AI module architecture but broke several critical frontend-backend integration points: the new implementation lacks proper SSE streaming events for tool execution, the SecurityAspect is not wired into the tool execution flow, and the Resume logic needs completion for approval-based interruptions. Additionally, heartbeat events are needed to maintain long-lived SSE connections.

## What Changes

1. **Add SSE streaming events to AIAgent.Stream()**:
   - Emit `tool_result` events when tool execution completes
   - Ensure `delta`, `thinking_delta`, `tool_call` events are properly streamed
   - Handle interrupt events (`approval_required`) from ApprovableTool/ReviewableTool wrappers

2. **Integrate SecurityAspect into tool execution**:
   - Wire SecurityAspect.Middleware() into the react.Agent's ToolsNode
   - Ensure approval interruption flows through the SSE stream
   - Map tool.StatefulInterrupt to `approval_required` SSE event

3. **Complete Resume logic for approval flow**:
   - Implement proper checkpoint storage for interrupted sessions
   - Add `ResumePayload` endpoint integration with checkpoint-based resume
   - Support multi-turn approval scenarios

4. **Add heartbeat to SSE stream**:
   - Emit periodic `heartbeat` events during long-running operations
   - Configure heartbeat interval via RunnerConfig

## Capabilities

### New Capabilities

- `ai-streaming-events`: SSE event emission protocol for AI chat sessions including delta, thinking_delta, tool_call, tool_result, approval_required, heartbeat

### Modified Capabilities

- `ai-tool-execution`: Integrate SecurityAspect middleware for approval-based tool interruption

## Impact

- **Backend Files**:
  - `internal/ai/agent.go` - Add tool_result streaming, SecurityAspect integration
  - `internal/ai/orchestrator.go` - Heartbeat coordination
  - `internal/ai/aspect/security.go` - Wire middleware into react.Agent

- **Frontend Integration**:
  - Frontend already expects these events (see `web/src/api/modules/ai.ts`)
  - No frontend changes required - only backend emission fixes

- **API Endpoints**:
  - `/api/v1/ai/chat/stream` - SSE event emission
  - `/api/v1/ai/chat/resume` - Checkpoint-based resume
