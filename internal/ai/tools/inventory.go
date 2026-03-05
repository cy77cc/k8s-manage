package tools

func buildInventoryTools(all []RegisteredTool) []RegisteredTool {
	return filterToolsByName(all,
		"host_list_inventory",
		"cluster_list_inventory",
		"service_list_inventory",
	)
}
