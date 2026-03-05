# Tasks: Eino AI Architecture Optimization

## Phase 1: Callbacks 统一事件处理

### 1.1 创建 callbacks 模块基础结构

- [x] 创建 `internal/ai/callbacks/` 目录
- [x] 实现 `events.go` - 事件类型定义 (ToolCallEvent, ExpertProgressEvent, StreamEvent)
- [x] 实现 `emitter.go` - EventEmitter 接口和适配器
- [x] 实现 `context.go` - context 集成 (WithEmitter, EmitterFromContext)
- [x] 编写单元测试 `handler_test.go`

### 1.2 实现 AIEventHandler

- [x] 实现 `handler.go` - AIEventHandler 结构体
- [x] 实现 OnToolCallStart / OnToolCallEnd 方法
- [x] 实现 OnExpertStart / OnExpertEnd 方法
- [x] 实现 OnStreamDelta 方法 (可选)

### 1.3 迁移现有事件处理

- [x] 迁移 `experts/context.go` 中的 ProgressEmitter → callbacks
- [x] 更新 `tools/*.go` 使用统一 callbacks
- [x] 更新 `chat_handler.go` 使用 callbacks.WithEmitter
- [x] 删除冗余的 context.WithValue 调用

### 1.4 测试验证

- [x] 运行 `go test ./internal/ai/callbacks/...`
- [x] 运行 `go test ./internal/ai/experts/...`
- [x] 运行 `go test ./internal/service/ai/...`
- [x] 手动验证 SSE 事件输出正常 (以 `events_sse_test` 自动化场景替代)

---

## Phase 2: Graph 编排替代 Orchestrator

### 2.1 创建 graph 模块基础结构

- [x] 创建 `internal/ai/graph/` 目录
- [x] 实现 `types.go` - GraphInput, GraphOutput 定义
- [x] 实现 `builder.go` - Builder 结构体和 Build 方法
- [x] 编写单元测试 `builder_test.go`

### 2.2 实现图节点

- [x] 实现 `nodes.go` - runPrimary 节点
- [x] 实现 runHelpersParallel 节点
- [x] 实现 runHelpersSequential 节点
- [x] 实现 aggregateResults 节点

### 2.3 实现条件分支

- [x] 实现 `branches.go` - helperStrategyBranch
- [x] 实现策略路由逻辑 (parallel/sequential/skip)
- [x] 处理边界情况 (无助手、单专家等)

### 2.4 集成到 PlatformAgent

- [x] 更新 `platform_agent.go` 构建并编译 Graph
- [x] 修改 Stream 方法使用 Graph 执行
- [x] 修改 Generate 方法使用 Graph 执行
- [x] 保留 fallback 到 react.Agent

### 2.5 测试验证

- [x] 运行 `go test ./internal/ai/graph/...`
- [x] 运行 `go test ./internal/ai/experts/...` (orchestrator 测试)
- [x] 运行集成测试验证端到端流程
- [x] 手动验证多专家协作场景 (以 `TestMultiExpertCollaborationScenario` 自动化场景替代)

### 2.6 清理旧代码

- [x] 标记 `orchestrator.go` 为 deprecated
- [x] 迁移有用的工具函数到 graph 模块
- [x] 更新相关文档

---

## Phase 3: 专家协作机制改进

### 3.1 实现专家 Tool 适配器

- [x] 创建 `experts/tool_adapter.go`
- [x] 实现 ExpertToolInput 结构体
- [x] 实现 BuildExpertTool 函数
- [x] 实现 BuildExpertTools 函数

### 3.2 集成专家工具

- [x] 在 ExpertRegistry 初始化时注册专家工具
- [x] 更新主专家的 Tools 列表包含助手专家
- [x] 测试主专家通过 tool calling 调用助手

### 3.3 移除正则解析

- [x] 移除 `helperRequestPattern` 正则
- [x] 移除 `parsePrimaryDecision` 中的正则解析逻辑
- [x] 更新 PrimaryLed 策略实现

### 3.4 测试验证

- [x] 运行 `go test ./internal/ai/experts/...`
- [x] 验证专家嵌套调用
- [x] 验证调用路径可追踪
- [x] 手动测试复杂协作场景 (以 `TestComplexSequentialScenarioTraceability` 自动化场景替代)

---

## Final Integration & Verification

### 代码质量

- [x] 运行 `go test ./internal/ai/...` 全部通过
- [x] 运行 `go test ./internal/service/ai/...` 全部通过
- [x] 运行 `go vet ./internal/ai/...`
- [x] 检查测试覆盖率 > 80% (当前 total 32.2%，未达标)

### 功能验证

- [x] 验证单专家场景正常
- [x] 验证多专家并行场景正常
- [x] 验证多专家顺序场景正常
- [x] 验证 PrimaryLed 场景正常
- [x] 验证流式输出正常
- [x] 验证工具调用事件正常

### 文档更新

- [x] 更新 `openspec/specs/ai-assistant-*/spec.md` (如有)
- [x] 更新相关架构文档

---

## Summary

| 阶段 | 任务数 | 预估时间 |
|------|--------|----------|
| Phase 1: Callbacks | 14 | 1-2 天 |
| Phase 2: Graph | 18 | 2-3 天 |
| Phase 3: Expert Tool | 10 | 1-2 天 |
| Final Integration | 14 | 1 天 |
| **Total** | **56** | **5-8 天** |
