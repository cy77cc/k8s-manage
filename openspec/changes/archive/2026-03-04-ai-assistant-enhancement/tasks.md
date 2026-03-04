# AI助手模块优化增强 - 实施任务

## Phase 1: 工具能力扩展

### 1.1 服务管理域工具

- [x] **T1.1.1** 扩展 `tools_service.go`，新增 `service_catalog_list` 工具
  - 输入: keyword?, category_id?, limit
  - 输出: 服务目录列表，包含分类、可见性信息
  - 权限: `ai:tool:read`

- [x] **T1.1.2** 新增 `service_category_tree` 工具
  - 输入: 无
  - 输出: 服务分类树结构
  - 权限: `ai:tool:read`

- [x] **T1.1.3** 新增 `service_visibility_check` 工具
  - 输入: service_id (必填)
  - 输出: 服务可见性配置详情
  - 权限: `ai:tool:read`

### 1.2 部署目标域工具

- [x] **T1.2.1** 新建 `tools_deployment.go`，实现 `deployment_target_list` 工具
  - 输入: env?, status?, keyword?, limit
  - 输出: 部署目标列表
  - 权限: `ai:tool:read`

- [x] **T1.2.2** 实现 `deployment_target_detail` 工具
  - 输入: target_id (必填)
  - 输出: 目标详情，包括环境、集群、命名空间配置
  - 权限: `ai:tool:read`

- [x] **T1.2.3** 实现 `deployment_bootstrap_status` 工具
  - 输入: target_id (必填)
  - 输出: 环境引导进度、状态
  - 权限: `ai:tool:read`

### 1.3 基础设施域工具

- [x] **T1.3.1** 新建 `tools_infrastructure.go`，实现 `credential_list` 工具
  - 输入: type?, keyword?, limit
  - 输出: 凭证列表
  - 权限: `ai:tool:read`

- [x] **T1.3.2** 实现 `credential_test` 工具
  - 输入: credential_id (必填)
  - 输出: 连通性测试结果
  - 权限: `ai:tool:read`

### 1.4 CI/CD域工具

- [x] **T1.4.1** 新建 `tools_cicd.go`，实现 `cicd_pipeline_list` 工具
  - 输入: status?, keyword?, limit
  - 输出: 流水线列表
  - 权限: `ai:tool:read`

- [x] **T1.4.2** 实现 `cicd_pipeline_status` 工具
  - 输入: pipeline_id (必填)
  - 输出: 流水线状态、最近执行记录
  - 权限: `ai:tool:read`

- [x] **T1.4.3** 实现 `cicd_pipeline_trigger` 工具
  - 输入: pipeline_id (必填), branch (必填), params?
  - 输出: 触发结果、执行ID
  - 权限: `ai:tool:execute`
  - 风险: high，需审批

### 1.5 任务调度域工具

- [x] **T1.5.1** 新建 `tools_job.go`，实现 `job_list` 工具
  - 输入: status?, keyword?, limit
  - 输出: 任务列表
  - 权限: `ai:tool:read`

- [x] **T1.5.2** 实现 `job_execution_status` 工具
  - 输入: job_id (必填), execution_id?
  - 输出: 任务执行状态
  - 权限: `ai:tool:read`

- [x] **T1.5.3** 实现 `job_run` 工具
  - 输入: job_id (必填), params?
  - 输出: 执行ID、状态
  - 权限: `ai:tool:execute`
  - 风险: medium，需审批

### 1.6 配置中心域工具

- [x] **T1.6.1** 新建 `tools_config.go`，实现 `config_app_list` 工具
  - 输入: keyword?, env?, limit
  - 输出: 配置应用列表
  - 权限: `ai:tool:read`

- [x] **T1.6.2** 实现 `config_item_get` 工具
  - 输入: app_id (必填), key (必填), env?
  - 输出: 配置项值
  - 权限: `ai:tool:read`

- [x] **T1.6.3** 实现 `config_diff` 工具
  - 输入: app_id (必填), env_a (必填), env_b (必填)
  - 输出: 配置差异对比结果
  - 权限: `ai:tool:read`

### 1.7 监控告警域工具

- [x] **T1.7.1** 新建 `tools_monitor.go`，实现 `monitor_alert_rule_list` 工具
  - 输入: status?, keyword?, limit
  - 输出: 告警规则列表
  - 权限: `ai:tool:read`

- [x] **T1.7.2** 实现 `monitor_alert_active` 工具
  - 输入: severity?, service_id?, limit
  - 输出: 活跃告警列表
  - 权限: `ai:tool:read`

- [x] **T1.7.3** 实现 `monitor_metric_query` 工具
  - 输入: query (必填), time_range?, step?
  - 输出: 指标数据点
  - 权限: `ai:tool:read`

### 1.8 拓扑审计域工具

- [x] **T1.8.1** 新建 `tools_topology.go`，实现 `topology_get` 工具
  - 输入: service_id?, depth?
  - 输出: 服务拓扑关系图数据
  - 权限: `ai:tool:read`

- [x] **T1.8.2** 实现 `audit_log_search` 工具
  - 输入: time_range?, resource_type?, action?, user_id?, limit
  - 输出: 审计日志列表
  - 权限: `ai:tool:read`

### 1.9 治理域工具

- [x] **T1.9.1** 新建 `tools_governance.go`，实现 `user_list` 工具
  - 输入: keyword?, status?, limit
  - 输出: 用户列表
  - 权限: `ai:tool:read`

- [x] **T1.9.2** 实现 `role_list` 工具
  - 输入: keyword?, limit
  - 输出: 角色列表
  - 权限: `ai:tool:read`

- [x] **T1.9.3** 实现 `permission_check` 工具
  - 输入: user_id (必填), resource (必填), action (必填)
  - 输出: 权限检查结果
  - 权限: `ai:tool:read`

## Phase 2: 参数智能解析

### 2.1 工具Schema增强

- [x] **T2.1.1** 扩展 `ToolMeta` 结构体，新增 `EnumSources`、`ParamHints`、`RelatedTools`、`SceneScope` 字段

- [x] **T2.1.2** 更新现有工具的 Description，统一格式：
  - 功能描述 + 必填参数说明 + 默认值说明 + 示例 + 参数来源

- [x] **T2.1.3** 为ID类参数配置 `EnumSources`，指向对应的 inventory 工具

### 2.2 参数提示接口

- [x] **T2.2.1** 新建 `param_hints.go`，实现参数提示解析器
  - 从 inventory 工具获取可选值列表
  - 缓存高频参数值

- [x] **T2.2.2** 新增 API 路由 `GET /api/v1/ai/tools/:name/params/hints`
  - 返回工具参数的可选值、提示信息、枚举来源

- [x] **T2.2.3** 前端 API 模块新增 `getToolParamHints` 方法

### 2.3 场景上下文提取

- [x] **T2.3.1** 更新 `GlobalAIAssistant.tsx`，实现页面数据提取
  - 从路由参数提取 cluster_id, service_id, host_id 等
  - 从表格选中项提取选中的资源ID

- [x] **T2.3.2** 更新 `ChatInterface.tsx`，在聊天请求中携带页面上下文

- [x] **T2.3.3** 后端 `chat_handler.go` 解析并注入上下文到工具调用

### 2.4 参数校验与提示

- [x] **T2.4.1** 增强 `tool_param_resolver.go`，添加参数校验逻辑

- [x] **T2.4.2** 实现参数缺失时的友好提示
  - 返回缺失参数列表
  - 提供参数获取方式建议

- [x] **T2.4.3** 实现参数格式错误时的修正建议

## Phase 3: 场景细分增强

### 3.1 一级场景扩展

- [x] **T3.1.1** 更新 `sceneFromPath` 函数，新增一级场景识别：
  - services, configcenter, jobs, cicd, governance, cmdb, automation

- [x] **T3.1.2** 后端 `scene_router.go` 新增场景注册表

- [x] **T3.1.3** 更新前端场景上下文传递逻辑

### 3.2 二级场景细分

- [x] **T3.2.1** 部署管理二级场景细分：
  - deployment:clusters, credentials, hosts, targets, releases, approvals, topology, metrics, audit, aiops

- [x] **T3.2.2** 服务管理二级场景细分：
  - services:list, detail, provision, deploy, catalog

- [x] **T3.2.3** 治理管理二级场景细分：
  - governance:users, roles, permissions

### 3.3 场景上下文关联

- [x] **T3.3.1** 新建 `scene_context.go`，定义场景元数据结构

- [x] **T3.3.2** 实现场景与工具的映射关系

- [x] **T3.3.3** 新增 API `GET /api/v1/ai/scene/:scene/tools` 返回场景推荐工具

## Phase 4: 补充功能

### 4.1 专家路由增强

- [x] **T4.1.1** 新增专家 Agent：service_expert, monitor_expert, deployment_expert

- [x] **T4.1.2** 实现场景优先的专家选择逻辑

- [x] **T4.1.3** 为不同专家配置工具子集

### 4.2 工具发现与引导

- [x] **T4.2.1** CommandPanel 新增工具分类浏览面板

- [x] **T4.2.2** 实现场景工具推荐组件

- [x] **T4.2.3** 实现命令自动补全功能

### 4.3 快捷指令系统

- [x] **T4.3.1** 实现内置别名映射

- [x] **T4.3.2** 新增自定义别名存储接口

- [x] **T4.3.3** 实现参数模板保存与应用

### 4.4 错误恢复与重试

- [x] **T4.4.1** 定义 `ToolExecutionError` 结构，包含恢复建议

- [x] **T4.4.2** 实现智能重试逻辑

- [x] **T4.4.3** 前端展示错误恢复建议

### 4.5 工具结果可视化

- [x] **T4.5.1** 实现结果类型自动识别（表格/图表/拓扑）

- [x] **T4.5.2** 新增表格结果展示组件（排序、筛选）

- [x] **T4.5.3** 新增图表结果展示组件（指标数据）

### 4.6 多轮对话增强

- [x] **T4.6.1** 实现对话上下文记忆（提及的资源ID）

- [x] **T4.6.2** 实现引用回指解析（"刚才那个集群"）

- [x] **T4.6.3** 实现对话分支功能

## 测试任务

- [x] **T.TEST.1** 新增工具单元测试覆盖

- [x] **T.TEST.2** 参数提示接口集成测试

- [x] **T.TEST.3** 场景路由端到端测试

- [x] **T.TEST.4** 前端组件测试更新

## 文档任务

- [x] **T.DOC.1** 更新 `docs/ai/tool-catalog-and-policy.md`

- [x] **T.DOC.2** 更新 `docs/ai/api-contracts.md`

- [x] **T.DOC.3** 新增工具使用指南
