# Design: 服务管理职责边界重构

## Context

### 当前状态

服务管理模块存在职责边界模糊的问题，主要体现在：

1. **数据模型层面**: `Service` 表包含 `default_target_id` 和 `default_deployment_target_id` 字段，试图在服务配置中存储部署目标
2. **UI 层面**: 服务详情页的"配置"Tab 包含"部署目标"卡片，混淆了配置与部署的边界
3. **交互层面**: 服务列表页的"编辑配置"按钮只显示提示，实际功能缺失

### 涉及模块

| 模块 | 路径 | 变更类型 |
|------|------|----------|
| Cluster Model | `internal/model/` | 新增字段 |
| Service Model | `internal/model/project.go` | 字段废弃 |
| 迁移 | `storage/migrations/` | 新增迁移 |
| 服务详情页 | `web/src/pages/Services/ServiceDetailPage.tsx` | UI 重构 |
| 服务列表页 | `web/src/pages/Services/ServiceListPage.tsx` | 交互修复 |
| 服务 API | `internal/service/service_studio/` | 校验逻辑 |

### 约束

- 不改变现有的部署 API 路径和参数结构
- 保持向后兼容，`default_target_id` 字段暂时保留但不再使用
- 前端改动需保持现有功能不变

## Goals / Non-Goals

**Goals:**

1. 给 Cluster 添加 `env_type` 字段，用于部署时的环境匹配约束
2. 重构服务详情页"配置"Tab，移除"部署目标"卡片
3. 将"编辑配置"Modal 改为"配置"Tab 的内联编辑
4. 修复服务列表页"编辑配置"按钮的交互

**Non-Goals:**

1. 部署流程的具体交互设计（部署弹窗、集群选择器等）
2. CI/CD 功能的设计和实现
3. 日志和监控功能的开发
4. 删除 `ServiceDeployTarget` 表（保留供历史数据查询）

## Decisions

### Decision 1: Cluster.env_type 字段设计

**决定**: 在 Cluster 表添加 `env_type VARCHAR(32)` 字段

```sql
ALTER TABLE clusters ADD COLUMN env_type VARCHAR(32) NOT NULL DEFAULT 'development';
```

**值域**: `development | staging | production`

**理由**:
- 集群按环境隔离是运维的基本实践
- 通过数据库约束可以确保部署时只选择匹配环境的集群
- 比 `ClusterNamespaceBinding.env` 更直接，因为一个集群通常只服务于一个环境

**替代方案考虑**:

| 方案 | 优点 | 缺点 |
|------|------|------|
| A. Cluster.env_type (选用) | 简单直接，查询高效 | 集群只能属于一个环境 |
| B. ClusterNamespaceBinding.env | 灵活，同集群可多环境 | 查询复杂，部署校验繁琐 |
| C. 独立 Environment 表 | 更完整的模型 | 过度设计，当前场景不需要 |

### Decision 2: Service.env 与 Cluster.env_type 的关系

**决定**: 两者独立，但在部署时校验匹配

```
Service.env (模板引擎用) ──┐
                          ├──▶ 部署时校验匹配
Cluster.env_type (约束用) ─┘
```

**理由**:
- `Service.env` 已经存在，用于模板引擎变量替换，不应改变其语义
- `Cluster.env_type` 是新增的约束字段，职责清晰
- 部署时检查 `Service.env == Cluster.env_type`，不匹配则拒绝

**校验时机**: 在 `POST /services/:id/deploy` 接口中校验

### Decision 3: 配置 Tab 内联编辑

**决定**: 将"编辑配置"Modal 的内容直接展示在"配置"Tab 中，支持内联编辑

**理由**:
- 减少交互层级，用户不需要打开 Modal 才能编辑
- 配置内容较多，Modal 中滚动体验不佳
- 符合现代 SaaS 产品的设计趋势

**实现方式**:
- 每个 Card 内使用 `Form` 组件，默认为只读状态
- 点击"编辑"按钮切换为可编辑状态
- 保存后切换回只读状态

### Decision 4: 服务列表"编辑配置"交互

**决定**: 点击"编辑配置"跳转到详情页并激活"配置"Tab

**理由**:
- 编辑配置需要展示完整信息，列表页空间不足
- 复用详情页的配置 Tab，避免代码重复
- URL 参数 `?tab=config` 可直接定位

**替代方案考虑**:

| 方案 | 优点 | 缺点 |
|------|------|------|
| A. 跳转详情页 (选用) | 复用现有组件，信息完整 | 多一次页面跳转 |
| B. 列表页弹 Modal | 操作快 | Modal 内容复杂，体验差 |
| C. 行内快速编辑 | 最快 | 只能编辑少量字段 |

## Risks / Trade-offs

### Risk 1: 现有集群数据迁移

**风险**: 现有集群没有 `env_type`，默认值 `development` 可能不正确

**缓解措施**:
- 迁移后提供管理界面修改 `env_type`
- 在 UI 上明确标注未设置的集群，提示管理员配置

### Risk 2: 部署 API 兼容性

**风险**: 现有调用方可能依赖 `default_target_id`

**缓解措施**:
- 保留字段但不使用，API 文档标注废弃
- 部署接口参数不变，但 `cluster_id` 变为必填

### Risk 3: 用户习惯改变

**风险**: 用户习惯了"预设部署目标"的工作流

**缓解措施**:
- 部署弹窗记住上次选择的集群/命名空间 (localStorage)
- 在文档中说明新的工作流

## Migration Plan

### Phase 1: 数据模型变更 (后端)

1. 创建迁移文件 `storage/migrations/YYYYMMDD_0000xx_add_cluster_env_type.sql`
2. 更新 `Cluster` model 添加 `EnvType` 字段
3. 部署并执行迁移

### Phase 2: UI 重构 (前端)

1. 重构 `ServiceDetailPage.tsx`:
   - 删除"部署目标"卡片
   - 将配置内容改为内联编辑
2. 修复 `ServiceListPage.tsx`:
   - "编辑配置"改为跳转链接

### Phase 3: 部署校验 (后端)

1. 在部署接口中添加环境匹配校验
2. 更新 API 文档

### Rollback Plan

- 迁移回滚: 删除 `env_type` 列
- 前端回滚: 恢复"部署目标"卡片和编辑 Modal
- API 回滚: 移除环境匹配校验
