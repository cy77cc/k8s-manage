package ai_test

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestAIToolsRootGoFileCountWithinConvention(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve current file")
	}

	toolsRoot := filepath.Join(filepath.Dir(currentFile), "tools")
	files, err := filepath.Glob(filepath.Join(toolsRoot, "*.go"))
	if err != nil {
		t.Fatalf("glob tools root: %v", err)
	}

	if len(files) > 10 {
		t.Fatalf("expected at most 10 Go files in %s, got %d", toolsRoot, len(files))
	}
}
