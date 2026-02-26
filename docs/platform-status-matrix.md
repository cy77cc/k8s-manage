# Platform Status Matrix

Last Updated: 2026-02-26  
Update Cadence: once per sprint close or major release milestone

## Matrix Template

| Domain | Status (`Done|In Progress|Risk`) | Evidence | Risks | Next Actions | Owner |
| --- | --- | --- | --- | --- | --- |
| `<domain>` | `<Done/In Progress/Risk>` | `<code/docs path>` | `<key risk>` | `<next step>` | `<team/role>` |

## Snapshot (2026-02-26)

| Domain | Status (`Done|In Progress|Risk`) | Evidence | Risks | Next Actions | Owner |
| --- | --- | --- | --- | --- | --- |
| Auth & User | In Progress | `internal/service/user/routes.go`, `docs/roadmap.md` | 权限模型与业务权限码仍在收敛 | 统一权限字典并补充回归用例 | Backend-RBAC |
| Host Management | In Progress | `internal/service/host/routes.go`, `docs/progress.md` | 多来源接入（cloud/kvm）边界场景多 | 补齐批量执行与文件操作回归测试 | Backend-Host |
| Cluster / K8s | In Progress | `internal/service/cluster/routes.go`, `docs/roadmap.md` | 依赖 kubeconfig 与集群可达性 | 增强 namespace/rollout 的集成验证 | Backend-K8s |
| Service Management | In Progress | `internal/service/service/routes.go`, `docs/roadmap.md` | 模板变量与发布治理耦合高 | 收敛变量/发布契约并补齐 e2e | Backend-Service |
| Deployment Management | In Progress | `internal/service/deployment/routes.go`, `docs/codebase-architecture.md` | 多目标（`k8s|compose`）策略一致性需持续验证 | 完善 compose 与 k8s 的发布回滚基线用例 | Backend-Deploy |
| CI/CD | In Progress | `internal/service/cicd/routes.go` | 审批与审计链路仍需跨模块联调 | 建立 CI 配置到发布审批的端到端回归 | Backend-CICD |
| CMDB | In Progress | `internal/service/cmdb/routes.go` | 资产同步与拓扑一致性风险 | 补充 sync job 失败恢复与审计策略 | Backend-CMDB |
| Monitoring | In Progress | `internal/service/monitoring/routes.go`, `internal/service/monitoring/logic.go`, `internal/service/monitoring/notifier.go` | 当前通知适配器为 skeleton，外部渠道可靠性策略待完善 | 引入真实通知 provider 与重试/退避策略，并补充规则评估可视化 | Backend-Monitoring |
| RBAC | In Progress | `internal/service/rbac/routes.go`, `docs/progress.md` | `admin` 临时全量放行策略待收敛 | 分阶段替换为最小权限模型 | Backend-RBAC |
| AI Assistant | In Progress | `internal/service/ai/routes.go`, `docs/ai/vision-and-capability-map.md` | 审批与执行审计仍有内存态路径 | 推进审批/执行状态持久化与审计查询 | Backend-AI |
| AIOPS | In Progress | `internal/service/aiops/routes.go` | 诊断建议到动作闭环尚不完整 | 完成巡检建议执行闭环与风险分级 | Backend-AIOPS |
| Automation & Topology | In Progress | `internal/service/automation/routes.go`, `internal/service/topology/routes.go`, `openspec/changes/roadmap-phase-automation-topology/specs/automation-topology-phase/spec.md` | 当前为 API skeleton，编排引擎与可视化仍需迭代 | 完善执行编排、图谱聚合策略与前端拓扑可视化联调 | Backend-Platform |
| Project & Legacy Node Adapter | Risk | `internal/service/project/routes.go`, `internal/service/node/routes.go` | Node 兼容路由处于过渡态，治理边界不清晰 | 明确迁移下线计划并补充兼容告警 | Backend-Core |

## Governance Triggers

- PR trigger: when API/domain behavior changes, update this matrix row and evidence in the same PR.
- Release trigger: before each release milestone closes, refresh all `In Progress`/`Risk` rows.
- Review gate: status values must be only `Done`, `In Progress`, or `Risk`.
