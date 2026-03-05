-- +migrate Up
CREATE TABLE IF NOT EXISTS ai_confirmations (
  id VARCHAR(64) PRIMARY KEY,
  request_user_id BIGINT UNSIGNED NOT NULL,
  trace_id VARCHAR(96) NOT NULL DEFAULT '',
  tool_name VARCHAR(128) NOT NULL,
  tool_mode VARCHAR(32) NOT NULL DEFAULT 'mutating',
  risk_level VARCHAR(16) NOT NULL DEFAULT 'medium',
  params_json LONGTEXT,
  preview_json LONGTEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  reason VARCHAR(255) NOT NULL DEFAULT '',
  expires_at TIMESTAMP NOT NULL,
  confirmed_at TIMESTAMP NULL,
  cancelled_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_ai_confirmation_user_created (request_user_id, created_at),
  KEY idx_ai_confirmation_trace (trace_id),
  KEY idx_ai_confirmation_tool (tool_name),
  KEY idx_ai_confirmation_status (status),
  KEY idx_ai_confirmation_expires (expires_at)
);

CREATE TABLE IF NOT EXISTS ai_approval_tickets (
  id VARCHAR(64) PRIMARY KEY,
  confirmation_id VARCHAR(64) NOT NULL,
  request_user_id BIGINT UNSIGNED NOT NULL,
  approval_token VARCHAR(128) NOT NULL,
  tool_name VARCHAR(128) NOT NULL,
  target_resource_type VARCHAR(64) NOT NULL DEFAULT '',
  target_resource_id VARCHAR(128) NOT NULL DEFAULT '',
  risk_level VARCHAR(16) NOT NULL DEFAULT 'medium',
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  approver_user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reject_reason VARCHAR(255) NOT NULL DEFAULT '',
  params_json LONGTEXT,
  preview_json LONGTEXT,
  expires_at TIMESTAMP NOT NULL,
  approved_at TIMESTAMP NULL,
  rejected_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_ai_approval_token (approval_token),
  KEY idx_ai_approval_confirmation (confirmation_id),
  KEY idx_ai_approval_request_created (request_user_id, created_at),
  KEY idx_ai_approval_tool (tool_name),
  KEY idx_ai_approval_target (target_resource_type, target_resource_id),
  KEY idx_ai_approval_status (status),
  KEY idx_ai_approval_expires (expires_at)
);

-- +migrate Down
DROP TABLE IF EXISTS ai_approval_tickets;
DROP TABLE IF EXISTS ai_confirmations;
