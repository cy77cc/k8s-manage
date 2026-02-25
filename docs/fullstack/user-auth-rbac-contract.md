# User/Auth/RBAC Contract (2026-02-25)

## 1. Scope

- 对齐现有前端认证和设置页。
- 主路由保持 `/api/v1/rbac/*`。
- 认证接口保持 `/api/v1/auth/*`。
- 响应协议保持 `xcode`（`code/msg/data`）。

## 2. Auth API

### POST `/api/v1/auth/login`

- 入参：`{ username, password }`
- 出参 `data`：
  - `accessToken`
  - `refreshToken`
  - `expires`
  - `uid`
  - `roles: string[]`
  - `user?: { id, username, name, email, status, roles, permissions }`
  - `permissions?: string[]`

### POST `/api/v1/auth/register`

- 入参：`{ username, password, email, phone?, avatar? }`
- 密码使用 `bcrypt` 存储。
- 若存在 `viewer` 角色，自动绑定。
- 出参与 login 保持一致。

### POST `/api/v1/auth/logout`

- 入参可空：`{ refreshToken? }`
- 有 `refreshToken` 时删除白名单 token；无则直接成功。

### GET `/api/v1/auth/me`

- 需 JWT。
- 出参 `data`：`{ id, username, name, email, status, roles, permissions }`
- `roles` 来源：`user_roles + roles.code`
- `permissions` 来源：`role_permissions + permissions.code`
- admin 场景补全 `*:*`

## 3. RBAC API

### GET `/api/v1/rbac/users`

- 出参：`{ code, msg, data: { list: UserItem[], total } }`
- `UserItem.roles` 为真实角色 code 列表。

### GET `/api/v1/rbac/roles`

- 出参：`{ code, msg, data: { list: RoleItem[], total } }`
- `RoleItem.permissions` 为真实权限 code 列表。

### GET `/api/v1/rbac/permissions`

- 出参：`{ code, msg, data: { list: PermissionItem[], total } }`

### POST/PUT `/api/v1/rbac/users`

- 创建/更新用户时，`roles` 写入 `user_roles`。
- 密码字段（若传入）统一 bcrypt。

### POST/PUT `/api/v1/rbac/roles`

- 创建/更新角色时，`permissions` 写入 `role_permissions`。

### GET `/api/v1/rbac/me/permissions`

- admin 用户（用户名 admin 或角色 admin）返回全量权限并含 `*:*`。

## 4. Frontend Mapping

- `web/src/api/modules/auth.ts`
  - 兼容 `token/accessToken`
  - 新增 `refreshToken` 存储映射
  - 新增 `logout(refreshToken?)`
- `web/src/components/Auth/AuthContext.tsx`
  - 登录/注册保存 `refreshToken`
  - `logout()` 调用后端接口后清理本地会话
- Settings 页面继续消费：
  - `res.data.list`
  - `res.data.total`

## 5. Error Model

- 保持现有 `xcode` 机制。
- 参数错误返回 `ErrInvalidParam`。
- token 无效返回 `TokenInvalid` / `Unauthorized`。
