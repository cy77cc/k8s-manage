-- +migrate Up
ALTER TABLE ai_chat_messages
  ADD COLUMN status VARCHAR(32) NOT NULL DEFAULT '';

ALTER TABLE ai_chat_messages
  ADD COLUMN metadata_json LONGTEXT NULL;

ALTER TABLE ai_chat_messages
  ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- +migrate Down
ALTER TABLE ai_chat_messages DROP COLUMN updated_at;
ALTER TABLE ai_chat_messages DROP COLUMN metadata_json;
ALTER TABLE ai_chat_messages DROP COLUMN status;
