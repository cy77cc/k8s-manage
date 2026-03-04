# Tasks: 多专家协作机制重构

## 阶段1: 类型定义与配置调整

- [x] 更新 `internal/ai/experts/types.go`
  - [x] 添加 `StrategyPrimaryLed` 策略常量
  - [x] 添加 `ExpertProgressEvent` 类型
  - [x] 添加 `ProgressEmitter` 类型
  - [x] 添加 `HelperRequest` 类型
  - [x] 添加 `PrimaryDecision` 类型
  - [x] `RouteDecision` 字段改名 `HelperExperts` → `OptionalHelpers`
- [x] 更新 `internal/ai/experts/config.go`
  - [x] `SceneMapping` 字段改名 `HelperExperts` → `OptionalHelpers`
  - [x] 保持旧配置兼容（自动转换）

## 阶段2: Router 调整

- [x] 修改 `internal/ai/experts/router.go`
  - [x] `routeByScene()` 返回 `OptionalHelpers`
  - [x] 默认策略改为 `StrategyPrimaryLed`
  - [x] 保持旧配置兼容

## 阶段3: Orchestrator 核心重写

- [x] 重写 `internal/ai/experts/orchestrator.go`
  - [x] 实现 `streamPrimaryLed()` 主从协作流程
  - [x] 实现 `primaryDecisionPhase()` 决策阶段
    - [x] 构建决策 prompt
    - [x] 解析 `[REQUEST_HELPER: expert: task]` 格式
  - [x] 实现 `helperExecutionPhase()` 助手执行阶段
    - [x] 发送 `expert_progress` 事件
    - [x] 助手并行执行 (使用 `Generate`，更快)
    - [x] 静默收集结果
  - [x] 实现 `primarySummaryPhase()` 汇总阶段
    - [x] 构建汇总 prompt
    - [x] 流式输出最终回答
  - [x] 实现 `buildMessagesWithHistory()` 历史传递
- [x] 编写单元测试
  - [x] 测试决策解析
  - [x] 测试助手执行
  - [x] 测试历史传递

## 阶段4: Executor 调整

- [x] 修改 `internal/ai/experts/executor.go`
  - [x] `StreamStep()` 接收 `ExecuteRequest` 而非独立参数
  - [x] `buildExpertMessages()` 包含历史上下文
- [x] 编写单元测试

## 阶段5: PlatformAgent 适配

- [x] 修改 `internal/ai/platform_agent.go`
  - [x] `buildExecuteRequest()` 设置 `EventEmitter`
  - [x] 确保 `History` 正确传递
  - [x] 适配新的 `StrategyPrimaryLed` 策略

## 阶段6: 配置文件更新

- [x] 更新 `configs/scene_mappings.yaml`
  - [x] `helper_experts` → `optional_helpers`
  - [x] 审查每个场景的助手列表是否合理
  - [x] 调整默认策略
- [x] 保持向后兼容

## 阶段7: 前端适配

- [x] 修改 `web/src/components/AI/ChatInterface.tsx`
  - [x] 添加 `expert_progress` 事件处理
  - [x] 实现进度动画组件
  - [x] 显示助手执行状态
- [x] 编写前端测试

## 阶段8: 集成测试

- [x] 端到端测试
  - [x] 测试单专家场景不受影响
  - [x] 测试主专家直接回答（不需要助手）
  - [x] 测试主专家调用助手后汇总输出
  - [x] 测试多轮对话上下文保持
  - [x] 测试助手执行失败降级
- [x] 性能测试
  - [x] 测试助手并行执行效率
  - [x] 测试历史上下文大小影响

## 文件变更清单

| 文件 | 变更类型 | 优先级 |
|------|----------|--------|
| `internal/ai/experts/types.go` | 修改 | P0 |
| `internal/ai/experts/config.go` | 修改 | P0 |
| `internal/ai/experts/router.go` | 修改 | P1 |
| `internal/ai/experts/orchestrator.go` | 重写 | P0 |
| `internal/ai/experts/executor.go` | 修改 | P1 |
| `internal/ai/platform_agent.go` | 修改 | P1 |
| `configs/scene_mappings.yaml` | 修改 | P1 |
| `web/src/components/AI/ChatInterface.tsx` | 修改 | P2 |

## 依赖关系

```
阶段1 (类型定义) ────────────────────────────────────────────────────┐
                                                                    │
阶段2 (Router) ─────────────────────────────────────────────────────┤
                                                                    │
阶段3 (Orchestrator) ◀──────────────────────────────────────────────┤
         │                                                          │
         ▼                                                          │
阶段4 (Executor) ───────────────────────────────────────────────────┤
                                                                    │
阶段5 (PlatformAgent) ──────────────────────────────────────────────┤
                                                                    │
阶段6 (配置更新) ───────────────────────────────────────────────────┤
                                                                    │
阶段7 (前端适配) ───────────────────────────────────────────────────┤
                                                                    │
阶段8 (集成测试) ◀──────────────────────────────────────────────────┘
```

## 风险缓解

| 风险 | 缓解措施 |
|------|----------|
| 主专家决策超时 | 设置 5s 超时，超时后默认不调用助手 |
| 助手执行失败 | 单个助手失败不影响其他助手，主专家仍可汇总 |
| 前端不支持新事件 | 渐进增强，旧版前端正常显示最终输出 |
| 历史上下文过大 | 限制最近 10 轮对话 |
