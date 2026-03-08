# Multi-Domain Agent Architecture Tasks

## Phase 1: Foundation

### 1.1 Tool Registry Refactoring

- [x] 1.1.1 创建 `internal/ai/tools/registry.go`，按领域组织工具
- [x] 1.1.2 创建 `internal/ai/tools/classification.go`，定义 Discovery/Action 分类
- [x] 1.1.3 为现有工具添加 `domain` 和 `category` 元数据
- [x] 1.1.4 编写单元测试验证工具分类正确

### 1.2 DSL Types Definition

- [x] 1.2.1 创建 `internal/ai/types/plan.go`，定义 `DomainPlan`, `PlanStep`, `StepResult`
- [x] 1.2.2 创建 `internal/ai/types/reference.go`，定义变量引用解析逻辑
- [x] 1.2.3 编写 DSL 序列化/反序列化测试

---

## Phase 2: Core Components

### 2.1 Orchestrator Planner

- [x] 2.1.1 创建 `internal/ai/orchestrator/types.go`，定义 `DomainRequest`, `DomainDescriptor`
- [x] 2.1.2 创建 `internal/ai/orchestrator/prompt.go`，编写领域选择 Prompt
- [x] 2.1.3 创建 `internal/ai/orchestrator/planner.go`，实现 `OrchestratorPlanner`
- [x] 2.1.4 编写 Orchestrator 单元测试
- [x] 2.1.5 编写 Orchestrator 集成测试（使用 mock LLM）

### 2.2 Domain Planners

- [x] 2.2.1 创建 `internal/ai/planner/interface.go`，定义 `DomainPlanner` 接口
- [x] 2.2.2 创建 `internal/ai/planner/registry.go`，实现 Planner 注册表
- [x] 2.2.3 创建 `internal/ai/planner/prompt.go`，编写规划 Prompt 模板
- [x] 2.2.4 实现 `internal/ai/planner/infrastructure.go` (InfraPlanner)
- [x] 2.2.5 实现 `internal/ai/planner/service.go` (ServicePlanner)
- [x] 2.2.6 实现 `internal/ai/planner/cicd.go` (CICDPlanner)
- [x] 2.2.7 实现 `internal/ai/planner/monitor.go` (MonitorPlanner)
- [x] 2.2.8 实现 `internal/ai/planner/config.go` (ConfigPlanner)
- [x] 2.2.9 实现 `internal/ai/planner/user.go` (UserPlanner)
- [x] 2.2.10 编写各 Planner 单元测试

### 2.3 Executor

- [x] 2.3.1 创建 `internal/ai/executor/types.go`，定义 `ExecutionContext`, `ExecutionResult`
- [x] 2.3.2 创建 `internal/ai/executor/dag.go`，实现 DAG 构建与拓扑排序
- [x] 2.3.3 创建 `internal/ai/executor/resolver.go`，实现变量引用解析
- [x] 2.3.4 创建 `internal/ai/executor/executor.go`，实现 `Executor`
- [x] 2.3.5 实现步骤执行与结果记录
- [x] 2.3.6 实现错误处理策略
- [x] 2.3.7 编写 Executor 单元测试
- [x] 2.3.8 编写 DAG 构建测试（含循环依赖检测）

### 2.4 Replanner

- [x] 2.4.1 创建 `internal/ai/replanner/types.go`，定义 `ReplanDecision`
- [x] 2.4.2 创建 `internal/ai/replanner/validator.go`，实现结果验证逻辑
- [x] 2.4.3 创建 `internal/ai/replanner/replanner.go`，实现 `Replanner`
- [x] 2.4.4 编写 Replanner 单元测试

---

## Phase 3: Graph Integration

### 3.1 Graph Construction

- [x] 3.1.1 创建 `internal/ai/graph/orchestrator_graph.go`，构建主 Graph
- [x] 3.1.2 创建 `internal/ai/graph/planners_node.go`，实现并行 Planner 节点
- [x] 3.1.3 编写 Graph 编译测试
- [x] 3.1.4 编写 Graph 执行测试

### 3.2 HybridAgent Integration

- [x] 3.2.1 修改 `internal/ai/hybrid.go`，添加配置开关
- [x] 3.2.2 实现 `queryMultiDomain` 方法
- [x] 3.2.3 修改 `internal/ai/modes/agentic.go`，适配新架构入口
- [x] 3.2.4 编写 HybridAgent 集成测试

---

## Phase 4: Validation

### 4.1 Testing

- [x] 4.1.1 编写端到端测试：单领域任务
- [x] 4.1.2 编写端到端测试：多领域并行任务
- [x] 4.1.3 编写端到端测试：跨领域依赖任务
- [x] 4.1.4 编写错误场景测试：工具不存在
- [x] 4.1.5 编写错误场景测试：变量引用解析失败
- [x] 4.1.6 编写错误场景测试：循环依赖

### 4.2 Performance

- [x] 4.2.1 编写性能基准测试
- [x] 4.2.2 对比新旧架构延迟
- [x] 4.2.3 优化并行 Planner 执行效率

### 4.3 Compatibility

- [x] 4.3.1 验证 SSE 事件类型兼容
- [x] 4.3.2 验证前端交互无感知
- [x] 4.3.3 编写迁移验证脚本

---

## Phase 5: Rollout

### 5.1 Configuration

- [x] 5.1.1 添加配置项 `ai.use_multi_domain_arch`
- [x] 5.1.2 更新配置文档

### 5.2 Documentation

- [x] 5.2.1 更新 `openspec/specs/ai-assistant-command-bridge/spec.md`
- [x] 5.2.2 更新项目 memory（如有，仓库内未发现项目 memory 工件）  
- [x] 5.2.3 编写开发者文档

### 5.3 Final Verification

- [x] 5.3.1 运行 `go test ./...`
- [x] 5.3.2 运行 `cd web && npm run build`
- [x] 5.3.3 运行 `openspec validate --json`
- [ ] 5.3.4 手动测试典型场景（已提供 `docs/ai-multi-domain-manual-test.md` 检查清单，待人工执行）

---

## Estimated Effort

| Phase | Tasks | Priority |
|-------|-------|----------|
| Phase 1 | 7 | High |
| Phase 2 | 24 | High |
| Phase 3 | 4 | Medium |
| Phase 4 | 9 | Medium |
| Phase 5 | 7 | Low |

**Total**: 51 tasks
