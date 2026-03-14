# Tasks: 修复 AI 流式输出和审批功能问题

## Phase 1: 修复流式增量输出 (P0)

- [x] 1.1 修改 `internal/ai/orchestrator.go` 的 `streamExecution` 函数
  - 添加 `lastContent` 变量跟踪已发送内容
  - 计算增量内容：`chunk := text[len(lastContent):]`
  - 只发送增量部分

- [x] 1.2 更新单元测试
  - 验证 delta 事件包含增量内容
  - 验证多次 delta 事件的内容可以正确拼接

## Phase 2: 增强思维链事件数据 (P0)

- [x] 2.1 修改 `internal/ai/runtime/sse_converter.go`
  - 为 `OnPlannerStart` 添加 title 和 description
  - 为 `OnPlanCreated` 添加 steps 数组
  - 新增 `OnExecuteStart` 方法发送工具调用详情

- [x] 2.2 更新前端类型定义
  - 添加 `steps` 字段到 `SSEStageDeltaEvent`
  - 更新思维链渲染逻辑

- [x] 2.3 更新单元测试

## Phase 3: 修复用户消息持久化 (P0)

- [x] 3.1 检查 `internal/service/ai/handler.go` 中 chatRecorder 初始化
  - 确保 chatStore 正确传递
  - 添加日志记录

- [x] 3.2 修改 `internal/service/ai/session_recorder.go`
  - 确保 `handleMeta` 正确处理缺失的 session_id
  - 添加持久化失败的日志记录

- [x] 3.3 验证持久化流程
  - 发送消息后检查数据库记录
  - 刷新页面后检查消息恢复

## Phase 4: 修复审批面板显示 (P0)

- [x] 4.1 检查 `internal/ai/tools/tools.go` 中工具包装逻辑
  - 添加日志记录工具包装过程
  - 验证工具名称与 registry 匹配

- [x] 4.2 确保所有高风险工具正确注册
  - 检查 registry 中所有 `ModeMutating` 工具
  - 验证 `Info().Name` 与注册名称一致

- [x] 4.3 添加回退逻辑
  - 对于找不到 spec 的工具，根据名称推断模式
  - 对推断为 mutating 的工具包装审批门

- [ ] 4.4 端到端测试
  - 执行高风险工具验证审批面板显示
  - 验证审批确认后继续执行
  - 验证审批拒绝后不执行

## Phase 5: 过滤 AI 原始 JSON 输出 (P1)

- [x] 5.1 修改 `internal/ai/orchestrator.go`
  - 检查内容是否是工具参数 JSON
  - 过滤掉 `{"steps": [...]}` 格式的输出

- [x] 5.2 验证前端显示
  - 确认工具参数不再显示在文本中

## Testing

- [x] T.1 流式输出单元测试
- [x] T.2 思维链事件单元测试
- [x] T.3 持久化流程集成测试
- [ ] T.4 审批流程端到端测试
- [x] T.5 前端 SSE 事件处理测试

## Dependencies

- Eino ADK
- GORM
- Gin
- Redis
