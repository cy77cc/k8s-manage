# Proposal: 修复 AI 流式输出和审批功能问题

## Summary

修复 AI 模块的 SSE 流式输出、思维链显示、对话持久化和审批功能的多个问题。

## Motivation

用户报告了以下问题，严重影响 AI 助手的使用体验：

1. **流式输出不是增量的** - 用户收到的是一个大 chunk 而不是逐字输出
2. **思维链数据不完整** - 前端显示的是硬编码的标题，而不是后端动态生成的内容
3. **AI 原始 JSON 输出未过滤** - 工具参数 `{"steps": [...]}` 被当作文本显示
4. **用户消息消失** - 对话持久化有问题，用户输入丢失
5. **审批面板不显示** - 高风险工具执行时审批确认面板未出现

## Root Cause Analysis

### 问题 1: 流式输出不是增量

**根因**: `internal/ai/orchestrator.go:328-333`

ADK 返回的 `msg.Content` 是累积的完整内容，后端每次发送完整内容而不是增量 chunk。

### 问题 2: 思维链事件缺少详细数据

**根因**: `internal/ai/runtime/sse_converter.go`

`OnPlannerStart` 和 `OnPlanCreated` 只发送基本信息，没有发送 title、description、steps 等详细数据。

### 问题 3: 用户消息消失

**根因**: `internal/service/ai/session_recorder.go`

用户消息持久化依赖于 `Meta` 事件中的 `session_id`，如果事件处理不正确，用户消息不会被保存。

### 问题 4: 审批面板不显示

**根因**: `internal/ai/tools/tools.go:51-109`

在 `NewAllTools` 函数中，如果工具的 `Info().Name` 与 registry 中注册的名称不匹配，工具就不会被 ApprovalGate 包装，导致审批流程不被触发。

## Proposed Solution

### Phase 1: 修复流式增量输出

修改 `internal/ai/orchestrator.go` 中的 `streamExecution` 函数：
- 维护 `lastContent` 变量跟踪已发送的内容
- 计算增量内容并只发送增量部分

### Phase 2: 增强思维链事件数据

修改 `internal/ai/runtime/sse_converter.go`：
- 为 `OnPlannerStart` 添加 title 和 description
- 为 `OnPlanCreated` 添加 steps 数组

### Phase 3: 修复用户消息持久化

修改 `internal/service/ai/session_recorder.go`：
- 确保 `handleMeta` 正确处理 `session_id`
- 添加日志记录帮助调试

### Phase 4: 修复审批面板显示

修改 `internal/ai/tools/tools.go`：
- 添加日志记录工具包装过程
- 确保所有高风险工具都被正确包装

## Affected Files

- `internal/ai/orchestrator.go`
- `internal/ai/runtime/sse_converter.go`
- `internal/service/ai/handler.go`
- `internal/service/ai/session_recorder.go`
- `internal/ai/tools/tools.go`

## Risks

- 流式输出修改可能影响现有 SSE 客户端兼容性
- 审批门修改需要确保所有工具都被正确包装

## Success Criteria

1. AI 对话逐字输出，而不是一个大 chunk
2. 思维链显示动态的标题和描述
3. 用户消息正确持久化，刷新页面后可恢复
4. 高风险工具执行时显示审批确认面板
