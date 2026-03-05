-- +migrate Up
CREATE TABLE IF NOT EXISTS ai_checkpoints (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `key` VARCHAR(255) NOT NULL,
  value MEDIUMBLOB NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_ai_checkpoints_key (`key`),
  KEY idx_ai_checkpoints_updated_at (updated_at)
);

-- +migrate Down
DROP TABLE IF EXISTS ai_checkpoints;
