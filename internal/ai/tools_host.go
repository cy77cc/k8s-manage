package ai

import (
	"context"
	"strings"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/model"
)

func hostSSHReadonly(ctx context.Context, deps PlatformDeps, input HostSSHReadonlyInput) (ToolResult, error) {
	return runWithPolicyAndEvent(ctx, ToolMeta{Name: "host.ssh_exec_readonly", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read"}, input, func(in HostSSHReadonlyInput) (any, string, error) {
		hostID := in.HostID
		cmd := strings.TrimSpace(in.Command)
		if hostID <= 0 {
			return nil, "host_ssh", NewMissingParam("host_id", "host_id is required")
		}
		if cmd == "" {
			return nil, "host_ssh", NewMissingParam("command", "command is required")
		}
		if !isReadonlyHostCommand(cmd) {
			return nil, "host_ssh", NewInvalidParam("command", "command not allowed")
		}
		var node model.Node
		if err := deps.DB.First(&node, hostID).Error; err != nil {
			return nil, "db", err
		}
		cli, err := sshclient.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, "")
		if err != nil {
			return nil, "host_ssh", err
		}
		defer cli.Close()
		out, err := sshclient.RunCommand(cli, cmd)
		if err != nil {
			return map[string]any{"stdout": out, "stderr": err.Error(), "exit_code": 1}, "host_ssh", nil
		}
		return map[string]any{"stdout": out, "stderr": "", "exit_code": 0}, "host_ssh", nil
	})
}

func isReadonlyHostCommand(cmd string) bool {
	switch strings.TrimSpace(cmd) {
	case "hostname", "uptime", "df -h", "free -m", "ps aux --sort=-%cpu":
		return true
	default:
		return false
	}
}
