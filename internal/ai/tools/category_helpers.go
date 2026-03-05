package tools

import "strings"

func filterToolsByPrefix(all []RegisteredTool, prefixes ...string) []RegisteredTool {
	out := make([]RegisteredTool, 0)
	for _, item := range all {
		for _, prefix := range prefixes {
			if strings.HasPrefix(item.Meta.Name, prefix) {
				out = append(out, item)
				break
			}
		}
	}
	return out
}

func filterToolsByName(all []RegisteredTool, names ...string) []RegisteredTool {
	want := make(map[string]struct{}, len(names))
	for _, name := range names {
		want[name] = struct{}{}
	}
	out := make([]RegisteredTool, 0)
	for _, item := range all {
		if _, ok := want[item.Meta.Name]; ok {
			out = append(out, item)
		}
	}
	return out
}
