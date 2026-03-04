# Spec: Service Template Management

模板创建、编辑、版本管理功能。

## ADDED Requirements

### Requirement: Create template

系统 SHALL 允许用户创建服务模板。

#### Scenario: Create template flow
- **WHEN** 用户点击"创建模板"
- **THEN** 系统显示模板创建向导
- **AND** 引导用户填写基本信息、K8s 模板、Compose 模板、变量定义

### Requirement: Basic info input

系统 SHALL 收集模板基本信息。

#### Scenario: Input basic info
- **WHEN** 用户填写模板基本信息
- **THEN** 系统要求填写名称、显示名称、描述
- **AND** 系统要求选择分类
- **AND** 系统可选填写图标和标签

#### Scenario: Unique template name
- **WHEN** 用户输入已存在的模板名称
- **THEN** 系统提示"模板名称已存在"

### Requirement: K8s template input

系统 SHALL 支持输入 K8s YAML 模板。

#### Scenario: Input K8s template
- **WHEN** 用户在 K8s 模板编辑器中输入 YAML
- **THEN** 系统提供代码高亮和语法检查
- **AND** 系统自动检测模板中的变量

#### Scenario: Validate K8s YAML
- **WHEN** 用户输入无效的 K8s YAML
- **THEN** 系统显示语法错误提示

### Requirement: Compose template input

系统 SHALL 支持输入 Docker Compose YAML 模板。

#### Scenario: Input Compose template
- **WHEN** 用户在 Compose 模板编辑器中输入 YAML
- **THEN** 系统提供代码高亮和语法检查
- **AND** 系统自动检测模板中的变量

### Requirement: Variable definition

系统 SHALL 支持定义模板变量。

#### Scenario: Auto-detect variables
- **WHEN** 用户输入包含 `{{ var_name }}` 的模板
- **THEN** 系统自动检测并列出变量
- **AND** 用户可设置变量类型、默认值、是否必填、说明

#### Scenario: Manual add variable
- **WHEN** 用户手动添加变量定义
- **THEN** 系统将变量定义添加到列表

### Requirement: Save as draft

系统 SHALL 支持保存草稿。

#### Scenario: Save draft
- **WHEN** 用户点击"保存草稿"
- **THEN** 系统保存模板为 `status=draft`
- **AND** 模板仅对创建者可见

### Requirement: Edit template

系统 SHALL 允许用户编辑自己的模板。

#### Scenario: Edit own template
- **WHEN** 模板作者点击"编辑"
- **THEN** 系统显示模板编辑页面
- **AND** 预填充现有模板内容

#### Scenario: Edit permission check
- **WHEN** 非作者尝试编辑模板
- **THEN** 系统拒绝并提示"无权限编辑此模板"

### Requirement: Delete template

系统 SHALL 允许删除模板。

#### Scenario: Delete own template
- **WHEN** 模板作者点击"删除"
- **THEN** 系统提示确认对话框
- **AND** 确认后删除模板

#### Scenario: Cannot delete published template
- **WHEN** 用户尝试删除已发布的模板
- **THEN** 系统提示"已发布的模板无法删除，请先下架"

### Requirement: Template readme

系统 SHALL 支持编辑模板使用说明。

#### Scenario: Edit readme
- **WHEN** 用户编辑模板 readme
- **THEN** 系统提供 Markdown 编辑器
- **AND** 在模板详情页渲染 Markdown 内容
