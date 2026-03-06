# Tasks: Token 无感刷新

## 1. TokenManager 工具类

- [x] 1.1 创建 `web/src/utils/tokenManager.ts`
  - 实现 JWT 解析函数 `parseJwtExpiresAt`
  - 实现 token 过期检查函数 `isTokenExpiringSoon`
  - 实现定时器管理：`startExpiryCheck`、`stopExpiryCheck`
  - 定义刷新事件：`tokenRefreshed`、`tokenExpired`

- [x] 1.2 编写 TokenManager 单元测试
  - 文件: `web/src/utils/tokenManager.test.ts`
  - 测试用例：JWT 解析、过期判断、定时器启停

## 2. ApiService 增强

- [x] 2.1 增强 `web/src/api/api.ts` 刷新逻辑
  - 暴露 `refreshAccessToken` 为公共方法
  - 刷新成功后触发 `tokenRefreshed` 自定义事件
  - 刷新失败后触发 `tokenExpired` 自定义事件
  - 确保并发刷新只发起一次请求

- [x] 2.2 编写 ApiService 刷新逻辑测试
  - 文件: `web/src/api/api.test.ts`（扩展现有测试）
  - 测试用例：刷新成功、刷新失败、并发刷新合并

## 3. AuthContext 集成

- [x] 3.1 更新 `web/src/components/Auth/AuthContext.tsx`
  - 监听 `tokenRefreshed` 事件，更新 token 状态
  - 监听 `tokenExpired` 事件，清除状态并跳转登录页
  - 登录成功后启动 TokenManager 定时检查
  - 登出时停止 TokenManager 定时检查
  - 组件卸载时清理定时器

- [x] 3.2 编写 AuthContext 刷新集成测试
  - 文件: `web/src/components/Auth/AuthContext.test.tsx`
  - 测试用例：刷新事件处理、状态同步、定时器清理

## 4. 集成测试

- [x] 4.1 编写端到端刷新流程测试
  - 文件: `web/src/__tests__/auth/tokenRefresh.test.ts`
  - 测试场景：token 过期前主动刷新、刷新失败跳转登录

## 5. 验证

- [x] 5.1 手动验证刷新流程
  - 模拟 token 即将过期场景
  - 验证请求无感知继续
  - 验证刷新失败跳转登录页

- [x] 5.2 运行前端测试
  - 运行 `cd web && npm test`
  - 确保所有测试通过
