## Context

当前系统已有完善的监控告警能力（monitoring API），包括 Alert、AlertRule、AlertChannel 等数据模型。Header 中已有通知铃铛图标（BellOutlined + Badge），但功能未实现。

本设计需要：
- 集成现有告警系统，复用 Alert 数据模型
- 实现 WebSocket 实时推送
- 后端持久化用户通知状态（已读/未读/忽略/确认）
- 前端通知面板组件

## Goals / Non-Goals

**Goals:**
- 实现 WebSocket 实时通知推送
- 后端持久化通知状态，支持多端同步
- 前端通知铃铛 + 下拉面板组件
- 支持确认告警、忽略通知、标记已读操作
- 点击通知跳转到告警详情

**Non-Goals:**
- Phase 4 之前不实现通知偏好设置
- 不实现邮件/短信等外部通知渠道（已有 AlertChannel）
- 不实现通知分组和静默时段

## Decisions

### 1. 通知数据模型设计

**决策：** 新建独立的 `notifications` 和 `user_notifications` 表，而非扩展现有 `alerts` 表。

**理由：**
- 通知是用户维度的概念，同一告警可能通知多个用户
- 支持 future 扩展（任务通知、系统公告等）
- 解耦告警逻辑和通知逻辑

**数据模型：**
```sql
-- 通知主体
CREATE TABLE notifications (
  id BIGSERIAL PRIMARY KEY,
  type VARCHAR(32) NOT NULL,           -- alert/task/system/approval
  title VARCHAR(255) NOT NULL,
  content TEXT,
  severity VARCHAR(16),                 -- critical/warning/info
  source VARCHAR(128),                  -- 来源标识
  source_id VARCHAR(128),               -- 来源 ID (如 alert_id)
  action_url VARCHAR(512),              -- 跳转链接
  action_type VARCHAR(32),              -- confirm/approve/view
  created_at TIMESTAMP DEFAULT NOW()
);

-- 用户通知关联
CREATE TABLE user_notifications (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  notification_id BIGINT NOT NULL REFERENCES notifications(id),
  read_at TIMESTAMP,
  dismissed_at TIMESTAMP,
  confirmed_at TIMESTAMP,
  UNIQUE(user_id, notification_id)
);

CREATE INDEX idx_user_notifications_user ON user_notifications(user_id);
CREATE INDEX idx_user_notifications_unread ON user_notifications(user_id) WHERE read_at IS NULL;
```

### 2. 实时推送方案

**决策：** 使用 WebSocket 实现实时推送。

**备选方案：**
- SSE (Server-Sent Events)：单向推送足够，但 WebSocket 更灵活
- 轮询：简单但延迟高，不满足实时性要求

**实现方式：**
- 后端：WebSocket Hub 管理用户连接
- 前端：NotificationContext 管理连接状态
- 消息格式：`{ type: 'new' | 'update', notification: {...} }`

### 3. 告警与通知集成

**决策：** 告警触发时，由告警服务创建通知并推送。

**流程：**
```
Alert Triggered → Alert Service
                     ↓
              Create Notification
                     ↓
              Query Target Users (by rule channels)
                     ↓
              Create UserNotification records
                     ↓
              Push via WebSocket Hub
```

### 4. 前端组件结构

**决策：** 使用 Context + Hook 模式管理通知状态。

```
NotificationContext
├── notifications: UserNotification[]   // 通知列表
├── unreadCount: number                 // 未读数量
├── wsConnected: boolean                // WebSocket 状态
├── markAsRead(id)                      // 标记已读
├── dismiss(id)                         // 忽略通知
├── confirm(id)                         // 确认告警
└── markAllAsRead()                     // 全部已读

NotificationBell (Header 入口)
└── Popover
    └── NotificationPanel
        ├── NotificationFilter (Tab: 全部/告警/任务)
        ├── NotificationList
        │   └── NotificationItem
        └── NotificationFooter
```

### 5. API 设计

**REST API:**
```
GET    /notifications              # 获取通知列表 (分页)
GET    /notifications/unread-count # 获取未读数量
POST   /notifications/:id/read     # 标记已读
POST   /notifications/:id/dismiss  # 忽略通知
POST   /notifications/:id/confirm  # 确认告警
POST   /notifications/read-all     # 全部已读
```

**WebSocket:**
```
/ws/notifications
  - 连接时携带 JWT Token
  - 消息: { type: 'new', notification: {...} }
  - 消息: { type: 'update', id, read_at, ... }
```

## Risks / Trade-offs

### 风险 1: WebSocket 连接管理复杂度
- **风险：** 多标签页、网络断开重连等场景
- **缓解：**
  - 使用指数退避重连策略
  - 心跳检测 (30s)
  - 重连后自动同步未读状态

### 风险 2: 通知风暴
- **风险：** 大规模告警触发时，可能产生大量通知
- **缓解：**
  - 通知聚合（同一规则的告警合并）
  - 限流策略（每用户每分钟最多 N 条）
  - 前端防抖处理

### 风险 3: 后端改动范围
- **风险：** 需要后端新增表和 API
- **缓解：**
  - Phase 1 先用 Mock 数据实现前端
  - 后端 API 独立于告警核心逻辑

## Migration Plan

**Phase 1: 前端 MVP (可独立部署)**
- 使用 Mock API 实现前端组件
- 轮询方式获取通知
- 验证 UI 交互

**Phase 2: 后端 API + 持久化**
- 创建数据库表
- 实现 REST API
- 替换 Mock 为真实 API

**Phase 3: WebSocket 实时推送**
- 实现 WebSocket Hub
- 前端集成 WebSocket
- 告警触发集成通知创建

**回滚策略：**
- 前端：隐藏通知铃铛（Feature Flag）
- 后端：API 返回空列表，不影响告警核心功能

## Open Questions

1. **通知目标用户确定逻辑**
   - 告警规则关联的通知渠道如何映射到用户？
   - 是否需要用户订阅机制？

2. **通知保留策略**
   - 已读通知保留多久？
   - 是否需要定时清理任务？

3. **告警确认后的行为**
   - 确认告警后，告警状态如何变化？
   - 是否需要写入告警历史？
