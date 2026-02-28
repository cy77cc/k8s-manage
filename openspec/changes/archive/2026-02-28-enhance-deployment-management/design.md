## Context

现有部署链路以“目标已存在、直接发布”为主，缺少新环境从 0 到 1 的交付能力。当前产品诉求是补齐环境部署闭环：通过 SSH 在目标节点安装运行时（二进制方式），并支持 Kubernetes 与 Compose 双运行时；同时统一平台托管集群与外部集群导入的接入模型（证书或 kubeconfig），使发布治理（preview/approval/audit）在环境创建与发布执行间保持一致。

约束：
- API 统一走 `/api/v1`，受 JWT + Casbin 保护。
- 涉及变更执行的动作必须纳入审批与时间线审计。
- 数据持久化通过 GORM + migration，必须包含回滚路径。

## Goals / Non-Goals

**Goals:**
- 提供环境部署引擎：支持 SSH 远程执行、二进制安装、安装状态回传。
- 统一集群接入模型：支持平台托管证书直连与外部 kubeconfig/证书导入。
- 为 k8s/compose 的目标创建增加安装前后校验（连通性、版本、健康）。
- 建立 `script/` 物料规范（包、校验、执行器、版本元数据），减少环境漂移。
- 将环境部署和凭据操作纳入统一审计与 RBAC。

**Non-Goals:**
- 不引入完整配置管理系统（如 Ansible inventory/role DSL）。
- 不覆盖所有 Linux 发行版差异，仅定义支持矩阵与失败回退。
- 不在本次设计中实现多云托管控制面创建（仅远程主机/现有集群接入）。

## Decisions

1. 引入“环境部署任务（EnvironmentInstallJob）”作为长任务实体。
- 方案：创建任务状态机（queued/running/succeeded/failed/rolled_back），支持步骤级日志与错误码。
- Why: SSH 安装耗时且易失败，需要可重试、可观测、可审计。
- Alternative: 同步 API 阻塞执行；缺点是超时与诊断能力差。

2. 采用“安装包 + 清单 + 校验”三件套管理 `script/` 物料。
- 方案：`script/runtime/<runtime>/<version>/` 目录下存放二进制包、sha256、install/uninstall 脚本与 manifest。
- Why: 可复现、可审计，便于离线与受限网络环境部署。
- Alternative: 在线下载即用；缺点是版本不可控且受外网波动影响。

3. 集群接入采用双模型统一抽象。
- 方案：`cluster_source = platform_managed | external_managed`。
  - platform_managed：平台创建并保管证书材料，可直接生成访问配置。
  - external_managed：支持导入 kubeconfig 或证书组（ca/cert/key + endpoint）。
- Why: 区分凭据来源和生命周期，避免混用导致权限与审计混乱。
- Alternative: 单一 kubeconfig 字段；缺点是平台创建场景无法表达证书生命周期管理。

4. 安全策略：凭据加密存储 + 最小暴露。
- 方案：凭据入库前加密，API 仅返回脱敏元数据，下载/明文查看需额外权限与审批。
- Why: 凭据高敏感，必须满足最小披露原则。
- Alternative: 明文配置只做 RBAC；缺点是泄漏风险不可接受。

5. 与发布流程解耦但可关联。
- 方案：环境安装完成后产生可引用的 deployment target；发布流程继续走 preview -> approval -> apply。
- Why: 避免把环境安装与业务发布耦合成单次事务，提高重用性与可维护性。
- Alternative: 在 apply 前隐式安装；缺点是行为不透明，失败恢复复杂。

## Risks / Trade-offs

- [Risk] SSH 网络抖动导致安装任务高失败率。
  - Mitigation: 增加连接重试、幂等步骤标记、断点重试与超时分级。
- [Risk] 二进制包版本漂移导致环境不一致。
  - Mitigation: 强制 manifest + checksum 校验，安装时写入版本指纹。
- [Risk] 外部导入 kubeconfig 格式差异导致解析失败。
  - Mitigation: 导入前静态校验并返回字段级错误，提供模板与示例。
- [Risk] 凭据权限配置不当引发越权访问。
  - Mitigation: 新增 Casbin 细粒度策略（导入、测试连接、读取元数据、下载原文）。

## Migration Plan

1. 数据层：新增环境安装任务、安装日志、凭据元数据表及索引，提供 Up/Down migration。
2. API 层：新增环境部署/安装任务查询/凭据导入与连接测试接口。
3. 执行层：引入 SSH 执行器与 runtime 安装适配器（k8s/compose）。
4. 前端：新增环境部署向导、凭据导入表单、安装进度与失败诊断视图。
5. 回滚：保留旧目标创建路径；新能力通过开关逐步放量，异常时切回仅“绑定已有集群”模式。

## Open Questions

- SSH 认证首期是否仅支持密钥，还是同时支持密码+MFA 跳板场景？
- 平台托管证书的轮换周期与自动续签策略由谁触发？
- `script/` 物料由仓库托管还是外部制品库托管并镜像到本地？
