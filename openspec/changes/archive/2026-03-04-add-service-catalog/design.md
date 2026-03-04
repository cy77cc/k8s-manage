# Design: 服务目录 (Service Catalog)

## Context

### 当前状态

系统已具备以下可复用能力：

- **模板变量系统** (`internal/service/service/template_vars.go`): 支持 `{{ var_name }}` 和 `{{ var_name|default:value }}` 语法
- **Helm 部署能力** (`/services/:id/deploy/helm`): Helm Chart 部署
- **K8s 部署流程** (`internal/service/service/logic_deploy.go`): K8s YAML 部署
- **RBAC 权限系统** (`internal/service/rbac/`): 权限控制基础

缺失：
- 服务目录入口和页面
- 模板存储和管理
- 审核发布工作流
- 分类管理

### 约束

- 复用现有模板变量语法，保持一致性
- 复用现有部署流程，不重复实现
- 新模块遵循项目架构规范 (routes/handler/logic 分层)
- 支持单进程交付模式
- **不使用数据库外键**: 所有表间关联关系由业务层处理，不依赖数据库外键约束

## Goals / Non-Goals

**Goals:**
- 用户可浏览服务目录，按分类筛选模板
- 用户可从模板一键部署到 K8s 或 Compose 环境
- 用户可创建私有模板，提交审核后发布到公共目录
- 管理员可管理分类，审核模板发布
- 模板支持变量定义，部署时动态填充

**Non-Goals:**
- 不实现 Helm Chart 在线编辑器 (使用 YAML 编辑器即可)
- 不实现模板版本历史对比 (MVP 阶段)
- 不实现模板市场/付费功能
- 不实现跨集群模板同步

## Decisions

### D1: 模板数据模型设计

**决策**: 使用双表设计，`service_categories` 存储分类，`service_templates` 存储模板

**理由**:
- 分类和模板分离，便于独立管理
- 支持分类的层级扩展（预留 parent_id 字段）
- 模板可关联多个标签，便于搜索
- 表间关联由业务层处理，不使用数据库外键

**备选方案**:
- ❌ 单表 JSON 存储: 查询性能差，难以索引
- ❌ 纯文件系统存储: 无法做权限控制和审核流程
- ❌ 使用数据库外键: 影响数据库性能和扩展性，本项目统一由业务层处理关联

**数据模型**:

```sql
-- service_categories 表
CREATE TABLE service_categories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,           -- 分类标识 (database/cache/...)
    display_name VARCHAR(128) NOT NULL,  -- 显示名称
    icon VARCHAR(256),                   -- 图标
    description TEXT,                    -- 描述
    sort_order INT DEFAULT 0,            -- 排序
    is_system BOOLEAN DEFAULT FALSE,     -- 是否系统预置 (不可删除)
    created_by BIGINT UNSIGNED,          -- 创建者 (null 表示系统预置)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_name (name)
);

-- service_templates 表
CREATE TABLE service_templates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,          -- 模板标识 (mysql-single)
    display_name VARCHAR(256) NOT NULL,  -- 显示名称
    description TEXT,                    -- 描述
    icon VARCHAR(256),                   -- 图标
    category_id BIGINT UNSIGNED NOT NULL,-- 分类 ID
    version VARCHAR(32) DEFAULT '1.0.0', -- 模板版本
    owner_id BIGINT UNSIGNED NOT NULL,   -- 创建者 ID

    -- 可见性和状态
    visibility ENUM('private', 'public') DEFAULT 'private',
    status ENUM('draft', 'pending_review', 'published', 'rejected') DEFAULT 'draft',

    -- 模板内容
    k8s_template MEDIUMTEXT,             -- K8s YAML 模板
    compose_template MEDIUMTEXT,         -- Docker Compose YAML 模板
    variables_schema JSON,               -- 变量定义 [{name, type, default, required, description}]
    readme TEXT,                         -- 使用说明 Markdown

    -- 元数据
    tags JSON,                           -- 标签 ["mysql", "database"]
    deploy_count INT DEFAULT 0,          -- 部署次数统计

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_category (category_id),
    INDEX idx_owner (owner_id),
    INDEX idx_status (status),
    INDEX idx_visibility_status (visibility, status)
    -- 关联关系由业务层处理，不使用数据库外键
);
```

### D2: 变量系统设计

**决策**: 复用现有 `{{ var_name|default:value }}` 语法，扩展变量类型定义

**变量类型**:
| 类型 | 说明 | 前端组件 |
|------|------|----------|
| `string` | 字符串 | Input |
| `number` | 数字 | InputNumber |
| `password` | 密码 | Input.Password |
| `boolean` | 布尔 | Switch |
| `select` | 枚举 | Select |
| `textarea` | 多行文本 | Input.TextArea |

**variables_schema 示例**:
```json
[
  {
    "name": "mysql_version",
    "type": "select",
    "default": "8.0",
    "required": false,
    "description": "MySQL 版本",
    "options": ["5.7", "8.0", "8.1"]
  },
  {
    "name": "root_password",
    "type": "password",
    "default": "",
    "required": true,
    "description": "root 用户密码"
  }
]
```

### D3: 审核发布工作流

**决策**: 简化的单步审核流程

```
draft → pending_review → published
  │           │
  │           └── rejected
  └── private (用户自用)
```

**状态转换规则**:
| 当前状态 | 允许操作 | 目标状态 | 权限 |
|---------|---------|---------|------|
| draft | submit | pending_review | 作者 |
| draft | save | draft | 作者 |
| pending_review | publish | published | 管理员 |
| pending_review | reject | rejected | 管理员 |
| rejected | resubmit | pending_review | 作者 |
| published | unpublish | draft | 管理员 |

### D4: 部署流程设计

**决策**: 复用现有服务部署能力，新增 catalog → service 转换层

```
┌─────────────────────────────────────────────────────────────────────┐
│                        部署流程                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. 用户选择模板 + 填写变量值                                        │
│  2. 系统渲染 YAML (变量替换)                                         │
│  3. 用户预览渲染结果                                                 │
│  4. 用户确认部署                                                     │
│  5. 系统创建 Service 记录 (复用现有服务模型)                         │
│  6. 系统调用现有部署 API (K8s/Compose)                              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**关键**: 不新建部署逻辑，而是将模板部署转换为标准服务部署

### D5: 权限设计

**决策**: 使用 Casbin 策略控制关键操作

| 操作 | 权限码 | 说明 |
|------|--------|------|
| 浏览公共模板 | `catalog:read` | 所有认证用户 |
| 创建模板 | `catalog:write` | 所有认证用户 |
| 审核发布 | `catalog:approve` | 管理员 |
| 管理分类 | `catalog:manage` | 管理员 |

**模板编辑权限**: 仅作者和管理员可编辑

### D6: 预置分类

**决策**: 系统启动时自动初始化预置分类

| name | display_name | icon |
|------|-------------|------|
| database | 数据库 | DatabaseOutlined |
| cache | 缓存 | ThunderboltOutlined |
| message-queue | 消息队列 | MessageOutlined |
| web-server | Web 服务器 | GlobalOutlined |
| monitoring | 监控告警 | MonitorOutlined |
| dev-tools | 开发工具 | CodeOutlined |
| custom | 自定义服务 | AppstoreOutlined |

## Risks / Trade-offs

### R1: 模板安全性

**风险**: 恶意用户可能提交包含危险命令的模板

**缓解**:
- 模板发布需管理员审核
- 可选: 部署前 YAML 语法校验
- 可选: 敏感操作拦截 (如 `rm -rf`, `exec`)

### R2: 变量复杂度

**风险**: 复杂变量嵌套可能导致渲染失败

**缓解**:
- 限制变量嵌套层级 (最多 1 层)
- 前端实时预览渲染结果
- 后端渲染失败时返回明确错误

### R3: 部署目标一致性

**风险**: K8s 和 Compose 模板可能不一致

**缓解**:
- 前端提示用户选择部署目标
- 模板可只定义 K8s 或 Compose 其中之一
- 变量定义共用，模板内容分开

## Migration Plan

### 部署步骤

1. 执行数据库迁移 (创建 service_categories, service_templates 表)
2. 启动服务时自动初始化预置分类
3. 前端部署新页面
4. 配置导航菜单入口

### 回滚策略

1. 删除数据库表
2. 移除导航菜单入口
3. 回滚前端代码

### 数据迁移

无需迁移现有数据 (新增功能)

## API Contract

### 分类 API

```
GET    /api/v1/catalog/categories          # 获取分类列表
POST   /api/v1/catalog/categories          # 创建分类 (管理员)
PUT    /api/v1/catalog/categories/:id      # 更新分类
DELETE /api/v1/catalog/categories/:id      # 删除分类 (非系统预置)
```

### 模板 API

```
GET    /api/v1/catalog/templates           # 获取模板列表 (支持 category_id, status 筛选)
GET    /api/v1/catalog/templates/:id       # 获取模板详情
POST   /api/v1/catalog/templates           # 创建模板
PUT    /api/v1/catalog/templates/:id       # 更新模板
DELETE /api/v1/catalog/templates/:id       # 删除模板

POST   /api/v1/catalog/templates/:id/submit   # 提交审核
POST   /api/v1/catalog/templates/:id/publish  # 发布 (管理员)
POST   /api/v1/catalog/templates/:id/reject   # 驳回 (管理员)
```

### 部署 API

```
POST   /api/v1/catalog/preview             # 预览渲染后的 YAML
POST   /api/v1/catalog/deploy              # 从模板部署服务
```

## File Changes Summary

```
后端新增:
├── api/catalog/v1/catalog.go           # API 类型定义
└── internal/service/catalog/
    ├── routes.go                        # 路由注册
    ├── handler.go                       # HTTP Handler
    ├── logic.go                         # 业务逻辑
    └── types.go                         # 内部类型

前端新增:
├── web/src/api/modules/catalog.ts      # API 调用
└── web/src/pages/Catalog/
    ├── CatalogListPage.tsx              # 服务目录首页
    ├── CatalogDetailPage.tsx            # 模板详情
    ├── CatalogDeployPage.tsx            # 部署向导
    ├── TemplateListPage.tsx             # 我的模板
    ├── TemplateCreatePage.tsx           # 创建模板
    ├── TemplateEditPage.tsx             # 编辑模板
    ├── ReviewListPage.tsx               # 审核列表
    └── CategoryManagePage.tsx           # 分类管理

数据库迁移:
└── storage/migrations/YYYYMMDD_XXXXXX_service_catalog.sql
```
