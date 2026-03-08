# Domain Plan DSL

## Status

PROPOSED

## Summary

定义 DomainPlan DSL 规范，包括步骤定义、依赖声明、变量引用语法。这是 Domain Planner 与 Executor 之间的契约。

## ADDED Requirements

### Requirement: DomainPlan SHALL have valid domain and steps

DomainPlan is the output format of Domain Planner. The `domain` field SHALL be a valid domain name. The `steps` field SHALL contain at least one step. Step IDs SHALL be unique within the same DomainPlan. Step IDs SHALL use snake_case naming convention.

#### Scenario: 单领域部署 Plan 格式

- **WHEN** Service Planner 输出部署计划
- **THEN** DomainPlan 包含 domain: "service"
- **AND** steps 包含 get_service、get_cluster、deploy 步骤
- **AND** deploy 的 depends_on 包含 get_service 和 get_cluster

### Requirement: PlanStep SHALL define tool and dependencies

PlanStep defines a single execution step. The `tool` field SHALL be a registered tool name. The `params` field MAY contain static values or variable references. The `depends_on` field SHALL only contain step IDs within the same DomainPlan. The `produces` field SHALL declare output variable names. The `requires` field SHALL declare variables needed from external sources.

#### Scenario: 步骤声明输出字段

- **WHEN** 步骤 get_service 声明 produces: ["service_id"]
- **THEN** 执行结果必须包含 service_id 字段

### Requirement: Variable reference SHALL use $ref syntax

Variable references SHALL use the format `{$ref: "<domain>.<step_id>.<field>"}` for cross-domain references. When referencing steps within the same domain, the domain MAY be omitted. The reference path SHALL contain three parts: domain, step_id, and field. Executor SHALL resolve all variable references before executing a step. When the target step of a variable reference has not been executed, Executor SHALL return an error.

#### Scenario: 跨领域变量引用

- **WHEN** Monitor Planner 需要引用 Service 领域的输出
- **THEN** 参数使用 {"$ref": "service.deploy.service_id"} 格式

#### Scenario: 领域内变量引用

- **WHEN** 步骤引用同一领域内其他步骤的输出
- **THEN** 参数使用 {"$ref": "get_service.service_id"} 格式（省略 domain）

### Requirement: Executor SHALL build and validate DAG

After merging all DomainPlans, Executor SHALL build a global DAG. Each global step ID SHALL be unique in format domain.step_id. Executor SHALL detect circular dependencies and return an error if found. Before execution, Executor SHALL validate all tool names are registered. Executor SHALL validate all depends_on references exist. Executor SHALL validate all $ref references exist and will be executed first.

#### Scenario: 循环依赖检测

- **GIVEN** 步骤 A 依赖 B，步骤 B 依赖 A
- **WHEN** Executor 构建 DAG
- **THEN** 返回错误 "检测到循环依赖"

#### Scenario: 未注册工具验证

- **GIVEN** Plan 包含 tool: "unknown_tool"
- **WHEN** Executor 执行前验证
- **THEN** 返回错误 "工具未注册: unknown_tool"

#### Scenario: 缺失依赖验证

- **GIVEN** 步骤 deploy 的 depends_on 包含不存在的 "missing_step"
- **WHEN** Executor 执行前验证
- **THEN** 返回错误 "依赖步骤不存在: missing_step"

### Requirement: StepResult SHALL record execution output

StepResult records step execution results. The Output field SHALL contain all fields declared in produces. When step execution fails, the Error field SHALL contain error information.

#### Scenario: 步骤输出字段缺失

- **GIVEN** 步骤声明 produces: ["service_id"] 但执行结果未包含
- **WHEN** Executor 记录结果
- **THEN** 记录警告但继续执行
