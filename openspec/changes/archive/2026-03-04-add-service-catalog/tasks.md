# Tasks: 服务目录 (Service Catalog)

## 1. 数据库迁移

- [x] 1.1 创建 `storage/migrations/YYYYMMDD_XXXXXX_service_catalog.sql`
  - 创建 `service_categories` 表
  - 创建 `service_templates` 表
  - 包含 Up/Down 迁移脚本
  - **注意**: 不使用数据库外键，表间关联由业务层处理

## 2. 后端 Model 层

- [x] 2.1 创建 `internal/model/service_category.go`
  - 定义 ServiceCategory 结构体
  - 定义 CRUD 方法

- [x] 2.2 创建 `internal/model/service_template.go`
  - 定义 ServiceTemplate 结构体
  - 定义 CRUD 方法
  - 定义状态枚举常量

## 3. 后端 API 类型定义

- [x] 3.1 创建 `api/catalog/v1/catalog.go`
  - 定义 Category 相关类型 (CategoryCreateRequest, CategoryResponse, etc.)
  - 定义 Template 相关类型 (TemplateCreateRequest, TemplateResponse, etc.)
  - 定义 Deploy 相关类型 (DeployRequest, PreviewRequest, etc.)

## 4. 后端 Logic 层

- [x] 4.1 创建 `internal/service/catalog/logic.go`
  - 实现 Category 相关业务逻辑 (ListCategories, CreateCategory, UpdateCategory, DeleteCategory)
  - 实现 Template 相关业务逻辑 (ListTemplates, GetTemplate, CreateTemplate, UpdateTemplate, DeleteTemplate)
  - 实现审核流程逻辑 (SubmitForReview, PublishTemplate, RejectTemplate)
  - 实现预览逻辑 (PreviewYAML)
  - 实现部署逻辑 (DeployFromTemplate)

- [x] 4.2 实现变量渲染
  - 复用 `internal/service/service/template_vars.go` 的变量检测和渲染
  - 支持变量类型验证

- [x] 4.3 实现部署计数
  - 部署成功后递增 `deploy_count`

## 5. 后端 Handler 层

- [x] 5.1 创建 `internal/service/catalog/handler.go`
  - 实现分类 Handler (ListCategories, CreateCategory, UpdateCategory, DeleteCategory)
  - 实现模板 Handler (ListTemplates, GetTemplate, CreateTemplate, UpdateTemplate, DeleteTemplate)
  - 实现审核 Handler (SubmitForReview, PublishTemplate, RejectTemplate)
  - 实现部署 Handler (Preview, Deploy)

## 6. 后端路由注册

- [x] 6.1 创建 `internal/service/catalog/routes.go`
  - 注册分类 API 路由
  - 注册模板 API 路由
  - 注册部署 API 路由
  - 应用 JWT 认证中间件
  - 应用 Casbin 权限检查

- [x] 6.2 在 `internal/service/service.go` 中注册 catalog 模块

## 7. 后端权限配置

- [x] 7.1 添加 Casbin 策略
  - `catalog:read` - 浏览模板
  - `catalog:write` - 创建/编辑模板
  - `catalog:approve` - 审核发布
  - `catalog:manage` - 分类管理

## 8. 后端预置数据初始化

- [x] 8.1 创建预置分类初始化逻辑
  - database (数据库)
  - cache (缓存)
  - message-queue (消息队列)
  - web-server (Web 服务器)
  - monitoring (监控告警)
  - dev-tools (开发工具)
  - custom (自定义服务)

## 9. 前端 API 模块

- [x] 9.1 创建 `web/src/api/modules/catalog.ts`
  - 定义 TypeScript 类型
  - 实现 Category API 调用
  - 实现 Template API 调用
  - 实现 Deploy API 调用

## 10. 前端页面 - 服务目录

- [x] 10.1 创建 `web/src/pages/Catalog/CatalogListPage.tsx`
  - 实现分类侧边栏
  - 实现模板卡片列表
  - 实现搜索功能
  - 实现分类筛选

- [x] 10.2 创建 `web/src/pages/Catalog/CatalogDetailPage.tsx`
  - 显示模板详情
  - 显示变量定义
  - 显示使用说明
  - 提供"部署"按钮入口

## 11. 前端页面 - 部署向导

- [x] 11.1 创建 `web/src/pages/Catalog/CatalogDeployPage.tsx`
  - 实现部署目标选择 (K8s 集群 / Compose 环境)
  - 实现变量输入表单 (根据 variables_schema 动态生成)
  - 实现 YAML 预览
  - 实现部署确认和执行

- [x] 11.2 实现变量类型组件映射
  - string → Input
  - number → InputNumber
  - password → Input.Password
  - boolean → Switch
  - select → Select
  - textarea → Input.TextArea

## 12. 前端页面 - 模板管理

- [x] 12.1 创建 `web/src/pages/Catalog/TemplateListPage.tsx`
  - 显示"我的模板"列表
  - 支持按状态筛选 (draft/pending_review/published/rejected)
  - 提供编辑/删除操作

- [x] 12.2 创建 `web/src/pages/Catalog/TemplateCreatePage.tsx`
  - 实现 Step 1: 基本信息 (名称、描述、分类、图标)
  - 实现 Step 2: K8s 模板编辑器 (Monaco Editor)
  - 实现 Step 3: Compose 模板编辑器 (可选)
  - 实现 Step 4: 变量定义 (自动检测 + 手动添加)
  - 实现 Step 5: 使用说明 (Markdown 编辑器)
  - 提供预览和保存功能

- [x] 12.3 创建 `web/src/pages/Catalog/TemplateEditPage.tsx`
  - 预填充现有模板内容
  - 复用创建页面的表单组件

## 13. 前端页面 - 审核管理

- [x] 13.1 创建 `web/src/pages/Catalog/ReviewListPage.tsx`
  - 显示待审核模板列表
  - 显示提交者和提交时间
  - 提供查看详情入口

- [x] 13.2 创建审核操作组件
  - 实现发布/驳回按钮
  - 实现驳回原因输入

## 14. 前端页面 - 分类管理

- [x] 14.1 创建 `web/src/pages/Catalog/CategoryManagePage.tsx`
  - 显示分类列表
  - 实现创建/编辑分类
  - 实现删除分类 (仅非系统预置)
  - 实现排序调整

## 15. 前端路由和导航

- [x] 15.1 添加路由配置
  - `/catalog` - 服务目录
  - `/catalog/:id` - 模板详情
  - `/catalog/:id/deploy` - 部署向导
  - `/catalog/my-templates` - 我的模板
  - `/catalog/templates/create` - 创建模板
  - `/catalog/templates/:id/edit` - 编辑模板
  - `/catalog/reviews` - 审核管理
  - `/catalog/categories` - 分类管理

- [x] 15.2 添加导航菜单入口
  - 在侧边栏添加"服务目录"菜单项

## 16. 测试

- [x] 16.1 后端单元测试
  - `internal/service/catalog/logic_test.go`
  - 测试变量渲染
  - 测试审核流程状态转换

- [x] 16.2 前端组件测试
  - CatalogListPage 组件测试
  - CatalogDeployPage 组件测试

## 17. 文档和验证

- [x] 17.1 更新 API 文档
- [x] 17.2 运行 `openspec validate` 确保变更规范
