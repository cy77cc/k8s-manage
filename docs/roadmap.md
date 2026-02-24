# k8s-manage Roadmap

## 1. Product Goals

- 单进程交付：后端启动后直接提供前端 UI 与 API。
- 核心 MVP：打通 `登录 + 主机 + K8s + 服务 + RBAC`。
- 以 README 规划为主线分阶段演进，持续记录状态与风险。

## 2. Phase Plan

### Phase 1 (MVP) - In Progress

- Frontend hosting in backend (`Go Embed web/dist`)
- Auth & session bootstrap
- Hosts resource APIs + SSH command execution + 3-step onboarding (`probe/create/credentials`)
- Clusters resource APIs + K8s read/deploy preview/apply
- Services CRUD + deploy
- RBAC users/roles/permissions minimum APIs
- AI minimum APIs for K8s pages and global assistant
- Versioned DB migration framework (`migrate up/down/status`)

### Phase 2 - Pending

- 任务调度（Jobs / Executions / Logs）
- 配置中心（Apps / Configs / History）
- 监控告警（Metrics / Alerts / AlertRules）

### Phase 3 - Pending

- CI/CD、CMDB、Automation、Topology
- 多租户配额治理、审计完善
- 生产级 AI Agent 执行动作闭环

## 3. Module Status Matrix

| Module | Status | Notes |
| --- | --- | --- |
| Frontend Embed Serving | Done | `web/dist` embed + SPA fallback |
| Auth | In Progress | login/register/refresh/logout + `auth/me` |
| Hosts | In Progress | CRUD + action + ssh exec + batch + onboarding token flow + cloud/kvm/credentials mvp |
| Clusters / K8s | In Progress | list/create/detail + nodes/pods/deployments/services/ingresses/events/logs + deploy preview/apply |
| Services | In Progress | CRUD + deploy + quota/events/rollback(mvp stub) |
| RBAC | In Progress | admin 全量放行（含 `*:*`），非 admin 继续走数据库权限 |
| AI | In Progress | Eino + Ollama(`glm-5:cloud`) + 控制面（capabilities/tools/approvals/executions）+ SSE tool events |

## 4. API Coverage Matrix (MVP)

| Frontend Page | Backend API Group | Status |
| --- | --- | --- |
| Login/Register | `/api/v1/auth/*` | In Progress |
| Hosts | `/api/v1/hosts/*` | In Progress |
| K8s | `/api/v1/clusters/*` | In Progress |
| Services | `/api/v1/services/*` | In Progress |
| Settings-RBAC | `/api/v1/rbac/*` | In Progress |
| Global AI Assistant | `/api/v1/ai/*` | In Progress |
| Node legacy adapter | `/api/v1/node/add` | In Progress (deprecated, delegating to host domain) |
| Host Credentials | `/api/v1/credentials/ssh_keys*` | In Progress |
| Host Cloud Import | `/api/v1/hosts/cloud/*` | In Progress |
| Host Virtualization | `/api/v1/hosts/virtualization/*` | In Progress |

## 5. Current Decisions (2026-02-24)

- AI chat protocol统一为 SSE（`meta/delta/done/error`）。
- LLM 统一走 `ollama` provider，默认模型 `glm-5:cloud`。
- RBAC admin 采用临时全量放行策略，保障平台管理页可用性。
- AI function calling 默认只读放行，变更动作必须审批后执行。
- AI 工具实现采用 “本地 Eino Tool + mcp-go MCP Tool Proxy” 混合模式。
