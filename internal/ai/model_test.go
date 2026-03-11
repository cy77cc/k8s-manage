package ai

import (
	"context"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/config"
)

func TestCheckStartupModelHealthIncludesAllAILayers(t *testing.T) {
	original := config.CFG
	t.Cleanup(func() {
		config.CFG = original
	})
	config.CFG.LLM.Enable = false

	results := CheckStartupModelHealth(context.Background())
	if len(results) != 4 {
		t.Fatalf("health results = %d, want 4", len(results))
	}

	expected := map[string]bool{
		"planner":    false,
		"rewrite":    false,
		"expert":     false,
		"summarizer": false,
	}
	for _, result := range results {
		if _, ok := expected[result.Name]; ok {
			expected[result.Name] = true
		}
		if result.Err == nil {
			t.Fatalf("result %+v, want error when llm disabled", result)
		}
	}
	for name, seen := range expected {
		if !seen {
			t.Fatalf("missing health result for layer %q: %+v", name, results)
		}
	}
}
