## Why

当前系统已有完善的监控告警能力，但缺少统一的通知入口。用户无法实时感知告警状态变化，需要主动刷新监控页面才能发现新告警。随着系统规模增长，实时告警通知成为运维效率的关键需求。

## What Changes

- 新增**通知中心**功能，集成现有告警系统
- Header 通知铃铛入口，实时显示未读数量
- WebSocket 实时推送，确保告警即时触达
- 后端持久化已读/未读状态，支持多端同步
- 通知面板支持**确认告警**、**忽略通知**、**标记已读**操作
- 点击通知跳转到告警详情页

## Capabilities

### New Capabilities

- `notification-center`: 通知中心核心能力，包括通知数据模型、API、WebSocket 推送、前端组件
- `notification-realtime`: 实时通知推送能力，WebSocket 连接管理与消息处理
- `notification-preferences`: 用户通知偏好设置（预留，Phase 4 实现）

### Modified Capabilities

- `monitoring-alerting-phase`: 告警触发时需要创建通知并推送

## Impact

**前端新增：**
- `src/components/Notification/` 通知组件目录
- `src/contexts/NotificationContext.tsx` 通知状态上下文
- `src/hooks/useNotification.ts` 通知 Hook
- `src/api/modules/notification.ts` 通知 API

**前端修改：**
- `src/components/Layout/AppLayout.tsx` Header 集成通知铃铛

**后端新增：**
- `notifications` 通知数据表
- `user_notifications` 用户通知关联表
- `/notifications` API 端点
- WebSocket Hub 实时推送

**后端修改：**
- 告警触发逻辑集成通知创建
