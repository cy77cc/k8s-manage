# Service Permission Model

## 1. Permission Codes

- `service:read`: view list/detail/render outputs
- `service:write`: create/update/delete/preview/transform/helm import/render
- `service:deploy`: deploy/rollback actions
- `service:approve`: production deploy approval gate

## 2. Evaluation Order

1. JWT authentication
2. RBAC permission code check
3. Ownership guard (`project_id + team_id`)
4. Environment policy (`production` requires `service:approve`)

## 3. Default Role Strategy

- `viewer`: `service:read`
- `operator`: `service:read`, `service:write`, `service:deploy`
- `admin`: `*:*` (current global admin bypass strategy)

## 4. Migration Impact

- migration `20260225_000005_service_management_upgrade.sql` seeds service permissions and role bindings.
- first phase uses RBAC table joins + admin fast path in handler.
