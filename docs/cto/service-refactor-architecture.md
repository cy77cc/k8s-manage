# Service Refactor Architecture (CTO)

## Scope

- 目录重构标准：`routes + handler + logic`，对齐 `internal/service/user`。
- Host 统一为外部域，Node 进入兼容代理层。
- 引入版本化 SQL migration，替代生产 AutoMigrate。

## Decisions

- 服务启动前执行 `RunMigrations(db)`；生产默认 `app.auto_migrate=false`。
- `storage/gorm.go` 仅负责连接；开发态可显式启用 `RunDevAutoMigrate`。
- Host 新增三步入网链路：`probe -> create(confirm) -> credentials rotate`。
- `node/add` 不再直连 node 逻辑，统一委托 host 逻辑，返回 `Deprecation/Sunset` Header。

## Compatibility Strategy

- `/api/v1/hosts/*` 为主路径。
- `/api/v1/node/add` 兼容保留到 `Sunset: 2026-06-30`。
- `POST /hosts` 在无 `probe_token` 时保留 legacy 创建路径，避免旧前端立即中断。

## Risk Control

- probe token 仅服务端存 hash，TTL=10m，一次性消费。
- mutating 路由保留 JWT/Casbin 既有边界，不在本次重构扩散变更。
- 迁移脚本采用幂等 `IF NOT EXISTS`。

## Rollback

- 通过 `make migrate-down` 回滚最近版本化迁移（测试环境）。
- 若回滚业务层，仅需恢复 `host` 路由与 `node` 旧逻辑绑定。
