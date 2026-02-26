-- +migrate Up
CREATE TABLE IF NOT EXISTS ai_command_executions (
  id VARCHAR(64) PRIMARY KEY,
  user_id BIGINT NOT NULL DEFAULT 0,
  scene VARCHAR(128) NOT NULL DEFAULT 'global',
  command_text TEXT,
  intent VARCHAR(128) NOT NULL DEFAULT '',
  plan_hash VARCHAR(96) NOT NULL DEFAULT '',
  risk VARCHAR(16) NOT NULL DEFAULT 'low',
  status VARCHAR(32) NOT NULL DEFAULT 'previewed',
  trace_id VARCHAR(96) NOT NULL DEFAULT '',
  params_json LONGTEXT,
  missing_json LONGTEXT,
  plan_json LONGTEXT,
  result_json LONGTEXT,
  approval_context LONGTEXT,
  execution_summary TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_ai_cmd_user_created ON ai_command_executions(user_id, created_at);
CREATE INDEX idx_ai_cmd_trace_id ON ai_command_executions(trace_id);
CREATE INDEX idx_ai_cmd_intent ON ai_command_executions(intent);

ALTER TABLE cicd_audit_events ADD COLUMN command_id VARCHAR(96) NOT NULL DEFAULT '';
ALTER TABLE cicd_audit_events ADD COLUMN intent VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE cicd_audit_events ADD COLUMN plan_hash VARCHAR(96) NOT NULL DEFAULT '';
ALTER TABLE cicd_audit_events ADD COLUMN trace_id VARCHAR(96) NOT NULL DEFAULT '';
ALTER TABLE cicd_audit_events ADD COLUMN approval_context LONGTEXT;
ALTER TABLE cicd_audit_events ADD COLUMN execution_summary TEXT;

CREATE INDEX idx_cicd_audit_command_id ON cicd_audit_events(command_id);
CREATE INDEX idx_cicd_audit_trace_id ON cicd_audit_events(trace_id);

-- +migrate Down
DROP INDEX idx_cicd_audit_trace_id ON cicd_audit_events;
DROP INDEX idx_cicd_audit_command_id ON cicd_audit_events;
ALTER TABLE cicd_audit_events DROP COLUMN execution_summary;
ALTER TABLE cicd_audit_events DROP COLUMN approval_context;
ALTER TABLE cicd_audit_events DROP COLUMN trace_id;
ALTER TABLE cicd_audit_events DROP COLUMN plan_hash;
ALTER TABLE cicd_audit_events DROP COLUMN intent;
ALTER TABLE cicd_audit_events DROP COLUMN command_id;

DROP INDEX idx_ai_cmd_intent ON ai_command_executions;
DROP INDEX idx_ai_cmd_trace_id ON ai_command_executions;
DROP INDEX idx_ai_cmd_user_created ON ai_command_executions;
DROP TABLE IF EXISTS ai_command_executions;
