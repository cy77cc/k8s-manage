# permission-approval Specification

## Purpose

实现"确认-执行"模式的操作审批机制。对于所有变更类操作（mutating tools），系统在执行前必须先向用户展示操作预览并请求确认；对于高风险操作，在用户确认后还需要第三方审批人批准。

## Requirements

### Requirement: 变更操作确认机制

系统 SHALL 对所有变更类操作（mode=mutating）执行前请求用户确认。

#### Scenario: 变更操作预览
- **GIVEN** AI 准备执行一个变更类工具（如 host_batch_exec_apply）
- **WHEN** 系统检测到该工具为 mutating 类型
- **THEN** 系统 MUST 暂停执行
- **AND** 系统 MUST 向用户展示操作预览（工具名称、参数、影响范围、风险级别）

#### Scenario: 用户确认执行
- **GIVEN** 系统已展示操作预览
- **WHEN** 用户确认执行
- **THEN** 系统 MUST 检查是否需要第三方审批
- **AND** 无需审批时直接执行，需要审批时进入审批流程

#### Scenario: 用户取消操作
- **GIVEN** 系统已展示操作预览
- **WHEN** 用户拒绝确认
- **THEN** 系统 MUST 取消该操作
- **AND** 系统 MUST 返回取消消息给用户

#### Scenario: 只读操作跳过确认
- **GIVEN** AI 准备执行一个只读工具（mode=readonly）
- **WHEN** 系统检测到该工具为 readonly 类型
- **THEN** 系统 SHOULD 直接执行（无需确认）
- **AND** 用户可在配置中设置为所有操作都需要确认

### Requirement: 风险分级审批

系统 SHALL 根据操作风险级别决定审批流程。

#### Scenario: 低风险操作 - 用户确认即可执行
- **GIVEN** 操作风险级别为 low
- **WHEN** 用户确认执行
- **THEN** 系统 MUST 直接执行
- **AND** 系统 MUST 记录操作审计日志

#### Scenario: 中风险操作 - 需要权限验证
- **GIVEN** 操作风险级别为 medium
- **WHEN** 用户确认执行
- **THEN** 系统 MUST 验证用户对目标资源的操作权限
- **AND** 有权限时直接执行，无权限时需要第三方审批

#### Scenario: 高风险操作 - 必须第三方审批
- **GIVEN** 操作风险级别为 high
- **WHEN** 用户确认执行
- **THEN** 系统 MUST 创建审批工单
- **AND** 系统 MUST 通知有审批权限的用户

### Requirement: 审批工单管理

系统 SHALL 创建和管理审批工单。

#### Scenario: 创建审批工单
- **GIVEN** 操作需要第三方审批
- **WHEN** 系统创建审批工单
- **THEN** 系统 MUST 将工单状态设为 pending
- **AND** 工单 MUST 包含：工具名称、参数、风险级别、目标资源、过期时间、请求用户

#### Scenario: 实时推送审批通知
- **GIVEN** 审批工单已创建
- **WHEN** 系统确定审批人列表
- **THEN** 系统 MUST 通过 WebSocket 向审批人推送通知
- **AND** 通知 MUST 包含：工单ID、操作描述、影响范围、过期时间

#### Scenario: 审批人选择逻辑
- **GIVEN** 操作涉及特定资源
- **WHEN** 系统确定审批人
- **THEN** 系统 MUST 按以下顺序查找：
  1. 对该资源有 admin/approve 权限的用户
  2. 该资源的 owner
  3. 系统管理员

### Requirement: 确认/审批超时机制

系统 SHALL 支持确认和审批的超时。

#### Scenario: 用户确认超时
- **GIVEN** 系统请求用户确认操作
- **WHEN** 用户在指定时间内未响应
- **THEN** 系统 MUST 自动取消操作
- **AND** 系统 MUST 通知用户操作已超时取消

#### Scenario: 审批超时
- **GIVEN** 审批工单处于 pending 状态
- **WHEN** 到达过期时间
- **THEN** 系统 MUST 将工单状态更新为 expired
- **AND** 系统 MUST 通知请求用户审批已超时

#### Scenario: 默认超时配置
- **GIVEN** 未指定超时时间
- **WHEN** 系统设置过期时间
- **THEN** 用户确认超时 MUST 默认 5 分钟
- **AND** 审批超时 MUST 默认 30 分钟

### Requirement: 审批状态流转

系统 SHALL 支持审批工单的状态管理。

#### Scenario: 审批通过
- **GIVEN** 审批人批准操作
- **WHEN** 系统接收审批结果
- **THEN** 系统 MUST 将工单状态更新为 approved
- **AND** 系统 MUST 自动执行待执行的操作

#### Scenario: 审批拒绝
- **GIVEN** 审批人拒绝操作
- **WHEN** 系统接收审批结果
- **THEN** 系统 MUST 将工单状态更新为 rejected
- **AND** 系统 MUST 记录拒绝理由并通知请求用户

#### Scenario: 审批 token 验证
- **GIVEN** 用户携带 approval_token 执行操作
- **WHEN** 系统验证 token
- **THEN** 系统 MUST 检查 token 对应的工单状态为 approved
- **AND** 系统 MUST 检查 token 未过期且工具匹配

### Requirement: 操作预览内容

系统 SHALL 为变更操作提供详细的预览信息。

#### Scenario: 预览内容格式
- **GIVEN** 系统准备展示操作预览
- **WHEN** 构建预览内容
- **THEN** 预览 MUST 包含：
  - 工具名称和描述
  - 完整参数列表
  - 目标资源（主机/服务/集群）
  - 风险级别
  - 预期影响（如可能影响的服务数量）
- **AND** 预览 SHOULD 包含：
  - 类似操作的历史记录
  - 相关告警或依赖信息

#### Scenario: 差异预览（适用于部署类操作）
- **GIVEN** 操作为部署或配置变更
- **WHEN** 展示预览
- **THEN** 系统 SHOULD 展示变更差异（diff）
- **AND** 系统 SHOULD 展示回滚方案

## Constraints

- 用户确认响应时间 MUST < 5秒
- 审批响应时间 MUST < 5秒
- WebSocket 连接断开后 MUST 支持重连
- 审批状态 MUST 持久化到数据库
- 所有变更操作 MUST 记录审计日志

## Dependencies

- 现有 Casbin 权限系统
- WebSocket 基础设施
- MySQL/PostgreSQL 数据库
