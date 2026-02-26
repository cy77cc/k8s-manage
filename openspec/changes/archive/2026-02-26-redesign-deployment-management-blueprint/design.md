## Context

现有部署管理已覆盖目标管理、发布、审批、回滚与部分诊断，但能力定义仍按功能点散落在 deployment/cicd/ai 页面与接口中，缺少统一蓝图来约束跨运行时语义一致性。产品侧希望把部署管理升级为统一操作面，支持用户通过页面与 AI 命令中心两种入口完成同一条发布链路，并保持审批、审计、观测一致。

## Goals / Non-Goals

**Goals:**
- 形成部署管理蓝图能力分层：Target、Release、Governance、Observability、AI Command Bridge。
- 统一 Compose 与 K8s 的发布生命周期语义和 API 状态词汇。
- 明确生产环境审批门禁、审计追踪、时间线可视化要求。
- 给出分阶段落地路径和验收标准，作为后续实现 change 的输入。

**Non-Goals:**
- 本 change 不直接完成所有代码重构与页面重写。
- 不在本阶段引入新的部署执行引擎（如 ArgoCD/Spinnaker）。
- 不替换既有 RBAC/认证体系，只定义蓝图下的权限要求。

## Decisions

1. 采用“蓝图能力域 + 运行时适配层”模型。
- Why: 能把产品能力与运行时实现解耦，避免每次新增 runtime 都复制一套流程。
- Alternative: 按 K8s/Compose 分别维护全栈能力；缺点是交互与接口持续分叉。

2. 生命周期统一为 `preview -> pending_approval/approved -> applying -> applied|failed -> rollback`。
- Why: 可同时覆盖 UI 状态、审计事件和 API 返回，便于前后端对齐。
- Alternative: 继续沿用各模块自定义状态；缺点是跨模块追踪困难。

3. 生产变更必须走审批票据并写入审计事件流。
- Why: 满足高风险环境可追溯与合规要求。
- Alternative: 仅基于角色放行；缺点是缺少动作级审计链。

4. AI 命令中心作为同级操作入口，复用同一审批与审计链路。
- Why: 避免“页面操作可审计、AI 操作不可审计”的双轨风险。
- Alternative: AI 使用独立审批机制；缺点是心智与治理成本高。

5. 所有发布动作采用“先预览再确认”策略，禁止绕过预览直接执行 apply。
- Why: 先暴露风险与变更影响，降低误发布概率，并让审批依据统一。
- Alternative: 允许高权限用户直接 apply；缺点是高风险环境误操作成本过高。

6. 审批入口采用 AI/全局统一审批收件箱，不按页面来源拆分审批流。
- Why: 审批治理需要跨入口统一可见、可追溯、可接管。
- Alternative: Deployment 页面与 AI 页面各自维护审批列表；缺点是审批割裂和漏审风险高。

7. 默认入口面向普通项目组用户，控制台高级能力按需下沉。
- Why: 大多数用户目标是“完成发布任务”，不是“管理全部平台配置”。
- Alternative: 默认进入全量运维控制台；缺点是学习成本高、路径长、误触发概率增加。

## Risks / Trade-offs

- [Risk] 蓝图覆盖范围大，短期需求会与既有实现产生过渡期不一致。
  - Mitigation: 使用阶段化任务与 capability delta spec，先统一术语与契约，再逐步重构实现。
- [Risk] 状态机升级可能影响历史数据展示兼容。
  - Mitigation: 定义状态映射与兼容读策略，在 API 响应提供 lifecycle_state。
- [Risk] AI 入口接入审批后，操作链路变长，用户感知变慢。
  - Mitigation: 提供审批前预览与批量审批能力，减少重复确认成本。
- [Risk] 强制 preview 后，发布时效性可能下降。
  - Mitigation: 引入预览结果时效和缓存策略，并提供快速重预览能力。
- [Risk] 全局审批收件箱在多项目并行场景下可能信息过载。
  - Mitigation: 增加按项目/环境/风险等级过滤与默认“与我相关”视图。

## Migration Plan

1. 先合并蓝图与 capability specs，冻结术语与状态机。
2. 后端按 capability 分步改造：release API、approval/audit、timeline 查询。
3. 前端同步升级普通用户默认入口、部署页与命令中心状态呈现。
4. 发布时保留旧状态兼容映射，完成历史数据回放校验后再收敛。
5. 审批入口收敛到全局收件箱后，再移除分散审批入口。

## Open Questions

- 是否需要在 Phase-2 引入多级审批（owner + SRE）策略。
- Compose 运行时是否在蓝图阶段定义标准健康探针模型。
- 预览结果的默认有效期应设置为 15 分钟还是 30 分钟。
