# Notification Center

通知中心核心能力，提供统一的通知入口和管理功能。

## Requirements

### REQ-001: 通知铃铛入口
系统 Header 右上角显示通知铃铛图标，带有未读数量 Badge。

- Badge 显示未读通知数量
- 数量为 0 时隐藏 Badge
- 数量超过 99 显示 "99+"
- 点击铃铛打开通知面板

### REQ-002: 通知面板
点击铃铛后显示下拉通知面板。

- 面板宽度 380px，最大高度 480px
- 支持滚动加载历史通知
- 显示最近 20 条通知
- 底部显示"查看全部"链接

### REQ-003: 通知列表展示
通知列表按时间倒序排列，每条通知显示：

- 严重级别图标（🔴 严重 / 🟡 警告 / 🔵 信息）
- 通知标题
- 来源标识
- 时间（相对时间，如"2分钟前"）
- 已读/未读状态

### REQ-004: 通知过滤
面板顶部提供 Tab 过滤：

- 全部
- 告警
- 任务
- 系统

### REQ-005: 通知操作
每条通知支持以下操作：

- **标记已读**：将通知标记为已读状态
- **忽略**：忽略此通知，不再显示在列表中
- **确认告警**（仅告警类型）：确认告警，更新告警状态

### REQ-006: 批量操作
面板顶部提供批量操作：

- **全部已读**：将所有未读通知标记为已读

### REQ-007: 通知跳转
点击通知标题跳转到详情页：

- 告警通知 → 告警详情页 `/monitor?alert_id={id}`
- 任务通知 → 任务详情页 `/tasks/{id}`
- 系统通知 → 相关页面

### REQ-008: 通知持久化
用户通知状态需要持久化存储：

- 已读状态（read_at）
- 忽略状态（dismissed_at）
- 确认状态（confirmed_at）
- 支持多设备同步

## Data Model

```typescript
interface Notification {
  id: string;
  type: 'alert' | 'task' | 'system' | 'approval';
  title: string;
  content?: string;
  severity: 'critical' | 'warning' | 'info';
  source: string;
  source_id: string;
  action_url?: string;
  action_type?: 'confirm' | 'approve' | 'view';
  created_at: string;
}

interface UserNotification {
  id: string;
  user_id: string;
  notification_id: string;
  read_at?: string;
  dismissed_at?: string;
  confirmed_at?: string;
  notification: Notification;
}
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /notifications | 获取通知列表 |
| GET | /notifications/unread-count | 获取未读数量 |
| POST | /notifications/:id/read | 标记已读 |
| POST | /notifications/:id/dismiss | 忽略通知 |
| POST | /notifications/:id/confirm | 确认告警 |
| POST | /notifications/read-all | 全部已读 |
