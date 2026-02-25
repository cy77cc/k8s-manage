package ai

import "strings"

// NormalizeToolName keeps tool naming compatible with OpenAI-style function names.
// Legacy dotted names are converted to underscore style.
func NormalizeToolName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	return strings.ReplaceAll(trimmed, ".", "_")
}
