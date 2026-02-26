## Why

当前部署管理能力偏向“绑定已有集群并执行发布”，缺少“环境创建与运行时部署”闭环，导致新环境上线依赖手工操作且治理链路不完整。基于现有 task.md 的蓝图方向，需要补齐 SSH 远程安装、二进制分发、证书接入与外部集群导入能力，以统一 Kubernetes 与 Compose 的环境交付流程。

## What Changes

- 新增环境部署能力：支持通过 SSH 在目标主机执行二进制安装，完成 `k8s`/`compose` 运行时初始化与校验。
- 新增安装物料规范：定义 `script/` 目录下安装包、校验文件、执行脚本与版本清单的标准组织方式。
- 新增集群接入双模型：
  - 平台托管集群（平台创建并保存证书，可直接访问）
  - 外部集群（导入证书或 kubeconfig 连接）
- 扩展部署目标与发布前校验：发布前必须验证目标环境可连通、凭据可用、运行时状态健康。
- 统一审计与诊断：环境创建、安装、证书导入、连接测试与失败原因进入统一时间线。

## Capabilities

### New Capabilities
- `environment-runtime-bootstrap`: 定义环境创建、SSH 远程执行、二进制安装与运行时初始化的统一能力边界。
- `cluster-credential-ingestion`: 定义平台托管与外部导入两类集群凭据模型（平台证书、证书导入、kubeconfig 导入）及校验规则。

### Modified Capabilities
- `deployment-management-blueprint`: 扩展蓝图能力域，纳入“环境部署引擎 + 集群接入模型”并明确跨入口治理一致性。
- `k8s-runtime-deployment-management`: 扩展 K8s 目标模型与连接策略，支持平台证书直连与外部 kubeconfig/证书接入。
- `compose-runtime-deployment-management`: 扩展 Compose 目标模型，支持 SSH 安装前置校验、节点连通性与安装后健康验证。

## Impact

- Backend:
  - `internal/service/deployment`、`internal/service/cluster` 增加环境部署与凭据导入流程。
  - 新增/调整 `api/*/v1` 中环境部署、凭据导入、连通性检测与安装状态查询契约。
  - `storage/migrations` 新增环境安装任务、凭据元数据与审计关联字段。
- Frontend:
  - `web/src/pages/Deployment` 与 API 模块增加“创建环境/导入集群/安装状态”流程。
- Ops/Scripts:
  - 新增 `script/` 下二进制物料与安装脚本规范（版本、校验、回滚策略）。
- Governance:
  - 环境部署与凭据操作纳入 RBAC、审批与时间线审计。
