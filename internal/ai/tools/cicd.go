package tools

func buildCICDTools(all []RegisteredTool) []RegisteredTool {
	return filterToolsByPrefix(all, "cicd_", "job_")
}
