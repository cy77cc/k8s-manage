# Spec: Service Template Review

模板审核发布工作流 (私有 → 审核 → 发布)。

## ADDED Requirements

### Requirement: Submit for review

系统 SHALL 允许用户提交模板审核。

#### Scenario: Submit template
- **WHEN** 用户点击"提交审核"
- **THEN** 系统将模板状态改为 `pending_review`
- **AND** 模板进入审核队列

#### Scenario: Cannot submit draft with errors
- **WHEN** 用户尝试提交存在语法错误的模板
- **THEN** 系统拒绝并提示"请修复模板错误后再提交"

### Requirement: Review queue display

系统 SHALL 为管理员显示待审核模板列表。

#### Scenario: View review queue
- **WHEN** 管理员访问审核管理页面
- **THEN** 系统显示所有 `status=pending_review` 的模板
- **AND** 显示提交者、提交时间

### Requirement: Review template detail

系统 SHALL 显示待审核模板的完整内容。

#### Scenario: View template for review
- **WHEN** 管理员点击待审核模板
- **THEN** 系统显示模板的完整内容
- **AND** 显示 K8s/Compose YAML
- **AND** 显示变量定义
- **AND** 显示提交者信息

### Requirement: Approve template

系统 SHALL 允许管理员批准发布模板。

#### Scenario: Approve template
- **WHEN** 管理员点击"发布"
- **THEN** 系统将模板状态改为 `published`
- **AND** 模板对所有用户可见
- **AND** 系统通知提交者审核通过

### Requirement: Reject template

系统 SHALL 允许管理员驳回模板。

#### Scenario: Reject template
- **WHEN** 管理员点击"驳回"并填写驳回原因
- **THEN** 系统将模板状态改为 `rejected`
- **AND** 系统通知提交者审核驳回及原因

### Requirement: Resubmit rejected template

系统 SHALL 允许用户重新提交被驳回的模板。

#### Scenario: Resubmit after rejection
- **WHEN** 用户修改被驳回的模板后点击"重新提交"
- **THEN** 系统将模板状态改为 `pending_review`
- **AND** 模板重新进入审核队列

### Requirement: Unpublish template

系统 SHALL 允许管理员下架已发布的模板。

#### Scenario: Unpublish template
- **WHEN** 管理员点击"下架"
- **THEN** 系统将模板状态改为 `draft`
- **AND** 模板从服务目录中隐藏

### Requirement: Review notification

系统 SHALL 通知用户审核结果。

#### Scenario: Notify on approval
- **WHEN** 模板审核通过
- **THEN** 系统向提交者发送通知

#### Scenario: Notify on rejection
- **WHEN** 模板审核被驳回
- **THEN** 系统向提交者发送通知，包含驳回原因

### Requirement: Review permission

系统 SHALL 限制审核操作权限。

#### Scenario: Non-admin cannot approve
- **WHEN** 非管理员用户尝试发布模板
- **THEN** 系统拒绝并提示"无权限执行此操作"
