package tools

func buildK8sTools(all []RegisteredTool) []RegisteredTool {
	return filterToolsByPrefix(all, "k8s_", "cluster_")
}
