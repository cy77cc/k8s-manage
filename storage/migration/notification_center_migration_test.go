package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationCenterMigrationUsesMySQLSyntax(t *testing.T) {
	path := filepath.Join("..", "migrations", "20260301_000018_notification_center.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration file failed: %v", err)
	}

	sql := strings.ToLower(string(content))

	forbidden := []string{
		"bigserial",
		"create index if not exists",
		"where read_at is null",
	}
	for _, item := range forbidden {
		if strings.Contains(sql, item) {
			t.Fatalf("migration contains non-mysql syntax %q", item)
		}
	}
}
