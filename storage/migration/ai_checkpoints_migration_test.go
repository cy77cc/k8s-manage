package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAICheckPointsMigrationContainsRequiredSQL(t *testing.T) {
	path := filepath.Join("..", "migrations", "20260305_000031_ai_checkpoints.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration file failed: %v", err)
	}

	sql := strings.ToLower(string(content))
	required := []string{
		"create table if not exists ai_checkpoints",
		"unique key uk_ai_checkpoints_key",
		"drop table if exists ai_checkpoints",
	}
	for _, item := range required {
		if !strings.Contains(sql, item) {
			t.Fatalf("migration missing required sql %q", item)
		}
	}
}
