package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/utils"
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
		out, err := executeHostCommand(deps, &node, cmd)
		if err != nil {
			return map[string]any{"stdout": out, "stderr": err.Error(), "exit_code": 1}, "host_ssh", nil
		}
		return map[string]any{"stdout": out, "stderr": "", "exit_code": 0}, "host_ssh", nil
	})
}

func hostListInventory(ctx context.Context, deps PlatformDeps, input HostInventoryInput) (ToolResult, error) {
	return runWithPolicyAndEvent(ctx, ToolMeta{Name: "host.list_inventory", Mode: ToolModeReadonly, Risk: ToolRiskLow, Provider: "local", Permission: "ai:tool:read"}, input, func(in HostInventoryInput) (any, string, error) {
		if deps.DB == nil {
			return nil, "db", fmt.Errorf("db unavailable")
		}
		limit := in.Limit
		if limit <= 0 {
			limit = 50
		}
		if limit > 200 {
			limit = 200
		}
		query := deps.DB.Model(&model.Node{})
		if status := strings.TrimSpace(in.Status); status != "" {
			query = query.Where("status = ?", status)
		}
		if kw := strings.TrimSpace(in.Keyword); kw != "" {
			pattern := "%" + kw + "%"
			query = query.Where("name LIKE ? OR ip LIKE ? OR hostname LIKE ?", pattern, pattern, pattern)
		}
		var nodes []model.Node
		if err := query.Order("id desc").Limit(limit).Find(&nodes).Error; err != nil {
			return nil, "db", err
		}

		items := make([]map[string]any, 0, len(nodes))
		for _, node := range nodes {
			items = append(items, map[string]any{
				"id":         uint64(node.ID),
				"name":       node.Name,
				"ip":         node.IP,
				"hostname":   node.Hostname,
				"status":     node.Status,
				"ssh_user":   node.SSHUser,
				"port":       node.Port,
				"cpu_cores":  node.CpuCores,
				"memory_mb":  node.MemoryMB,
				"disk_gb":    node.DiskGB,
				"labels":     parseHostLabels(node.Labels),
				"updated_at": node.UpdatedAt,
			})
		}

		return map[string]any{
			"total": len(items),
			"list":  items,
		}, "db", nil
	})
}

func hostBatchExecPreview(ctx context.Context, deps PlatformDeps, input HostBatchExecPreviewInput) (ToolResult, error) {
	return runWithPolicyAndEvent(ctx, ToolMeta{Name: "host.batch_exec_preview", Mode: ToolModeReadonly, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:read"}, input, func(in HostBatchExecPreviewInput) (any, string, error) {
		hostIDs, err := normalizeHostIDs(in.HostIDs)
		if err != nil {
			return nil, "host_batch_preview", err
		}
		cmd := strings.TrimSpace(in.Command)
		if cmd == "" {
			return nil, "host_batch_preview", NewMissingParam("command", "command is required")
		}

		class, risk, blocked := classifyHostCommand(cmd)
		targets, missing, err := loadHostBatchTargets(deps, hostIDs)
		if err != nil {
			return nil, "db", err
		}
		return map[string]any{
			"command":        cmd,
			"reason":         strings.TrimSpace(in.Reason),
			"command_class":  class,
			"risk":           risk,
			"blocked":        blocked,
			"target_count":   len(hostIDs),
			"resolved_count": len(targets),
			"missing_host_ids": func() []uint64 {
				out := make([]uint64, 0, len(missing))
				for _, id := range missing {
					out = append(out, uint64(id))
				}
				return out
			}(),
			"targets": targets,
		}, "host_batch_preview", nil
	})
}

func hostBatchExecApply(ctx context.Context, deps PlatformDeps, input HostBatchExecApplyInput) (ToolResult, error) {
	return runWithPolicyAndEvent(ctx, ToolMeta{Name: "host.batch_exec_apply", Mode: ToolModeMutating, Risk: ToolRiskHigh, Provider: "local", Permission: "ai:tool:execute"}, input, func(in HostBatchExecApplyInput) (any, string, error) {
		hostIDs, err := normalizeHostIDs(in.HostIDs)
		if err != nil {
			return nil, "host_batch_exec", err
		}
		cmd := strings.TrimSpace(in.Command)
		if cmd == "" {
			return nil, "host_batch_exec", NewMissingParam("command", "command is required")
		}

		class, risk, blocked := classifyHostCommand(cmd)
		if blocked {
			return nil, "host_batch_exec", NewInvalidParam("command", "dangerous command is blocked")
		}

		nodesByID, missing, err := loadHostNodesMap(deps, hostIDs)
		if err != nil {
			return nil, "db", err
		}

		results := map[string]any{}
		succeeded := 0
		failed := 0
		for _, id := range hostIDs {
			key := strconv.FormatUint(id, 10)
			node, ok := nodesByID[id]
			if !ok {
				results[key] = map[string]any{"stdout": "", "stderr": "host not found", "exit_code": 1}
				failed++
				continue
			}
			out, execErr := executeHostCommand(deps, node, cmd)
			if execErr != nil {
				results[key] = map[string]any{"stdout": out, "stderr": execErr.Error(), "exit_code": 1}
				failed++
				continue
			}
			results[key] = map[string]any{"stdout": out, "stderr": "", "exit_code": 0}
			succeeded++
		}

		return map[string]any{
			"reason":           strings.TrimSpace(in.Reason),
			"command":          cmd,
			"command_class":    class,
			"risk":             risk,
			"target_count":     len(hostIDs),
			"missing_host_ids": missing,
			"succeeded":        succeeded,
			"failed":           failed,
			"results":          results,
		}, "host_batch_exec", nil
	})
}

func hostBatchStatusUpdate(ctx context.Context, deps PlatformDeps, input HostBatchStatusInput) (ToolResult, error) {
	return runWithPolicyAndEvent(ctx, ToolMeta{Name: "host.batch_status_update", Mode: ToolModeMutating, Risk: ToolRiskMedium, Provider: "local", Permission: "ai:tool:execute"}, input, func(in HostBatchStatusInput) (any, string, error) {
		hostIDs, err := normalizeHostIDs(in.HostIDs)
		if err != nil {
			return nil, "host_batch_status", err
		}
		action := strings.ToLower(strings.TrimSpace(in.Action))
		if action == "" {
			return nil, "host_batch_status", NewMissingParam("action", "action is required")
		}
		if action != "online" && action != "offline" && action != "maintenance" {
			return nil, "host_batch_status", NewInvalidParam("action", "action must be online/offline/maintenance")
		}
		if deps.DB == nil {
			return nil, "db", fmt.Errorf("db unavailable")
		}
		res := deps.DB.Model(&model.Node{}).Where("id IN ?", hostIDs).Update("status", action)
		if res.Error != nil {
			return nil, "db", res.Error
		}
		return map[string]any{
			"action":        action,
			"reason":        strings.TrimSpace(in.Reason),
			"target_count":  len(hostIDs),
			"updated_count": res.RowsAffected,
		}, "host_batch_status", nil
	})
}

func executeHostCommand(deps PlatformDeps, node *model.Node, command string) (string, error) {
	privateKey, err := loadNodePrivateKey(deps, node)
	if err != nil {
		return "", err
	}
	cli, err := sshclient.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, privateKey)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	return sshclient.RunCommand(cli, command)
}

func loadNodePrivateKey(deps PlatformDeps, node *model.Node) (string, error) {
	if deps.DB == nil || node == nil || node.SSHKeyID == nil {
		return "", nil
	}
	var key model.SSHKey
	if err := deps.DB.Select("id", "private_key", "encrypted").Where("id = ?", uint64(*node.SSHKeyID)).First(&key).Error; err != nil {
		return "", err
	}
	pk := strings.TrimSpace(key.PrivateKey)
	if !key.Encrypted {
		return pk, nil
	}
	return utils.DecryptText(pk, config.CFG.Security.EncryptionKey)
}

func normalizeHostIDs(raw []int) ([]uint64, error) {
	if len(raw) == 0 {
		return nil, NewMissingParam("host_ids", "host_ids is required")
	}
	uniq := map[uint64]struct{}{}
	out := make([]uint64, 0, len(raw))
	for _, id := range raw {
		if id <= 0 {
			continue
		}
		uid := uint64(id)
		if _, ok := uniq[uid]; ok {
			continue
		}
		uniq[uid] = struct{}{}
		out = append(out, uid)
	}
	if len(out) == 0 {
		return nil, NewInvalidParam("host_ids", "host_ids must contain at least one positive id")
	}
	return out, nil
}

func loadHostNodesMap(deps PlatformDeps, hostIDs []uint64) (map[uint64]*model.Node, []uint64, error) {
	if deps.DB == nil {
		return nil, nil, fmt.Errorf("db unavailable")
	}
	var nodes []model.Node
	if err := deps.DB.Where("id IN ?", hostIDs).Find(&nodes).Error; err != nil {
		return nil, nil, err
	}
	byID := make(map[uint64]*model.Node, len(nodes))
	for i := range nodes {
		byID[uint64(nodes[i].ID)] = &nodes[i]
	}
	missing := make([]uint64, 0)
	for _, id := range hostIDs {
		if _, ok := byID[id]; !ok {
			missing = append(missing, id)
		}
	}
	return byID, missing, nil
}

func loadHostBatchTargets(deps PlatformDeps, hostIDs []uint64) ([]map[string]any, []int, error) {
	byID, _, err := loadHostNodesMap(deps, hostIDs)
	if err != nil {
		return nil, nil, err
	}
	targets := make([]map[string]any, 0, len(byID))
	missing := make([]int, 0)
	for _, id := range hostIDs {
		node, ok := byID[id]
		if !ok {
			missing = append(missing, int(id))
			continue
		}
		targets = append(targets, map[string]any{
			"id":       id,
			"name":     node.Name,
			"ip":       node.IP,
			"status":   node.Status,
			"hostname": node.Hostname,
		})
	}
	return targets, missing, nil
}

func classifyHostCommand(cmd string) (class string, risk string, blocked bool) {
	trimmed := strings.ToLower(strings.TrimSpace(cmd))
	if isReadonlyHostCommand(cmd) {
		return "readonly", string(ToolRiskLow), false
	}
	dangerous := []string{
		"rm -rf /", "mkfs", "shutdown", "poweroff", "reboot", "init 0",
		"dd if=", "iptables -f", "userdel", "chown -r /", "chmod -r 777 /",
	}
	for _, keyword := range dangerous {
		if strings.Contains(trimmed, keyword) {
			return "dangerous", string(ToolRiskHigh), true
		}
	}
	return "mutating", string(ToolRiskMedium), false
}

func parseHostLabels(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			out := make([]string, 0, len(arr))
			for _, item := range arr {
				if s := strings.TrimSpace(item); s != "" {
					out = append(out, s)
				}
			}
			return out
		}
	}
	parts := strings.Split(trimmed, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func isReadonlyHostCommand(cmd string) bool {
	switch strings.TrimSpace(cmd) {
	case "hostname", "uptime", "df -h", "free -m", "ps aux --sort=-%cpu":
		return true
	default:
		return false
	}
}
