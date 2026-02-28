# RBAC Governance Endpoint Policy Map

## Read (`rbac:read`)
- `GET /api/v1/rbac/me/permissions`
- `POST /api/v1/rbac/check`
- `GET /api/v1/rbac/users`
- `GET /api/v1/rbac/users/:id`
- `GET /api/v1/rbac/roles`
- `GET /api/v1/rbac/roles/:id`
- `GET /api/v1/rbac/permissions`
- `GET /api/v1/rbac/permissions/:id`

## Write (`rbac:write`)
- `POST /api/v1/rbac/users`
- `PUT /api/v1/rbac/users/:id`
- `DELETE /api/v1/rbac/users/:id`
- `POST /api/v1/rbac/roles`
- `PUT /api/v1/rbac/roles/:id`
- `DELETE /api/v1/rbac/roles/:id`

## Deny Audit Fields
On denied access, middleware writes audit context with:
- `actor`
- `resource`
- `action`
- `timestamp`
