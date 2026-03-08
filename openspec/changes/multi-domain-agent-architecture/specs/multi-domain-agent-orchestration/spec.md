# Multi-Domain Agent Orchestration

## Status

PROPOSED

## Summary

定义多领域 Agent 编排能力，包括 Orchestrator Planner、Domain Planner、Executor、Replanner 的职责边界、交互协议与行为规范。

## ADDED Requirements

### Requirement: Orchestrator Planner SHALL analyze and dispatch domains

Orchestrator Planner SHALL analyze user requests to identify involved domains. Orchestrator Planner SHALL output a list of DomainRequest, each corresponding to one domain. Orchestrator Planner SHALL NOT execute any tools, only responsible for domain selection. When unable to determine the domain, Orchestrator Planner SHALL default to the `general` domain.

#### Scenario: 单领域请求

- **WHEN** 用户请求 "查询主机 10.0.0.1 的 CPU 使用率"
- **THEN** Orchestrator Planner 输出 [{"domain": "infrastructure", "context": {"focus": "主机状态"}}]

#### Scenario: 多领域请求

- **WHEN** 用户请求 "部署支付服务到生产环境并检查告警"
- **THEN** Orchestrator Planner 输出 [{"domain": "service"}, {"domain": "monitor"}]

#### Scenario: 无法识别领域

- **WHEN** 用户请求 "你好"
- **THEN** Orchestrator Planner 输出 [{"domain": "general", "context": {}}]

### Requirement: Domain Planner SHALL plan without execution

Domain Planner SHALL receive DomainRequest and output DomainPlan. DomainPlan SHALL contain a list of execution steps, each step containing id, tool, params, depends_on, produces, and requires. Domain Planner MAY call Discovery tools during planning phase to complete parameters. Domain Planner SHALL NOT call Action tools. Domain Planner SHALL declare dependencies between steps. Multiple Domain Planners SHALL support parallel execution.

#### Scenario: 服务部署规划

- **WHEN** Service Planner 接收 {"domain": "service", "user_intent": "部署支付服务到生产环境"}
- **THEN** 输出 DomainPlan 包含步骤 get_service、get_cluster、deploy
- **AND** deploy 依赖于 get_service 和 get_cluster

#### Scenario: 跨领域依赖声明

- **WHEN** Monitor Planner 需要检查服务告警
- **THEN** 输出 DomainPlan 包含 requires: [service_id]
- **AND** 参数使用 $ref 引用 service.deploy.service_id

### Requirement: Executor SHALL merge and execute plans by DAG

Executor SHALL merge steps from all DomainPlans. Executor SHALL build a global DAG including intra-domain and cross-domain dependencies. Executor SHALL execute steps in topological order. Executor SHALL resolve variable references ($ref) from results of executed steps. When variable reference cannot be resolved, Executor SHALL return an error and stop execution. Executor SHALL have execution permission for all tools (Discovery + Action). Executor SHALL record execution result for each step.

#### Scenario: 合并多领域 Plan 并执行

- **GIVEN** Service Plan 包含步骤 [get_service, deploy]
- **AND** Monitor Plan 包含步骤 [check_alerts] 依赖 service.deploy.service_id
- **WHEN** Executor 执行
- **THEN** 执行顺序为 get_service -> deploy -> check_alerts

#### Scenario: 变量引用解析

- **GIVEN** 步骤 get_service 输出 {service_id: "pay-001"}
- **AND** 步骤 deploy 参数 service_id 为 $ref 引用
- **WHEN** Executor 解析参数
- **THEN** deploy.service_id = "pay-001"

#### Scenario: 变量引用解析失败

- **GIVEN** 步骤 check_alerts 参数 service_id 引用尚未执行的步骤
- **WHEN** Executor 尝试解析
- **THEN** 返回错误

### Requirement: Replanner SHALL validate results and decide re-planning

Replanner SHALL validate the completeness of execution results. When there are failed steps, Replanner SHALL analyze the failure reasons. Replanner MAY decide to re-trigger Orchestrator Planner. Replanner SHALL output ReplanDecision containing need_replan, reason, and suggestions.

#### Scenario: 执行成功无需重规划

- **GIVEN** 所有步骤执行成功
- **WHEN** Replanner 评估
- **THEN** 输出 need_replan: false

#### Scenario: 部分失败需要重规划

- **GIVEN** 步骤 deploy 执行失败
- **WHEN** Replanner 评估
- **THEN** 输出 need_replan: true
