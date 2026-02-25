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

## 2026-02-24 (AI Control Plane + Function Calling)

### Scope

- 实施 AI 全域融合第 1 批：控制面、工具运行时、审批执行链、OS/K8s/Service/Host function calling。

### Completed

- 后端 AI 控制面新增接口：
  - `GET /api/v1/ai/capabilities`
  - `POST /api/v1/ai/tools/preview`
  - `POST /api/v1/ai/tools/execute`
  - `GET /api/v1/ai/executions/:id`
  - `POST /api/v1/ai/approvals`
  - `POST /api/v1/ai/approvals/:id/confirm`
- `/api/v1/ai/chat` SSE 新增事件：
  - `tool_call`
  - `tool_result`
  - `approval_required`
- 工具注册与执行：
  - OS: cpu/mem, disk/fs, net stat, process top, journal tail, container runtime
  - K8s: list resources, events, pod logs
  - Services: detail, deploy preview, deploy apply(审批)
  - Host: ssh readonly command
- 策略基线：
  - 默认只读
  - mutating 工具必须 `approval_token`
  - 目标白名单（localhost 或数据库已登记节点）
- 前端 API 扩展：
  - `AICapability`, `ToolExecution`, `ApprovalTicket`, `RiskLevel`
  - 工具预览/执行/审批/执行查询接口

### Known Gaps / Risks

- 控制面状态当前为内存态（单实例友好，多实例需落库或Redis共享状态）。
- 审批与执行审计表尚未落 DB（已保留后续迁移接口与文档）。
- SSE 聊天中的工具触发暂为显式触发（`/tool` 或 `context.tool_name`）。

### Next Actions

1. 落地 `ai_sessions/ai_messages/ai_tool_calls/ai_approvals` 数据表与迁移脚本。
2. 前端新增审批中心与执行时间线 UI。
3. 将 tool runtime 从内存迁移到持久化审计存储并增加分页查询。

## 2026-02-24 (AI Refactor Supplement: Eino Spec + mcp-go + UI)

### Scope

- 按 Eino 规范重构 function calling 主链路，并新增平台级单一 `PlatformAgent`。
- MCP 接入改为 `mcp-go client`，本地工具与 MCP 工具混合注册。
- 前端 AI 聊天窗口样式与渲染升级：加宽抽屉 + Markdown/GFM + 代码高亮 + tool trace 可见。

### Completed

- 后端结构化重构：
  - `internal/ai/` 拆分为：
    - `platform_agent.go`
    - `mcp_client.go`
    - `tool_contracts.go`
    - `tools_registry.go`
    - `tools_os.go`
    - `tools_k8s.go`
    - `tools_service.go`
    - `tools_host.go`
    - `tools_mcp_proxy.go`
    - `tools_common.go`
  - 删除历史大文件 `tools_local.go`，按功能拆分实现。
- Eino + Agent：
  - `react.NewAgent` 作为唯一平台 agent（`PlatformAgent`）。
  - `/api/v1/ai/chat` 内部统一调用 `PlatformAgent.Runnable.Stream`。
- MCP（mcp-go）：
  - 支持 `sse|stdio` 初始化、`Initialize`、`ListTools`、`CallTool`。
  - MCP tool 动态封装为 Eino tool，命名 `mcp.default.<tool>`。
- SSE 事件链：
  - 保持并验证 `meta/delta/tool_call/tool_result/approval_required/done/error`。
- 前端 AI 聊天：
  - `GlobalAIAssistant` 抽屉放大（桌面更宽，移动端全宽）。
  - `ChatInterface` 改为 Markdown-only 渲染（GFM）+ 代码块高亮。
  - 对话中可见 tool calling 轨迹系统消息。
- 工程回归：
  - `go test ./...` 通过。
  - `cd web && npm run build` 通过。

### Known Gaps / Risks

- tool trace 当前以系统消息块展示，后续可升级为独立时间线组件。
- AI 审批/执行状态目前仍为内存态，重启后不保留。
- 前端产物 chunk 体积偏大（vite 构建有 >500k 警告），后续建议按页面拆包。

### Next Actions

1. 将 `ai_approvals/ai_tool_calls` 迁移到 DB，并新增审计查询 API。
2. 补充 E2E：审批前拦截、审批后重放、MCP 不可达降级。
3. 优化前端 tool trace UI（时间线 + 折叠 JSON + 风险标签）。

## 2026-02-24 (Team Follow-up: Long Stream + Scroll + Thinking + Memory + MoE)

### Team

- CTO(`cto-vogels`): 长流式与Agent能力边界设计（MaxStep/NumPredict/MoE路由）。
- Product(`product-norman`): 思考过程可见性与可折叠体验。
- Fullstack(`fullstack-dhh`): 前后端实现（SSE事件、滚动行为、会话记忆注入）。
- QA(`qa-bach`): 回归关注点（中断恢复、长输出、滚动交互）。

### Completed

- 修复长回答中途中断：
  - `PlatformAgent` 的 `MaxStep` 提升到 `20`。
  - Ollama `NumPredict` 提升到 `1024`。
  - SSE接收错误时不立即丢弃已生成内容，改为先发 `error` 再落 `done`。
- 修复“不能上下翻页”：
  - 聊天区加入“仅在接近底部时自动滚动”策略；用户上滑阅读时不再被强制拉回底部。
- 新增“思考过程”展示：
  - 后端新增 `thinking_delta` SSE 事件（来自 `schema.Message.ReasoningContent`）。
  - 前端新增可折叠 Thinking 消息块（Reasoning 标签）。
- 增加会话记忆：
  - 聊天请求会自动携带最近 20 条会话历史（user/assistant）作为上下文输入，而非单轮问答。
- 强化 Agent（MoE）：
  - 新增专家路由：`default / ops / k8s / security`。
  - 根据用户问题关键词自动选择专家 agent 执行 `Stream/Generate`。

### Verification

- `go test ./...` 通过。
- `cd web && npm run build` 通过。

## 2026-02-24 (Team + Skill Check: Suggestion Pipeline / Persistence / Scene Chat / Trigger UX)

### Team & Skills

- 按 `team` 协作方式执行（CTO/Product/Fullstack/QA 分工）。
- 按要求检查 `find-skills`：本轮无需新增安装技能，直接基于现有代码栈实现。

### Completed

- Suggestion 链路改造：
  - 用户对话完成后，后端将 assistant 输出送入 suggestion 智能体（LLM prompt）提炼建议。
  - 建议按 `user + scene` 缓存，`/ai/recommendations` 直接返回对应场景建议。
  - 前端在对话完成后自动刷新 recommendation。
- 对话持久化：
  - 新增表：`ai_chat_sessions`、`ai_chat_messages`。
  - 通过 Gorm AutoMigrate 自动迁移。
  - 会话与消息落库，替换原纯内存会话。
- 分场景会话：
  - 新增 `GET /api/v1/ai/sessions/current?scene=...`。
  - 聊天发送携带 `context.scene`，默认复用该场景最近会话。
  - 前端打开 AI 助手时自动加载当前场景历史；支持“当前场景/全局”切换。
- 触发按钮 UX：
  - AI 助手按钮改为 Header 内联触发，不再固定悬浮。

### Verification

- `go test ./...` 通过。
- `cd web && npm run build` 通过。

## 2026-02-24 (Team Refactor Phase-1: Migration + Host Domain + Onboarding)

### Team

- CTO(`cto-vogels`): 重构边界、兼容策略、迁移与回滚方案。
- Fullstack(`fullstack-dhh`): 后端目录重排、API 实现、前端主机向导落地。
- Product(`product-norman`): 主机三步接入流与错误交互模型。
- QA(`qa-bach`): 迁移与兼容回归矩阵。

### Completed

- 版本化迁移框架落地：
  - 新增 `storage/migration/runner.go`，支持 `up/down/status`。
  - 新增 migration SQL：
    - `20260224_000001_create_ai_chat_tables.sql`
    - `20260224_000002_host_onboarding_tables.sql`
  - 新增命令：`make migrate-up|migrate-status|migrate-down`。
  - 启动流程改为先执行版本化迁移（`internal/cmd/root.go`）。
- 配置与数据库职责调整：
  - `app.auto_migrate=false` 默认关闭。
  - `storage/gorm.go` 移除生产迁移职责，仅保留连接。
- service 结构重排：
  - `host` 改为 `routes + handler + logic`。
  - `cluster/rbac` handler 移入子目录（统一组织方式）。
- Host/Node 收敛：
  - 新增 `POST /api/v1/hosts/probe`。
  - 新增 `PUT /api/v1/hosts/:id/credentials`。
  - `POST /api/v1/hosts` 支持 `probe_token` 创建。
  - `/api/v1/node/add` 改为委托 host 逻辑，并返回 `Deprecation/Sunset` Header。
- 模型统一：
  - 新增 `internal/model/host_probe.go`（`host_probe_sessions`）。
- 前端主机向导：
  - `HostOnboardingPage` 改为 3-step：连接信息 -> 探测结果 -> 入库确认。

## 2026-02-25 (AI Tool Calling Hardening: Typed Schema + Param Resolver)

### Scope

- 对齐 Eino Tool Calling 最佳实践，修复工具空参数调用导致失败的问题。
- 全量本地工具强类型化，并引入参数自动补全与一次重试。

### Completed

- `internal/ai` 工具契约增强：
  - 本地工具输入从 `map[string]any` 迁移到强类型 struct（含 `jsonschema`）。
  - `ToolMeta` 扩展：`required/default_hint/examples/schema`。
  - 新增输入错误码：`missing_param/invalid_param/param_conflict`。
- 参数自动补全：
  - 新增 `tool_param_resolver.go`。
  - 优先级：`runtime context > session memory > safety defaults`。
  - 仅对白名单参数补全，不覆盖用户显式参数。
- 重试与可观测性：
  - `runWithPolicyAndEvent` 支持缺参后一次重试。
  - SSE tool 事件携带 `param_resolution` 和 `retry`。
- 会话参数记忆：
  - `internal/service/ai/store.go` 增加按 `user+scene+tool` 记忆上次成功参数。
  - `chat_handler` 注入 runtime context + memory accessor 到 tool context。
- Agent 策略强化：
  - `platform_agent.go` 增加 tool calling 约束指令，降低空参数调用概率。
- MCP 工具补充：
  - 缓存并透出 MCP tool input schema/required（保持代理执行模式）。
- 文档沉淀：
  - `docs/cto/ai-tool-contract-hardening.md`
  - `docs/fullstack/ai-tool-typing-migration.md`
  - `docs/product/ai-tool-call-ux-rules.md`
  - `docs/qa/ai-tool-call-regression.md`
  - `docs/ai/tool-schema-catalog.md`
  - `docs/ai/tool-error-codes.md`

### Verification

- `go test ./...` 通过。
- `npm run build` 通过。
  - `hosts API` 新增 `probeHost`、`updateCredentials`，`createHost` 支持 probe token。

### Verification

- `go test ./...` 通过。
- `cd web && npm run build` 通过。

### Known Gaps / Risks

- `rbac` 与 `cluster` 已迁移到 handler 子目录，但仍可继续细分为更小 handler 文件。
- `POST /hosts` 仍保留无 `probe_token` 的 legacy 路径以兼容旧调用，后续应逐步收敛。
- 迁移 `down` 仅建议测试环境使用。

### Next Actions

1. 将 `rbac` 拆分为 `permission.go + user_role.go` 等更细文件。
2. 在 HostList 页面移除 legacy 弹窗创建，统一跳转三步向导。
3. 为 onboarding 增加后端集成测试（token 一次性消费/过期/非 admin force）。

## 2026-02-24 (Team: Host Platform Expansion - SSH/Credentials/Cloud/KVM)

### Team

- CTO(`cto-vogels`): 多入口主机接入架构与安全边界。
- Product(`product-norman`): 主机多入口流程与页面路径规划。
- Fullstack(`fullstack-dhh`): 后端接口与前端页面落地。
- QA(`qa-bach`): 回归矩阵与发布验收项。

### Completed

- Host 领域新增能力（后端）：
  - `GET|POST|DELETE /api/v1/credentials/ssh_keys*`
  - `POST /api/v1/credentials/ssh_keys/:id/verify`
  - `GET|POST /api/v1/hosts/cloud/accounts`
  - `POST /api/v1/hosts/cloud/providers/:provider/accounts/test`
  - `POST /api/v1/hosts/cloud/providers/:provider/instances/query`
  - `POST /api/v1/hosts/cloud/providers/:provider/instances/import`
  - `GET /api/v1/hosts/cloud/import_tasks/:task_id`
  - `POST /api/v1/hosts/virtualization/kvm/hosts/:id/preview`
  - `POST /api/v1/hosts/virtualization/kvm/hosts/:id/provision`
  - `GET /api/v1/hosts/virtualization/tasks/:task_id`
- 安全：
  - 新增 `internal/utils/secret.go`（AES-GCM）
  - SSH 私钥、云账号密钥加密落库（依赖 `security.encryption_key`）
- 数据模型：
  - `nodes` 扩展：`source/provider/provider_instance_id/parent_host_id`
  - `ssh_keys` 扩展：`fingerprint/algorithm/encrypted/usage_count`
  - 新增表模型：`host_cloud_accounts/host_import_tasks/host_virtualization_tasks`
- 迁移：
  - 新增 `storage/migrations/20260224_000003_host_platform_and_key_management.sql`
- 前端：
  - 新增页面：`/hosts/keys`、`/hosts/cloud-import`、`/hosts/virtualization`
  - `HostListPage` 新增“新增主机”下拉入口，支持三种接入方式 + 密钥管理
  - `hosts.ts` API 模块扩展 cloud/kvm/credentials 方法

### Verification

- `go test ./...` 通过。
- `cd web && npm run build` 通过。

### Known Gaps / Risks

- 云平台 provider 当前是 MVP mock 适配，未接真实云SDK调用。
- KVM 创建当前是任务模型与主机纳管闭环，未接 libvirt 实际执行器。
- migration 中 `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` 依赖较新 MySQL 版本。

### Next Actions

1. 接入阿里云/腾讯云 SDK 的真实实例查询与导入。
2. 将 KVM 流程切换到 libvirt 执行器并补充任务状态机。
3. 为密钥管理和云导入补充后端集成测试与 RBAC 细粒度权限码。

## 2026-02-25 (User/Auth/RBAC Contract Completion)

### Scope

- 按既定方案补全用户与认证能力，重点修复 auth/rbac 与前端设置页对接断点。

### Completed

- Auth 后端：
  - `Login/Register/Refresh` 回填 `roles` 与 `permissions`。
  - `auth/me` 返回完整用户信息（含角色与权限）。
  - `logout` 支持空入参容忍。
- 密码安全：
  - 新增 `internal/utils/password.go`，统一使用 `bcrypt` hash。
  - 登录校验支持 `bcrypt` + 旧 hash 兼容校验。
  - RBAC 创建/更新用户密码改为 `bcrypt`。
- RBAC 后端：
  - `users/roles/permissions` 列表接口统一返回 `data.list + data.total`。
  - 用户与角色关联（`user_roles`）及角色与权限关联（`role_permissions`）改为事务化写入。
  - `GetUser/GetRole` 回填真实 `roles/permissions`。
- 前端对接：
  - `authApi` 新增 `refreshToken` 映射与 `logout()` 调用。
  - `AuthContext` 登录/注册存储 `refreshToken`，退出时请求后端再清理会话。
- 迁移：
  - 新增 `storage/migrations/20260225_000004_user_rbac_baseline.sql`，补齐用户/RBAC基线表结构与基础种子。
- 文档：
  - 新增 `docs/fullstack/user-auth-rbac-contract.md`。

### Verification

- `go test ./...` 通过。
- `cd web && npm run build` 待本轮执行。

### Known Gaps / Risks

- 新增基线迁移使用 `UNIX_TIMESTAMP()`，当前以 MySQL 为默认目标；跨方言部署前需补充方言兼容迁移。
- 登录旧密码兼容当前仅校验，不自动回写升级为 bcrypt。

### Next Actions

1. 增加 auth/rbac 接口集成测试（含返回结构断言）。
2. 增加旧密码登录后自动 rehash 的可选开关。
3. 清理 `internal/service/user/handler/{roles,permissions,rbac}.go` 空壳文件或补齐实现。

## 2026-02-25 (Service Management: Standard/Custom + Preview + Ownership + Helm)

### Scope

- 落地服务管理能力升级：通用配置/自定义配置、实时渲染预览、多维筛选、权限分层与 Helm 导入渲染。

### Completed

- 后端服务域重构：
  - 新增 `internal/service/service/`（`routes/handler/logic/render/types`）。
  - `internal/service/service.go` 注册新服务域路由。
  - `internal/service/project/routes.go` 移除旧 services 路由挂载。
- API 能力补齐：
  - `POST /api/v1/services/render/preview`
  - `POST /api/v1/services/transform`
  - 扩展 `POST/PUT /api/v1/services`
  - 扩展 `GET /api/v1/services`（team/runtime/env/label_selector/q）
  - `POST /api/v1/services/:id/deploy`（支持 `k8s|compose|helm`）
  - Helm: `POST /api/v1/services/helm/import`, `POST /api/v1/services/helm/render`, `POST /api/v1/services/:id/deploy/helm`
  - 保留 `rollback/events/quota` 的 MVP 接口兼容。
- 数据模型与迁移：
  - 扩展 `internal/model/project.go` 的 `Service` 字段（ownership/config/render/labels）。
  - 新增 `ServiceHelmRelease`、`ServiceRenderSnapshot` 模型。
  - 新增 migration：`20260225_000005_service_management_upgrade.sql`。
  - 修复 migration 执行兼容：将 `PREPARE/EXECUTE/DEALLOCATE` 拆成独立语句并加索引幂等检查。
- 权限模型：
  - 新权限码：`service:read|write|deploy|approve`（migration seed）。
  - handler 中增加 `service:approve` 的 production deploy 校验。
  - 前端服务路由接入 `Authorized('service', ...)`。
- 前端页面：
  - `ServiceProvisionPage` 升级为双模式（standard/custom）+ 目标渲染切换（k8s/compose）+ 实时预览 + 一键转换。
  - `ServiceListPage` 增加 team/runtime/env/label_selector/q 筛选与标签展示。
  - `ServiceDetailPage` 展示归属与配置摘要，支持 Helm 导入渲染与部署按钮。
  - `web/src/api/modules/services.ts` 对齐新接口与字段映射。
- 测试补充：
  - 新增 `internal/service/service/render_test.go`，覆盖 standard->k8s/compose 渲染主路径。

### Verification

- `go test ./...` 通过。
- `cd web && npm run build` 通过。

### Known Gaps / Risks

- Helm 渲染当前优先使用传入 `rendered_yaml` 或系统 `helm template`，未完成 Helm SDK 全链路托管。
- 标签筛选当前基于 `labels_json LIKE`，数据量上升后需要专用索引策略。
- `rollback/events/quota` 仍为 MVP 兼容实现，后续需接真实审计与配额统计源。

### Next Actions

1. 将 Helm 渲染切换为 SDK 路径并增加 chart 依赖/私仓异常分类。
2. 为服务权限补充接口级测试（owner/team + prod approve）。
3. 将服务详情页的部署审批流程接入 AI 审批中心统一入口。

## 2026-02-25 (Cluster Management Phase-1: Lifecycle + Namespace + RBAC)

### Scope

- 按既定计划落地 cluster management Phase-1：生命周期管理、多租户命名空间隔离、发布策略、HPA、Quota/LimitRange 与生产审批门禁。

### Completed

- 数据模型与迁移：
  - 新增模型：
    - `ClusterNamespaceBinding`
    - `ClusterReleaseRecord`
    - `ClusterHPAPolicy`
    - `ClusterQuotaPolicy`
    - `ClusterDeployApproval`
    - `ClusterOperationAudit`
  - 新增 migration：`20260225_000006_cluster_lifecycle_phase1.sql`
  - 补充 dev auto migrate 注册。
- 后端 API 扩展（`/api/v1/clusters/:id/*`）：
  - Namespaces: `GET/POST/DELETE` + bindings `GET/PUT`
  - Rollouts: `GET/POST preview/apply` + `promote/abort/rollback`
  - HPA: `GET/POST/PUT/DELETE`
  - Quota/LimitRange: `GET/POST/PUT/DELETE`（Quota）+ `GET/POST`（LimitRange）
  - Approval: `POST approvals`, `POST approvals/:ticket/confirm`
- 后端策略层：
  - 新增 `policy.go`：
    - `k8s:*` 与兼容 `kubernetes:*` 权限判定
    - team-namespace 读写校验（含 readonly）
    - 生产环境动作审批门禁（`k8s:approve` 或 approved token）
    - 操作审计落库
- 旧集群接口回归增强：
  - 基础读操作（nodes/pods/services/ingresses/events/logs）加 `k8s:read` 与 namespace 绑定校验。
  - legacy deploy/apply 增加生产审批校验与审计。
- 前端能力落地：
  - `web/src/api/modules/kubernetes.ts` 扩展 phase-1 新接口与类型。
  - 新增组件：
    - `web/src/components/K8s/ClusterOverview.tsx`
    - `web/src/components/K8s/NamespacePolicyPanel.tsx`
    - `web/src/components/K8s/RolloutPanel.tsx`
    - `web/src/components/K8s/HPAEditor.tsx`
    - `web/src/components/K8s/QuotaEditor.tsx`
  - `K8sPage.tsx` 接入新面板 tabs：Namespaces/Rollouts/HPA/Quotas。

### Verification

- `go test ./...` 通过。
- `cd web && npm run build` 通过。

### Known Gaps / Risks

- Rollout promote/abort/rollback 依赖 `kubectl argo rollouts` plugin；缺失时返回 `rollout_cli_missing`。
- 若集群未安装 Argo Rollouts CRD，rollout 接口返回 `rollout_crd_missing`。
- 前端权限按钮仍以后端校验为主，后续可补精细按钮级隐藏策略。

### Next Actions

1. 增加 rollout/hpa/quota 的接口级单测与 e2e 场景回归。
2. 补充 network/observability/storage 的 Phase-C 接口骨架与页面占位。
3. 将 rollout 动作从 CLI 依赖逐步迁移到可控 controller/API 路径。
