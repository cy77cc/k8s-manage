package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAIChatTurnBlocksMigrationContainsRequiredTables(t *testing.T) {
	path := filepath.Join("..", "migrations", "20260312_000035_ai_chat_turn_blocks.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration file failed: %v", err)
	}

	sql := strings.ToLower(string(content))
	required := []string{
		"create table if not exists ai_chat_turns",
		"create table if not exists ai_chat_blocks",
		"drop table if exists ai_chat_blocks",
		"drop table if exists ai_chat_turns",
	}
	for _, item := range required {
		if !strings.Contains(sql, item) {
			t.Fatalf("migration missing required sql %q", item)
		}
	}
}
