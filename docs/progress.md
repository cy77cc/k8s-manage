# Development Progress Log

## 2026-02-24

### Scope

- 实施核心 MVP 第 1 轮：后端渲染前端、基础接口对接、关键模块最小可用。

### Completed

- 修复构建阻塞：移除 `internal/service/service.go` 对缺失 swagger docs 包的硬依赖。
- 新增 `web/embed.go`，使用 `Go Embed` 打包 `web/dist`。
- 新增静态资源路由与 SPA fallback（非 `/api*` 路径回落到 `index.html`）。
- 新增并接入路由组：
  - `/api/v1/hosts/*`
  - `/api/v1/clusters/*`
  - `/api/v1/rbac/*`
  - `/api/v1/ai/*`（chat/sessions/analyze/recommendations/k8s analyze/actions）
- Auth：
  - 新增 `GET /api/v1/auth/me`
  - 修复 JWT 中间件错误分支未及时 return。
- Services：
  - 增加 `POST /api/v1/services/:id/deploy`
  - 增加 `POST /api/v1/services/:id/rollback`（MVP 占位）
  - 增加 `GET /api/v1/services/:id/events`
  - 增加 `GET /api/v1/services/quota`
  - `GET /services` 支持无 `project_id` 查询全部。
- SSH：
  - 修复 SSH 客户端，支持“仅密码”认证。
- 前端 API 适配：
  - API base 默认改为 `/api/v1`
  - Axios 成功码兼容 `xcode=1000`
  - auth/services 模块做后端字段映射适配。
- 构建流程：
  - Makefile 新增 `web-build` 与 `build-all`。

### Verification

- `go test ./...` 通过。
- `cd web && npm run build` 通过。

### Known Gaps / Risks

- RBAC 当前为最小实现，角色-权限写入链路仍需强化（目前以基础 CRUD 为主）。
- Clusters/K8s 能力依赖有效 kubeconfig；无有效配置时返回最小兜底数据。
- AI 动作执行为 MVP 级别，占位逻辑较多，未接完整运维动作引擎。
- Services 领域模型与前端展示字段（env/owner/tags）存在语义映射，后续需统一模型。

### Next Actions

1. 将 services/hosts/rbac 的返回协议进一步统一为同一套结构（含分页 total 与 data_source）。
2. 完成主机详情页与 K8s 页面联调问题修复（字段补齐与边界错误处理）。
3. 为 RBAC 与 Clusters 补充基础集成测试。
4. 在 `README.md` 补充 `build-all` 启动说明与 MVP 当前状态说明。

## 2026-02-24 (AI + RBAC Fix Round)

### Scope

- AI 模块从 mock 升级为 Eino + Ollama(`glm-5:cloud`) 实际调用。
- `/api/v1/ai/chat` 切换为 SSE 流式响应，并完成前端流式消费。
- 修复 admin 权限导致页面 403 的可用性问题。

### Completed

- 后端 AI：
  - `internal/service/ai/handler.go` 改为 SSE 聊天（`meta/delta/done/error`）。
  - 对接 `svcCtx.AI.Runnable.Stream/Generate`，并保留 sessions 接口 JSON 读写。
  - `analyze/recommendations/k8s/analyze` 接入 LLM 摘要能力，失败回退到 fallback。
- LLM 初始化：
  - `internal/ai/client.go` 强制 `provider=ollama`，并注入 temperature 到 Ollama options。
  - `internal/svc/svc.go` 增加 AI 初始化失败结构化日志（base_url/model/provider）。
  - `configs/config.yaml` 默认模型改为 `glm-5:cloud`。
- 前端 AI：
  - `web/src/api/modules/ai.ts` 新增 `chatStream()`（POST + ReadableStream 解析 SSE）。
  - `web/src/components/AI/ChatInterface.tsx` 改为流式增量渲染 assistant 消息。
- RBAC：
  - `internal/service/rbac/handler.go`：admin 用户（用户名=admin 或角色 code=admin）返回全量权限并含 `*:*`。
  - `rbac/check` 支持通配判定（`resource:*` 与 `*:*`）。
  - `web/src/components/RBAC/PermissionContext.tsx` 支持通配权限匹配。

### Known Gaps / Risks

- admin 全量放行是临时策略；页面权限码与 Casbin/API 权限码仍未完全统一。
- AI 依赖本地/远端 Ollama 可达与模型可用性；不可达时返回 error/fallback，不会中断服务。

### Next Actions

1. 增加 AI SSE handler 的单元测试（流式事件顺序、错误分支、会话落盘）。
2. 将 RBAC 页面权限码与后端 Casbin 权限模型整理为统一字典。
3. 对非 admin 用户补齐角色-权限初始化脚本，降低“有账号无权限”问题。
