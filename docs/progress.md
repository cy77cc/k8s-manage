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
