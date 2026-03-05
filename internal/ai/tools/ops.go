package tools

func buildOpsTools(all []RegisteredTool) []RegisteredTool {
	return filterToolsByPrefix(all, "os_", "host_")
}
