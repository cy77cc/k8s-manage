-- +migrate Up
CREATE TABLE IF NOT EXISTS ai_chat_sessions (
  id VARCHAR(64) PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  scene VARCHAR(128) NOT NULL,
  title VARCHAR(128) DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_ai_session_user_scene (user_id, scene)
);

CREATE TABLE IF NOT EXISTS ai_chat_messages (
  id VARCHAR(64) PRIMARY KEY,
  session_id VARCHAR(64) NOT NULL,
  role VARCHAR(32) NOT NULL,
  content LONGTEXT,
  thinking LONGTEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_ai_msg_session_created (session_id, created_at),
  INDEX idx_ai_msg_role (role)
);

-- +migrate Down
DROP TABLE IF EXISTS ai_chat_messages;
DROP TABLE IF EXISTS ai_chat_sessions;
