# Notification Center - Implementation Tasks

## Phase 1: 前端 MVP (Mock 数据)

### 1. 基础设施

- [x] 1.1 创建 `src/api/modules/notification.ts` 通知 API 模块
- [x] 1.2 创建 `src/types/notification.ts` 通知类型定义
- [x] 1.3 创建 `src/contexts/NotificationContext.tsx` 通知上下文
- [x] 1.4 创建 `src/hooks/useNotification.ts` 通知 Hook

### 2. 组件开发

- [x] 2.1 创建 `src/components/Notification/` 目录
- [x] 2.2 创建 `NotificationBell.tsx` 通知铃铛组件
- [x] 2.3 创建 `NotificationPanel.tsx` 下拉面板组件
- [x] 2.4 创建 `NotificationList.tsx` 通知列表组件
- [x] 2.5 创建 `NotificationItem.tsx` 单条通知组件
- [x] 2.6 创建 `notification.css` 样式文件
- [x] 2.7 创建 `index.ts` 导出文件

### 3. 集成

- [x] 3.1 在 `AppLayout.tsx` 中集成 NotificationBell
- [x] 3.2 在 App 根组件包裹 NotificationProvider
- [x] 3.3 添加 Mock 数据用于开发测试

## Phase 2: 后端 API + 持久化

### 4. 数据库

- [x] 4.1 创建 `notifications` 表
- [x] 4.2 创建 `user_notifications` 表
- [x] 4.3 添加必要的索引

### 5. 后端 API

- [x] 5.1 实现 `GET /notifications` 获取通知列表
- [x] 5.2 实现 `GET /notifications/unread-count` 获取未读数量
- [x] 5.3 实现 `POST /notifications/:id/read` 标记已读
- [x] 5.4 实现 `POST /notifications/:id/dismiss` 忽略通知
- [x] 5.5 实现 `POST /notifications/:id/confirm` 确认告警
- [x] 5.6 实现 `POST /notifications/read-all` 全部已读

### 6. 前端对接

- [x] 6.1 替换 Mock API 为真实 API
- [x] 6.2 添加加载状态处理
- [x] 6.3 添加错误处理

## Phase 3: WebSocket 实时推送

### 7. 后端 WebSocket

- [x] 7.1 实现 WebSocket Hub 连接管理
- [x] 7.2 实现 JWT Token 认证
- [x] 7.3 实现心跳检测 (ping/pong)
- [x] 7.4 实现通知推送逻辑

### 8. 告警集成

- [x] 8.1 告警触发时创建通知记录
- [x] 8.2 查询目标用户列表
- [x] 8.3 创建 UserNotification 关联
- [x] 8.4 通过 WebSocket 推送通知

### 9. 前端 WebSocket

- [x] 9.1 实现 WebSocket 连接管理
- [x] 9.2 实现自动重连机制 (指数退避)
- [x] 9.3 实现心跳检测
- [x] 9.4 实现断线降级为轮询
- [x] 9.5 实现多标签页同步 (BroadcastChannel)

## Phase 4: 测试与优化

### 10. 测试

- [x] 10.1 编写单元测试 (API 层)
- [x] 10.2 编写组件测试 (NotificationPanel)
- [ ] 10.3 测试 WebSocket 连接稳定性
- [ ] 10.4 测试多标签页同步
- [x] 10.5 测试移动端响应式布局

### 11. 优化

- [x] 11.1 添加通知面板动画效果
- [x] 11.2 优化大量通知时的渲染性能
- [x] 11.3 添加通知声音提示 (可选)
- [x] 11.4 添加浏览器通知权限请求 (可选)
