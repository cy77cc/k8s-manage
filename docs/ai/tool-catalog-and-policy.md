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

## OS Target Policy

- 目标仅允许：`localhost` 或 DB 中存在的节点（ID/IP/Name/Hostname）。
- 命令白名单：固定参数，禁止用户拼接 shell。
