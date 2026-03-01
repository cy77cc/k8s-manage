package ai

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var serviceUnitRegexp = regexp.MustCompile(`^[a-zA-Z0-9_.@-]+$`)

func osGetCPUMem(ctx context.Context, deps PlatformDeps, input OSCPUMemInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "os_get_cpu_mem",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in OSCPUMemInput) (any, string, error) {
			target := strings.TrimSpace(in.Target)
			loadavg, _, _ := runOnTarget(ctx, deps, target, "cat", []string{"/proc/loadavg"}, "cat /proc/loadavg")
			mem, source, err := runOnTarget(ctx, deps, target, "cat", []string{"/proc/meminfo"}, "cat /proc/meminfo")
			if err != nil {
				return nil, source, err
			}
			uptime, _, _ := runOnTarget(ctx, deps, target, "uptime", nil, "uptime")
			return map[string]any{"loadavg": loadavg, "meminfo": mem, "uptime": uptime}, source, nil
		})
}

func osGetDiskFS(ctx context.Context, deps PlatformDeps, input OSDiskInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "os_get_disk_fs",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in OSDiskInput) (any, string, error) {
			target := strings.TrimSpace(in.Target)
			out, source, err := runOnTarget(ctx, deps, target, "df", []string{"-h"}, "df -h")
			if err != nil {
				return nil, source, err
			}
			return map[string]any{"filesystem": out}, source, nil
		})
}

func osGetNetStat(ctx context.Context, deps PlatformDeps, input OSNetInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "os_get_net_stat",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in OSNetInput) (any, string, error) {
			target := strings.TrimSpace(in.Target)
			dev, source, err := runOnTarget(ctx, deps, target, "cat", []string{"/proc/net/dev"}, "cat /proc/net/dev")
			if err != nil {
				return nil, source, err
			}
			listen, _, _ := runOnTarget(ctx, deps, target, "ss", []string{"-ltn"}, "ss -ltn")
			return map[string]any{"net_dev": dev, "listening_ports": listen}, source, nil
		})
}

func osGetProcessTop(ctx context.Context, deps PlatformDeps, input OSProcessTopInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "os_get_process_top",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in OSProcessTopInput) (any, string, error) {
			target := strings.TrimSpace(in.Target)
			limit := in.Limit
			if limit <= 0 {
				limit = 10
			}
			if limit > 50 {
				limit = 50
			}
			cmd := fmt.Sprintf("ps aux --sort=-%%cpu | head -n %d", limit+1)
			out, source, err := runOnTarget(ctx, deps, target, "ps", []string{"aux", "--sort=-%cpu"}, cmd)
			if err != nil {
				return nil, source, err
			}
			return map[string]any{"top_processes": out, "limit": limit}, source, nil
		})
}

func osGetJournalTail(ctx context.Context, deps PlatformDeps, input OSJournalInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "os_get_journal_tail",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskMedium,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in OSJournalInput) (any, string, error) {
			target := strings.TrimSpace(in.Target)
			service := strings.TrimSpace(in.Service)
			if service == "" {
				return nil, "validation", NewMissingParam("service", "service is required")
			}
			if !serviceUnitRegexp.MatchString(service) {
				return nil, "validation", NewInvalidParam("service", "invalid service name")
			}
			lines := in.Lines
			if lines <= 0 {
				lines = 200
			}
			if lines > 500 {
				lines = 500
			}
			localArgs := []string{"-u", service, "-n", strconv.Itoa(lines), "--no-pager"}
			remoteCmd := fmt.Sprintf("journalctl -u %s -n %d --no-pager", service, lines)
			out, source, err := runOnTarget(ctx, deps, target, "journalctl", localArgs, remoteCmd)
			if err != nil {
				return nil, source, err
			}
			return map[string]any{"service": service, "lines": lines, "logs": out}, source, nil
		})
}

func osGetContainerRuntime(ctx context.Context, deps PlatformDeps, input OSContainerRuntimeInput) (ToolResult, error) {
	return runWithPolicyAndEvent(
		ctx,
		ToolMeta{
			Name:       "os_get_container_runtime",
			Mode:       ToolModeReadonly,
			Risk:       ToolRiskLow,
			Provider:   "local",
			Permission: "ai:tool:read",
		},
		input,
		func(in OSContainerRuntimeInput) (any, string, error) {
			target := strings.TrimSpace(in.Target)
			out, source, err := runOnTarget(ctx, deps, target, "docker", []string{"ps", "--format", "{{.ID}} {{.Image}} {{.Status}}"}, "docker ps --format '{{.ID}} {{.Image}} {{.Status}}'")
			if err == nil {
				return map[string]any{"runtime": "docker", "containers": out}, source, nil
			}
			out2, source2, err2 := runOnTarget(ctx, deps, target, "ctr", []string{"-n", "k8s.io", "containers", "list"}, "ctr -n k8s.io containers list")
			if err2 == nil {
				return map[string]any{"runtime": "containerd", "containers": out2}, source2, nil
			}
			return nil, source2, fmt.Errorf("docker/containerd unavailable: %v / %v", err, err2)
		})
}
