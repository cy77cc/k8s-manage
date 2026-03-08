-- +migrate Up
ALTER TABLE ai_approval_tickets
  ADD COLUMN target_resource_name VARCHAR(128) NOT NULL DEFAULT '' AFTER target_resource_id,
  ADD COLUMN task_detail_json LONGTEXT NULL AFTER preview_json,
  ADD COLUMN tool_calls_json LONGTEXT NULL AFTER task_detail_json,
  ADD COLUMN executed_at TIMESTAMP NULL AFTER rejected_at;

CREATE INDEX idx_ai_approval_resource_name ON ai_approval_tickets (target_resource_name);

-- +migrate Down
DROP INDEX idx_ai_approval_resource_name ON ai_approval_tickets;

ALTER TABLE ai_approval_tickets
  DROP COLUMN executed_at,
  DROP COLUMN tool_calls_json,
  DROP COLUMN task_detail_json,
  DROP COLUMN target_resource_name;
