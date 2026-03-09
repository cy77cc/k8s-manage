## 1. SSE Tool Result Events

- [x] 1.1 Create `streamingCallbackHandler` in `internal/ai/callbacks.go` implementing `callbacks.Handler` with `OnEnd` for tool result emission
- [x] 1.2 Modify `AIAgent.Stream()` to use callback handler for `tool_result` event emission
- [x] 1.3 Add unit tests for callback handler tool result emission

## 2. SecurityAspect Integration

- [x] 2.1 Modify `NewAIAgent()` to accept optional `*aspect.SecurityAspect` parameter
- [x] 2.2 Wire `SecurityAspect.Middleware()` into `compose.ToolsNodeConfig.Middlewares`
- [x] 2.3 Ensure `approval_required` interrupt is mapped to SSE event emission
- [ ] 2.4 Add integration tests for permission-denied and approval-required scenarios

## 3. Resume Logic Enhancement

- [x] 3.1 Add `CheckpointStore` field to `RunnerConfig` in `internal/ai/agent.go`
- [x] 3.2 Configure `compose.WithCheckPointStore` when creating `react.Agent`
- [ ] 3.3 Update `AIAgent.Resume()` to properly handle checkpoint-based resume
- [ ] 3.4 Ensure `ResumePayload` in orchestrator passes approval result correctly
- [ ] 3.5 Add tests for resume after approval and rejection scenarios

## 4. Heartbeat Events

- [x] 4.1 Add `HeartbeatInterval` field to `RunnerConfig` (default 10s)
- [x] 4.2 Implement heartbeat ticker in `AIAgent.Stream()` streaming loop
- [x] 4.3 Emit `heartbeat` events with timestamp every interval
- [ ] 4.4 Add configuration test for custom heartbeat interval

## 5. Integration & Verification

- [x] 5.1 Run all existing tests to ensure no regressions
- [ ] 5.2 Test SSE stream with real frontend to verify event order
- [ ] 5.3 Test approval flow end-to-end (high-risk tool → approval_required → resume)
- [ ] 5.4 Verify heartbeat keeps connection alive during long operations
