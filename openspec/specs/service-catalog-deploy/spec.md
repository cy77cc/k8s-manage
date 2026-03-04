# Spec: Service Catalog Deploy

从模板部署服务到 K8s/Compose 环境。

## ADDED Requirements

### Requirement: Deploy target selection

系统 SHALL 支持选择部署目标类型。

#### Scenario: Select K8s cluster
- **WHEN** 用户选择 K8s 部署目标
- **THEN** 系统显示可选的 K8s 集群列表
- **AND** 显示可选的命名空间

#### Scenario: Select Compose environment
- **WHEN** 用户选择 Compose 部署目标
- **THEN** 系统显示可选的主机/环境列表

### Requirement: Variable input form

系统 SHALL 根据模板变量定义生成输入表单。

#### Scenario: Render variable form
- **WHEN** 用户进入部署配置页面
- **THEN** 系统根据 `variables_schema` 生成对应的表单字段
- **AND** 必填字段标记为必填
- **AND** 显示字段的默认值和说明

#### Scenario: Variable type mapping
- **WHEN** 变量类型为 `string`
- **THEN** 系统显示文本输入框
- **WHEN** 变量类型为 `number`
- **THEN** 系统显示数字输入框
- **WHEN** 变量类型为 `password`
- **THEN** 系统显示密码输入框
- **WHEN** 变量类型为 `select`
- **THEN** 系统显示下拉选择框

### Requirement: YAML preview

系统 SHALL 在部署前显示渲染后的 YAML 预览。

#### Scenario: Preview rendered YAML
- **WHEN** 用户填写变量值后点击预览
- **THEN** 系统显示渲染后的 K8s YAML 或 Compose YAML
- **AND** 系统提示未填充的必填变量（如有）

### Requirement: Deploy execution

系统 SHALL 执行模板部署。

#### Scenario: Deploy to K8s
- **WHEN** 用户确认部署到 K8s 集群
- **THEN** 系统创建 Service 记录
- **AND** 系统调用 K8s 部署 API
- **AND** 系统增加模板的部署次数

#### Scenario: Deploy to Compose
- **WHEN** 用户确认部署到 Compose 环境
- **THEN** 系统创建 Service 记录
- **AND** 系统调用 Compose 部署 API
- **AND** 系统增加模板的部署次数

### Requirement: Deploy validation

系统 SHALL 在部署前验证必填变量。

#### Scenario: Missing required variable
- **WHEN** 用户未填写必填变量
- **THEN** 系统阻止部署并提示"请填写所有必填字段"

#### Scenario: Invalid variable value
- **WHEN** 用户输入的变量值不符合类型要求
- **THEN** 系统提示具体的验证错误

### Requirement: Deploy result

系统 SHALL 显示部署结果。

#### Scenario: Deploy success
- **WHEN** 部署成功完成
- **THEN** 系统显示成功消息
- **AND** 提供跳转到服务详情页的链接

#### Scenario: Deploy failure
- **WHEN** 部署过程中发生错误
- **THEN** 系统显示错误消息
- **AND** 允许用户修改配置后重试
