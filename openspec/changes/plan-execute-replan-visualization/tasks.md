# 任务清单

## 概述

本文档记录 Plan-Execute-Replan 思考过程可视化功能的开发任务。

**预估总工时**: 11-12 小时

---

## Phase 1: 后端事件增强

**预估工时**: 3 小时

### 1.1 事件定义

- [ ] `internal/ai/events/events.go`: 新增事件常量
  - `EventPhaseStarted`
  - `EventPhaseComplete`
  - `EventPlanGenerated`
  - `EventStepStarted`
  - `EventStepComplete`
  - `EventReplanTriggered`
  - `EventReplanComplete`

### 1.2 数据结构

- [ ] `internal/ai/runtime/runtime.go`: 新增数据结构
  - `PlanStep` 结构体
  - `PhaseInfo` 结构体
  - `StepInfo` 结构体
  - `ToolCallInfo` 结构体
  - `ToolResultInfo` 结构体
  - `ReplanInfo` 结构体

### 1.3 SSE 转换器

- [ ] `internal/ai/runtime/sse_converter.go`: 新增转换方法
  - `OnPhaseStarted(phase, title string) StreamEvent`
  - `OnPhaseComplete(phase, status string) StreamEvent`
  - `OnPlanGenerated(planID string, steps []PlanStep) []StreamEvent`
  - `OnStepStarted(info *StepInfo) StreamEvent`
  - `OnToolCall(info *ToolCallInfo) StreamEvent`
  - `OnToolResult(info *ToolResultInfo) StreamEvent`
  - `OnStepComplete(stepID, status, summary string) StreamEvent`
  - `OnReplanTriggered(info *ReplanInfo) StreamEvent`

### 1.4 阶段检测器

- [ ] `internal/ai/runtime/phase_detector.go`: 新建文件
  - `PhaseDetector` 结构体
  - `Detect(event *adk.AgentEvent) string` 方法
  - `NextStepID() string` 方法

### 1.5 计划解析器

- [ ] `internal/ai/runtime/plan_parser.go`: 新建文件
  - `PlanParser` 结构体
  - `Parse(event *adk.AgentEvent) *ParsedPlan` 方法
  - JSON 格式解析
  - 编号列表解析

### 1.6 Orchestrator 增强

- [ ] `internal/ai/orchestrator.go`: 修改 streamExecution
  - 集成 PhaseDetector
  - 集成 PlanParser
  - 发送阶段事件
  - 发送步骤事件
  - 发送工具事件

### 1.7 单元测试

- [ ] `internal/ai/runtime/phase_detector_test.go`: 新建测试
- [ ] `internal/ai/runtime/plan_parser_test.go`: 新建测试

---

## Phase 2: 前端状态管理

**预估工时**: 2 小时

### 2.1 类型定义

- [ ] `web/src/components/AI/types.ts`: 新增类型
  - `PlanStep` 接口
  - `ToolExecution` 接口
  - 扩展 `ThoughtStageItem` 添加 `steps`, `executions`, `replanReason` 字段
  - `ApprovalRequest` 接口扩展
  - SSE 事件类型: `SSEPhaseStartedEvent`, `SSEPlanGeneratedEvent` 等

### 2.2 事件处理器

- [ ] `web/src/components/AI/hooks/useAIChat.ts`: 新增处理器
  - `handlePhaseStarted`
  - `handlePlanGenerated`
  - `handleStepStarted`
  - `handleToolResult`
  - `handleStepComplete`
  - `handleReplanTriggered`

### 2.3 辅助函数

- [ ] `web/src/components/AI/hooks/useAIChat.ts`: 新增辅助函数
  - `mapPhaseToStage(phase: string): ThoughtStageKey`
  - `upsertMultipleStages(stages, updates): ThoughtStageItem[]`

### 2.4 SSE API 扩展

- [ ] `web/src/api/modules/ai.ts`: 新增 SSE 事件类型
- [ ] `web/src/api/modules/ai.ts`: 扩展 `AIChatStreamHandlers` 接口

---

## Phase 3: UI 组件实现

**预估工时**: 4 小时

### 3.1 思考过程面板

- [ ] `web/src/components/AI/components/ThinkingProcessPanel.tsx`: 新建组件
  - Props: `stages`, `isStreaming`, `defaultExpanded`
  - 使用 Ant Design Collapse
  - 集成 StageTimeline

### 3.2 阶段时间线

- [ ] `web/src/components/AI/components/StageTimeline.tsx`: 新建组件
  - 使用 Ant Design Timeline
  - 渲染各阶段状态图标
  - 集成子组件 (PlanStepsList, ToolExecutionTimeline)

### 3.3 步骤列表

- [ ] `web/src/components/AI/components/PlanStepsList.tsx`: 新建组件
  - Props: `steps`, `currentStepIndex`
  - 渲染步骤状态图标
  - 显示工具提示标签

### 3.4 工具执行时间线

- [ ] `web/src/components/AI/components/ToolExecutionTimeline.tsx`: 新建组件
  - Props: `executions`
  - 显示工具名称、参数、结果
  - 状态图标指示

### 3.5 审批弹窗

- [ ] `web/src/components/AI/components/ApprovalModal.tsx`: 新建组件
  - Props: `visible`, `request`, `onConfirm`, `onCancel`, `loading`
  - 风险等级标签
  - 详情折叠面板
  - 确认/取消按钮

### 3.6 样式文件

- [ ] `web/src/components/AI/styles/thinking-process.css`: 新建样式
  - 脉冲动画
  - 时间线样式
  - 弹窗样式

---

## Phase 4: 集成与测试

**预估工时**: 2 小时

### 4.1 组件集成

- [ ] `web/src/components/AI/Copilot.tsx`: 集成新组件
  - 导入 ThinkingProcessPanel
  - 导入 ApprovalModal
  - 修改 AssistantMessage 渲染逻辑

### 4.2 后端集成测试

- [ ] 启动后端服务
- [ ] 测试 SSE 事件流
- [ ] 验证事件序列正确

### 4.3 前端集成测试

- [ ] 启动前端服务
- [ ] 测试 UI 渲染
- [ ] 验证步骤状态更新
- [ ] 验证审批弹窗功能

### 4.4 样式调整

- [ ] 调整响应式布局
- [ ] 调整颜色和间距
- [ ] 测试暗色主题

### 4.5 文档更新

- [ ] 更新 README
- [ ] 更新 API 文档

---

## 验收检查

### 功能验收

- [ ] 前端能正确显示 Planning/Executing/Replanning 阶段
- [ ] 步骤列表能实时更新状态 (pending → running → completed)
- [ ] 工具调用参数和结果能正确展示
- [ ] 审批弹窗能正常工作
- [ ] 用户批准/拒绝后流程能继续

### 性能验收

- [ ] SSE 事件延迟 < 100ms
- [ ] 前端渲染无明显卡顿
- [ ] 内存无泄漏

### 代码验收

- [ ] 后端新增代码有单元测试
- [ ] 前端组件有 TypeScript 类型检查
- [ ] 代码符合项目规范

---

## 风险与缓解

| 风险 | 缓解措施 | 负责人 |
|------|---------|--------|
| Planner 输出格式不稳定 | 使用 structured output 或明确 Prompt 格式要求 | 后端 |
| RunPath 格式变化 | 多重检测手段 (AgentName + 消息内容特征) | 后端 |
| SSE 事件过多 | 节流处理，合并高频事件 | 前端 |
| 审批流程卡死 | 添加超时机制和重试按钮 | 前后端 |

---

## 时间线

| 阶段 | 开始日期 | 结束日期 | 状态 |
|------|---------|---------|------|
| Phase 1: 后端 | - | - | 待开始 |
| Phase 2: 前端状态 | - | - | 待开始 |
| Phase 3: UI 组件 | - | - | 待开始 |
| Phase 4: 集成测试 | - | - | 待开始 |

---

## 相关文档

- [proposal.md](./proposal.md) - 提案文档
- [design-backend.md](./design-backend.md) - 后端实现方案
- [design-frontend.md](./design-frontend.md) - 前端实现方案
- [specs/ai-streaming-events/spec.md](./specs/ai-streaming-events/spec.md) - SSE 事件规范变更
- [specs/ai-runtime-core/spec.md](./specs/ai-runtime-core/spec.md) - Runtime 规范变更
