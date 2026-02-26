 当前在 explore 模式，已发现你正在推进的 change：

  - redesign-deployment-management-blueprint（in-progress, 0/16）

  部署管理蓝图
  ├─ 能力边界
  │  ├─ Target
  │  └─ AI Command Bridge
  ├─ 生命周期语义
  │  └─ preview → approval → applying/applied → rollback
  ├─ 入口体验
  │  ├─ Deployment 页面
  │  └─ AI 命令中心
  └─ 治理一致性
     ├─ RBAC
     ├─ 审批票据
     └─ 审计时间线

  你想先聊哪一块？
  我建议先从“入口体验”开始：定义用户在页面和 AI 命令中心执行同一发布动作时，最小一致流程应该是什么。


› 入口体验


• 好，我们聚焦“入口体验”。我先快速看现有两个入口（Deployment 页面 + AI Command Center）的结构与已知 UX 约束，再给你
  一个统一体验草图。

• Explored
  └ Read DeploymentPage.tsx, AICommandCenterPage.tsx, ai-tool-call-ux-rules.md

────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

• 很好，这块现在明显是“两套入口，不同心智”：

  - DeploymentPage：功能全，但信息密度高、偏操作台
  - AICommandCenterPage：更像容器页，核心靠 CommandPanel，部署语义不够显式

  我建议把入口体验定为“同一任务，两种入口，同一状态机”。

  用户目标（发布一次）
          │
          ├─ 入口A：Deployment 页面（表单驱动）
          └─ 入口B：AI 命令中心（意图驱动）
                  │
           统一发布编排层（同一 API / 同一审批 / 同一审计）
                  │
         preview → pending_approval → applying → applied/failed → rollback

  入口体验蓝图（先定义，不实现）

  1. 入口定位

  - Deployment 页面：控制台模式（批量、配置、对比）
  - AI 命令中心：导航模式（快速达成、问答解释、一步到位）

  2. 跨入口一致对象

  - Release Draft（草稿）
  - Release Run（执行实例）
  - Approval Ticket（审批票据）
  - Timeline（统一事件时间线）

  3. 关键体验原则

  - 同一 release_id 可在两个入口无缝跳转
  - 任一入口触发审批，另一入口必须可继续处理
  - 状态文案和颜色完全一致（避免“一个叫 applying，一个叫 executing”）

  建议先讨论的 3 个产品决策

  1. AI 入口是否允许“直接 apply”，还是必须“先 preview 再确认”？
  2. 审批中心放哪里：Deployment 内嵌，还是 AI/全局统一审批收件箱？
  3. 用户第一入口默认给谁：运维（Deployment）还是普通用户（AI 命令）？

  如果你愿意，我下一步可以把“入口体验”拆成一页更具体的交互蓝图（用户旅程 + 页面信息架构 + 失败态策略），并建议应写入
  哪几个 OpenSpec artifact。



• 很好，这三个决策组合非常清晰，方向是“低风险默认 + 统一治理 + 普通用户优先”。
  1. 先 preview 再确认（禁止跳过）
  2. AI/全局统一审批收件箱
  3. 默认入口面向普通用户（项目组成员）

  这会形成一个很稳的体验主线：

  普通用户
    │
    ├─ AI 命令中心发起意图
    │      ↓
    │   生成 Release Draft（必须 preview）
    │      ↓
    │   用户确认提交
    │      ↓
    └──> 全局审批收件箱（跨项目/跨入口统一）
               ↓
          approved/rejected
               ↓
        执行与时间线回流到两端入口

  直接影响（关键）

  - 需要强约束：apply 前必须有对应 preview artifact/hash
  - 审批对象应是统一实体（不要按页面来源拆两套 ticket）
  - 普通用户首页应是“我能做什么 + 我的待办审批结果”，不是“全量运维控制台”

  建议补充的两条规则

  - preview 有时效（如 30 分钟），过期必须重预览
  - preview 与 apply 参数必须一致（防止预览与执行不一致）

  如果你同意，我建议把这 3 项正式写入当前 change 的：

  - design.md（入口策略与审批架构）
  - specs/deployment-cd-management/spec.md（preview-before-apply 强约束）
  - specs/deployment-management-blueprint/spec.md（普通用户默认入口与全局收件箱）


› 同意


• 我会把你刚确认的 3 个决策写入当前 change 的 design 与两个 spec 文件，只做规范更新，不做代码实现。先读取这三个文件定位插入点。

• Explored
  └ Read design.md, spec.md

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

• 我将把这三项决策落成明确条款：design 加关键决策，deployment-cd-management 加“preview 强约束”，deployment-management-blueprint 加“普通用户默认入口 + 全局审批收件箱”。

• Edited openspec/changes/redesign-deployment-management-blueprint/design.md (+19 -2)
    35
    36 +5. 所有发布动作采用“先预览再确认”策略，禁止绕过预览直接执行 apply。
    37 +- Why: 先暴露风险与变更影响，降低误发布概率，并让审批依据统一。
    38 +- Alternative: 允许高权限用户直接 apply；缺点是高风险环境误操作成本过高。
    39 +
    40 +6. 审批入口采用 AI/全局统一审批收件箱，不按页面来源拆分审批流。
    41 +- Why: 审批治理需要跨入口统一可见、可追溯、可接管。
    42 +- Alternative: Deployment 页面与 AI 页面各自维护审批列表；缺点是审批割裂和漏审风险高。
    43 +
    44 +7. 默认入口面向普通项目组用户，控制台高级能力按需下沉。
    45 +- Why: 大多数用户目标是“完成发布任务”，不是“管理全部平台配置”。
    46 +- Alternative: 默认进入全量运维控制台；缺点是学习成本高、路径长、误触发概率增加。
    47 +
    48  ## Risks / Trade-offs
       ⋮
    55    - Mitigation: 提供审批前预览与批量审批能力，减少重复确认成本。
    56 +- [Risk] 强制 preview 后，发布时效性可能下降。
    57 +  - Mitigation: 引入预览结果时效和缓存策略，并提供快速重预览能力。
    58 +- [Risk] 全局审批收件箱在多项目并行场景下可能信息过载。
    59 +  - Mitigation: 增加按项目/环境/风险等级过滤与默认“与我相关”视图。
    60
       ⋮
    64  2. 后端按 capability 分步改造：release API、approval/audit、timeline 查询。
    49 -3. 前端同步升级部署页与命令中心状态呈现。
    65 +3. 前端同步升级普通用户默认入口、部署页与命令中心状态呈现。
    66  4. 发布时保留旧状态兼容映射，完成历史数据回放校验后再收敛。
    67 +5. 审批入口收敛到全局收件箱后，再移除分散审批入口。
    68
       ⋮
    72  - Compose 运行时是否在蓝图阶段定义标准健康探针模型。
    24
    25 +### Requirement: Preview MUST be confirmed before apply
    26 +The system MUST require a valid preview result before release apply, and SHALL reject apply requests that do not reference a valid preview artifact generated from the same
         release draft context.
    27 +
    28 +#### Scenario: Reject apply without preview
    29 +- **WHEN** a user submits apply for a release draft without a prior valid preview
    30 +- **THEN** the system MUST reject the request and return a preview-required response
    31 +
    32 +#### Scenario: Reject apply with stale preview
    33 +- **WHEN** a user submits apply with a preview artifact that has expired based on platform preview TTL policy
    34 +- **THEN** the system MUST reject the request and require re-preview before confirmation
    35 +
    36 +#### Scenario: Reject apply with mismatched parameters
    37 +- **WHEN** a user confirms apply with parameters or target context different from the referenced preview artifact
    38 +- **THEN** the system MUST reject the request and require a new preview for the changed draft

• Edited openspec/changes/redesign-deployment-management-blueprint/specs/deployment-management-blueprint/spec.md (+14 -0)
    22  - **THEN** the system SHALL verify that required lifecycle APIs, UI states, approval controls, and audit timelines satisfy the defined acceptance checklist
    23 +
        hs.
    26 +
    32 +The blueprint SHALL define a global approval inbox shared across UI and AI command entry points, and MUST ensure approval tickets are managed in one unified queue with con
        sistent scope and audit linkage.
    33 +
    34 +#### Scenario: Unified approval handling across entry points
    35 +- **WHEN** a release request is submitted from Deployment UI or AI command center
    36 +- **THEN** the system SHALL route the approval ticket into the same global inbox and SHALL allow approvers to process it without depending on the source entry point

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

• 我已完成规范更新，最后做一次快速校验（文件内容与 change 状态）确保探索结论已正确落盘。

• Explored
  └ Search 先预览再确认|全局|默认入口|Preview MUST be confirmed before apply|global approval inbox|default entry in design.md
    Read spec.md

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

• 已按你的确认，把入口体验决策写入当前 change 的 artifacts（仅规范更新，无实现）：

  - design.md (/root/project/k8s-manage/openspec/changes/redesign-deployment-management-blueprint/design.md)
  - deployment-cd-management/spec.md (/root/project/k8s-manage/openspec/changes/redesign-deployment-management-blueprint/specs/deployment-cd-management/spec.md)
  - deployment-management-blueprint/spec.md (/root/project/k8s-manage/openspec/changes/redesign-deployment-management-blueprint/specs/deployment-management-blueprint/spec.md)

  新增/强化的关键约束：

  1. 所有发布必须 preview -> confirm -> apply，禁止跳过 preview
  2. 审批统一进入 AI/全局审批收件箱，不按入口拆分
  3. 默认入口面向普通项目组用户（任务导向），高级控制台能力下沉
  4. apply 需绑定有效 preview（含过期与参数一致性校验）

  当前 change 仍是 apply-ready（artifacts 全部 done）。
  如果你想继续探索，我建议下一步聊“全局审批收件箱的信息架构”（分组、过滤、优先级、SLA 展示）。


› 再一个就是应该添加环境部署的功能，现在基本上只能添加现有的集群。希望加上一个环境部署的功能，支持部署docker compose和k8s，用二进制的方式安装，相关的安装文件你告诉我需要什么，我
  放到script目录下规范存放的形式，通过远程ssh的方式部署。对于k8s现在有生成证书的demo需要完善一下，用平台创建的集群可以直接通过保存的证书访问，对于外部的集群就通过引入证书或者
  kubeconfig的方式连接。这样的话就比较完整了。


• 这个方向非常关键，我先在代码里快速定位现有“建群/证书/kubeconfig/SSH 部署”能力边界，然后给你一份可直接落地的“环境部署蓝图 + script 目录物料清单”。

• Explored
  └ Read cluster.go, logic.go
    Search kubeconfig|certificate|cert|bootstrap|ssh|compose|deploy in service
    Read cluster_phase1.go

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

• 这个需求非常好，实际上是在补齐“环境部署闭环”。你现在要的是：

  1. 不只是绑定现有集群，而是平台可“创建环境并部署运行时”
  2. 支持 compose 和 k8s，通过 SSH + 二进制安装
  3. 集群访问打通两类来源：

  - 平台创建集群：自动保存证书/配置并可直接访问
  - 外部集群：导入证书或 kubeconfig 连接

  我建议蓝图拆成 3 条主线：

  [环境部署引擎]
     ├─ SSH 执行编排
     ├─ 二进制安装(Compose/K8s)
     └─ 安装结果回传(资产+审计)

  [集群接入模型]
     ├─ platform_managed (平台创建)
     └─ external_managed (外部导入)
          ├─ kubeconfig
          └─ cert+endpoint+token

  [统一部署入口]
     ├─ 选择环境/运行时
     ├─ preview -> confirm -> apply
     └─ approval + timeline

  ## 你要准备到 script/ 的物料（建议规范）

  建议按这个结构放，平台只做编排和校验：

  script/
    manifests/
      index.yaml                 # 版本、sha256、适用架构
    bin/
      linux-amd64/
        kubectl
        kubeadm
        kubelet
        crictl
        helm                     # 可选
        docker
        docker-compose           # 或 compose plugin
        containerd
        runc
        cni-plugins.tgz
    pkgs/
      linux-amd64/
      k8s-v1.xx/
        kubeadm-images.tar       # kube-apiserver/controller/scheduler/proxy/coredns/pause/etcd
        cni-images.tar           # calico/flannel 对应镜像
    templates/
      kubeadm/
        init-config.yaml.tpl
        join-config.yaml.tpl
      compose/
        docker-compose.yml.tpl
    hooks/
      preflight.sh
      install-compose.sh
      install-k8s-controlplane.sh
      install-k8s-worker.sh
      post-install-verify.sh

  ## K8s 证书与连接模型（关键）

  你提的点完全正确，建议定死两种模式：

  1. platform_managed

  - 平台建群后保存：kubeconfig + ca_cert + endpoint（必要时 token/client cert）
  - 默认可直接被平台 clientset 使用
  - 证书轮转要有任务入口（后续 phase）

  2. external_managed

  - 只支持两种导入：
      - 完整 kubeconfig 导入
      - endpoint + ca + token/client-cert 导入
  - 入库前做连通与权限探测（namespace list / version）

  ## 你当前需求里最容易踩坑的 4 点

  1. 二进制版本矩阵不统一（kubeadm/kubelet/kubectl/containerd/CNI）
  2. 离线镜像与二进制版本不匹配
  3. SSH 执行幂等性不足（重跑安装会炸）
  4. 平台创建与外部导入的凭据字段不统一，后续 API 混乱

  如果你愿意，我下一步可以在 explore 模式里给你一版“最小可执行 Phase-1 范围”（先做哪些、暂时不做哪些），让你上线风险最低。


› 不用最小可执行，我这个是学习项目，完全拉上来就行


• 可以，学习项目就走“全量方案”，不压缩范围。

  全量目标图

  环境部署中心
    -> 基础设施准备(SSH/主机分组/凭据)
    -> 运行时安装(Compose/K8s binary+offline artifacts)
    -> 集群接入(platform_managed / external_managed)
    -> 发布执行(preview->confirm->apply->approval->timeline->rollback)
    -> 运维能力(证书轮转/版本升级/健康检查/审计)

  建议一次性拉齐的能力包

  1. 环境模型

  - environment 实体：name、project/team、runtime_type、mode(platform/external)、status、region、labels
  - environment_nodes：控制平面/worker/compose-manager/compose-worker
  - environment_credentials_ref：ssh_key、kubeconfig、cert bundle 引用

  2. 安装编排

  - 统一 install task 状态机：pending -> running -> succeeded|failed|rollbacked
  - 每一步存日志、耗时、exit code、stderr 摘要
  - 支持重试与断点续跑（幂等）

  3. Compose 部署管理

  - 二进制安装 docker/containerd/compose plugin
  - compose 环境初始化（目录、systemd、日志路径）
  - deploy/preview/apply/rollback/ps/logs 全链路

  4. K8s 部署管理

  - kubeadm 二进制安装 + control-plane/worker join
  - CNI 安装与检查
  - kubeconfig 生成、回传、入库
  - 平台创建集群直接可访问（clientset 可用）

  5. 外部集群接入

  - import by kubeconfig
  - import by endpoint+ca+token|client-cert
  - 导入即验证：API 可达、权限最小探测、namespace 枚举

  6. 证书与访问闭环

  - 平台创建集群保存 kubeconfig/ca/token(or cert)
  - 外部集群保存导入凭据（建议加密）
  - 提供证书过期检查与轮转任务

  7. 全局审批与审计

  - 全局审批收件箱（UI 与 AI 共用）
  - 审批票据挂到 release/install/import
  - 统一 timeline 事件模型

  8. UI 全量入口

  - 环境中心页（创建环境、选择 runtime、安装进度）
  - 集群接入页（平台创建/外部导入）

  9. AI 入口打通
  你需要放到 script/ 的完整物料

  1. script/manifests/index.yaml

  - 记录每个版本的 binary、image、sha256、os/arch、兼容矩阵

  2. script/bin/linux-amd64/

  - kubectl kubeadm kubelet crictl containerd runc docker docker-compose(or plugin)

  3. script/images/

  - K8s 核心镜像离线包
  - CNI 镜像离线包

  4. script/templates/

  - kubeadm init/join 模板
  - compose 模板

  5. script/hooks/

  - preflight.sh
  - install-compose.sh
  - install-k8s-controlplane.sh
  - install-k8s-worker.sh
  - verify.sh
  - rollback.sh

  6. script/checksums/

  - 全部物料 hash 清单

  如果你同意，我下一步可以继续在 explore 模式给你一份“数据库字段草案 + API 草案清单（按 endpoint 列出来）”，这样你后面 /opsx:apply 就能直接按表施工。


› 同意


• 我先基于现有路由快速做对齐（避免和当前 /deploy、/clusters 冲突），然后给你一版可直接落地的“数据库字段草案 + API 清单”。

• Explored
  └ Read routes.go
    Search Register.*Handlers|Group\(\"/deploy|Group\(\"/clusters|/api/v1 in service
    Read routes.go

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

• 非常好，这里给你一版可直接用的“全量数据库草案 + API 草案”。

  1) 数据库字段草案
  建议新增这些表（不改代码，只是蓝图）：

  - environments
      - id
      - name
      - project_id
      - team_id
      - runtime_type (k8s|compose|hybrid)
      - mode (platform_managed|external_managed)
      - status (draft|installing|ready|failed|offline)
      - region
      - labels_json
      - description
      - created_by
      - created_at
      - updated_at
  - environment_nodes
      - id
      - environment_id
      - host_id
      - node_role (control_plane|worker|compose_manager|compose_worker)
      - weight
      - status
      - created_at
      - updated_at
  - environment_install_tasks
      - id (task_id 字符串)
      - environment_id
      - runtime_type
      - action (preflight|install|upgrade|repair|rollback)
      - status (pending|running|succeeded|failed|rolled_back)
      - step_index
      - steps_json
      - result_json
      - error_message
      - operator_id
      - created_at
      - updated_at
  - environment_install_logs
      - id
      - task_id
      - step_name
      - host_id
      - level
      - content
      - created_at
  - environment_artifacts
      - id
      - environment_id
      - artifact_type (binary|image|template|checksum)
      - artifact_name
      - version
      - path
      - sha256
      - meta_json
      - created_at
  - cluster_connections（建议新增，不把连接信息全塞 clusters）
      - id
      - cluster_id
      - source_type (platform_created|external_imported)
      - auth_type (kubeconfig|cert_token|cert_key)
      - endpoint
      - kubeconfig_encrypted
      - ca_cert_encrypted
      - client_cert_encrypted
      - client_key_encrypted
      - token_encrypted
      - expires_at
      - last_check_at
      - last_check_status
      - created_at
      - updated_at
  - cluster_cert_rotation_tasks
      - id
      - cluster_id
      - status
      - rotation_type
      - result_json
      - error_message
      - operator_id
      - created_at
      - updated_at

  2) API 草案（全量）
  建议新增 /api/v1/environments/*，保留现有 /deploy 与 /clusters 兼容。

  - 环境管理
      - GET /api/v1/environments
      - POST /api/v1/environments
      - GET /api/v1/environments/:id
      - PUT /api/v1/environments/:id
      - DELETE /api/v1/environments/:id
      - PUT /api/v1/environments/:id/nodes
  - 安装与运维
      - POST /api/v1/environments/:id/install/preview
      - POST /api/v1/environments/:id/install/apply
      - POST /api/v1/environments/:id/install/rollback
      - GET /api/v1/environments/:id/install/tasks
      - GET /api/v1/environments/:id/install/tasks/:task_id
      - GET /api/v1/environments/:id/install/tasks/:task_id/logs
      - POST /api/v1/environments/:id/health/check
  - 集群接入（补齐平台/外部双模式）
      - POST /api/v1/clusters/create-from-environment（platform_managed）
      - POST /api/v1/clusters/import/kubeconfig
      - POST /api/v1/clusters/import/cert
      - POST /api/v1/clusters/:id/connect/test
      - GET /api/v1/clusters/:id/connection
      - PUT /api/v1/clusters/:id/connection
      - POST /api/v1/clusters/:id/certs/rotate/preview
      - POST /api/v1/clusters/:id/certs/rotate/apply
  - 部署执行（与现有 /deploy/releases/* 对齐）
      - POST /api/v1/deploy/releases/preview（必须先走）
      - POST /api/v1/deploy/releases/apply（必须带 preview_ref）
      - POST /api/v1/deploy/releases/:id/approve
      - POST /api/v1/deploy/releases/:id/reject
      - POST /api/v1/deploy/releases/:id/rollback
      - GET /api/v1/deploy/releases/:id/timeline

  3) script 目录契约（接口要依赖）

  - script/manifests/index.yaml 必须有：
      - name
      - version
      - os
      - arch
      - runtime
      - artifacts[]（path, sha256, type）
      - images[]
      - steps[]（调用哪个 hooks/*.sh）

  4) 关键流程

  Create Environment
   -> Bind Hosts
   -> Install Preview (检查 script 物料 + SSH连通 + 版本矩阵)
   -> Install Apply (逐步执行 hooks)
   -> Create/Import Cluster Connection
   -> Deploy Preview
   -> Deploy Apply (approval)
   -> Timeline & Audit

  如果你同意，我可以下一步继续在 explore 模式给你“请求/响应 JSON 草案”（每个核心接口给一个样例），这样你实现时基本不需要再补协议。


› 同意


• 下面给你核心接口的 JSON 草案（全量版，直接可当 contract 起稿）。

  1) 创建环境
  POST /api/v1/environments

  {
    "name": "prod-cn-hz",
    "project_id": 11,
    "team_id": 7,
    "runtime_type": "k8s",
    "mode": "platform_managed",
    "region": "cn-hz",
    "labels": {
      "tier": "prod",
      "owner": "platform"
    },
    "description": "production k8s environment"
  }

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "id": 301,
      "name": "prod-cn-hz",
      "status": "draft",
      "runtime_type": "k8s",
      "mode": "platform_managed",
      "created_at": "2026-02-26T12:00:00Z"
    }
  }

  2) 绑定环境节点
  PUT /api/v1/environments/:id/nodes

  {
    "nodes": [
      { "host_id": 101, "node_role": "control_plane", "weight": 100 },
      { "host_id": 102, "node_role": "worker", "weight": 100 },
      { "host_id": 103, "node_role": "worker", "weight": 100 }
    ]
  }

  3) 安装预览
  POST /api/v1/environments/:id/install/preview

  {
    "runtime_type": "k8s",
    "manifest_version": "v1.30.1-offline.1",
    "script_root": "script",
    "install_options": {
      "cni": "calico",
      "pod_cidr": "10.244.0.0/16",
      "service_cidr": "10.96.0.0/12"
    }
  }

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "preview_id": "inst-prev-987",
      "checks": [
        { "level": "info", "code": "ssh_ok", "message": "all nodes reachable" },
        { "level": "info", "code": "artifacts_ok", "message": "all binaries/images found" }
      ],
      "warnings": [],
      "steps": [
        "preflight.sh",
        "install-k8s-controlplane.sh",
        "install-k8s-worker.sh",
        "post-install-verify.sh"
      ],
      "expires_at": "2026-02-26T12:30:00Z"
    }
  }

  4) 安装执行（必须带 preview_ref）
  POST /api/v1/environments/:id/install/apply

  {
    "preview_ref": "inst-prev-987",
    "approval_token": "approve-xxx",
    "confirm": true
  }

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "task_id": "inst-task-123",
      "status": "running",
      "environment_id": 301
    }
  }

  5) 安装任务详情
  GET /api/v1/environments/:id/install/tasks/:task_id

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "task_id": "inst-task-123",
      "status": "running",
      "step_index": 2,
      "steps": [
        { "name": "preflight.sh", "status": "succeeded" },
        { "name": "install-k8s-controlplane.sh", "status": "running" },
        { "name": "install-k8s-worker.sh", "status": "pending" }
      ],
      "result": {},
      "error_message": ""
    }
  }

  6) 外部集群导入（kubeconfig）
  POST /api/v1/clusters/import/kubeconfig

  {
    "name": "external-prod-a",
    "project_id": 11,
    "team_id": 7,
    "kubeconfig": "BASE64_OR_RAW",
    "labels": { "source": "external" }
  }

  7) 外部集群导入（证书）
  POST /api/v1/clusters/import/cert

  {
    "name": "external-prod-b",
    "project_id": 11,
    "team_id": 7,
    "endpoint": "https://10.0.0.10:6443",
    "auth_type": "cert_token",
    "ca_cert": "-----BEGIN CERTIFICATE-----...",
    "token": "eyJhbGci...",
    "client_cert": "",
    "client_key": ""
  }

  8) 集群连通测试
  POST /api/v1/clusters/:id/connect/test

  {
    "check_permissions": true
  }

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "reachable": true,
      "server_version": "v1.30.1",
      "permissions": {
        "list_namespaces": true,
        "list_deployments": true
      }
    }
  }

  9) 发布预览
  POST /api/v1/deploy/releases/preview

  {
    "service_id": 501,
    "target_id": 801,
    "env": "production",
    "strategy": "rolling",
    "variables": { "image_tag": "v2.0.1" }
  }

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "preview_ref": "rel-prev-555",
      "resolved_manifest": "...",
      "checks": [],
      "warnings": [],
      "runtime": "k8s",
      "expires_at": "2026-02-26T12:30:00Z"
    }
  }

  10) 发布执行（强制 preview_ref）
  POST /api/v1/deploy/releases/apply

  {
    "service_id": 501,
    "target_id": 801,
    "env": "production",
    "strategy": "rolling",
  }

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "release_id": 9001,
      "status": "pending_approval",
      "approval_required": true,
      "approval_ticket": "dep-appr-9001",
      "lifecycle_state": "preview"
    }
  }

  11) 审批通过/拒绝
  POST /api/v1/deploy/releases/:id/approve

  { "comment": "approved for prod window" }

  POST /api/v1/deploy/releases/:id/reject

  { "comment": "risk too high" }

  12) 时间线
  GET /api/v1/deploy/releases/:id/timeline

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "list": [
        { "id": 1, "action": "release.previewed", "actor": 12, "created_at": "..." },
        { "id": 2, "action": "release.pending_approval", "actor": 12, "created_at": "..." },
        { "id": 3, "action": "release.approved", "actor": 3, "created_at": "..." },
        { "id": 4, "action": "release.applied", "actor": 12, "created_at": "..." }
      ],
      "total": 4
    }
  }

  如果你要，我下一步可以继续给你一份“错误码草案”（比如 preview_required, preview_expired, artifact_missing, ssh_unreachable, kubeconfig_invalid）。


› 继续


• 给你一版可直接用的错误码草案。

  错误响应统一格式

  {
    "code": 32001,
    "msg": "preview required",
    "data": {
      "error_key": "preview_required",
      "hint": "请先执行预览并确认后再发布"
    }
  }

  A. 环境与安装（32xxx）

  - 32001 preview_required：缺少 preview_ref
  - 32002 preview_expired：预览过期
  - 32003 preview_mismatch：预览与 apply 参数不一致
  - 32004 environment_not_found
  - 32005 environment_status_invalid：状态不允许当前动作
  - 32006 install_task_not_found
  - 32007 install_task_state_invalid
  - 32008 artifact_manifest_missing：script/manifests/index.yaml 缺失
  - 32009 artifact_missing：二进制/镜像文件缺失
  - 32010 artifact_checksum_mismatch
  - 32011 artifact_version_unsupported
  - 32012 script_hook_missing
  - 32013 script_hook_exec_failed
  - 32014 preflight_failed
  - 32015 ssh_unreachable
  - 32016 ssh_auth_failed
  - 32017 node_role_invalid
  - 32018 insufficient_nodes

  B. K8s 接入与证书（33xxx）

  - 33001 cluster_not_found
  - 33002 cluster_connection_not_found
  - 33003 kubeconfig_invalid
  - 33004 certificate_invalid
  - 33005 endpoint_invalid
  - 33006 cluster_connect_failed
  - 33007 cluster_permission_insufficient
  - 33008 cert_rotation_not_allowed
  - 33009 cert_rotation_failed
  - 33010 kubeadm_init_failed
  - 33011 kubeadm_join_failed
  - 33012 cni_install_failed
  - 33013 k8s_binary_missing
  - 33014 k8s_version_incompatible

  C. Compose 运行时（34xxx）

  - 34001 compose_binary_missing
  - 34004 compose_up_failed
  - 34005 compose_ps_failed
  - 34006 compose_rollback_failed
  - 34007 compose_project_conflict
  - 34008 compose_host_unhealthy

  D. 发布与审批（35xxx）

  - 35001 release_not_found
  - 35002 release_state_invalid
  - 35003 approval_required
  - 35004 approval_ticket_not_found
  - 35005 approval_ticket_expired
  - 35006 approval_ticket_scope_mismatch
  - 35007 approval_ticket_not_approved
  - 35008 approval_permission_denied
  - 35009 release_apply_failed
  - 35010 release_rollback_failed
  - 35011 release_timeline_unavailable

  E. 权限与通用（36xxx）

  - 36001 permission_denied
  - 36002 project_scope_denied
  - 36003 team_scope_denied
  - 36004 validation_error
  - 36005 resource_conflict
  - 36006 resource_locked
  - 36007 rate_limited
  - 36008 internal_error

  前端文案建议（短句）

  - preview_required：请先预览再确认发布
  - preview_expired：预览已过期，请重新预览
  - artifact_missing：安装物料缺失，请检查 script 目录
  - ssh_unreachable：目标主机不可达，请检查网络与凭据
  - kubeconfig_invalid：kubeconfig 无效或格式错误
  - approval_required：该操作需要审批，已进入审批收件箱

  如果继续，我可以给你“状态机事件字典”（release.previewed、install.step.failed 等）和每个事件的 detail 字段规范。


› 继续


• 下面是“状态机事件字典 + detail 字段规范”。

  统一事件包格式

  {
    "event_id": "evt-123",
    "domain": "release",
    "action": "release.previewed",
    "resource_type": "deployment_release",
    "resource_id": "9001",
    "environment_id": 301,
    "project_id": 11,
    "team_id": 7,
    "actor_id": 12,
    "actor_type": "user",
    "source": "deployment_ui",
    "status": "success",
    "occurred_at": "2026-02-26T12:00:00Z",
    "detail": {}
  }

  ## 1) Release 事件字典

  - release.draft_created
  - release.previewed
  - release.preview_failed
  - release.apply_requested
  - release.pending_approval
  - release.approved
  - release.rejected
  - release.applying
  - release.applied
  - release.failed
  - release.rollback_requested
  - release.rollback_completed
  - release.rollback_failed

  detail 规范（release）

  {
    "release_id": 9001,
    "service_id": 501,
    "target_id": 801,
    "runtime_type": "k8s",
    "env": "production",
    "strategy": "rolling",
    "preview_ref": "rel-prev-555",
    "approval_ticket": "dep-appr-9001",
    "error_key": "",
    "summary": "optional short summary"
  }

  ## 2) Approval 事件字典

  - approval.ticket_created
  - approval.ticket_assigned
  - approval.ticket_viewed
  - approval.ticket_approved
  - approval.ticket_rejected
  - approval.ticket_expired
  - approval.ticket_scope_mismatch

  detail 规范（approval）

  {
    "ticket": "dep-appr-9001",
    "scope_type": "release",
    "scope_id": 9001,
    "required_role": "service:approve",
    "requested_by": 12,
    "reviewed_by": 3,
    "comment": "approved for prod window",
    "expires_at": "2026-02-26T13:00:00Z"
  }

  ## 3) Environment 安装事件字典

  - environment.install_previewed
  - environment.install_requested
  - environment.install_started
  - environment.install_step_started
  - environment.install_step_succeeded
  - environment.install_step_failed
  - environment.install_succeeded
  - environment.install_failed
  - environment.install_rollback_started
  - environment.install_rollback_succeeded
  - environment.install_rollback_failed

  detail 规范（environment/install）

  {
    "environment_id": 301,
    "task_id": "inst-task-123",
    "runtime_type": "k8s",
    "manifest_version": "v1.30.1-offline.1",
    "step_name": "install-k8s-controlplane.sh",
    "step_index": 2,
    "host_id": 101,
    "exit_code": 1,
    "stderr_excerpt": "kubeadm init failed ...",
    "error_key": "kubeadm_init_failed"
  }

  ## 4) Cluster 接入事件字典

  - cluster.create_from_environment_requested
  - cluster.create_from_environment_succeeded
  - cluster.import_kubeconfig_requested
  - cluster.import_kubeconfig_succeeded
  - cluster.import_cert_requested
  - cluster.import_cert_succeeded
  - cluster.connection_test_succeeded
  - cluster.connection_test_failed
  - cluster.cert_rotation_requested
  - cluster.cert_rotation_succeeded
  - cluster.cert_rotation_failed

  detail 规范（cluster）

  {
    "cluster_id": 88,
    "source_type": "platform_created",
    "auth_type": "kubeconfig",
    "endpoint": "https://10.0.0.10:6443",
    "server_version": "v1.30.1",
    "namespaces_checked": ["default", "kube-system"],
    "error_key": ""
  }

  ## 5) Compose 运行时事件字典

  - compose.preflight_succeeded
  - compose.preflight_failed
  - compose.config_rendered
  - compose.up_started
  - compose.up_succeeded
  - compose.up_failed
  - compose.ps_collected
  - compose.rollback_started
  - compose.rollback_succeeded
  - compose.rollback_failed

  detail 规范（compose）

  {
    "environment_id": 401,
    "target_id": 901,
    "host_id": 201,
    "project_name": "svc-a-prod",
    "compose_file": "/opt/opspilot/compose/svc-a/docker-compose.yml",
    "error_key": "",
    "summary": "docker compose up -d completed"
  }

  - k8s.preflight_succeeded
  - k8s.preflight_failed
  - k8s.kubeadm_init_started
  - k8s.kubeadm_init_succeeded
  - k8s.kubeadm_init_failed
  - k8s.kubeadm_join_started
  - k8s.kubeadm_join_succeeded
  - k8s.kubeadm_join_failed
  - k8s.cni_install_succeeded
  - k8s.cni_install_failed
  - k8s.kubeconfig_generated
  - k8s.deploy_apply_succeeded
  - k8s.deploy_apply_failed

  detail 规范（k8s）

  {
    "cluster_id": 88,
    "control_plane_host_id": 101,
    "worker_host_id": 102,
    "k8s_version": "v1.30.1",
    "cni": "calico",
    "kubeconfig_ref": "cluster-88-admin",
    "error_key": ""
  }

  ## 7) Source 枚举（入口来源）

  - deployment_ui
  - ai_command_center
  - approval_inbox
  - system_scheduler
  - api_direct

  ## 8) 状态字段约定

  - status: success|failed|warning|running
  - severity: info|low|medium|high|critical
  - error_key: 必须来自错误码字典（如 preview_expired）

  如果你继续，我下一步给你“审批收件箱数据模型 + 列表筛选协议”（包括默认“与我相关”视图、优先级排序、SLA）。


› 继续


• 下面是“全局审批收件箱”数据模型与筛选协议草案。

  ## 1) 审批收件箱主表

  approval_inbox_items

  - id
  - ticket（唯一）
  - scope_type（release|install|cluster_import|cert_rotation|compose_apply|k8s_apply）
  - scope_id
  - project_id
  - team_id
  - environment_id
  - runtime_type（k8s|compose|mixed）
  - risk_level（low|medium|high|critical）
  - priority（p0|p1|p2|p3）
  - status（pending|approved|rejected|expired|cancelled）
  - requested_by
  - assigned_to（可空）
  - required_role（如 service:approve）
  - source（deployment_ui|ai_command_center|api_direct）
  - summary
  - detail_json
  - expires_at
  - sla_due_at
  - approved_at
  - approved_by
  - rejected_at
  - rejected_by
  - created_at
  - updated_at

  ## 2) 审批动作表

  approval_inbox_actions

  - id
  - ticket
  - action（created|assigned|viewed|approved|rejected|expired|commented|escalated）
  - actor_id
  - comment
  - meta_json
  - created_at

  ## 3) 默认“与我相关”视图规则

  满足任一即入：

  - assigned_to = me
  - requested_by = me
  - required_role 在我的角色权限内 且 status=pending
  - 我所属 project_id/team_id 且 status=pending

  ## 4) 列表查询 API

  GET /api/v1/approvals/inbox

  Query 参数：

  - view：mine|pending|all（默认 mine）
  - status：多值
  - scope_type：多值
  - project_id
  - team_id
  - environment_id
  - runtime_type
  - risk_level
  - priority
  - assigned_to
  - requested_by
  - q（ticket/summary 模糊）
  - sort_by：priority|expires_at|sla_due_at|created_at
  - sort_order：asc|desc
  - page
  - page_size

  ## 5) 响应结构（列表）

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "list": [
        {
          "ticket": "dep-appr-9001",
          "scope_type": "release",
          "scope_id": 9001,
          "project_id": 11,
          "team_id": 7,
          "runtime_type": "k8s",
          "risk_level": "high",
          "priority": "p1",
          "status": "pending",
          "summary": "prod release svc-a v2.0.1",
          "requested_by": 12,
          "assigned_to": 3,
          "required_role": "service:approve",
          "expires_at": "2026-02-26T13:00:00Z",
          "sla_due_at": "2026-02-26T12:30:00Z",
          "source": "ai_command_center",
          "created_at": "2026-02-26T12:00:00Z"
        }
      ],
      "total": 1,
      "stats": {
        "pending": 1,
        "overdue": 0,
        "expiring_15m": 1
      }
    }
  }

  ## 6) 详情 API

  GET /api/v1/approvals/inbox/:ticket
  返回：

  - 基本信息（同列表）
  - detail_json
  - timeline（actions）
  - related_release / related_task 摘要

  ## 7) 审批操作 API

  - POST /api/v1/approvals/inbox/:ticket/approve
  - POST /api/v1/approvals/inbox/:ticket/reject
  - POST /api/v1/approvals/inbox/:ticket/assign
  - POST /api/v1/approvals/inbox/:ticket/comment

  Approve 请求：

  {
    "comment": "approved in maintenance window",
    "expected_version": "2.0.1"
  Reject 请求：

  {
    "comment": "risk not mitigated"
  }

  Assign 请求：

  {
    "assigned_to": 3,
    "comment": "oncall reviewer"
  }

  ## 8) 排序优先级建议

  默认排序：

  1. status=pending
  2. priority（p0 > p1 > p2 > p3）
  3. expires_at 最近
  4. created_at 最早

  ## 9) SLA 规则建议

  - critical: 15m
  - high: 30m
  - medium: 2h
  - low: 8h

  逾期条件：now > sla_due_at && status=pending

  ## 10) 前端筛选栏（最小字段）

  - 视图切换：我的待办 | 全部待办 | 我发起的
  - 状态筛选
  - 项目/团队筛选
  - 风险等级 + 优先级
  - 到期窗口：15m|1h|today
  - 搜索框（ticket/摘要）

  如果继续，我可以给你“AI 命令中心与审批收件箱的联动交互脚本”（用户说一句话到审批完成的完整 UI 事件序列）。


› 继续


• 下面给你“AI 命令中心 <-> 审批收件箱”完整交互脚本（端到端）。

  ## 1) 主流程：一句话到上线

  [AI 命令输入]
    "把 svc-a 发布到生产，镜像 v2.0.1"
          │
          ▼
  [参数补全与校验]
    - service_id/target_id/env/strategy/variables
          │
          ▼
  [强制 Preview]
    - 调 /deploy/releases/preview
    - 返回 preview_ref + checks/warnings + expires_at
          │
          ▼
  [用户确认]
    - “确认发布”
          │
          ▼
  [Apply 请求]
    - 带 preview_ref 调 /deploy/releases/apply
    - 返回 pending_approval + approval_ticket
          │
          ▼
  [审批收件箱]
    - ticket 进入 /approvals/inbox
    - 审批人处理 approve/reject
          │
     ┌────┴─────┐
     ▼          ▼
  approved   rejected
     │          │
     ▼          ▼
  [执行 apply]  [结束并回显原因]
     │
     ▼
  [时间线回流]
  AI 对话 + Deployment 页 + Inbox 全部同步状态

  ## 2) AI 对话事件序列（SSE/消息流）

  - intent_detected
  - param_resolution
  - preview_started
  - preview_result
  - user_confirmation_required
  - apply_requested
  - approval_required（含 approval_ticket）
  - approval_status_changed（approved/rejected）
  - release_status_changed（applying/applied/failed）
  - timeline_synced

  ## 3) AI 侧消息文案模板

  - Preview 成功：
      - 已完成预览：2 项检查通过，0 项阻断。是否提交发布审批？
  - 审批中：
      - 已提交审批（ticket: dep-appr-9001）。我会在审批结果出来后继续。
  - 审批通过：
      - 审批已通过，开始执行发布。
  - 审批拒绝：
      - 审批被拒绝：{comment}。如需我可生成修复后重提方案。
  - 发布完成：
  ## 4) Inbox 与 AI 联动规则

  - Inbox 每次状态变化都发事件给 AI 会话（同 user/project scope）
  - AI 会话恢复时可按 ticket 拉取最新状态（防断线）
  - AI 操作“批准/拒绝”必须走 inbox API，不允许绕过

  ## 5) 失败分支脚本

  - preview_required：
      - AI 自动回到 preview 步骤，不直接报错终止
  - preview_expired：
      - AI 提示“预览过期，正在重新预览”
  - approval_ticket_expired：
      - AI 提示“审批已过期，建议重新提交”
  - scope_mismatch：
      - AI 提示“审批票据与当前发布上下文不匹配，需重提审批”

  ## 6) 关键上下文键（AI runtime context）

  - release_draft_id
  - preview_ref
  - preview_expires_at
  - approval_ticket
  - release_id
  - environment_id
  - project_id
  - team_id

  ## 7) 安全护栏

  - mutating 操作必须携带 approval_ticket（通过后）
  - AI 不展示敏感证书原文，只展示引用 ID
  - 跨项目 ticket 不可见、不可操作

  ## 8) 前端联动点（页面）

  - AI 命令中心：显示“审批中卡片 + 继续执行状态”
  - 审批收件箱：显示来源 source=ai_command_center
  - Deployment 详情：时间线含 AI 来源事件

  如果继续，我下一步给你“测试用例矩阵”（AI、Inbox、Deployment 三端一致性的回归清单）。


› 继续


• 下面是三端一致性回归矩阵（AI / Inbox / Deployment）。

  ## 1) 主成功路径

  1. AI 发起生产发布 -> 必须先 preview -> confirm -> apply -> pending_approval
  2. Inbox 审批通过
  3. Deployment 显示 applying -> applied
  4. 三端时间线一致（同 release_id + ticket）

  验收：

  - AI 中出现 approval_required
  - Inbox 有对应 ticket 且 source=ai_command_center
  - Deployment 详情含 release.previewed -> pending_approval -> approved -> applied

  ## 2) 审批拒绝路径

  1. AI 发起发布进入 pending_approval
  2. Inbox reject
  3. AI 收到拒绝回推
  4. Deployment 状态 rejected，无 apply 执行事件

  验收：

  - approval.ticket_rejected 在三端可见
  - 无 release.applying 事件

  ## 3) Preview 约束路径

  1. 直接 apply（无 preview_ref）
  2. 返回 preview_required
  3. AI 自动回到 preview
  4. 用户确认后再提交

  验收：

  - 错误码稳定：preview_required
  - 不产生审批 ticket

  ## 4) Preview 过期路径

  1. preview 成功
  2. 等待超 TTL
  3. apply
  4. 返回 preview_expired

  验收：

  - 错误码：preview_expired
  - 需重新 preview 后才可 apply

  ## 5) 参数不一致路径

  1. preview 用 image_tag=v1
  2. apply 改为 v2
  3. 返回 preview_mismatch

  验收：

  - 拒绝执行
  - 提示需重 preview

  ## 6) Inbox 指派路径

  1. ticket 创建未指派
  2. 指派 oncall 审批人
  3. 审批人批准
  4. 事件链完整

  验收：

  - actions: created -> assigned -> approved
  - assigned_to 变更可追踪

  ## 7) 权限路径

  1. 无 service:approve 用户访问 inbox approve
  2. 返回 approval_permission_denied

  验收：

  - 状态不变
  - 写入审计拒绝事件

  ## 8) 断线恢复路径

  1. AI 发起后用户离线
  2. Inbox 完成审批
  3. 用户回到 AI 会话
  4. 会话自动补齐最新状态

  验收：

  - AI 能恢复到最终 release 状态
  - 无重复 apply

  ## 9) 回滚路径

  3. AI 与 Inbox 可见对应事件（如需审批则进入 inbox）
  4. 最终 rollback 成功

  验收：

  - 事件链闭环：rollback_requested -> rollback_completed

  ## 10) 外部集群路径

  1. 外部 kubeconfig 导入成功
  2. AI 发起发布
  3. 审批通过后应用成功
  4. Deployment 详情可查询诊断与时间线

  验收：

  - cluster.import_kubeconfig_succeeded 与 release 时间线可关联

  ## 11) Compose 路径

  1. AI 发起 compose 发布
  2. preview/apply/approval
  3. compose up 成功
  4. Deployment 与 Inbox 同步

  验收：

  - runtime_type=compose 全链路事件完整

  ## 12) 异常注入路径

  - SSH 不可达 -> ssh_unreachable
  - 物料缺失 -> artifact_missing
  - kubeconfig 无效 -> kubeconfig_invalid

  验收：

  - 错误码一致
  - 三端提示语一致且可操作（有下一步建议）

  如果你继续，我可以给你一份“测试数据种子模板”（最小主机/项目/服务/目标/角色集），方便一次性跑完这 12 组用例。


› 继续


• 最小测试数据种子模板如下，可覆盖你前面 12 组回归路径。

  ## 1) 项目与团队

  projects:
    - id: 11
      name: "proj-alpha"
  teams:
    - id: 7
      name: "team-app"
      project_id: 11

  ## 2) 用户与角色

  users:
    - id: 12
      username: "dev_user"
      team_id: 7
    - id: 3
      username: "approver_user"
      team_id: 7
    - id: 1
      username: "admin"

  permissions:
    - "deploy:release:read"
    - "deploy:release:apply"
    - "deploy:release:approve"
    - "deploy:k8s:apply"
    - "deploy:compose:apply"

  role_bindings:
    - user: 12
      roles: ["operator"]      # apply/read，不含 approve
    - user: 3
      roles: ["approver"]      # read/approve
    - user: 1
      roles: ["admin"]

  ## 3) 主机（SSH）

  hosts:
    - id: 101
      name: "cp-1"
      ip: "10.0.0.11"
      ssh_user: "root"
      auth: "key"
      status: "active"
    - id: 102
      name: "wk-1"
      ip: "10.0.0.12"
      ssh_user: "root"
      auth: "key"
      status: "active"
    - id: 103
      name: "wk-2"
      ip: "10.0.0.13"
      ssh_user: "root"
      auth: "key"
      status: "active"
    - id: 201
      name: "compose-1"
      ip: "10.0.1.21"
      ssh_user: "root"
      auth: "key"
      status: "active"

  ## 4) 环境

  environments:
    - id: 301
      name: "env-k8s-prod"
      runtime_type: "k8s"
      mode: "platform_managed"
      project_id: 11
      team_id: 7
      status: "ready"
    - id: 401
      name: "env-compose-prod"
      runtime_type: "compose"
      mode: "platform_managed"
      project_id: 11
      team_id: 7
      status: "ready"

  ## 5) 环境节点映射

  environment_nodes:
    - environment_id: 301
      host_id: 101
      node_role: "control_plane"
    - environment_id: 301
      host_id: 102
      node_role: "worker"
    - environment_id: 301
      host_id: 103
      node_role: "worker"
    - environment_id: 401
      host_id: 201
      node_role: "compose_manager"

  ## 6) 集群连接

  clusters:
    - id: 88
      name: "k8s-prod-platform"
      source_type: "platform_created"
      auth_type: "kubeconfig"
      endpoint: "https://10.0.0.11:6443"

  cluster_connections:
    - cluster_id: 88
      source_type: "platform_created"
      auth_type: "kubeconfig"
      kubeconfig_ref: "cluster-88-admin"

  ## 7) 外部集群样本

  external_clusters:
    - id: 99
      name: "k8s-prod-external"
      source_type: "external_imported"
      auth_type: "kubeconfig"
      endpoint: "https://10.2.0.10:6443"

  ## 8) 服务与目标

  services:
    - id: 501
      name: "svc-a"
      env: "production"
      runtime_type: "k8s"
    - id: 502
      name: "svc-b"
      env: "production"
      runtime_type: "compose"

  deploy_targets:
    - id: 801
      name: "target-k8s-prod"
      target_type: "k8s"
      runtime_type: "k8s"
      cluster_id: 88
      project_id: 11
      team_id: 7
      env: "production"
    - id: 901
      name: "target-compose-prod"
      target_type: "compose"
      runtime_type: "compose"
      project_id: 11
      team_id: 7
  ## 9) 物料清单（用于安装）

  script_manifest:
    version: "v1.30.1-offline.1"
    runtime:
      - "k8s"
      - "compose"
    artifacts:
      - "script/bin/linux-amd64/kubeadm"
      - "script/bin/linux-amd64/kubelet"
      - "script/bin/linux-amd64/kubectl"
      - "script/bin/linux-amd64/docker"
      - "script/bin/linux-amd64/docker-compose"
    hooks:
      - "script/hooks/preflight.sh"
      - "script/hooks/install-k8s-controlplane.sh"
      - "script/hooks/install-k8s-worker.sh"
      - "script/hooks/install-compose.sh"
      - "script/hooks/post-install-verify.sh"

  ## 10) 预置异常样本（负例）

  fault_injection:
    - case: "ssh_unreachable"
      host_id: 103
      override_status: "offline"
    - case: "artifact_missing"
      remove_path: "script/bin/linux-amd64/kubeadm"
    - case: "kubeconfig_invalid"
      cluster_id: 99
      kubeconfig: "invalid-content"

  这套种子够你完整跑：

  - k8s/compose 双路径
  - platform/external 双集群来源
  - approval 成功/拒绝
  - preview required/expired/mismatch
  - ssh/artifact/kubeconfig 三类失败注入

  如果你继续，我给你最后一份“回归执行顺序脚本（按步骤跑，不写代码）”，保证一次跑完不乱序。


› 继续


• 下面是“回归执行顺序脚本（人工执行版）”，按这个顺序跑最稳。

  ## Phase 0: 基础健康检查

  1. 校验数据种子已就绪（项目、团队、用户、服务、目标、主机、环境）。
  2. 校验 script/manifests/index.yaml 与 hooks/binaries 路径存在。
  3. 校验 admin、dev_user、approver_user 三账号可登录。

  通过标准：

  - 能列出环境、目标、服务、主机
  - 无缺失物料报错

  ## Phase 1: 环境安装链路（k8s + compose）

  1. 对 env-k8s-prod 执行 install preview。
  2. 执行 install apply，观察 task 分步状态直至 succeeded。
  3. 对 env-compose-prod 重复 1-2。

  通过标准：

  - install task 全部步骤有状态与日志
  - 失败时返回明确 error_key

  ## Phase 2: 集群接入链路（平台/外部）

  1. 平台创建集群后检查 connection 信息可读。
  2. 外部 kubeconfig 导入并 connect test。
  3. 外部 cert 导入并 connect test。

  通过标准：

  - platform_created 与 external_imported 都能通过连接探测
  - 连通失败给出 cluster_connect_failed|kubeconfig_invalid|certificate_invalid

  ## Phase 3: 发布主路径（AI 发起）

  1. 用 dev_user 在 AI 命令中心发起 svc-a 生产发布。
  2. 验证必须先 preview，再 confirm。
  3. apply 后进入 pending_approval，生成 ticket。
  4. approver_user 在 Inbox approve。
  5. 回到 AI 与 Deployment 查看最终 applied。

  通过标准：

  - 三端 ticket/release_id 一致
  - 事件链完整：previewed -> pending_approval -> approved -> applying -> applied

  ## Phase 4: 审批拒绝路径

  1. AI 发起 svc-a 新一次发布。
  2. approver_user reject。
  3. AI/Deployment 同步看到 rejected。

  通过标准：

  - 无 applying 事件
  - reject comment 可见

  ## Phase 5: Preview 强约束负例

  1. 直接 apply（无 preview_ref）→ 预期 preview_required
  2. preview 后等待过期再 apply → 预期 preview_expired
  3. preview 后改参数 apply → 预期 preview_mismatch

  通过标准：

  - 三个错误码稳定
  - 不生成执行任务

  ## Phase 6: Compose 发布路径

  1. AI 发起 svc-b compose 生产发布。
  2. 走审批并 approve。
  3. 验证 compose 侧事件与 Deployment 展示。

  通过标准：

  - runtime_type=compose 全链路状态一致
  - timeline 有 compose 事件

  1. 选择已 applied 的 release 发起 rollback。
  2. 观察 rollback 结果与事件链。

  通过标准：

  - rollback 状态正确（成功或失败均有明确事件）
  - 事件链含 rollback_requested / completed|failed

  ## Phase 8: 权限与隔离

  1. 用 dev_user 尝试 approve ticket → 预期 approval_permission_denied
  2. 验证跨项目/跨团队 ticket 不可见不可操作（如有多租户种子）。

  通过标准：

  - 权限拒绝准确
  - 无越权可见性

  ## Phase 9: 异常注入

  1. 注入 ssh_unreachable 再跑 install/apply。
  2. 注入 artifact_missing 再跑 install preview。
  3. 注入 kubeconfig_invalid 再跑 connect test。

  通过标准：

  - 错误码与 hint 正确
  - timeline 记录失败原因

  ## Phase 10: 一致性收尾

  1. 抽样 3 个 release 对比 AI/Inbox/Deployment timeline。
  2. 抽样 2 个 install task 对比 task 详情与日志顺序。
  3. 统计 pending/expired ticket 数与 inbox stats 对齐。

  通过标准：

  - 三端数据一致
  - 无“状态不一致”或“孤儿 ticket”

  如果你还要继续，我可以给你一个“验收打分表模板（Pass/Fail + 证据链接）”，你跑完后直接形成学习项目的发布报告。


› 继续


• 验收打分表模板如下，直接可用。

  ## 1) 总览页（Release Readiness）

  项目: k8s-manage
  版本/分支: __________________
  测试日期: ___________________
  执行人: _____________________

  总分: ____ / 100
  结论:
  [ ] Go  Live
  [ ] Conditional Go (需修复低风险项)
  [ ] No Go

  ## 2) 维度权重

  - 环境安装链路（k8s+compose）: 20
  - 集群接入（平台+外部）: 15
  - 发布主路径（preview->approval->apply）: 20
  - 三端一致性（AI/Inbox/Deployment）: 15
  - 权限与审批治理: 10
  - 异常处理与错误码: 10
  - 回滚与可恢复性: 10

  ## 3) 评分规则

  - Pass: 满分
  - Pass with issue: 50%
  - Fail: 0

  ## 4) 用例记录模板

  [Case ID] C-01
  名称: AI 发起生产发布主路径
  优先级: P0
  结果: Pass / Pass with issue / Fail
  得分: __ / __

  前置条件:
  - ___________________________

  执行步骤:
  1. _________________________
  2. _________________________
  3. _________________________

  预期结果:
  - ___________________________

  实际结果:
  - ___________________________

  证据:
  - 日志/截图/请求ID: __________________
  - release_id / ticket: ________________

  问题单:
  - BUG-xxx (可选)
  备注:
  - ___________________________

  ## 5) 关键 P0 清单（必须全 Pass）

  - C-01: AI 主路径（preview->apply->pending_approval->approved->applied）
  - C-02: 审批拒绝路径（无 apply）
  - C-03: preview 强约束三负例（required/expired/mismatch）
  - C-04: 权限拒绝（无 approve 权限不可批）
  - C-05: 三端时间线一致（AI/Inbox/Deployment）
  - C-06: 回滚路径可执行并有审计

  门禁规则

  - 任一 P0 Fail => No Go

  ## 6) 缺陷分级建议

  - S0: 数据破坏/越权审批/错误发布
  - S1: 主流程阻断（无法发布或审批）
  - S2: 次流程问题（日志、展示、弱一致）
  - S3: 文案/样式/非阻断体验问题

  ## 7) 风险登记模板

  风险ID: R-01
  描述: ________________________
  影响: 高/中/低
  概率: 高/中/低
  缓解措施: ____________________
  负责人: ______________________
  截止时间: ____________________
  状态: Open / Mitigated / Closed

  ## 8) 最终发布摘要模板

  发布结论: Go / Conditional Go / No Go

  通过情况:
  - P0: __/__ 通过
  - P1: __/__ 通过
  - 总缺陷: __ (S0:_, S1:_, S2:_, S3:_)

  主要风险:
  1. ___________________________
  2. ___________________________

  上线前必做:
  1. ___________________________
  2. ___________________________

  建议上线窗口:
  - 日期: ______________________
  - 回滚窗口: __________________


› 再加一个需求就是对于一个集群通常都会用LVS和VIP，所以前端应该提供这个功能，如果用户选了之后就在集群创建之后部署相应的应用


• 这个需求很合理，而且是生产集群常见刚需。可以把它定义成“可选网络高可用引导能力”：

  创建集群
    -> 选择网络入口模式
       [A] NodePort/Ingress(默认)
       [B] LVS + VIP
    -> 若选 B:
       自动创建“后置部署任务”
         - keepalived / lvs 组件
         - VIP 漂移配置
         - 健康检查与回切策略

  建议加三条硬规则：

  1. LVS+VIP 必须是可选项，不默认开启
  2. 只有在满足前置条件时允许勾选（至少 2 节点、同网段、已配置 SSH）
  3. 集群创建成功后才触发部署（异步 task），失败不应回滚已创建集群，只标记“网络入口未完成”

  你这个需求应落到当前 change 的这些位置：

  - design.md
      - 增加决策：网络入口模式（默认/HA）
      - 增加风险：VIP 冲突、漂移脑裂、健康检查误判
  - specs/deployment-management-blueprint/spec.md
      - 新增 requirement：Cluster creation SHALL support optional LVS+VIP bootstrap
  - tasks.md
      - 新增任务组：LVS/VIP 前端配置、后置部署任务、状态回传与告警

  还需要你确认 4 个关键参数（不然后续实现会反复改）：

  1. VIP 是单个还是按环境多 VIP？
  2. 用 keepalived+ipvsadm 还是 kube-vip？
  3. 健康检查目标是 API Server :6443 还是 Ingress :80/:443？
  4. 失败策略是“自动重试 N 次”还是“人工确认再重试”？

  如果你同意，我下一步可以把这 4 点整理成一页“LVS/VIP 决策表”，你选项后就能直接写入 OpenSpec。


› 同意


• 好，给你一版可直接选的 LVS/VIP 决策表。

  ## LVS/VIP 决策表（建议默认项已标注）

  1. VIP 模型

  - A. 单 VIP（推荐）
      - 一个环境一个 VIP，先跑通主链路
  - B. 多 VIP（API/Ingress 分离）
      - API 一个 VIP，Ingress 一个/多个 VIP

  2. 组件方案

  - A. keepalived + ipvsadm（推荐）
      - 更云原生，配置相对简化

  3. 健康检查目标

  - A. 先保 API Server :6443（推荐）
      - 集群基础可用优先
  - B. 同时覆盖 Ingress :80/:443
      - 更完整但复杂度更高

  4. 失败恢复策略

  - A. 自动重试 3 次后人工介入（推荐）
      - 平衡自动化与风险
  - B. 全自动持续重试
      - 简单但可能隐藏故障
  - C. 完全人工重试
      - 可控但效率低

  5. 前端创建页交互

  - A. “网络入口模式”单选（推荐）
      - 标准模式 / LVS+VIP
  - B. 高级开关 + 折叠配置
      - 适合参数多的场景

  6. 最小参数集（前端必填）

  - vip_address
  - interface
  - virtual_router_id
  - auth_pass
  - backend_nodes[]
  - health_check_port（默认 6443）

  ———

  如果你认可，我建议直接定这套默认组合作为 Phase-1：

  - 单 VIP + keepalived+ipvsadm + 检查 6443 + 重试3次后人工

  回复“按推荐组合”，我就继续给你“这套方案的 API 请求/响应草案”。


› 按推荐组合


• 按推荐组合，下面是 API 草案。

  ## 1) 创建集群（带网络入口模式）

  POST /api/v1/clusters

  {
    "name": "prod-k8s-a",
    "project_id": 11,
    "team_id": 7,
    "mode": "platform_managed",
    "bootstrap": {
      "runtime": "k8s",
      "control_plane_host_id": 101,
      "worker_host_ids": [102, 103],
      "cni": "calico"
    },
    "network_entry": {
      "mode": "lvs_vip",
      "lvs_vip": {
        "vip_address": "10.0.0.100",
        "interface": "eth0",
        "virtual_router_id": 51,
        "auth_pass": "ReplaceMeStrong",
        "backend_nodes": [101, 102, 103],
        "health_check_port": 6443,
        "retry_policy": {
          "max_retries": 3,
          "backoff_seconds": 20
        }
      }
    }
  }

  响应（集群先创建成功，LVS 走后置任务）：

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "cluster_id": 88,
      "status": "provisioning",
      "network_entry_mode": "lvs_vip",
      "post_tasks": [
        {
          "task_type": "lvs_vip_bootstrap",
          "task_id": "net-task-9001",
          "status": "pending"
        }
      ]
    }
  }

  ## 2) 网络入口预览

  POST /api/v1/clusters/:id/network-entry/preview

  {
    "mode": "lvs_vip",
    "lvs_vip": {
      "vip_address": "10.0.0.100",
      "interface": "eth0",
      "virtual_router_id": 51,
      "backend_nodes": [101, 102, 103],
      "health_check_port": 6443
    }
  }

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "preview_ref": "net-prev-777",
      "checks": [
        { "level": "info", "code": "vip_format_ok", "message": "vip format valid" },
        { "level": "info", "code": "same_subnet_ok", "message": "nodes in same subnet" }
      ],
      "warnings": [],
      "expires_at": "2026-02-26T13:30:00Z"
    }
  }

  ## 3) 网络入口应用（必须 preview_ref）

  POST /api/v1/clusters/:id/network-entry/apply

  {
    "preview_ref": "net-prev-777",
    "mode": "lvs_vip",
    "approval_token": "approve-net-123",
    "confirm": true
  }

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "task_id": "net-task-9001",
      "status": "running",
      "retry_policy": { "max_retries": 3, "backoff_seconds": 20 }
    }
  }

  ## 4) 网络入口任务查询

  GET /api/v1/clusters/:id/network-entry/tasks/:task_id

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "task_id": "net-task-9001",
      "attempt": 1,
      "max_retries": 3,
      "steps": [
        { "name": "preflight.sh", "status": "succeeded" },
        { "name": "install-keepalived.sh", "status": "running" },
        { "name": "configure-ipvs.sh", "status": "pending" },
        { "name": "verify-vip.sh", "status": "pending" }
      ],
      "last_error": ""
    }
  }

  ## 5) 网络入口状态总览

  GET /api/v1/clusters/:id/network-entry

  {
    "code": 1000,
    "msg": "ok",
    "data": {
      "mode": "lvs_vip",
      "status": "ready",
      "vip_address": "10.0.0.100",
      "health_check_port": 6443,
      "active_node": 101,
      "last_check_at": "2026-02-26T12:40:00Z"
    }
  }

  ## 6) 失败码建议（网络入口）

  - 33101 vip_conflict
  - 33102 vip_unreachable
  - 33103 lvs_binary_missing
  - 33104 keepalived_config_invalid
  - 33105 backend_nodes_invalid
  - 33106 network_entry_preview_required
  - 33107 network_entry_preview_expired
  - 33108 network_entry_apply_failed

  如果继续，我可以给你对应的 script/hooks 参数约定（每个 shell 脚本接收哪些环境变量）。


› 继续


• 下面是 script/hooks 参数约定（可直接落地）。

  ## 1) 通用约定

  - 所有 hook 必须：
      - #!/usr/bin/env bash
      - set -euo pipefail
      - 输出结构化日志：[LEVEL] key=value ...
  - 退出码：
      - 0 成功
      - 非 0 失败（平台采集 stderr + exit code）

  ## 2) 平台注入的通用环境变量

  - OPS_TASK_ID
  - OPS_CLUSTER_ID
  - OPS_ENVIRONMENT_ID
  - OPS_RUNTIME_TYPE (k8s|compose)
  - OPS_PROJECT_ID
  - OPS_TEAM_ID
  - OPS_WORKDIR（如 /opt/opspilot/tasks/$OPS_TASK_ID）
  - OPS_SCRIPT_ROOT（如 /opt/opspilot/script）
  - OPS_RETRY_ATTEMPT
  - OPS_RETRY_MAX

  ## 3) LVS/VIP 专用变量

  - OPS_NET_MODE (lvs_vip)
  - OPS_VIP_ADDRESS（10.0.0.100）
  - OPS_VIP_INTERFACE（eth0）
  - OPS_VRID（51）
  - OPS_AUTH_PASS
  - OPS_HEALTH_PORT（6443）
  - OPS_BACKEND_NODES（逗号分隔 IP，例如 10.0.0.11,10.0.0.12,10.0.0.13）
  - OPS_PRIMARY_NODE_IP（主节点）
  - OPS_NODE_IP（当前执行节点 IP）

  ## 4) 建议 hooks 与参数需求

  ### hooks/preflight.sh

  用途：校验命令、网卡、端口、VIP 冲突、SSH 连通
  依赖变量：

  - 通用变量
  - OPS_VIP_ADDRESS
  - OPS_VIP_INTERFACE
  - OPS_HEALTH_PORT
  - OPS_BACKEND_NODES

  输出建议：

  [INFO] check=vip_conflict result=ok
  [INFO] check=interface_exists result=ok interface=eth0

  ### hooks/install-keepalived.sh

  用途：安装 keepalived 二进制/包
  依赖变量：

  - 通用变量
  - OPS_SCRIPT_ROOT
  - OPS_NODE_IP

  ### hooks/configure-ipvs.sh

  用途：写 keepalived 配置 + ipvs 规则
  依赖变量：

  - 通用变量
  - LVS 全部专用变量

  建议平台写模板文件路径变量：

  - OPS_KEEPALIVED_TEMPLATE（如 script/templates/lvs/keepalived.conf.tpl）

  ### hooks/reload-network-entry.sh

  用途：重载 keepalived，应用配置
  依赖变量：

  - 通用变量

  ### hooks/verify-vip.sh

  用途：验证 VIP 漂移与 6443 健康检查
  依赖变量：

  - OPS_VIP_ADDRESS
  - OPS_HEALTH_PORT
  - OPS_BACKEND_NODES

  输出建议：

  [INFO] check=vip_bind result=ok active_node=10.0.0.11
  [INFO] check=health_port result=ok port=6443

  ### hooks/rollback-network-entry.sh

  用途：回滚 keepalived/ipvs 配置
  依赖变量：

  - 通用变量
  - OPS_WORKDIR（读取备份配置）

  除 env 外，平台额外下发 JSON：

  - 路径：$OPS_WORKDIR/context.json

  示例：

  {
    "task_id": "net-task-9001",
    "cluster_id": 88,
    "network_entry": {
      "mode": "lvs_vip",
      "vip_address": "10.0.0.100",
      "interface": "eth0",
      "virtual_router_id": 51,
      "health_check_port": 6443,
      "backend_nodes": [
        {"host_id": 101, "ip": "10.0.0.11"},
        {"host_id": 102, "ip": "10.0.0.12"},
        {"host_id": 103, "ip": "10.0.0.13"}
      ]
    }
  }

  ## 6) 目录建议

  script/
    hooks/
      preflight.sh
      install-keepalived.sh
      configure-ipvs.sh
      reload-network-entry.sh
      verify-vip.sh
      rollback-network-entry.sh
    templates/
      lvs/
        keepalived.conf.tpl
    manifests/
      index.yaml

  如果继续，我可以给你 keepalived.conf.tpl 的字段规范（模板变量清单，不写实现内容）。


› 继续


• keepalived.conf.tpl 字段规范如下（模板变量清单）。

  ## 1) 模板目标

  用于生成每个节点的 keepalived.conf，支持：

  - VRRP 主备漂移
  - 对后端节点的 TCP 健康检查（默认 6443）
  - VIP 绑定指定网卡

  ## 2) 必填模板变量

  - {{ INSTANCE_NAME }}
  - {{ STATE }} (MASTER|BACKUP)
  - {{ INTERFACE }}
  - {{ VIRTUAL_ROUTER_ID }}
  - {{ PRIORITY }}
  - {{ AUTH_PASS }}
  - {{ VIP_ADDRESS }}
  - {{ VIP_CIDR }}（如 24）
  - {{ HEALTH_PORT }}（默认 6443）
  - {{ LB_KIND }}（NAT|DR|TUN，推荐 DR）
  - {{ LB_ALGO }}（rr|wrr|lc|wlc）
  - {{ REAL_SERVERS_BLOCK }}（平台拼接）

  ## 3) 可选模板变量

  - {{ ADVERT_INT }}（默认 1）
  - {{ UNICAST_SRC_IP }}
  - {{ UNICAST_PEERS_BLOCK }}
  - {{ GARP_MASTER_DELAY }}（默认 1）
  - {{ GARP_MASTER_REPEAT }}（默认 2）
  - {{ NOTIFY_MASTER_CMD }}

  ## 4) REAL_SERVERS_BLOCK 结构（建议）

  每个 backend 一段：

  - {{ REAL_SERVER_IP }}
  - {{ REAL_SERVER_PORT }}
  - {{ REAL_SERVER_WEIGHT }}
  - {{ CHECK_CONNECT_TIMEOUT }}
  - {{ CHECK_RETRY }}
  - {{ CHECK_DELAY_BEFORE_RETRY }}

  ## 5) 渲染规则建议

  1. MASTER 节点 priority 最大（如 150），备节点递减（如 140,130）
  2. vip_address 必须带 CIDR，最终输出 x.x.x.x/24 dev eth0
  3. 若配置了 unicast，则必须同时输出 unicast_src_ip + unicast_peer
  4. auth_pass 长度建议 8~32
  5. health check 仅检查 TCP 6443（你当前推荐方案）

  ## 6) 生成前校验（平台侧）

  - INTERFACE 在目标节点存在
  - VIP_ADDRESS 不与已有地址冲突
  - VIRTUAL_ROUTER_ID 在同网段不冲突
  - REAL_SERVER 列表至少 2 个
  - HEALTH_PORT 为合法端口

  ## 7) 产物与审计

  - 渲染后文件路径建议：/etc/keepalived/keepalived.conf
  - 备份旧配置：/etc/keepalived/keepalived.conf.bak.<ts>
  - 记录审计 detail：
      - vrid
      - vip
      - interface
      - backend_count
      - health_port
      - node_role(master|backup)

  如果继续，我给你下一份“index.yaml 元数据规范”（把 binary/image/template/checksum 一次性规范好）。


› 继续


• 下面是 script/manifests/index.yaml 元数据规范（建议作为唯一安装物料入口）。

  ## 1) 顶层字段

  - apiVersion: opspilot.manifest/v1
  - kind: InstallBundleIndex
  - bundleVersion: 如 v1.30.1-offline.1
  - generatedAt
  - defaultOS: linux
  - defaultArch: amd64
  - runtimes: k8s|compose
  - artifactsRoot: script/

  ## 2) runtime 结构

  每个 runtime 一段：

  - name: k8s 或 compose
  - version
  - supportedOSArch:
      - os
      - arch
  - dependencies:
      - 依赖 runtime 或组件名
  - binaries: 二进制列表
  - images: 离线镜像包列表
  - templates: 模板列表
  - hooks: hooks 执行顺序
  - checks: 预检规则列表

  ## 3) artifact 通用字段

  - id: 唯一键（如 kubeadm-linux-amd64-v1.30.1）
  - type: binary|image-archive|template|script|checksum
  - path: 相对 script/ 的路径
  - sha256
  - sizeBytes
  - mode: 文件权限（如 0755）
  - required: true|false
  - description

  ## 4) hooks 定义字段

  - name（如 preflight.sh）
  - path（如 hooks/preflight.sh）
  - timeoutSeconds
  - retryable (true|false)
  - maxRetries
  - runOn (control_plane|worker|all|compose_manager)
  - inputs（所需变量列表）
  - outputs（约定输出键）

  ## 5) checks 定义字段

  - id
  - type：cmd|file_exists|port_open|ip_conflict|subnet_match
  - target
  - expect
  - severity: error|warning|info
  - message

  ## 6) k8s runtime 最小条目建议

  - binaries:
      - kubectl, kubeadm, kubelet, crictl, containerd, runc, cni-plugins.tgz
  - images:
      - k8s-core-images.tar
      - cni-images.tar
  - templates:
      - templates/kubeadm/init-config.yaml.tpl
      - templates/kubeadm/join-config.yaml.tpl
  - hooks:
      - preflight.sh
      - post-install-verify.sh

  ## 7) compose runtime 最小条目建议

  - binaries:
      - docker, docker-compose（或 compose plugin）
  - templates:
      - templates/compose/docker-compose.yml.tpl
  - hooks:
      - preflight.sh
      - install-compose.sh
      - post-install-verify.sh

  ## 8) lvs_vip feature flag 段（你当前需求）

  - features.lvsVip.enabled: true
  - features.lvsVip.requiredArtifacts
      - hooks/install-keepalived.sh
      - hooks/configure-ipvs.sh
      - templates/lvs/keepalived.conf.tpl
  - features.lvsVip.requiredInputs
      - vip_address, interface, virtual_router_id, auth_pass, backend_nodes, health_check_port

  ## 9) 签名与完整性建议

  - checksums:
      - 指向 checksums/SHA256SUMS
  - signature:
      - algo: ed25519
      - path: checksums/SHA256SUMS.sig
      - pubkeyRef

  ## 10) 版本兼容策略

  - compatibility:
      - k8sApiMin
      - k8sApiMax
      - requiresKernelMin
      - notes

  如果继续，我下一步给你“完整 index.yaml 示例骨架”（可直接拷贝填值）。


› 继续


• apiVersion: opspilot.manifest/v1
  kind: InstallBundleIndex

  bundleVersion: v1.30.1-offline.1
  generatedAt: "2026-02-26T13:00:00Z"
  defaultOS: linux
  defaultArch: amd64
  artifactsRoot: script/

  checksums:
    algo: sha256
    path: checksums/SHA256SUMS

  signature:
    algo: ed25519
    path: checksums/SHA256SUMS.sig
    pubkeyRef: checksums/public.key

  features:
    lvsVip:
      enabled: true
      requiredArtifacts:
        - hooks/install-keepalived.sh
        - hooks/configure-ipvs.sh
        - hooks/reload-network-entry.sh
        - hooks/verify-vip.sh
        - hooks/rollback-network-entry.sh
        - templates/lvs/keepalived.conf.tpl
      requiredInputs:
        - vip_address
        - interface
        - virtual_router_id
        - auth_pass
        - backend_nodes
        - health_check_port

  runtimes:
    - name: k8s
      version: v1.30.1
      supportedOSArch:
        - os: linux
          arch: amd64
      dependencies:
        - containerd
        - cni-plugins

      binaries:
        - id: kubectl-linux-amd64-v1.30.1
          type: binary
          path: bin/linux-amd64/kubectl
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0755"
          required: true
          description: kubectl client

        - id: kubeadm-linux-amd64-v1.30.1
          type: binary
          path: bin/linux-amd64/kubeadm
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0755"
          required: true
          description: kubeadm installer

        - id: kubelet-linux-amd64-v1.30.1
          type: binary
          path: bin/linux-amd64/kubelet
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0755"
          required: true
          description: kubelet binary

        - id: crictl-linux-amd64-v1.30.1
          type: binary
          path: bin/linux-amd64/crictl
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0755"
          required: true
          description: cri client

        - id: containerd-linux-amd64-v1.7.x
          type: binary
          path: bin/linux-amd64/containerd
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0755"
          required: true
          description: container runtime

        - id: runc-linux-amd64-v1.1.x
          type: binary
          path: bin/linux-amd64/runc
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0755"
          required: true
          description: OCI runtime

        - id: cni-plugins-linux-amd64-v1.4.x
          type: image-archive
          path: bin/linux-amd64/cni-plugins.tgz
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0644"
          required: true
          description: cni plugins archive

      images:
        - id: k8s-core-images-v1.30.1
          type: image-archive
          path: images/k8s-v1.30.1/k8s-core-images.tar
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0644"
          required: true
          description: kube-apiserver/controller/scheduler/proxy/coredns/pause/etcd

        - id: cni-images-calico-v3.xx
          type: image-archive
          path: images/k8s-v1.30.1/cni-images.tar
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0644"
          required: true
          description: cni image bundle

      templates:
        - id: kubeadm-init-template
          type: template
          path: templates/kubeadm/init-config.yaml.tpl
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0644"
          required: true
          description: kubeadm init config template

        - id: kubeadm-join-template
          type: template
          path: templates/kubeadm/join-config.yaml.tpl
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0644"
          required: true
          description: kubeadm join config template

        - id: keepalived-template
          type: template
          path: templates/lvs/keepalived.conf.tpl
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0644"
          required: false
          description: keepalived config template for lvs+vip

      hooks:
        - name: preflight.sh
          path: hooks/preflight.sh
          timeoutSeconds: 300
          retryable: false
          maxRetries: 0
          runOn: all
          inputs:
            - OPS_TASK_ID
            - OPS_RUNTIME_TYPE
          outputs:
            - preflight_summary

        - name: install-k8s-controlplane.sh
          path: hooks/install-k8s-controlplane.sh
          timeoutSeconds: 1800
          retryable: true
          maxRetries: 1
          runOn: control_plane
          inputs:
            - OPS_TASK_ID
            - OPS_RUNTIME_TYPE
            - OPS_WORKDIR
          outputs:
            - kubeadm_init_result
            - kubeconfig_path

        - name: install-k8s-worker.sh
          path: hooks/install-k8s-worker.sh
          timeoutSeconds: 1200
          retryable: true
          maxRetries: 1
          runOn: worker
          inputs:
            - OPS_TASK_ID
            - OPS_RUNTIME_TYPE
            - OPS_WORKDIR
          outputs:
            - kubeadm_join_result

        - name: install-keepalived.sh
          path: hooks/install-keepalived.sh
          timeoutSeconds: 600
          retryable: true
          maxRetries: 1
          runOn: control_plane
          inputs:
            - OPS_NET_MODE
            - OPS_VIP_ADDRESS
          outputs:
            - keepalived_install_result

        - name: configure-ipvs.sh
          path: hooks/configure-ipvs.sh
          timeoutSeconds: 600
          retryable: true
          maxRetries: 1
          runOn: control_plane
          inputs:
            - OPS_VIP_ADDRESS
            - OPS_VIP_INTERFACE
            - OPS_VRID
            - OPS_AUTH_PASS
            - OPS_BACKEND_NODES
            - OPS_HEALTH_PORT
          outputs:
            - keepalived_conf_path

        - name: reload-network-entry.sh
          path: hooks/reload-network-entry.sh
          timeoutSeconds: 180
          retryable: true
          maxRetries: 2
          runOn: control_plane
          inputs:
            - OPS_TASK_ID
          outputs:
            - reload_result

        - name: verify-vip.sh
          path: hooks/verify-vip.sh
          timeoutSeconds: 180
          retryable: true
          maxRetries: 2
          runOn: control_plane
          inputs:
            - OPS_VIP_ADDRESS
            - OPS_HEALTH_PORT
          outputs:
            - active_node
            - vip_health

        - name: rollback-network-entry.sh
          path: hooks/rollback-network-entry.sh
          timeoutSeconds: 300
          retryable: false
          maxRetries: 0
          runOn: control_plane
          inputs:
            - OPS_WORKDIR
          outputs:
            - rollback_result

        - name: post-install-verify.sh
          path: hooks/post-install-verify.sh
          timeoutSeconds: 300
          retryable: true
          maxRetries: 2
          runOn: all
          inputs:
            - OPS_TASK_ID
            - OPS_RUNTIME_TYPE
          outputs:
            - verify_summary

      checks:
        - id: ssh-reachable
          type: cmd
          target: "ssh_connect"
          expect: "success"
          severity: error
          message: "all nodes must be reachable"

        - id: required-binaries
          type: file_exists
          target: "bin/linux-amd64"
          expect: "all_present"
          severity: error
          message: "required k8s binaries must exist"

        - id: required-images
          type: file_exists
          target: "images/k8s-v1.30.1"
          expect: "all_present"
          severity: error
          message: "required k8s image archives must exist"

        - id: vip-conflict
          type: ip_conflict
          target: "vip_address"
          expect: "no_conflict"
          severity: error
          message: "vip must not conflict with existing addresses"

    - name: compose
      version: v2.24.x
      supportedOSArch:
        - os: linux
          arch: amd64
      dependencies:
        - docker

      binaries:
        - id: docker-linux-amd64-v24.x
          type: binary
          path: bin/linux-amd64/docker
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0755"
          required: true
          description: docker engine binary

        - id: docker-compose-linux-amd64-v2.24.x
          type: binary
          path: bin/linux-amd64/docker-compose
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0755"
          required: true
          description: docker compose cli

      templates:
        - id: compose-template
          type: template
          path: templates/compose/docker-compose.yml.tpl
          sha256: "<FILL>"
          sizeBytes: 0
          mode: "0644"
          required: true
          description: compose deployment template

      hooks:
        - name: preflight.sh
          path: hooks/preflight.sh
          timeoutSeconds: 300
          retryable: false
          maxRetries: 0
          runOn: all
          inputs:
            - OPS_TASK_ID
            - OPS_RUNTIME_TYPE
          outputs:
            - preflight_summary

        - name: install-compose.sh
          path: hooks/install-compose.sh
          timeoutSeconds: 900
          retryable: true
          maxRetries: 1
          runOn: compose_manager
            - OPS_RUNTIME_TYPE
          retryable: true
          maxRetries: 2
          runOn: all
          inputs:
            - OPS_TASK_ID
            - OPS_RUNTIME_TYPE
          outputs:
            - verify_summary

      checks:
        - id: ssh-reachable
          type: cmd
          target: "ssh_connect"
          expect: "success"
          severity: error
          message: "compose nodes must be reachable"

        - id: required-binaries
          type: file_exists
          target: "bin/linux-amd64"
          expect: "all_present"
          severity: error
          message: "docker and compose binaries must exist"

  compatibility:
    k8sApiMin: "v1.28.0"
    k8sApiMax: "v1.31.x"
    requiresKernelMin: "4.19"
    notes:
      - "lvs_vip mode assumes keepalived+ipvsadm available or installable"
      - "preview/apply requires manifest and checksum validation before execution"

  如果你要，我下一步给你一份“你现在该放到 script 目录的最小文件清单（逐文件）”。