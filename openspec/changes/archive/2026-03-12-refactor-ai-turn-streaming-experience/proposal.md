## Why

当前 AI 模块虽然已经具备 SSE、阶段编排、工具执行和审批恢复能力，但用户实际看到的仍然接近“整段消息晚到”的聊天体验。后端多数阶段要等模型或步骤完整结束后才向前端暴露稳定内容，前端又把 assistant 输出建模为单条文本消息，导致打字机效果弱、阶段反馈不连续、审批恢复无法自然续写，整体体验与 AIOps 控制平面的产品目标不匹配。

## What Changes

- Refactor AI chat delivery around a turn-oriented streaming model so each assistant turn can grow incrementally instead of being treated as a single final message blob.
- Introduce block-native event projection for status, plan, tool activity, evidence, approval, thinking, and final answer content while preserving current SSE compatibility during rollout.
- Add a new persisted turn/block data model so chat history, partial streams, approvals, and resumed execution can be replayed as the same UI object instead of flattened message text.
- Define a stable AI session/detail API contract that exposes structured turn replay for new clients while preserving legacy message-compatible fields during rollout.
- Upgrade frontend AI state management from `content + thinking + tools` accumulation into a reducer-driven turn/block renderer with stronger typing, richer cards, and true streaming text presentation.
- Extend approval and resume flows so resumed execution continues on the original assistant turn rather than creating a disconnected control-path response, using a dedicated streaming resume endpoint while preserving the existing JSON resume API.
- Define explicit UX rules for what model output is shown to users, what remains collapsed or debug-only, and how operational evidence should be presented without exposing raw agent internals by default.

## Capabilities

### New Capabilities
- `ai-turn-lifecycle-storage`: Persist AI sessions as structured turns and blocks so streaming output, approvals, evidence, and resumed execution can be replayed and rendered consistently.
- `ai-chat-session-contract`: Define the HTTP contract for AI session replay so new consumers can read turn/block history while legacy consumers retain message-compatible fields.

### Modified Capabilities
- `ai-streaming-events`: Expand the SSE contract from stage-only compatibility events into turn/block-oriented streaming semantics while preserving current event compatibility during rollout.
- `aiops-card-event-stream`: Change platform event requirements so control-plane progress is emitted as block-native UI projections suitable for step cards, evidence cards, approval cards, and final-answer text blocks.
- `ai-assistant-drawer`: Change frontend rendering requirements so the drawer consumes turn/block state, supports true streaming text presentation, and renders richer operation cards with progressive disclosure.
- `ai-control-plane-baseline`: Change baseline AI lifecycle requirements so orchestration state, approvals, and resume behavior are anchored to a persisted turn identity in addition to execution state.
- `ai-assistant-command-bridge`: Change the gateway bridge contract so streaming resume and rollout flags are explicitly defined without breaking current chat and resume entrypoints.

## Impact

- Affected backend areas: `internal/ai`, `internal/service/ai`, `internal/ai/runtime`, `internal/ai/state`, `internal/model`, `storage/migration`.
- Affected frontend areas: `web/src/components/AI/**`, `web/src/api/modules/ai.ts`, message normalization and session replay logic.
- Affected API/runtime contracts: `/api/v1/ai/chat` SSE event family, `/api/v1/ai/resume/step` compatibility semantics, `/api/v1/ai/resume/step/stream` streaming resume semantics, persisted AI history shape returned by session APIs.
- Affected systems: database schema for AI chat persistence, Redis execution state linkage, SSE compatibility rollout, observability for turn/block lifecycle.
