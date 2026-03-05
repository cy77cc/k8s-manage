package tools

func buildMonitorTools(all []RegisteredTool) []RegisteredTool {
	return filterToolsByPrefix(all, "monitor_")
}
