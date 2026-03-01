-- +migrate Up
-- Notification Center Tables
-- Creates tables for notification center feature

-- 通知主体表
CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  type VARCHAR(32) NOT NULL,
  title VARCHAR(255) NOT NULL,
  content TEXT,
  severity VARCHAR(16) NOT NULL DEFAULT 'info',
  source VARCHAR(128),
  source_id VARCHAR(128),
  action_url VARCHAR(512),
  action_type VARCHAR(32),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_notifications_type (type),
  INDEX idx_notifications_severity (severity),
  INDEX idx_notifications_source (source),
  INDEX idx_notifications_source_id (source_id),
  INDEX idx_notifications_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 用户通知关联表
CREATE TABLE IF NOT EXISTS user_notifications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  notification_id BIGINT UNSIGNED NOT NULL,
  read_at TIMESTAMP NULL,
  dismissed_at TIMESTAMP NULL,
  confirmed_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_user_notification (user_id, notification_id),
  INDEX idx_user_notifications_user_id (user_id),
  INDEX idx_user_notifications_notification_id (notification_id),
  INDEX idx_user_unread (user_id, read_at, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS user_notifications;
DROP TABLE IF EXISTS notifications;
