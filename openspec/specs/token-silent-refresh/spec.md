# token-silent-refresh Specification

## Purpose
定义前端 token 无感刷新机制，确保用户在 token 过期时无感知地继续使用系统。

## ADDED Requirements

### Requirement: Token 过期前主动刷新
系统 SHALL 在 accessToken 即将过期前主动刷新，确保用户操作不被中断。

#### Scenario: Token 过期前5分钟主动刷新
- **WHEN** accessToken 距离过期时间少于 5 分钟
- **THEN** 系统 SHALL 自动调用刷新接口获取新 token

#### Scenario: 刷新成功后状态同步
- **WHEN** token 刷新成功
- **THEN** 系统 SHALL 更新 localStorage 中的 token 和 refreshToken
- **AND** 系统 SHALL 更新 AuthContext 中的 token 状态

#### Scenario: 刷新成功后新 token 生效
- **WHEN** token 刷新成功后的下一次 API 请求
- **THEN** 请求 SHALL 携带新的 accessToken

### Requirement: 并发刷新请求合并
系统 SHALL 确保多个并发请求触发的 token 刷新只发起一次刷新 API 调用。

#### Scenario: 多请求同时触发刷新
- **WHEN** 多个请求同时检测到 token 需要刷新
- **THEN** 系统 SHALL 只发起一次刷新请求
- **AND** 所有等待的请求 SHALL 在刷新成功后使用新 token 重试

#### Scenario: 刷新进行中的新请求等待
- **WHEN** 刷新请求正在进行中
- **AND** 新的 API 请求到达
- **THEN** 新请求 SHALL 等待刷新完成
- **AND** 使用新 token 发送请求

### Requirement: 刷新失败统一处理
系统 SHALL 在 token 刷新失败时清除用户状态并引导重新登录。

#### Scenario: 刷新失败清除状态
- **WHEN** token 刷新失败（refreshToken 无效或过期）
- **THEN** 系统 SHALL 清除 localStorage 中的 token 和 refreshToken
- **AND** 系统 SHALL 清除 AuthContext 中的用户状态
- **AND** 系统 SHALL 跳转到登录页

#### Scenario: 刷新失败保留重定向路径
- **WHEN** token 刷新失败并跳转登录页
- **THEN** 系统 SHALL 保存当前页面路径
- **AND** 登录成功后 SHALL 跳转回原页面

### Requirement: 定时检查 Token 状态
系统 SHALL 定时检查 token 过期状态，确保及时刷新。

#### Scenario: 登录后启动定时检查
- **WHEN** 用户登录成功
- **THEN** 系统 SHALL 启动定时器每分钟检查一次 token 过期状态

#### Scenario: 登出后停止定时检查
- **WHEN** 用户登出
- **THEN** 系统 SHALL 停止 token 过期检查定时器

#### Scenario: 组件卸载时清理定时器
- **WHEN** AuthProvider 组件卸载
- **THEN** 系统 SHALL 清理所有 token 相关的定时器

### Requirement: 被动刷新兜底机制
系统 SHALL 在请求因 token 过期失败时尝试刷新并重试，作为主动刷新的兜底。

#### Scenario: 请求401错误触发刷新
- **WHEN** API 请求返回 401 状态码
- **AND** 请求不是刷新接口本身
- **THEN** 系统 SHALL 尝试刷新 token
- **AND** 刷新成功后使用新 token 重试原请求

#### Scenario: 业务错误码触发刷新
- **WHEN** API 请求返回业务错误码 4005（token 无效）或 4006（token 过期）
- **THEN** 系统 SHALL 尝试刷新 token
- **AND** 刷新成功后使用新 token 重试原请求
