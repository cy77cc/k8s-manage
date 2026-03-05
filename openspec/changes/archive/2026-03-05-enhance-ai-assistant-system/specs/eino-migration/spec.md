# eino-migration Specification

## Purpose

将 AI 助手系统中使用 eino 框架的废弃 API 迁移到推荐的新 API，消除技术债务，确保与框架最新版本的兼容性。

## Requirements

### Requirement: 移除废弃的 Orchestrator

系统 SHALL 移除 `internal/ai/experts/orchestrator.go` 中标记为 Deprecated 的 Orchestrator 实现，并 SHALL 将所有调用点迁移到 Graph Runner。

#### Scenario: Graph Runner 完全替代 Orchestrator
- **GIVEN** Orchestrator 已被标记为废弃
- **WHEN** PlatformAgent 需要执行专家编排
- **THEN** 系统 MUST 使用 `internal/ai/graph` 中的 Graph Runner
- **AND** 所有 Orchestrator 的功能 MUST 在 Graph Runner 中有对应实现

#### Scenario: 流式执行支持
- **GIVEN** 用户发起需要专家协作的请求
- **WHEN** 使用 Graph Runner 执行
- **THEN** 系统 MUST 支持流式输出（StreamReader）
- **AND** 流式输出的行为 MUST 与原 Orchestrator 一致

### Requirement: 更新 react.NewAgent 配置方式

系统 SHALL 更新 React Agent 的创建方式，使用推荐选项函数模式。

#### Scenario: 使用 WithTools 选项函数
- **GIVEN** 需要创建新的 React Agent
- **WHEN** 调用 react.NewAgent
- **THEN** 系统 SHOULD 使用 `react.WithTools` 选项函数
- **AND** 工具配置 MUST 同时更新 chat model 和 tool registry

#### Scenario: Agent 配置兼容性
- **GIVEN** 现有 Agent 配置使用 AgentConfig 结构体
- **WHEN** 迁移到选项函数模式
- **THEN** 所有现有配置项 MUST 有对应的选项函数
- **AND** 功能行为 MUST 保持一致

### Requirement: 评估 eino v0.8 兼容性

系统 SHALL 评估 eino v0.8 的 breaking changes，确保平滑升级路径。

#### Scenario: Shell 接口重命名检查
- **GIVEN** 项目可能使用 ShellBackend 接口
- **WHEN** 升级到 v0.8
- **THEN** 系统 MUST 更新为 Shell 接口
- **AND** 所有相关调用 MUST 通过编译检查

#### Scenario: ReadRequest.Offset 基准变更
- **GIVEN** 可能存在文件读取操作
- **WHEN** 使用 ReadRequest.Offset
- **THEN** 系统 MUST 将 0-based 改为 1-based
- **AND** 所有相关逻辑 MUST 正确处理偏移量

## Constraints

- 迁移过程中 MUST 保持现有测试用例通过
- 废弃代码删除前 MUST 确保无其他模块依赖
- Graph Runner 的性能 MUST 不低于原 Orchestrator

## Dependencies

- eino v0.7.37 或更高版本
- 现有测试覆盖率达到迁移要求
