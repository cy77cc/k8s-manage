# Proposal: 服务目录 (Service Catalog)

## Why

README.md 承诺的"服务目录 (Service Catalog)"功能完全缺失。用户无法通过预置模板一键部署常用中间件 (MySQL/Redis/Kafka)，也无法自助创建和管理服务模板。这是 PaaS 平台"开发者门户"的核心功能，直接影响用户对平台价值的感知。

## What Changes

### 新增功能

- **服务目录浏览**: 用户可浏览、搜索、筛选预置和公共模板
- **模板部署**: 从模板一键部署服务到 K8s 集群或 Docker Compose 环境
- **模板创建**: 用户可创建私有模板，定义 K8s YAML 和/或 Compose YAML
- **变量系统**: 模板支持 `{{ var_name }}` 变量语法，部署时动态填充
- **审核发布**: 私有模板可提交审核，管理员审批后发布到公共目录
- **分类管理**: 系统预置分类 (数据库/缓存/消息队列等)，用户可自定义扩展

### 数据模型

- `service_categories` 表: 分类管理 (预置 + 自定义)
- `service_templates` 表: 模板存储 (含 K8s/Compose 模板、变量定义)

### API 端点

- `GET/POST /api/v1/catalog/categories` - 分类 CRUD
- `GET/POST /api/v1/catalog/templates` - 模板列表/创建
- `GET/PUT/DELETE /api/v1/catalog/templates/:id` - 模板详情/更新/删除
- `POST /api/v1/catalog/templates/:id/submit` - 提交审核
- `POST /api/v1/catalog/templates/:id/publish` - 发布 (管理员)
- `POST /api/v1/catalog/templates/:id/reject` - 驳回 (管理员)
- `POST /api/v1/catalog/deploy` - 从模板部署服务
- `POST /api/v1/catalog/preview` - 预览渲染后的 YAML

## Capabilities

### New Capabilities

- `service-catalog-browse`: 服务目录浏览、搜索、分类筛选功能
- `service-catalog-deploy`: 从模板部署服务到 K8s/Compose 环境
- `service-template-management`: 模板创建、编辑、版本管理功能
- `service-template-review`: 模板审核发布工作流 (私有 → 审核 → 发布)
- `service-category-management`: 分类管理 (系统预置 + 用户自定义)

### Modified Capabilities

无现有 spec 需要修改。

## Impact

### 后端新增

- `internal/service/catalog/` 模块
  - `routes.go` - 路由注册
  - `handler.go` - HTTP Handler
  - `logic.go` - 业务逻辑
- `api/catalog/v1/catalog.go` - API 类型定义

### 前端新增

- `web/src/api/modules/catalog.ts` - API 调用
- `web/src/pages/Catalog/` 页面
  - `CatalogListPage.tsx` - 服务目录首页
  - `CatalogDetailPage.tsx` - 模板详情 + 部署向导
  - `CatalogDeployPage.tsx` - 部署配置页面
  - `TemplateListPage.tsx` - 我的模板列表
  - `TemplateCreatePage.tsx` - 创建模板
  - `TemplateEditPage.tsx` - 编辑模板
  - `ReviewListPage.tsx` - 审核列表 (管理员)
  - `CategoryManagePage.tsx` - 分类管理 (管理员)

### 数据库迁移

- `storage/migrations/YYYYMMDD_XXXXXX_service_catalog.sql`
  - `service_categories` 表
  - `service_templates` 表

### 依赖

- 复用现有模板变量系统 (`internal/service/service/template_vars.go`)
- 复用现有 Helm 部署能力 (`/services/:id/deploy/helm`)
- 复用现有 K8s/Compose 部署流程
