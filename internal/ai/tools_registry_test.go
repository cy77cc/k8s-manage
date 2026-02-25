package ai

import "testing"

func TestBuildLocalTools_HasSchemaAndRequired(t *testing.T) {
	tools, err := BuildLocalTools(PlatformDeps{})
	if err != nil {
		t.Fatalf("BuildLocalTools failed: %v", err)
	}
	if len(tools) == 0 {
		t.Fatal("expected tools")
	}

	lookup := map[string]ToolMeta{}
	for _, item := range tools {
		lookup[item.Meta.Name] = item.Meta
	}

	cases := []string{
		"os_get_cpu_mem",
		"k8s_get_pod_logs",
		"service_deploy_apply",
		"host_ssh_exec_readonly",
		"cluster_list_inventory",
		"service_list_inventory",
	}
	for _, name := range cases {
		meta, ok := lookup[name]
		if !ok {
			t.Fatalf("missing tool %s", name)
		}
		if meta.Schema == nil {
			t.Fatalf("tool %s missing schema", name)
		}
	}
}
