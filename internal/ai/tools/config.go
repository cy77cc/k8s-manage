package tools

func buildConfigTools(all []RegisteredTool) []RegisteredTool {
	return filterToolsByPrefix(all, "config_")
}
