# Spec: Service Catalog Browse

服务目录浏览、搜索、分类筛选功能。

## ADDED Requirements

### Requirement: Catalog list display

系统 SHALL 显示服务目录列表，展示所有已发布的模板。

#### Scenario: Browse published templates
- **WHEN** 用户访问服务目录页面
- **THEN** 系统显示所有 `status=published` 的模板列表
- **AND** 每个模板显示名称、描述、分类、图标

#### Scenario: Empty catalog
- **WHEN** 服务目录中没有已发布的模板
- **THEN** 系统显示"暂无服务模板"的空状态

### Requirement: Category filter

系统 SHALL 支持按分类筛选模板。

#### Scenario: Filter by category
- **WHEN** 用户选择某个分类
- **THEN** 系统仅显示该分类下的模板

#### Scenario: Show all categories
- **WHEN** 用户未选择分类或选择"全部"
- **THEN** 系统显示所有分类的模板

### Requirement: Search functionality

系统 SHALL 支持按模板名称和标签搜索。

#### Scenario: Search by name
- **WHEN** 用户输入搜索关键词
- **THEN** 系统返回名称或描述中包含该关键词的模板

#### Scenario: Search by tag
- **WHEN** 用户输入标签关键词
- **THEN** 系统返回标签中包含该关键词的模板

### Requirement: Template detail view

系统 SHALL 显示模板详情页面。

#### Scenario: View template detail
- **WHEN** 用户点击模板卡片
- **THEN** 系统显示模板详情页
- **AND** 显示模板名称、描述、变量定义、使用说明
- **AND** 显示"部署"按钮

### Requirement: Template statistics

系统 SHALL 显示模板的部署统计。

#### Scenario: Display deploy count
- **WHEN** 用户查看模板详情
- **THEN** 系统显示该模板的部署次数
