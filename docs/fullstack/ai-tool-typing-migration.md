# AI Assistant V2 Migration Guide

## Scope

本指南说明 AI 助手从旧版命令中心/悬浮助手迁移到新版统一工作台后的前后端变化，覆盖：

- 页面与组件迁移
- API 路由迁移
- SSE 事件与审批交互
- 工具 typing 与统一工具名
- 旧实现移除后的替代方案

## Summary

新版 AI 助手已经统一到 `/ai` 工作台，并以单一路由组、统一会话模型和统一工具注册表为基础运行。

主要变化：

- 旧的 `internal/service/ai/v2` 包装层已删除
- 旧的前端 `web/src/components/AI/*` 组件已删除
- `/ai` 不再指向旧命令中心，而是新的 `AIChatPage`
- 工具体系从分散的旧工具名，收敛到新的统一工具名
- 会话、审批、SSE done payload 都以当前统一接口为准

## Frontend Migration

### Route Entry

旧版：

- `/ai` -> `AICommandCenterPage`
- 全局布局内存在悬浮 `GlobalAIAssistant`

新版：

- `/ai` -> `AIChatPage`
- 不再注入全局悬浮助手

### Removed Components

以下旧组件已移除：

- `web/src/components/AI/ChatInterface.tsx`
- `web/src/components/AI/GlobalAIAssistant.tsx`
- `web/src/components/AI/CommandPanel.tsx`
- `web/src/components/AI/AnalysisPanel.tsx`
- `web/src/components/AI/RecommendationPanel.tsx`

### New Page Structure

新版入口集中在 `web/src/pages/AIChat/`：

- `ChatPage`
- `components/ConversationSidebar`
- `components/ChatMain`
- `hooks/useChatSession`
- `hooks/useSSEConnection`
- `hooks/useConfirmation`
- `hooks/useAIChatShortcuts`

### Frontend Integration Notes

- 会话列表和详情都来自 `/api/v1/ai/sessions*`
- SSE `done` 事件会回写最新 `session` 和 `turn_recommendations`
- 审批/确认面板不再依赖旧的本地状态约定，而是绑定后端返回的 ask / interrupt 元数据

## Backend Migration

### Route Registration

旧版存在 `legacy compat` 和 `v2` 双路径注册。

新版统一为单一注册入口：

- `ai.RegisterAIHandlers`

当前核心接口：

- `POST /api/v1/ai/chat`
- `POST /api/v1/ai/chat/respond`
- `GET /api/v1/ai/sessions`
- `GET /api/v1/ai/sessions/:id`
- `DELETE /api/v1/ai/sessions/:id`
- `GET /api/v1/ai/tools`

仍保留的辅助接口：

- `GET /api/v1/ai/sessions/current`
- `POST /api/v1/ai/approval/respond`
- `POST /api/v1/ai/confirmations/:id/confirm`
- `GET /api/v1/ai/capabilities`

### Model and Agent Entry

- `internal/svc/svc.go` 已切到 `NewToolCallingChatModel`
- 旧的 `NewChatModel` 兼容别名已移除
- Agent 层统一为 `IntentClassifier + SimpleChatMode + AgenticMode + HybridAgent`

### Storage

- 会话持久化使用 `SessionStore`
- checkpoint 使用 `CheckPointStore`
- 原先承担多种职责的 `memoryStore` 已拆分，不再作为主会话存储入口

## Tool Migration

### Typed Tool Contracts

工具输入不再依赖松散的 `map[string]any`，而是以 typed struct 为主，并统一纳入注册表导出。

### Unified Tool Names

推荐迁移到以下统一工具名：

- `k8s_query`
- `k8s_logs`
- `k8s_events`
- `host_exec`
- `host_batch`
- `service_deploy`
- `service_status`
- `monitor_alert`
- `monitor_metric`

MCP 工具则统一采用前缀化名称，例如：

- `mcp_default_search_docs`

### Approval and Preview Semantics

- 高风险工具不应绕过审批
- preview/apply 语义在统一工具层中保留
- `service_deploy` 之类的统一工具支持 preview 与 apply 两条路径

## SSE and Interaction Migration

当前 SSE 事件流以统一工作台为准，前端应处理以下事件：

- `meta`
- `delta`
- `thinking_delta`
- `tool_call`
- `tool_result`
- `approval_required`
- `review_required`
- `interrupt_required`
- `heartbeat`
- `done`
- `error`

迁移时需要注意：

- 不要再把 `schema.Tool` 原始内容直接拼到 assistant 文本里
- 只应展示真实执行层的 `tool_call`
- `tool_result` 应按 `call_id` 关联，而不是按文本猜测
- `done` 事件中的 `turn_recommendations` 是新的推荐来源

## Recommended Migration Steps

1. 将前端 `/ai` 入口切到 `AIChatPage`
2. 移除对旧 `components/AI/*` 组件的引用
3. 对齐后端到统一 AI 路由组
4. 迁移前端数据流到 `sessions + SSE + approvals` 模型
5. 将旧工具名调用收敛到统一工具名
6. 删除遗留兼容层与 deprecated 代码

## Validation

- `go test ./internal/ai ./internal/service/ai ./internal/service ./internal/svc`
- `go test ./internal/ai/tools`
- `./node_modules/.bin/tsc -b`
- `npm run build`
