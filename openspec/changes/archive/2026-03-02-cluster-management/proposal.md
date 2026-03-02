# Proposal: 集群管理功能实现

## Summary

实现完整的 Kubernetes 集群管理功能，支持两种创建方式（自建/导入），提供集群全生命周期管理能力，包括节点管理、资源查看、策略配置等运维功能。

## Motivation

当前集群管理功能存在以下问题：
1. **后端返回 Mock 数据**：`/clusters` 接口返回硬编码数据，没有真实数据库交互
2. **创建流程未完成**：`ApplyClusterBootstrap` 只创建记录，未执行实际的 kubeadm 安装
3. **缺少管理功能**：无法管理集群（编辑、删除、节点操作）
4. **缺少资源查看**：无法查看集群部署的服务、工作负载、配置等
5. **主机未联动**：主机与集群的关联关系未激活

作为 PaaS 平台，需要提供完整的集群管理能力，支持开发和运维两种视角。

## Goals

### Phase 1: 集群创建基础 (MVP)

1. **自建集群自动化**
   - 通过 SSH 在选定主机上执行 kubeadm init/join
   - 安装 containerd、kubelet、kubectl
   - 安装 CNI 插件（Calico、Flannel、Cilium）
   - 自动采集并存储 kubeconfig

2. **导入外部集群**
   - 通过 kubeconfig 导入已存在的集群
   - 验证 kubeconfig 有效性
   - 测试集群连接

3. **数据模型扩展**
   - 新增 `cluster_nodes` 表存储集群节点信息
   - 扩展现有表字段

### Phase 2: 集群管理

1. **集群 CRUD**
   - 编辑集群信息
   - 删除集群（含确认流程）
   - 连接测试

2. **节点管理**
   - 节点列表（实时查询 K8s API）
   - 添加/移除节点
   - 节点详情查看

### Phase 3: 资源查看

1. **运维视角**（集群为中心）
   - Namespace 管理
   - 工作负载查看（Deployments、StatefulSets、DaemonSets、Jobs）
   - 服务和配置查看（Services、Ingress、ConfigMap、Secret）
   - 存储查看（PV、PVC）

2. **开发视角**（服务为中心）
   - 查看集群部署的服务列表
   - 查看服务的部署记录

### Phase 4: 高级运维

1. **集群升级**
2. **证书管理**
3. **HPA/Quota 策略配置**
4. **集群事件查看**

## Non-Goals

- 主机异常处理（后续单独处理）
- 多集群联邦管理
- 集群备份恢复（Phase 4 考虑）
- 监控告警集成（使用现有监控模块）

## Technical Design

### 架构方案

采用方案 C：保持 cluster service 和 deployment service 分离

```
cluster service:
  - 集群 CRUD + 状态管理
  - 节点管理
  - 资源查看
  路由: /api/v1/clusters/*

deployment service:
  - 部署目标管理
  - 发布管理
  路由: /api/v1/deploy/*

调用关系: deployment 创建 target 时关联 cluster
```

### 数据模型

**新增表：cluster_nodes**
```sql
CREATE TABLE cluster_nodes (
    id               BIGINT PRIMARY KEY AUTO_INCREMENT,
    cluster_id       BIGINT NOT NULL,
    host_id          BIGINT,
    name             VARCHAR(64) NOT NULL,
    ip               VARCHAR(45) NOT NULL,
    role             VARCHAR(32) NOT NULL,  -- control-plane/worker
    status           VARCHAR(32) NOT NULL,  -- ready/notready/unknown
    kubelet_version  VARCHAR(32),
    container_runtime VARCHAR(32),
    os_image         VARCHAR(128),
    kernel_version   VARCHAR(64),
    allocatable_cpu  VARCHAR(16),
    allocatable_mem  VARCHAR(16),
    labels           JSON,
    taints           JSON,
    conditions       JSON,
    joined_at        DATETIME,
    last_seen_at     DATETIME,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_cluster_id (cluster_id),
    INDEX idx_host_id (host_id),
    UNIQUE KEY uk_cluster_name (cluster_id, name)
);
```

**扩展现有表：**
- `clusters` 增加: source, credential_id, k8s_version, pod_cidr, service_cidr, last_sync_at
- `cluster_bootstrap_tasks` 增加: cluster_id, k8s_version, pod_cidr, service_cidr, steps_json

### 脚本结构

```
script/
├── runtime/              # 已有
└── cluster/              # 新增
    ├── common/
    │   ├── preflight.sh           # 系统预检查
    │   ├── containerd-install.sh  # 安装 containerd
    │   └── sysctl-config.sh       # 系统参数
    ├── kubeadm/
    │   └── v1.28/
    │       ├── install.sh         # 安装 kubeadm/kubelet/kubectl
    │       ├── init.sh            # kubeadm init
    │       ├── join.sh            # kubeadm join
    │       ├── reset.sh           # kubeadm reset
    │       └── fetch-kubeconfig.sh
    └── cni/
        ├── calico/v3.26/
        ├── flannel/v0.22/
        └── cilium/v1.14/
```

### API 设计

```
集群 CRUD:
  GET    /clusters                     列表
  POST   /clusters                     创建
  GET    /clusters/:id                 详情
  PUT    /clusters/:id                 更新
  DELETE /clusters/:id                 删除
  POST   /clusters/:id/test            测试连接

自建集群:
  POST   /clusters/bootstrap/preview   预览
  POST   /clusters/bootstrap/apply     执行
  GET    /clusters/bootstrap/:taskId   状态

导入集群:
  POST   /clusters/import              导入
  POST   /clusters/import/validate     验证

节点管理:
  GET    /clusters/:id/nodes           列表
  POST   /clusters/:id/nodes           添加
  DELETE /clusters/:id/nodes/:name     移除

资源查看:
  GET    /clusters/:id/namespaces
  GET    /clusters/:id/namespaces/:ns/workloads
  GET    /clusters/:id/namespaces/:ns/pods
  GET    /clusters/:id/namespaces/:ns/services
  GET    /clusters/:id/namespaces/:ns/configmaps
  GET    /clusters/:id/namespaces/:ns/secrets
  GET    /clusters/:id/namespaces/:ns/pvcs
  GET    /clusters/:id/services        # 部署的服务
```

### 自建集群流程

1. Preflight Check → 检查系统要求
2. Install Containerd → 安装容器运行时
3. Install Kubeadm → 安装 kubeadm/kubelet/kubectl
4. Init Control Plane → kubeadm init
5. Install CNI → 安装网络插件
6. Join Workers → worker 节点加入
7. Fetch Kubeconfig → 获取并存储凭证
8. Create Records → 创建 Cluster、ClusterNode、Credential 记录

## Risks and Mitigations

1. **SSH 执行失败**
   - 风险：网络不稳定、权限不足
   - 缓解：重试机制、详细错误日志、rollback 脚本

2. **K8s 版本兼容性**
   - 风险：不同版本 kubeadm 参数差异
   - 缓解：版本化脚本，支持多版本

3. **CNI 安装失败**
   - 风险：网络配置冲突
   - 缓解：预检查网络配置，提供排查指南

4. **凭证安全**
   - 风险：kubeconfig 泄露
   - 缓解：加密存储、RBAC 控制、审计日志

## Success Criteria

- [ ] 自建集群：选定主机后 10 分钟内完成集群创建
- [ ] 导入集群：kubeconfig 验证后 1 分钟内完成导入
- [ ] 资源查看：实时查询响应时间 < 3 秒
- [ ] 节点管理：添加/移除节点操作可追溯、可回滚
