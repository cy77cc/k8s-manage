# Codebase Architecture

## 1. Runtime Topology

- 单进程部署：Go 后端同时提供 API 与前端静态资源（`web/dist` embed）。
- API 前缀：`/api/v1`。
- 非 `/api/*` 请求走 SPA fallback 到 `index.html`，支持前端路由刷新。

## 2. AI Module (Eino + Ollama)

- 入口：`internal/service/ai/routes.go` + `internal/service/ai/chat_handler.go` + `internal/service/ai/capability_handler.go`。
- 能力层：`internal/ai/platform_agent.go` + `internal/ai/model_factory(client.go)` + `internal/ai/tools_*.go`。
- 初始化：`internal/svc/svc.go` 在服务启动时构建 ChatModel、PlatformAgent、MCP client manager。
- 模型约束：当前固定 `llm.provider=ollama`，默认模型 `glm-5:cloud`。

### 2.1 Chat Protocol

- `POST /api/v1/ai/chat` 使用 `text/event-stream`。
- 事件定义：
  - `meta`：返回 `sessionId/createdAt`
  - `delta`：返回增量文本 `contentChunk`
  - `tool_call`：返回工具调用请求体
  - `tool_result`：返回工具执行结果与 execution id
  - `approval_required`：返回审批 token 与预览
  - `done`：返回完整 `session`
  - `error`：返回错误信息
- 会话查询仍使用 JSON：`GET/DELETE /api/v1/ai/sessions*`。

### 2.2 Fallback Strategy

- AI 未初始化或模型不可达时：
  - 聊天接口通过 `error` 事件返回失败原因。
  - `analyze/recommendations/k8s/analyze` 返回 fallback 数据并标记 `data_source=fallback`。

## 3. RBAC Strategy

- 入口：`internal/service/rbac/handler.go` + `web/src/components/RBAC/PermissionContext.tsx`。
- admin 可用性修复策略（当前版本）：
  - admin 判定：用户名 `admin` 或角色 code=`admin`。
- 后端 `GET /api/v1/rbac/me/permissions` 直接追加全量权限并包含 `*:*`。
- 前端权限判断支持三类命中：
    - `${resource}:${action}`
    - `${resource}:*`
    - `*:*`

## 4. AI Control Plane

- 新增运行时能力：
  - `GET /api/v1/ai/capabilities`
  - `POST /api/v1/ai/tools/preview`
  - `POST /api/v1/ai/tools/execute`
  - `GET /api/v1/ai/executions/:id`
  - `POST /api/v1/ai/approvals`
  - `POST /api/v1/ai/approvals/:id/confirm`
- 规则：
  - readonly 工具可直接执行
  - mutating 工具必须审批通过且 token 未过期
  - 工具执行输出统一 `{ok,data,error,source,latency_ms}`
- 工具接入形态：
  - Eino 本地工具：`os.*` / `k8s.*` / `service.*` / `host.*`
  - MCP 代理工具：`mcp.default.*`（由 `mcp-go` client 动态发现并注册到 Eino Agent）

## 5. Known Technical Debt

- admin 全量放行为短期可用性策略，需要后续替换为精细权限配置。
- 页面权限码与 Casbin/API 权限码尚未形成单一规范字典。
- AI 控制面当前为进程内存态，后续需迁移到 DB/Redis 以支持多实例。
