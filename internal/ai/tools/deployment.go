package tools

func buildDeploymentTools(all []RegisteredTool) []RegisteredTool {
	return filterToolsByPrefix(all, "deployment_", "credential_")
}
