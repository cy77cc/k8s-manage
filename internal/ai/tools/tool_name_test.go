package tools

import "testing"

func TestNormalizeToolName(t *testing.T) {
	cases := map[string]string{
		"os.get_cpu_mem":         "os_get_cpu_mem",
		"host.batch_exec_apply":  "host_batch_exec_apply",
		"k8s_get_events":         "k8s_get_events",
		"  service.deploy_apply": "service_deploy_apply",
		"":                       "",
	}
	for in, want := range cases {
		got := NormalizeToolName(in)
		if got != want {
			t.Fatalf("NormalizeToolName(%q)=%q, want=%q", in, got, want)
		}
	}
}
