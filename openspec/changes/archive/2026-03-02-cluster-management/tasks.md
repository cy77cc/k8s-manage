# Tasks: 集群管理功能

## Phase 1: 集群创建基础 (MVP)

### 1.1 数据模型扩展

- [x] **T1.1.1** 创建 `cluster_nodes` 表迁移文件
  - 定义 ClusterNode model
  - 添加 GORM 标签和索引
  - 编写迁移 SQL

- [x] **T1.1.2** 扩展 `clusters` 表
  - 添加 source, credential_id, k8s_version, pod_cidr, service_cidr, last_sync_at 字段
  - 更新 Cluster model

- [x] **T1.1.3** 扩展 `cluster_bootstrap_tasks` 表
  - 添加 cluster_id, k8s_version, pod_cidr, service_cidr, steps_json 字段
  - 更新 ClusterBootstrapTask model

### 1.2 脚本文件

- [x] **T1.2.1** 创建通用脚本目录 `script/cluster/common/`
  - `preflight.sh` - 系统预检查
  - `containerd-install.sh` - 安装 containerd
  - `sysctl-config.sh` - 系统参数配置

- [x] **T1.2.2** 创建 kubeadm 脚本目录 `script/cluster/kubeadm/v1.28/`
  - `manifest.json` - 版本清单
  - `install.sh` - 安装 kubeadm/kubelet/kubectl
  - `init.sh` - kubeadm init
  - `join.sh` - kubeadm join
  - `reset.sh` - kubeadm reset
  - `fetch-kubeconfig.sh` - 获取 kubeconfig

- [x] **T1.2.3** 创建 CNI 脚本目录 `script/cluster/cni/`
  - `calico/v3.26/install.sh` - 安装 Calico
  - `flannel/v0.22/install.sh` - 安装 Flannel
  - `cilium/v1.14/install.sh` - 安装 Cilium

### 1.3 后端 API - 自建集群

- [x] **T1.3.1** 重构 `internal/service/cluster/` 模块
  - 替换 Mock 数据为真实数据库查询
  - 实现 GetClusters 从数据库读取
  - 实现 GetClusterDetail 从数据库读取
  - 实现 GetClusterNodes 查询

- [x] **T1.3.2** 实现 Bootstrap Workflow
  - 定义 BootstrapStep 结构
  - 实现 ExecuteBootstrap 主流程
  - 实现 executeStep 单步执行
  - 实现 executeRollback 回滚逻辑
  - 实现步骤状态记录

- [x] **T1.3.3** 实现 Bootstrap API handlers
  - POST /clusters/bootstrap/preview
  - POST /clusters/bootstrap/apply
  - GET /clusters/bootstrap/:taskId
  - POST /clusters/bootstrap/:taskId/cancel

- [x] **T1.3.4** 实现脚本执行引擎
  - 加载脚本文件
  - 构建执行环境变量
  - 通过 SSH 执行脚本
  - 解析脚本输出

### 1.4 后端 API - 导入集群

- [x] **T1.4.1** 实现导入逻辑
  - 解析 kubeconfig
  - 验证连接有效性
  - 创建 Cluster 和 Credential 记录
  - 同步节点信息

- [x] **T1.4.2** 实现导入 API handlers
  - POST /clusters/import
  - POST /clusters/import/validate

### 1.5 前端页面

- [x] **T1.5.1** 改进 ClusterListPage
  - 从真实 API 获取数据
  - 添加创建入口（自建/导入选择）

- [x] **T1.5.2** 改进 ClusterBootstrapWizard
  - 添加网络配置步骤
  - 添加执行进度展示
  - 添加错误处理和重试

- [x] **T1.5.3** 新建 ClusterImportWizard
  - kubeconfig 上传
  - 连接测试
  - 确认导入

## Phase 2: 集群管理

### 2.1 集群 CRUD

- [x] **T2.1.1** 实现集群更新 API
  - PUT /clusters/:id
  - 更新名称、描述、标签

- [x] **T2.1.2** 实现集群删除 API
  - DELETE /clusters/:id
  - 检查关联资源
  - 清理凭证

- [x] **T2.1.3** 实现连接测试 API
  - POST /clusters/:id/test
  - 返回连接状态和延迟

### 2.2 节点管理

- [x] **T2.2.1** 实现节点列表 API
  - GET /clusters/:id/nodes
  - 实时查询 K8s API
  - 同步到 cluster_nodes 表

- [x] **T2.2.2** 实现添加节点 API
  - POST /clusters/:id/nodes
  - 执行 kubeadm join
  - 更新节点状态

- [x] **T2.2.3** 实现移除节点 API
  - DELETE /clusters/:id/nodes/:name
  - 驱逐 Pod
  - 执行 kubeadm reset

### 2.3 前端页面

- [x] **T2.3.1** 改进 ClusterDetailPage
  - 基本信息展示
  - 状态指示器
  - 操作按钮

- [x] **T2.3.2** 实现节点管理 Tab
  - 节点列表表格
  - 节点详情抽屉
  - 添加/移除节点

## Phase 3: 资源查看

### 3.1 Namespace 管理

- [x] **T3.1.1** 实现 Namespace API
  - GET /clusters/:id/namespaces
  - POST /clusters/:id/namespaces
  - DELETE /clusters/:id/namespaces/:name

### 3.2 工作负载查看

- [x] **T3.2.1** 实现工作负载查询 API
  - GET /clusters/:id/namespaces/:ns/deployments
  - GET /clusters/:id/namespaces/:ns/statefulsets
  - GET /clusters/:id/namespaces/:ns/daemonsets
  - GET /clusters/:id/namespaces/:ns/jobs
  - GET /clusters/:id/namespaces/:ns/cronjobs
  - GET /clusters/:id/namespaces/:ns/pods

### 3.3 服务和配置查看

- [x] **T3.3.1** 实现服务查询 API
  - GET /clusters/:id/namespaces/:ns/services
  - GET /clusters/:id/namespaces/:ns/ingresses

- [x] **T3.3.2** 实现配置查询 API
  - GET /clusters/:id/namespaces/:ns/configmaps
  - GET /clusters/:id/namespaces/:ns/secrets

### 3.4 存储查看

- [x] **T3.4.1** 实现存储查询 API
  - GET /clusters/:id/pvs
  - GET /clusters/:id/namespaces/:ns/pvcs

### 3.5 服务部署视角

- [x] **T3.5.1** 实现集群服务列表 API
  - GET /clusters/:id/services
  - 关联 deployment 表查询

- [x] **T3.5.2** 实现前端页面
  - 工作负载 Tab
  - 服务 Tab
  - 配置 Tab
  - 存储 Tab

## Phase 4: 高级运维

### 4.1 集群升级

- [x] **T4.1.1** 实现升级预览 API
  - POST /clusters/:id/upgrade/preview
  - 返回升级计划

- [x] **T4.1.2** 实现升级执行 API
  - POST /clusters/:id/upgrade/apply
  - 滚动升级节点

### 4.2 证书管理

- [x] **T4.2.1** 实现证书查看 API
  - GET /clusters/:id/certificates
  - 返回证书到期时间

- [x] **T4.2.2** 实现证书更新 API
  - POST /clusters/:id/certificates/renew

### 4.3 事件查看

- [x] **T4.3.1** 实现事件查询 API
  - GET /clusters/:id/events
  - 支持过滤和分页

## Dependencies

- Phase 2 依赖 Phase 1 完成
- Phase 3 依赖 Phase 2 完成
- Phase 4 依赖 Phase 3 完成

## Notes

- 所有 SSH 操作需要处理超时和重试
- K8s API 查询需要缓存 client 连接
- 脚本执行需要详细日志记录
- 前端需要处理长时间操作的用户反馈
