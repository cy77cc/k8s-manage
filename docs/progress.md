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
