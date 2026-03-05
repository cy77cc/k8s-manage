package tools

func buildGovernanceTools(all []RegisteredTool) []RegisteredTool {
	return filterToolsByPrefix(all, "topology_", "audit_", "user_", "role_", "permission_")
}
