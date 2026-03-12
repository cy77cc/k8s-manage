package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAIChatTurnBlocksMigrationContainsRequiredSQL(t *testing.T) {
	path := filepath.Join("..", "migrations", "20260312_000035_ai_chat_turn_blocks.sql")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration file failed: %v", err)
	}
	sql := string(data)

	required := []string{
		"CREATE TABLE IF NOT EXISTS ai_chat_turns",
		"CREATE TABLE IF NOT EXISTS ai_chat_blocks",
		"INDEX idx_ai_turn_session_created",
		"INDEX idx_ai_block_turn_position",
		"DROP TABLE IF EXISTS ai_chat_blocks",
		"DROP TABLE IF EXISTS ai_chat_turns",
	}
	for _, item := range required {
		if !strings.Contains(sql, item) {
			t.Fatalf("migration missing required sql %q", item)
		}
	}
}
