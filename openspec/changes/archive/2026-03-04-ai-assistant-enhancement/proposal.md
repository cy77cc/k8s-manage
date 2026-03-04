## Why

AI助手模块当前存在三个核心痛点：

1. **工具能力缺口**：系统新增了大量功能模块（服务目录、部署目标、CI/CD、配置中心、任务调度等），但AI工具覆盖仍停留在主机、K8s、服务基础操作层面，导致AI无法有效协助用户操作新功能。

2. **参数传递困难**：大量工具的Schema描述不够清晰，AI调用时经常缺少必要参数报错。例如`cluster_id`、`service_id`等ID类参数，AI无法知道可用值；`namespace`等名称参数，AI不知道可选范围。

3. **场景分类粗糙**：当前场景分类仅细分到`deployment:infrastructure/targets/observability`一级，无法准确区分具体的业务上下文。服务管理、配置中心、任务调度、CI/CD等模块完全没有独立场景，导致AI上下文信息不足。

## What Changes

### 一、工具能力扩展

新增覆盖以下域的工具：

- **服务管理域**：服务目录查询、服务分类树、服务可见性配置
- **部署目标域**：部署目标列表、目标详情、环境引导状态
- **基础设施域**：凭证列表、凭证连通性测试
- **CI/CD域**：流水线列表、流水线状态、触发构建
- **任务调度域**：任务列表、执行状态、手动触发任务
- **配置中心域**：配置应用列表、配置项查询、配置差异对比
- **监控告警域**：告警规则列表、活跃告警、指标查询
- **拓扑审计域**：服务拓扑查询、审计日志搜索
- **治理域**：用户列表、角色列表、权限检查

### 二、参数智能解析

- **工具Schema增强**：统一Description格式，添加`enum_source`字段指示参数值来源，添加`param_hints`参数填写提示
- **参数自动发现**：新增`/ai/tools/:name/params/hints`接口，返回参数可选值列表
- **场景上下文自动注入**：从当前页面自动提取`cluster_id`、`service_id`、`host_id`等上下文
- **参数校验与提示**：缺失必填参数时返回友好提示而非报错，参数格式错误时给出修正建议

### 三、场景细分增强

- **一级场景扩展**：新增`services`、`configcenter`、`jobs`、`cicd`、`governance`、`cmdb`、`automation`
- **二级场景细分**：`deployment`下细分为`clusters`、`credentials`、`hosts`、`targets`、`releases`、`approvals`、`topology`、`metrics`、`audit`、`aiops`
- **场景上下文增强**：每个场景关联业务描述、可用操作、相关工具推荐

### 四、补充功能

- **专家路由增强**：基于场景自动选择专家，新增`deployment_expert`、`service_expert`、`monitor_expert`
- **工具发现与引导**：场景工具推荐、命令自动补全、工具分类浏览
- **工具结果可视化**：自动选择展示方式（表格/图表/拓扑图）
- **快捷指令系统**：内置别名、自定义别名、参数模板
- **多轮对话增强**：上下文保持、引用回指、对话分支
- **错误恢复与重试**：智能重试、参数修正提示、降级方案

## Capabilities

### New Capabilities

- `ai-tool-expansion`: 定义AI工具覆盖范围扩展，包括服务管理、部署目标、CI/CD、任务调度、配置中心、监控告警、拓扑审计、治理等域的工具契约
- `ai-param-intelligence`: 定义参数智能解析能力，包括Schema增强、参数自动发现、上下文注入、校验提示等
- `ai-scene-classification`: 定义场景细分能力，包括一级场景扩展、二级场景细分、场景上下文关联

### Modified Capabilities

- `ai-assistant-experience-optimization`: 体验优化需求变更，增加工具发现引导、结果可视化、快捷指令等增强功能
- `ai-assistant-command-bridge`: 命令桥接需求变更，扩展命令覆盖范围和参数解析能力

## Impact

### Backend Domains

- `internal/ai/tools/*`：新增约25个工具实现
- `internal/service/ai/*`：参数提示接口、场景路由逻辑、专家选择逻辑
- `internal/model/*`：可能需要新增参数提示相关的缓存结构

### Frontend Domains

- `web/src/components/AI/GlobalAIAssistant.tsx`：场景细分逻辑更新
- `web/src/components/AI/ChatInterface.tsx`：参数选择器、结果可视化
- `web/src/components/AI/CommandPanel.tsx`：命令补全、工具发现
- `web/src/api/modules/ai.ts`：新增参数提示接口

### Data/Contracts

- 新增API：`GET /api/v1/ai/tools/:name/params/hints`
- 扩展API：`GET /api/v1/ai/capabilities` 返回增强的工具Schema

### Security/Operations

- 新工具的权限控制配置
- 场景上下文注入的数据访问权限校验
