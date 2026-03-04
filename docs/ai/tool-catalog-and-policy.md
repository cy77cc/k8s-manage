# AI Tool Catalog and Policy

## Tool Levels

- L1 readonly: 默认放行（需 `ai:tool:read`）。
- L2 mutating: 需审批（`approval_token` + `ai:tool:execute`）。
- L3 high risk: 当前不自动执行，保留二次确认策略。

## Permission Codes

- `ai:chat`
- `ai:tool:read`
- `ai:tool:execute`
- `ai:approval:review`
- `ai:admin`

## Approval Policy

- 创建审批：`POST /api/v1/ai/approvals`
- 审批确认：`POST /api/v1/ai/approvals/:id/confirm`
- 执行时校验：工具名一致、申请人一致（或 admin）、状态 approved、未过期。

## Parameter Intelligence Policy

- 工具元数据扩展字段：`enum_sources`、`param_hints`、`related_tools`、`scene_scope`。
- ID 类参数优先配置 `enum_sources`，并指向 inventory/list 工具。
- 统一描述格式：功能描述 + 必填参数 + 默认值 + 参数来源 + 示例。
- 参数解析优先级：页面上下文 > 会话记忆 > 工具默认值 > 安全默认值。

## Error Recovery Policy

- 统一错误结构：`ToolExecutionError`（`code/message/recoverable/suggestions/hint_action`）。
- 缺参/格式错误优先返回修复建议，而不是直接失败。
- 只读工具的临时错误支持一次智能重试（timeout/canceled/tool_error）。

## Scene Tool Recommendation Policy

- 场景推荐接口：`GET /api/v1/ai/scene/:scene/tools`。
- 命令建议接口：`GET /api/v1/ai/commands/suggestions` 支持 `scene` 与 `q` 过滤。
- 推荐来源区分：`builtin`、`scene`、`builtin_alias`、`custom_alias`。

## Alias and Template Policy

- 内置别名：`hst/svc/cls/pl/job/cfg/alert/topo`。
- 自定义别名按 `user + scene` 作用域隔离存储。
- 参数模板按 `user + scene + name` 存储，可通过 `template` 参数注入。

## OS Target Policy

- 目标仅允许：`localhost` 或 DB 中存在的节点（ID/IP/Name/Hostname）。
- 命令白名单：固定参数，禁止用户拼接 shell。

## Host Batch Policy

- `host.list_inventory`、`host.batch_exec_preview` 属于 L1 readonly。
- `host.batch_exec_apply`、`host.batch_status_update` 属于 L2 mutating，必须审批。
- `host.batch_exec_apply` 内置危险命令拦截（如 `rm -rf /`、`mkfs`、`shutdown` 等）。
