# Tasks: 服务管理职责边界重构

## 1. 数据模型变更 (后端)

- [x] 1.1 创建迁移文件 `storage/migrations/YYYYMMDD_0000xx_add_cluster_env_type.sql`
  - 添加 `env_type VARCHAR(32) NOT NULL DEFAULT 'development'` 列
  - 添加索引 `idx_cluster_env_type`
  - 编写 Down 回滚 (删除列和索引)

- [x] 1.2 更新 Cluster model (`internal/model/`)
  - 添加 `EnvType string` 字段
  - 更新 JSON tag 为 `env_type`

- [x] 1.3 执行数据库迁移
  - 运行 `migrate -path storage/migrations -database "mysql://..." up`

## 2. UI 重构 - 服务详情页 (前端)

- [x] 2.1 重构 ServiceDetailPage.tsx 配置 Tab
  - 删除"部署目标"卡片及相关表单
  - 删除 `targetForm` 和 `saveDeployTarget` 函数

- [x] 2.2 实现配置内联编辑
  - 将"编辑配置"Modal 内容移入配置 Tab
  - 添加编辑状态管理 (`editing: boolean`)
  - 每个配置卡片添加"编辑"/"保存"/"取消"按钮
  - 保存成功后切换回只读模式

- [x] 2.3 更新配置 Tab 结构
  - 服务配置卡片 (基本信息 + 运行时配置 + 标签 + 配置内容)
  - 环境变量集卡片 (保留现有)
  - 渲染输出预览 (保留现有)

## 3. UI 修复 - 服务列表页 (前端)

- [x] 3.1 修复 ServiceListPage.tsx "编辑配置"按钮
  - 修改 dropdown menu 的 `onClick` 处理
  - 将 `message.info` 改为 `navigate('/services/${service.id}?tab=config')`

- [x] 3.2 实现详情页 Tab 参数解析
  - 解析 URL 参数 `?tab=config`
  - 初始化时激活对应的 Tab

## 4. 部署校验逻辑 (后端)

- [x] 4.1 添加部署环境匹配校验
  - 在 `POST /services/:id/deploy` handler 中添加校验
  - 查询 Service.env 和 Cluster.env_type
  - 不匹配时返回 `400 ENV_MISMATCH` 错误

- [x] 4.2 更新部署 API 文档
  - 标注 `cluster_id` 为必填参数
  - 说明 `ENV_MISMATCH` 错误码

## 5. 清理与测试

- [x] 5.1 移除废弃的前端 API 调用
  - 删除 `upsertDeployTarget` 相关调用 (保留 API 定义供其他用途)

- [x] 5.2 更新单元测试
  - 添加 Cluster.env_type 字段测试
  - 添加部署环境匹配校验测试

- [x] 5.3 手动测试验证
  - 验证配置 Tab 内联编辑功能
  - 验证"编辑配置"导航功能
  - 验证部署环境匹配校验
