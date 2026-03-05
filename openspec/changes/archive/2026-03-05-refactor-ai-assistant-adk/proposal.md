## Why

当前 AI 助手模块经过多轮迭代，存在严重的架构问题：

1. **自建框架与 eino ADK 不兼容**：自建了 `experts/`、`graph/`、`callbacks/` 三个目录的框架代码，这些代码与 eino ADK 标准架构不一致，导致无法使用 ADK 提供的高级能力。

2. **审批机制不标准**：使用自定义的 `ApprovalRequiredError` + 内存票据方式，而非 eino 标准的 `tool.StatefulInterrupt` 机制，导致中断恢复流程不标准。

3. **入口处理过于复杂**：`chat_handler.go` 超过 600 行，手动处理 SSE、消息构建、工具调用跟踪等底层细节，而非使用 `adk.Runner`。

4. **功能分散且不完整**：代码量大（~2500 行核心代码）但实际能力不强，缺乏 Plan-Execute-Replan、反思循环等高级 Agent 模式。

5. **维护成本高**：自建框架需要持续维护，且与 eino 生态割裂，无法享受社区更新。

需要完整重构 AI 助手模块，拥抱 eino ADK 标准架构，实现更强大、更易维护的 AI 能力。

## What Changes

- **删除自建框架**：移除 `internal/ai/experts/`、`internal/ai/graph/`、`internal/ai/callbacks/` 三个目录。
- **采用 ADK Agent 模式**：使用 `planexecute.New()` 构建 Plan-Execute-Replan Agent。
- **标准化审批机制**：使用 `tool.StatefulInterrupt` + `schema.Register` 实现标准的中断恢复流程。
- **简化入口处理**：使用 `adk.NewRunner()` + `runner.Query()` 替代手动事件处理。
- **实现 CheckPointStore**：使用 `compose.CheckPointStore` 接口实现数据库持久化。
- **保留并改造工具系统**：保留 50+ 平台工具，添加标准审批/审核包装器。

## Capabilities

### New Capabilities
- `ai-assistant-adk-architecture`: 定义基于 eino ADK 的 AI 助手架构能力，包括 Plan-Execute-Replan 模式、StatefulInterrupt 审批机制、CheckPointStore 持久化。

### Modified Capabilities
- `ai-assistant-command-bridge`: 适配新的 ADK 架构，保持命令桥接能力不变。
- `ai-assistant-experience-optimization`: 使用 `adk.AgentEvent` 标准事件格式。

## Impact

### Backend
- `internal/ai/`:
  - 删除 `experts/` 目录 (~800 行)
  - 删除 `graph/` 目录 (~500 行)
  - 删除 `callbacks/` 目录 (~200 行)
  - 重写 `agent.go` (使用 planexecute)
  - 新增 `tools/wrapper.go` (StatefulInterrupt 包装器)
  - 新增 `store.go` (CheckPointStore 实现)
- `internal/service/ai/`:
  - 重写 `chat_handler.go` (使用 adk.Runner，从 600+ 行简化到 ~150 行)
  - 保留 `store.go` 的会话持久化逻辑
  - 简化 `events_sse.go`

### API
- 保持现有 API 接口不变
- SSE 事件格式适配 `adk.AgentEvent`

### Frontend
- 无需改动，SSE 事件格式向后兼容

### Data
- 新增 `ai_checkpoints` 表（用于 CheckPointStore）
- 保留现有 `ai_chat_sessions` 和 `ai_chat_messages` 表

## Success Metrics

| 指标 | 迁移前 | 迁移后 |
|-----|-------|-------|
| AI 核心代码量 | ~2500 行 | ~400 行 |
| 自建框架目录 | 3 个 | 0 个 |
| Agent 模式 | 自定义 experts | planexecute (标准) |
| 审批机制 | 手动 token | StatefulInterrupt (标准) |
| 状态持久化 | 混乱 | CheckPointStore (标准) |
| HTTP 处理代码 | 600+ 行 | ~150 行 |
