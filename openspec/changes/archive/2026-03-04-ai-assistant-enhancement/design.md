# AI助手模块优化增强 - 技术设计

## 一、架构概览

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              前端层 (React)                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│  GlobalAIAssistant                                                               │
│  ├── SceneRouter          ─── 场景路由（细分场景识别）                             │
│  ├── ChatInterface         ├── ParamSelector     (参数选择器)                    │
│  │                        ├── ResultVisualizer   (结果可视化)                    │
│  │                        └── QuickCommandBar    (快捷指令栏)                    │
│  └── CommandPanel          ├── CommandCompleter  (命令补全)                      │
│                           ├── ToolDiscovery      (工具发现)                      │
│                           └── ParamHints         (参数提示)                      │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              API层 (Go/Gin)                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│  POST /ai/chat                    SSE流式对话（增强场景上下文）                    │
│  GET  /ai/tools/:name/params/hints    参数提示接口 (新增)                        │
│  GET  /ai/capabilities            工具能力列表（增强Schema）                      │
│  GET  /ai/scene/:scene/tools      场景工具推荐 (新增)                             │
│  GET  /ai/commands/suggestions    命令建议（场景关联）                            │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              内部AI层 (Go)                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│  platform_agent.go                                                               │
│  ├── ExpertRouter          ─── 专家路由（场景感知）                               │
│  │   ├── deployment_expert
│  │   ├── service_expert
│  │   ├── monitor_expert
│  │   └── ...
│  ├── tools/
│  │   ├── tools_registry.go        工具注册（增强Schema）
│  │   ├── tools_service.go         服务域工具 (扩展)
│  │   ├── tools_deployment.go      部署域工具 (新增)
│  │   ├── tools_cicd.go            CI/CD工具 (新增)
│  │   ├── tools_job.go             任务调度工具 (新增)
│  │   ├── tools_config.go          配置中心工具 (新增)
│  │   ├── tools_monitor.go         监控告警工具 (新增)
│  │   ├── tools_governance.go      治理工具 (新增)
│  │   └── param_hints.go           参数提示解析器 (新增)
│  └── scene/
│      ├── scene_router.go          场景路由器 (新增)
│      ├── scene_context.go         场景上下文提取 (新增)
│      └── scene_tools.go           场景工具映射 (新增)
└─────────────────────────────────────────────────────────────────────────────────┘
```

## 二、工具能力扩展设计

### 2.1 工具分类与命名规范

```
工具命名规范: {domain}_{action}_{target}

domain:    host, cluster, service, deployment, cicd, job, config, monitor, topology, audit, user, role
action:    get, list, create, update, delete, preview, apply, trigger, check, search
target:    具体资源名称

示例:
- service_catalog_list        服务目录列表
- service_category_tree       服务分类树
- deployment_target_list      部署目标列表
- cicd_pipeline_trigger       CI/CD流水线触发
- config_item_get             配置项查询
- monitor_alert_active        活跃告警查询
```

### 2.2 新增工具清单

#### 服务管理域 (tools_service.go 扩展)

| 工具名 | 模式 | 风险 | 必填参数 | 描述 |
|--------|------|------|----------|------|
| service_catalog_list | readonly | low | - | 服务目录列表，支持分类过滤 |
| service_category_tree | readonly | low | - | 服务分类树结构 |
| service_visibility_check | readonly | low | service_id | 服务可见性配置检查 |

#### 部署目标域 (tools_deployment.go 新增)

| 工具名 | 模式 | 风险 | 必填参数 | 描述 |
|--------|------|------|----------|------|
| deployment_target_list | readonly | low | - | 部署目标列表 |
| deployment_target_detail | readonly | low | target_id | 目标详情 |
| deployment_bootstrap_status | readonly | low | target_id | 环境引导状态 |

#### CI/CD域 (tools_cicd.go 新增)

| 工具名 | 模式 | 风险 | 必填参数 | 描述 |
|--------|------|------|----------|------|
| cicd_pipeline_list | readonly | low | - | 流水线列表 |
| cicd_pipeline_status | readonly | low | pipeline_id | 流水线状态 |
| cicd_pipeline_trigger | mutating | high | pipeline_id, branch | 触发构建 |

#### 任务调度域 (tools_job.go 新增)

| 工具名 | 模式 | 风险 | 必填参数 | 描述 |
|--------|------|------|----------|------|
| job_list | readonly | low | - | 任务列表 |
| job_execution_status | readonly | low | job_id | 执行状态 |
| job_run | mutating | medium | job_id | 手动触发任务 |

#### 配置中心域 (tools_config.go 新增)

| 工具名 | 模式 | 风险 | 必填参数 | 描述 |
|--------|------|------|----------|------|
| config_app_list | readonly | low | - | 配置应用列表 |
| config_item_get | readonly | low | app_id, key | 配置项查询 |
| config_diff | readonly | low | app_id, env_a, env_b | 配置差异对比 |

#### 监控告警域 (tools_monitor.go 新增)

| 工具名 | 模式 | 风险 | 必填参数 | 描述 |
|--------|------|------|----------|------|
| monitor_alert_rule_list | readonly | low | - | 告警规则列表 |
| monitor_alert_active | readonly | low | - | 活跃告警 |
| monitor_metric_query | readonly | low | query, time_range | 指标查询 |

#### 拓扑审计域 (tools_topology.go 新增)

| 工具名 | 模式 | 风险 | 必填参数 | 描述 |
|--------|------|------|----------|------|
| topology_get | readonly | low | service_id? | 服务拓扑查询 |
| audit_log_search | readonly | low | time_range?, resource_type? | 审计日志搜索 |

#### 治理域 (tools_governance.go 新增)

| 工具名 | 模式 | 风险 | 必填参数 | 描述 |
|--------|------|------|----------|------|
| user_list | readonly | low | - | 用户列表 |
| role_list | readonly | low | - | 角色列表 |
| permission_check | readonly | low | user_id, resource, action | 权限检查 |

### 2.3 工具Schema增强设计

```go
type ToolMeta struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Mode        ToolMode       `json:"mode"`
    Risk        ToolRisk       `json:"risk"`
    Provider    string         `json:"provider"`
    Permission  string         `json:"permission"`
    Schema      map[string]any `json:"schema,omitempty"`
    Required    []string       `json:"required,omitempty"`
    DefaultHint map[string]any `json:"default_hint,omitempty"`
    Examples    []string       `json:"examples,omitempty"`

    // 新增字段
    EnumSources  map[string]string `json:"enum_sources,omitempty"`  // 参数值来源
    ParamHints   map[string]string `json:"param_hints,omitempty"`   // 参数填写提示
    RelatedTools []string          `json:"related_tools,omitempty"` // 相关工具
    SceneScope   []string          `json:"scene_scope,omitempty"`   // 适用场景
}

// Description格式规范
// 格式: {功能描述}。{必填参数说明}。{默认值说明}。示例: {JSON示例}。参数来源: {来源说明}
// 示例:
// "查询服务详情。service_id必填。默认limit=50。示例: {\"service_id\":123}。service_id可从service_list_inventory获取。"
```

## 三、参数智能解析设计

### 3.1 参数提示接口

```
GET /api/v1/ai/tools/:name/params/hints

Response:
{
  "tool": "k8s_list_resources",
  "params": {
    "cluster_id": {
      "type": "integer",
      "required": true,
      "hint": "集群ID，可从cluster_list_inventory获取",
      "enum_source": "cluster_list_inventory",
      "values": [
        {"value": 1, "label": "prod-cluster-01"},
        {"value": 2, "label": "test-cluster-01"}
      ]
    },
    "namespace": {
      "type": "string",
      "required": false,
      "default": "default",
      "hint": "K8s命名空间",
      "enum_source": null  // 无固定可选值
    },
    "resource": {
      "type": "string",
      "required": true,
      "enum": ["pods", "services", "deployments", "nodes"]
    }
  }
}
```

### 3.2 场景上下文提取

```typescript
// 前端场景上下文提取
interface SceneContext {
  scene: string;           // 场景标识
  pageData: {              // 页面数据
    cluster_id?: number;
    service_id?: number;
    host_id?: number;
    target_id?: number;
    namespace?: string;
    env?: string;
  };
  selectedItems: {         // 用户选中项
    type: string;
    ids: number[];
  };
}

// 路由到场景映射
const routeSceneMap: Record<string, string> = {
  '/deployment/infrastructure/clusters': 'deployment:clusters',
  '/deployment/infrastructure/credentials': 'deployment:credentials',
  '/deployment/infrastructure/hosts': 'deployment:hosts',
  '/deployment/targets': 'deployment:targets',
  '/deployment/:id': 'deployment:releases',
  '/deployment/approvals': 'deployment:approvals',
  '/deployment/observability/topology': 'deployment:topology',
  '/deployment/observability/metrics': 'deployment:metrics',
  '/deployment/observability/audit-logs': 'deployment:audit',
  '/deployment/observability/aiops': 'deployment:aiops',
  '/services': 'services:list',
  '/services/:id': 'services:detail',
  '/services/provision': 'services:provision',
  '/configcenter': 'configcenter',
  '/jobs': 'jobs',
  '/cicd': 'cicd',
  '/governance/users': 'governance:users',
  '/governance/roles': 'governance:roles',
  '/governance/permissions': 'governance:permissions',
  '/cmdb': 'cmdb',
  '/automation': 'automation',
};
```

### 3.3 参数解析流程

```
┌────────────────────────────────────────────────────────────────────────────┐
│                            参数解析流程                                      │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  1. AI生成工具调用                                                          │
│     │                                                                      │
│     ▼                                                                      │
│  2. 检查必填参数                                                            │
│     ├── 全部存在 → 继续                                                     │
│     └── 有缺失 ↓                                                           │
│         ├── 场景上下文有值 → 自动注入                                       │
│         ├── 会话记忆有值 → 自动注入                                         │
│         ├── 默认值存在 → 自动注入                                           │
│         └── 仍缺失 → 返回参数提示给AI                                       │
│     │                                                                      │
│     ▼                                                                      │
│  3. 参数校验                                                                │
│     ├── 类型正确 → 继续                                                     │
│     ├── 枚举值校验 → 校验通过/返回修正建议                                   │
│     └── 格式错误 → 返回修正建议                                             │
│     │                                                                      │
│     ▼                                                                      │
│  4. 执行工具调用                                                            │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
```

## 四、场景细分设计

### 4.1 场景层级结构

```
scene:{module}:{submodule}:{detail?}

一级场景 (module):
├── home              首页
├── deployment        部署管理
├── services          服务管理
├── configcenter      配置中心
├── jobs              任务调度
├── cicd              CI/CD
├── governance        治理管理
├── cmdb              资产管理
├── automation        自动化
├── monitor           监控中心
└── tools             工具集成

二级场景 (submodule):
deployment:
├── clusters          集群管理
├── credentials       凭证管理
├── hosts             主机管理
├── targets           部署目标
├── releases          发布管理
├── approvals         审批中心
├── topology          拓扑视图
├── metrics           指标监控
├── audit             审计日志
└── aiops             智能运维

services:
├── list              服务列表
├── detail            服务详情
├── provision         服务供应
├── deploy            服务部署
└── catalog           服务目录

governance:
├── users             用户管理
├── roles             角色管理
└── permissions       权限管理
```

### 4.2 场景上下文关联

```go
type SceneMeta struct {
    Scene        string   `json:"scene"`
    Description  string   `json:"description"`
    Keywords     []string `json:"keywords"`
    Tools        []string `json:"tools"`
    ContextHints []string `json:"context_hints"`
    FAQ          []string `json:"faq"`
}

var sceneRegistry = map[string]SceneMeta{
    "deployment:clusters": {
        Description: "Kubernetes集群管理，包括集群列表、详情、导入、引导",
        Keywords:    []string{"集群", "k8s", "kubernetes", "cluster"},
        Tools:       []string{"cluster_list_inventory", "k8s_list_resources", "k8s_get_events"},
        ContextHints: []string{"cluster_id"},
    },
    "deployment:targets": {
        Description: "部署目标管理，包括环境、命名空间、部署配置",
        Keywords:    []string{"部署目标", "环境", "target", "env"},
        Tools:       []string{"deployment_target_list", "deployment_target_detail"},
        ContextHints: []string{"target_id", "env"},
    },
    "services:detail": {
        Description: "服务详情查看，包括服务配置、实例、发布历史",
        Keywords:    []string{"服务", "service", "微服务"},
        Tools:       []string{"service_get_detail", "service_deploy_preview", "k8s_get_pod_logs"},
        ContextHints: []string{"service_id"},
    },
    // ... 其他场景
}
```

## 五、补充功能设计

### 5.1 专家路由增强

```go
func (p *PlatformAgent) selectAgentByScene(scene string, messages []*schema.Message) *react.Agent {
    // 场景优先
    switch {
    case strings.HasPrefix(scene, "deployment:clusters"):
        return p.experts["k8s"]
    case strings.HasPrefix(scene, "deployment:hosts"):
        return p.experts["ops"]
    case strings.HasPrefix(scene, "services"):
        return p.experts["service"]
    case strings.HasPrefix(scene, "monitor") || strings.HasPrefix(scene, "deployment:metrics"):
        return p.experts["monitor"]
    case strings.HasPrefix(scene, "governance"):
        return p.experts["security"]
    }

    // 关键词回退
    return p.selectAgentByKeywords(messages)
}
```

### 5.2 工具发现与引导

```typescript
// 场景工具推荐
interface SceneToolRecommendation {
  scene: string;
  tools: Array<{
    name: string;
    description: string;
    quickCommand: string;
    category: 'query' | 'action' | 'diagnosis';
  }>;
}

// 命令自动补全
interface CommandCompletion {
  input: string;
  suggestions: Array<{
    command: string;
    description: string;
    params: string[];
  }>;
}
```

### 5.3 快捷指令系统

```typescript
// 内置别名
const builtInAliases: Record<string, string> = {
  'hst': 'host_list_inventory',
  'svc': 'service_list_inventory',
  'cls': 'cluster_list_inventory',
  'pl': 'pipeline_list',
  'job': 'job_list',
  'cfg': 'config_app_list',
  'alert': 'monitor_alert_active',
  'topo': 'topology_get',
};

// 自定义别名存储
interface UserAlias {
  alias: string;
  command: string;
  defaultParams?: Record<string, any>;
}
```

### 5.4 错误恢复与重试

```go
type ToolExecutionError struct {
    Code        string `json:"code"`
    Message     string `json:"message"`
    Recoverable bool   `json:"recoverable"`
    Suggestions []string `json:"suggestions,omitempty"`
    HintAction  string `json:"hint_action,omitempty"`  // 建议的下一步操作
}

// 错误码定义
const (
    ErrMissingParam    = "missing_param"
    ErrInvalidParam    = "invalid_param"
    ErrPermissionDenied = "permission_denied"
    ErrResourceNotFound = "resource_not_found"
    ErrTimeout         = "timeout"
    ErrInternal        = "internal_error"
)

// 错误恢复建议
var errorRecoveryHints = map[string][]string{
    ErrMissingParam: {
        "检查是否从上下文获取了必要参数",
        "调用 {enum_source} 获取可用值",
        "询问用户提供参数值",
    },
    ErrResourceNotFound: {
        "资源可能已删除，尝试刷新列表",
        "检查资源ID是否正确",
    },
}
```

## 六、实施阶段

### Phase 1: 工具能力扩展 (P0)
- 新增服务管理域工具
- 新增部署目标域工具
- 新增CI/CD域工具
- 新增配置中心域工具
- 新增监控告警域工具
- 新增治理域工具

### Phase 2: 参数智能解析 (P0)
- 增强工具Schema
- 实现参数提示接口
- 实现场景上下文提取
- 实现参数校验与提示

### Phase 3: 场景细分增强 (P0)
- 扩展一级场景
- 细分二级场景
- 实现场景上下文关联

### Phase 4: 补充功能 (P1-P3)
- 专家路由增强
- 工具发现与引导
- 快捷指令系统
- 错误恢复与重试
- 工具结果可视化
- 多轮对话增强
