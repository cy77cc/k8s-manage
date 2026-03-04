# Spec: Service Category Management

分类管理 (系统预置 + 用户自定义)。

## ADDED Requirements

### Requirement: System preset categories

系统 SHALL 预置常用分类。

#### Scenario: Initialize preset categories
- **WHEN** 系统首次启动
- **THEN** 系统自动创建预置分类: database, cache, message-queue, web-server, monitoring, dev-tools, custom

### Requirement: List categories

系统 SHALL 显示分类列表。

#### Scenario: Get category list
- **WHEN** 用户请求分类列表
- **THEN** 系统返回所有分类
- **AND** 按排序字段排序

### Requirement: Create category

系统 SHALL 允许管理员创建自定义分类。

#### Scenario: Create custom category
- **WHEN** 管理员创建新分类
- **THEN** 系统要求填写名称、显示名称
- **AND** 系统可选填写图标和描述

#### Scenario: Duplicate category name
- **WHEN** 管理员输入已存在的分类名称
- **THEN** 系统提示"分类名称已存在"

### Requirement: Update category

系统 SHALL 允许管理员更新分类。

#### Scenario: Update category
- **WHEN** 管理员更新分类信息
- **THEN** 系统保存更新
- **AND** 关联模板的分类显示同步更新

### Requirement: Delete category

系统 SHALL 允许删除非系统预置分类。

#### Scenario: Delete custom category
- **WHEN** 管理员删除自定义分类
- **THEN** 系统删除该分类

#### Scenario: Cannot delete system category
- **WHEN** 管理员尝试删除系统预置分类 (`is_system=true`)
- **THEN** 系统拒绝并提示"系统预置分类无法删除"

#### Scenario: Delete category with templates
- **WHEN** 管理员删除包含模板的分类
- **THEN** 系统提示"该分类下存在模板，请先移动或删除模板"

### Requirement: Category sort order

系统 SHALL 支持调整分类显示顺序。

#### Scenario: Update sort order
- **WHEN** 管理员调整分类排序
- **THEN** 系统保存排序值
- **AND** 分类列表按新顺序显示

### Requirement: Category permission

系统 SHALL 限制分类管理权限。

#### Scenario: Non-admin cannot manage
- **WHEN** 非管理员用户尝试创建/更新/删除分类
- **THEN** 系统拒绝并提示"无权限执行此操作"

### Requirement: Category icon

系统 SHALL 支持分类图标。

#### Scenario: Use Ant Design icon
- **WHEN** 管理员设置分类图标
- **THEN** 系统支持选择 Ant Design 内置图标名称

#### Scenario: Custom icon URL
- **WHEN** 管理员设置自定义图标 URL
- **THEN** 系统显示该 URL 的图标
