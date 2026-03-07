# Code Organization Convention

## Purpose

定义项目代码的组织规范，要求使用目录形式进行功能分类，避免单个目录下堆积大量文件，提高代码可维护性和可发现性。

## Requirements

### Requirement: Directory Organization SHALL Use Functional Grouping

代码目录 SHALL 使用功能分组，禁止在单个目录下放置超过 10 个同级文件。当文件数量超过阈值时，必须按功能域拆分为子目录。

**Threshold Rules:**
| Category | Max Files Per Directory | Action When Exceeded |
|----------|------------------------|----------------------|
| Models | 10 | Group by domain |
| Handlers | 8 | Group by resource |
| Components | 12 | Group by feature |
| Hooks | 10 | Group by concern |
| Utils | 8 | Group by purpose |
| API Modules | 15 | Group by domain |

#### Scenario: model directory exceeds threshold

- **GIVEN** `internal/model/` contains more than 10 model files
- **WHEN** maintainers organize the codebase
- **THEN** the models SHALL be split into subdirectories by domain (e.g., `model/ai/`, `model/user/`, `model/cluster/`)
- **AND** each subdirectory SHALL contain a `doc.go` or have clear naming

#### Scenario: component directory exceeds threshold

- **GIVEN** `web/src/components/` contains more than 12 component directories
- **WHEN** maintainers organize the codebase
- **THEN** related components SHALL be grouped into feature directories (e.g., `components/Feedback/`, `components/DataDisplay/`)

---

## Backend Organization

### Requirement: Go Models SHALL Be Organized By Domain

后端数据模型 SHALL 按领域域分组，每个领域一个子目录。

**Recommended Structure:**
```
internal/model/
├── ai/                     # AI 相关模型
│   ├── chat.go
│   ├── checkpoint.go
│   ├── approval.go
│   └── doc.go
├── user/                   # 用户与权限模型
│   ├── user.go
│   ├── role.go
│   ├── permission.go
│   └── doc.go
├── infrastructure/         # 基础设施模型
│   ├── host.go
│   ├── cluster.go
│   ├── node.go
│   └── doc.go
├── deployment/             # 部署管理模型
│   ├── deployment.go
│   ├── target.go
│   ├── environment.go
│   └── doc.go
├── service/                # 服务管理模型
│   ├── service.go
│   ├── catalog.go
│   └── doc.go
├── observability/          # 可观测性模型
│   ├── monitoring.go
│   ├── alert.go
│   └── doc.go
└── common/                 # 通用模型
    ├── audit_log.go
    ├── notification.go
    └── doc.go
```

**Acceptance Criteria:**
- [ ] 每个子目录文件数不超过 8 个
- [ ] 每个子目录包含 `doc.go` 说明包用途
- [ ] 跨领域引用通过导入路径显式声明

---

### Requirement: Service Handlers SHALL Be Organized By Resource

服务层处理器 SHALL 按资源分组，每个资源一个子目录，包含 routes、handler、logic。

**Recommended Structure:**
```
internal/service/
├── ai/                     # AI 服务
│   ├── routes.go           # 路由注册
│   ├── handler/            # HTTP 处理器
│   │   ├── chat.go
│   │   ├── session.go
│   │   └── approval.go
│   └── logic/              # 业务逻辑
│       ├── session_store.go
│       └── confirmation.go
├── user/                   # 用户服务
├── cluster/                # 集群服务
├── host/                   # 主机服务
├── deployment/             # 部署服务
└── ...
```

**Acceptance Criteria:**
- [ ] handler 目录文件数不超过 8 个
- [ ] 每个 service 子目录包含独立的 `routes.go`
- [ ] 业务逻辑与 HTTP 处理分离

---

## Frontend Organization

### Requirement: React Components SHALL Be Organized By Feature

前端组件 SHALL 按功能特性分组，相关组件放在同一目录下。

**Recommended Structure:**
```
web/src/components/
├── AI/                     # AI 助手组件
│   ├── index.ts
│   ├── Copilot.tsx
│   ├── hooks/
│   │   ├── useAIChat.ts
│   │   └── useSSEAdapter.ts
│   └── components/
│       ├── MessageActions.tsx
│       └── ToolCard.tsx
├── Layout/                 # 布局组件
├── Feedback/               # 反馈组件
│   ├── Notification/
│   └── Loading/
├── DataDisplay/            # 数据展示组件
│   ├── Table/
│   └── Charts/
├── Form/                   # 表单组件
└── common/                 # 通用小组件
```

**Acceptance Criteria:**
- [ ] 每个功能目录包含 `index.ts` 导出
- [ ] 复杂组件拆分为子组件目录
- [ ] 目录层级不超过 3 层

---

### Requirement: API Modules SHALL Be Organized By Domain

API 模块 SHALL 按领域分组，避免 `modules/` 目录下堆积过多文件。

**Recommended Structure:**
```
web/src/api/
├── modules/
│   ├── ai/                 # AI 领域 API
│   │   ├── index.ts
│   │   ├── chat.ts
│   │   ├── session.ts
│   │   └── tools.ts
│   ├── infrastructure/     # 基础设施 API
│   │   ├── index.ts
│   │   ├── hosts.ts
│   │   └── clusters.ts
│   ├── deployment/         # 部署管理 API
│   ├── user/               # 用户与权限 API
│   └── observability/      # 可观测性 API
├── api.ts                  # API 客户端
└── types.ts                # 公共类型
```

**Acceptance Criteria:**
- [ ] 每个领域目录文件数不超过 6 个
- [ ] 每个目录包含 `index.ts` 导出所有 API
- [ ] 类型定义与 API 放在同一目录

---

### Requirement: Hooks SHALL Be Organized By Concern

自定义 Hooks SHALL 按关注点分组，相关 hooks 放在同一目录。

**Recommended Structure:**
```
web/src/hooks/
├── data/                   # 数据相关 hooks
│   ├── usePolling.ts
│   └── useRetry.ts
├── ui/                     # UI 相关 hooks
│   ├── useKeyboardShortcuts.ts
│   └── useDebounce.ts
├── notification/           # 通知相关 hooks
│   ├── useNotification.ts
│   └── useNotificationWebSocket.ts
├── auth/                   # 认证相关 hooks
└── index.ts                # 统一导出
```

**Acceptance Criteria:**
- [ ] 每个分组目录文件数不超过 8 个
- [ ] 每个目录包含 `index.ts` 导出
- [ ] 通用 hook 放在根目录

---

### Requirement: Utils SHALL Be Organized By Purpose

工具函数 SHALL 按用途分组，避免 `utils/` 目录杂乱。

**Recommended Structure:**
```
web/src/utils/
├── http/                   # HTTP 相关
│   ├── apiErrorHandler.ts
│   └── requestCache.ts
├── performance/            # 性能相关
│   └── performanceMonitor.ts
├── browser/                # 浏览器相关
│   ├── browserNotification.ts
│   └── tokenManager.ts
├── animation/              # 动画相关
│   └── animationOptimization.ts
└── index.ts                # 统一导出
```

**Acceptance Criteria:**
- [ ] 每个分组目录文件数不超过 6 个
- [ ] 单一职责的工具函数可放根目录
- [ ] 测试文件与源文件放同一目录

---

## Migration Guidelines

### Requirement: Code Reorganization SHALL Be Incremental

代码重组 SHALL 采用增量迁移，每次迁移一个领域，保持代码可编译状态。

**Migration Steps:**
1. 创建目标目录结构
2. 移动文件到新位置
3. 更新所有导入路径
4. 运行测试验证
5. 提交变更

**Acceptance Criteria:**
- [ ] 每次迁移影响范围可控
- [ ] 迁移后测试通过
- [ ] 导入路径使用绝对路径

---

## Enforcement

### Requirement: CI SHALL Check Directory File Count

CI 流程 SHALL 检查目录文件数量，超过阈值时发出警告。

**Check Script Example:**
```bash
#!/bin/bash
# Check directory file count
THRESHOLD=10
find internal/model -maxdepth 1 -type f -name "*.go" | wc -l | \
  xargs -I {} bash -c 'if [ {} -gt $THRESHOLD ]; then echo "Warning: too many files"; exit 1; fi'
```

#### Scenario: CI detects threshold violation

- **GIVEN** a directory contains more files than allowed
- **WHEN** CI runs the organization check
- **THEN** the build SHALL fail with a clear message indicating which directory needs reorganization
