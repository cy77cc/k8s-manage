-- +migrate Up
CREATE TABLE IF NOT EXISTS ai_trace_spans (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  span_id VARCHAR(64) NOT NULL,
  span_type VARCHAR(32) NOT NULL,
  name VARCHAR(200) NOT NULL,
  session_id VARCHAR(64) NOT NULL,
  trace_id VARCHAR(64) NOT NULL,
  parent_span_id VARCHAR(64) DEFAULT '',
  start_time TIMESTAMP NOT NULL,
  end_time TIMESTAMP NULL,
  duration_ms BIGINT NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'success',
  error_msg TEXT NULL,
  input LONGTEXT NULL,
  output LONGTEXT NULL,
  tokens BIGINT NOT NULL DEFAULT 0,
  metadata_json TEXT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_ai_trace_spans_span_id (span_id),
  KEY idx_ai_trace_spans_span_type (span_type),
  KEY idx_ai_trace_spans_name (name),
  KEY idx_ai_trace_spans_session_id (session_id),
  KEY idx_ai_trace_spans_trace_id (trace_id),
  KEY idx_ai_trace_spans_start_time (start_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 追踪跨度表，记录 LLM/工具/Agent 调用详情';

-- +migrate Down
DROP TABLE IF EXISTS ai_trace_spans;
