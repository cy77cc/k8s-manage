## Why

当前前端已有 accessToken 和 refreshToken 机制，但缺少真正的无感刷新体验：
1. **被动刷新**：只在请求失败（401/4005/4006）后才尝试刷新，用户会感知到短暂的请求失败
2. **状态不同步**：刷新成功后只更新 localStorage，AuthContext 中的 token 状态未同步更新
3. **无主动刷新**：没有在 token 即将过期时主动刷新的机制

无感刷新的核心是：用户无感知地持续使用系统，token 过期对用户透明。

## What Changes

- 新增 Token 过期监控机制，在 token 即将过期时主动刷新
- 增强 AuthContext，支持 token 刷新后的状态同步
- 优化 axios 拦截器，确保刷新后重试请求使用新 token
- 新增刷新失败的统一处理（跳转登录页）

## Capabilities

### New Capabilities

- `token-silent-refresh`: 前端 token 无感刷新机制，包括主动刷新、状态同步、失败处理

### Modified Capabilities

- `user-access-governance`: 扩展现有用户访问治理，增加前端 token 生命周期管理

## Impact

- **前端文件**:
  - `web/src/api/api.ts`: 增强刷新逻辑，暴露刷新事件
  - `web/src/components/Auth/AuthContext.tsx`: 监听刷新事件，同步状态
- **新增文件**:
  - `web/src/utils/tokenManager.ts`: token 过期监控和主动刷新管理器
- **API 依赖**: 后端 `/auth/refresh` API 已存在，无需修改
