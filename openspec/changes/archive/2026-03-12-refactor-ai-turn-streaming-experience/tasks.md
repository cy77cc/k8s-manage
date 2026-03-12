## 1. Backend Turn/Block Runtime

- [x] 1.1 Introduce turn/block domain models and projector interfaces under `internal/ai` so orchestration results can be emitted as stable assistant turn lifecycle updates.
- [x] 1.2 Refactor the AI orchestrator streaming path to create `turn_started`, block lifecycle, turn-state, and compatibility SSE events from one projection flow.
- [x] 1.3 Extend execution runtime state in `internal/ai/runtime` to persist `turn_id` and resume linkage so approval/recovery continues the original assistant turn.
- [x] 1.4 Add backend tests covering successful stream, approval-required stream, resume continuation, and compatibility event coexistence.

## 2. Persistence and Migration

- [x] 2.1 Add database models and migrations for `ai_chat_turns`, `ai_chat_blocks`, and any required event/audit tables with forward and rollback support.
- [x] 2.2 Implement storage services that write and read structured turn/block state while preserving `ai_chat_sessions` ownership semantics.
- [x] 2.3 Update chat recording/session replay paths to dual-write legacy message records and new turn/block records during rollout.
- [x] 2.4 Add migration and repository tests validating turn/block persistence, replay ordering, and compatibility projections.

## 3. API and Resume Streaming

- [x] 3.1 Extend `/api/v1/ai/chat` SSE payload handling to expose turn/block-native events alongside existing compatibility events.
- [x] 3.2 Add `/api/v1/ai/resume/step/stream` for streaming continuation while preserving `/api/v1/ai/resume/step` as the compatibility JSON endpoint.
- [x] 3.3 Update `api/ai/v1` session and replay contracts to expose structured turn/block history while retaining compatibility fields for legacy consumers.
- [x] 3.4 Update AI session detail/list responses to return the new contract shape from persisted turn/block data.
- [x] 3.5 Add API-level tests for chat stream, session replay, and resume semantics across old and new clients.

## 4. Frontend Turn/Block UX

- [x] 4.1 Replace the current AI message accumulation logic with a reducer-driven turn/block store in `web/src/components/AI/**`.
- [x] 4.2 Implement renderers for status, text, plan, tool, approval, evidence, thinking, and error blocks with safe fallback behavior.
- [x] 4.3 Add a buffered typing renderer for final-answer text blocks and state-based updates for tool, approval, and evidence cards.
- [x] 4.4 Ensure approval actions, resumed execution, and historical session replay all operate on the same visible assistant turn.
- [x] 4.5 Add an explicit drawer-level display mode toggle with persisted client preference so `normal` and `debug` rendering policies are user-controlled and independent from rollout flags.
- [x] 4.6 Implement conditional auto-follow scrolling, a jump-to-latest affordance, and reduced-motion-safe fallback behavior for high-frequency streaming updates.
- [x] 4.7 Ensure approval controls, block expanders, and jump-to-latest interactions are keyboard accessible, touch friendly, and announced appropriately for assistive technologies.

## 5. Rollout and Validation

- [x] 5.1 Add rollout configuration `ai.use_turn_block_streaming` and compatibility guards so the new turn/block pipeline can be enabled gradually without breaking existing chat behavior.
- [x] 5.2 Verify observability and diagnostics capture turn/block lifecycle, approval transitions, and projection errors.
- [x] 5.3 Implement default, collapsed, and debug-only display policies for thinking, raw tool payloads, and evidence details in the frontend renderer.
- [x] 5.4 Validate reduced-motion, keyboard navigation, focus order, and touch-target behavior for the new streaming drawer interactions.
- [x] 5.5 Run OpenSpec validation and update any impacted AI docs or developer notes describing the new turn/block streaming architecture.
