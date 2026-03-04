# ai-tool-expansion Specification

## Purpose

定义AI工具能力覆盖范围，确保AI助手能够有效协助用户操作系统的各项功能模块。

## Requirements

### Requirement: 服务管理域工具支持

系统SHALL提供服务管理相关的AI工具，支持服务目录查询、分类管理和可见性检查。

#### Scenario: 查询服务目录列表
- **WHEN** 用户通过AI助手查询服务目录
- **THEN** 系统MUST返回包含分类、可见性信息的服务目录列表

#### Scenario: 获取服务分类树
- **WHEN** 用户请求服务分类结构
- **THEN** 系统MUST返回完整的服务分类树形结构

### Requirement: 部署目标域工具支持

系统SHALL提供部署目标相关的AI工具，支持目标列表查询、详情查看和环境引导状态追踪。

#### Scenario: 查询部署目标
- **WHEN** 用户查询部署目标列表
- **THEN** 系统MUST返回包含环境、状态、集群信息的目标列表

#### Scenario: 查看环境引导状态
- **WHEN** 用户查看特定部署目标的引导状态
- **THEN** 系统MUST返回引导进度和当前状态

### Requirement: CI/CD域工具支持

系统SHALL提供CI/CD相关的AI工具，支持流水线查询和构建触发，高风险操作需审批。

#### Scenario: 查询流水线状态
- **WHEN** 用户查询特定流水线状态
- **THEN** 系统MUST返回流水线当前状态和最近执行记录

#### Scenario: 触发构建需审批
- **WHEN** 用户通过AI触发流水线构建
- **THEN** 系统MUST要求审批确认后才能执行

### Requirement: 配置中心域工具支持

系统SHALL提供配置中心相关的AI工具，支持配置应用查询、配置项获取和差异对比。

#### Scenario: 配置差异对比
- **WHEN** 用户请求对比两个环境的配置
- **THEN** 系统MUST返回配置项的差异对比结果

### Requirement: 监控告警域工具支持

系统SHALL提供监控告警相关的AI工具，支持告警规则查询、活跃告警查看和指标查询。

#### Scenario: 查询活跃告警
- **WHEN** 用户查询当前活跃告警
- **THEN** 系统MUST返回未恢复的告警列表及其详情

### Requirement: 治理域工具支持

系统SHALL提供治理相关的AI工具，支持用户查询、角色查询和权限检查。

#### Scenario: 权限检查
- **WHEN** 用户请求检查特定权限
- **THEN** 系统MUST返回用户对指定资源的操作权限结果

---

# ai-param-intelligence Specification

## Purpose

定义AI工具参数智能解析能力，解决AI调用时参数缺失或错误的问题。

## Requirements

### Requirement: 工具Schema增强

系统SHALL为每个工具提供完整的Schema描述，包括参数来源提示、枚举值来源和填写示例。

#### Scenario: 获取参数来源
- **WHEN** AI需要确定参数值
- **THEN** 系统MUST提供参数值的获取来源（如对应的inventory工具）

### Requirement: 参数自动发现

系统SHALL提供参数提示接口，返回参数的可选值列表。

#### Scenario: 获取参数可选值
- **WHEN** AI或前端请求参数提示
- **THEN** 系统MUST返回该参数的可用值列表和标签

### Requirement: 场景上下文自动注入

系统SHALL从当前页面自动提取上下文信息，包括资源ID、命名空间等，并自动注入到工具调用。

#### Scenario: 页面数据提取
- **WHEN** 用户在特定页面打开AI助手
- **THEN** 系统MUST自动提取页面相关的资源ID和配置

#### Scenario: 路由参数映射
- **WHEN** 用户在详情页面（如 /services/:id）
- **THEN** 系统MUST自动将路由参数映射为工具参数

### Requirement: 参数校验与提示

系统SHALL在工具执行前校验参数，缺失或错误时返回友好的提示信息。

#### Scenario: 缺失参数提示
- **WHEN** 工具调用缺少必填参数
- **THEN** 系统MUST返回缺失参数列表和获取方式建议

#### Scenario: 参数修正建议
- **WHEN** 参数格式或值不正确
- **THEN** 系统MUST提供修正建议

---

# ai-scene-classification Specification

## Purpose

定义AI助手场景细分能力，提升上下文准确性和工具推荐的相关性。

## Requirements

### Requirement: 一级场景扩展

系统SHALL支持以下一级场景：services, configcenter, jobs, cicd, governance, cmdb, automation, monitor。

#### Scenario: 识别服务管理场景
- **WHEN** 用户访问 /services 路由
- **THEN** 系统MUST识别为 services 场景并加载相关上下文

### Requirement: 二级场景细分

系统SHALL支持部署管理和服务管理的二级场景细分，提供更精确的上下文。

#### Scenario: 部署管理二级场景
- **WHEN** 用户访问 /deployment/infrastructure/clusters
- **THEN** 系统MUST识别为 deployment:clusters 场景

#### Scenario: 服务管理二级场景
- **WHEN** 用户访问 /services/:id
- **THEN** 系统MUST识别为 services:detail 场景并注入 service_id

### Requirement: 场景上下文关联

系统SHALL为每个场景关联业务描述、可用工具和上下文提示。

#### Scenario: 获取场景工具推荐
- **WHEN** 用户在特定场景打开AI助手
- **THEN** 系统MUST推荐该场景相关的工具和快捷命令
