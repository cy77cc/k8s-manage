-- +migrate Up
CREATE TABLE IF NOT EXISTS ai_chat_turns (
  id VARCHAR(64) PRIMARY KEY,
  session_id VARCHAR(64) NOT NULL,
  role VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT '',
  phase VARCHAR(64) NOT NULL DEFAULT '',
  trace_id VARCHAR(64) NOT NULL DEFAULT '',
  parent_turn_id VARCHAR(64) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP NULL DEFAULT NULL,
  INDEX idx_ai_turn_session_created (session_id, created_at),
  INDEX idx_ai_turn_status (status),
  INDEX idx_ai_turn_trace_id (trace_id)
);

CREATE TABLE IF NOT EXISTS ai_chat_blocks (
  id VARCHAR(64) PRIMARY KEY,
  turn_id VARCHAR(64) NOT NULL,
  block_type VARCHAR(32) NOT NULL,
  position INT NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT '',
  title VARCHAR(255) NOT NULL DEFAULT '',
  content_text LONGTEXT,
  content_json LONGTEXT,
  streaming BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_ai_block_turn_position (turn_id, position),
  INDEX idx_ai_block_type (block_type)
);

-- +migrate Down
DROP TABLE IF EXISTS ai_chat_blocks;
DROP TABLE IF EXISTS ai_chat_turns;
