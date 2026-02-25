# Service Release Architecture (Phase-1)

## Core Design

- Service 作为主实体，关联四类扩展对象：
  - `service_revisions`
  - `service_variable_sets`
  - `service_deploy_targets`
  - `service_release_records`

## Data Flow

1. 编辑/保存服务配置
2. 生成 revision（记录配置快照与变量 schema）
3. 按 env 维护变量值
4. 根据默认 deploy target（可临时覆盖）执行 preview
5. apply 时产出 release record 并记录状态

## Template Engine

- 语法：`{{var}}`, `{{var|default:value}}`
- 值优先级：
  - request variables
  - env variable set
  - template default
- unresolved 变量在 preview 阶段返回 warning，在 apply 阶段阻断发布

## Permissions

- 继续沿用 `service:read|write|deploy|approve`
- 非 admin 受 project/team 归属限制
- production deploy 需 `service:approve`

## Current Limits

- compose deploy 仍是占位动作，未接主机编排执行器
- preflight 仅做轻量检查，资源冲突检查后续补齐
