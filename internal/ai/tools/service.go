package tools

func buildServiceTools(all []RegisteredTool) []RegisteredTool {
	return filterToolsByPrefix(all, "service_")
}
