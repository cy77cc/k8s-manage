package ai

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/service/cicd"
	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type commandRisk string

const (
	commandRiskReadonly commandRisk = "readonly"
	commandRiskLow      commandRisk = "low"
	commandRiskHigh     commandRisk = "high"
)

type commandContext struct {
	CommandID string                 `json:"command_id"`
	TraceID   string                 `json:"trace_id"`
	Scene     string                 `json:"scene"`
	Text      string                 `json:"text"`
	Intent    string                 `json:"intent"`
	Params    map[string]any         `json:"params"`
	Missing   []string               `json:"missing"`
	Prompts   map[string]string      `json:"prompts,omitempty"`
	PlanHash  string                 `json:"plan_hash"`
	Risk      commandRisk            `json:"risk"`
	Action    commandAction          `json:"-"`
	Plan      map[string]any         `json:"plan"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

type commandAction struct {
	Intent       string
	Domain       string
	Description  string
	Required     []string
	Permission   string
	Mode         tools.ToolMode
	Risk         tools.ToolRisk
	Tool         string
	NextActions  []string
	Exec         func(ctx context.Context, h *handler, uid uint64, cc commandContext, approvalToken string) (map[string]any, error)
	BuildPreview func(ctx context.Context, h *handler, uid uint64, cc commandContext) (map[string]any, error)
}

type commandPreviewRequest struct {
	Command string         `json:"command" binding:"required"`
	Scene   string         `json:"scene"`
	Params  map[string]any `json:"params"`
}

type commandExecuteRequest struct {
	CommandID     string         `json:"command_id"`
	Command       string         `json:"command"`
	Scene         string         `json:"scene"`
	Params        map[string]any `json:"params"`
	Confirm       bool           `json:"confirm"`
	ApprovalToken string         `json:"approval_token"`
}

type commandResult struct {
	Status      string            `json:"status"`
	Summary     string            `json:"summary"`
	Artifacts   map[string]any    `json:"artifacts"`
	TraceID     string            `json:"trace_id"`
	NextActions []string          `json:"next_actions"`
	Plan        map[string]any    `json:"plan,omitempty"`
	Missing     []string          `json:"missing,omitempty"`
	Prompts     map[string]string `json:"prompts,omitempty"`
	Risk        commandRisk       `json:"risk"`
}

type commandHistoryItem struct {
	ID               string         `json:"id"`
	Command          string         `json:"command"`
	Intent           string         `json:"intent"`
	Status           string         `json:"status"`
	Risk             string         `json:"risk"`
	TraceID          string         `json:"trace_id"`
	PlanHash         string         `json:"plan_hash"`
	CreatedAt        time.Time      `json:"created_at"`
	ExecutionSummary string         `json:"execution_summary"`
	Plan             map[string]any `json:"plan,omitempty"`
	Result           map[string]any `json:"result,omitempty"`
}

func (h *handler) buildCommandRoutes() map[string]commandAction {
	routes := map[string]commandAction{}
	register := func(a commandAction) {
		routes[a.Intent] = a
	}
	register(commandAction{
		Intent:      "service.status",
		Domain:      "service",
		Description: "查询服务状态",
		Required:    []string{"service_id"},
		Permission:  "ai:tool:read",
		Mode:        tools.ToolModeReadonly,
		Risk:        tools.ToolRiskLow,
		Tool:        "service_get_detail",
		NextActions: []string{"deployment.release.preview", "ops.aggregate.status"},
	})
	register(commandAction{
		Intent:      "deployment.release",
		Domain:      "deployment",
		Description: "触发发布",
		Required:    []string{"service_id", "deployment_id", "env", "version", "runtime_type"},
		Permission:  "cicd:release:run",
		Mode:        tools.ToolModeMutating,
		Risk:        tools.ToolRiskMedium,
		NextActions: []string{"cicd.release.approve", "ops.aggregate.status"},
		Exec:        executeTriggerRelease,
	})
	register(commandAction{
		Intent:      "deployment.rollback",
		Domain:      "deployment",
		Description: "回滚发布",
		Required:    []string{"release_id", "target_version"},
		Permission:  "cicd:release:rollback",
		Mode:        tools.ToolModeMutating,
		Risk:        tools.ToolRiskHigh,
		NextActions: []string{"ops.aggregate.status"},
		Exec:        executeRollbackRelease,
	})
	register(commandAction{
		Intent:      "cicd.release.approve",
		Domain:      "cicd",
		Description: "审批发布",
		Required:    []string{"release_id", "approve"},
		Permission:  "cicd:release:approve",
		Mode:        tools.ToolModeMutating,
		Risk:        tools.ToolRiskHigh,
		NextActions: []string{"ops.aggregate.status"},
		Exec:        executeReleaseApproval,
	})
	register(commandAction{
		Intent:      "cmdb.asset.search",
		Domain:      "cmdb",
		Description: "查询资产与关系",
		Required:    []string{},
		Permission:  "cmdb:read",
		Mode:        tools.ToolModeReadonly,
		Risk:        tools.ToolRiskLow,
		NextActions: []string{"ops.aggregate.status"},
		Exec:        executeCMDBSearch,
	})
	register(commandAction{
		Intent:      "monitor.alerts",
		Domain:      "monitoring",
		Description: "查询告警",
		Required:    []string{},
		Permission:  "monitoring:read",
		Mode:        tools.ToolModeReadonly,
		Risk:        tools.ToolRiskLow,
		NextActions: []string{"service.status"},
		Exec:        executeAlertSearch,
	})
	register(commandAction{
		Intent:      "ops.aggregate.status",
		Domain:      "aggregate",
		Description: "跨域聚合状态",
		Required:    []string{},
		Permission:  "ai:tool:read",
		Mode:        tools.ToolModeReadonly,
		Risk:        tools.ToolRiskLow,
		NextActions: []string{"deployment.release", "monitor.alerts"},
		Exec:        executeAggregate,
	})
	register(commandAction{
		Intent:      "host.batch.exec.preview",
		Domain:      "host",
		Description: "主机批量命令预览",
		Required:    []string{"host_ids", "command"},
		Permission:  "ai:tool:read",
		Mode:        tools.ToolModeReadonly,
		Risk:        tools.ToolRiskMedium,
		Tool:        "host_batch_exec_preview",
		NextActions: []string{"host.batch.exec.apply"},
		BuildPreview: func(ctx context.Context, h *handler, uid uint64, cc commandContext) (map[string]any, error) {
			return h.executeWithSpecificTool(ctx, uid, cc, "", "host_batch_exec_preview")
		},
	})
	register(commandAction{
		Intent:      "host.batch.exec.apply",
		Domain:      "host",
		Description: "主机批量命令执行",
		Required:    []string{"host_ids", "command"},
		Permission:  "ai:tool:execute",
		Mode:        tools.ToolModeMutating,
		Risk:        tools.ToolRiskHigh,
		Tool:        "host_batch_exec_apply",
		NextActions: []string{"ops.aggregate.status"},
		BuildPreview: func(ctx context.Context, h *handler, uid uint64, cc commandContext) (map[string]any, error) {
			return h.executeWithSpecificTool(ctx, uid, cc, "", "host_batch_exec_preview")
		},
	})
	register(commandAction{
		Intent:      "host.batch.status.update",
		Domain:      "host",
		Description: "主机批量状态更新",
		Required:    []string{"host_ids", "action"},
		Permission:  "ai:tool:execute",
		Mode:        tools.ToolModeMutating,
		Risk:        tools.ToolRiskMedium,
		Tool:        "host_batch_status_update",
		NextActions: []string{"host.batch.exec.preview"},
	})
	register(commandAction{
		Intent:      "host.script.exec.apply",
		Domain:      "host",
		Description: "主机脚本上传并执行",
		Required:    []string{"host_ids", "script_content"},
		Permission:  "ai:tool:execute",
		Mode:        tools.ToolModeMutating,
		Risk:        tools.ToolRiskHigh,
		NextActions: []string{"ops.aggregate.status"},
		Exec:        executeHostScriptApply,
		BuildPreview: func(ctx context.Context, h *handler, uid uint64, cc commandContext) (map[string]any, error) {
			ids := hostIDsFromParams(cc.Params)
			script := resolveScriptContent(cc.Params)
			hash := sha256.Sum256([]byte(script))
			return map[string]any{
				"mode":           "script",
				"target_count":   len(ids),
				"host_ids":       ids,
				"script_hash":    hex.EncodeToString(hash[:]),
				"script_lines":   len(strings.Split(script, "\n")),
				"execution_root": fmt.Sprintf("/tmp/opsx/%s", cc.CommandID),
				"timeout_sec":    scriptTimeoutSec(cc.Params),
			}, nil
		},
	})
	return routes
}

func classifyRisk(action commandAction) commandRisk {
	if action.Mode == tools.ToolModeReadonly {
		return commandRiskReadonly
	}
	if action.Risk == tools.ToolRiskHigh {
		return commandRiskHigh
	}
	return commandRiskLow
}

func detectIntent(command string) string {
	v := strings.ToLower(strings.TrimSpace(command))
	switch {
	case strings.Contains(v, "host.script.exec.apply") || strings.Contains(v, "脚本执行") || strings.Contains(v, "上传脚本执行"):
		return "host.script.exec.apply"
	case strings.Contains(v, "host.batch.exec.preview") || strings.Contains(v, "批量命令预览"):
		return "host.batch.exec.preview"
	case strings.Contains(v, "host.batch.exec.apply") || strings.Contains(v, "批量命令执行") || strings.Contains(v, "批量执行主机命令"):
		return "host.batch.exec.apply"
	case strings.Contains(v, "host.batch.status.update") || strings.Contains(v, "批量维护") || strings.Contains(v, "批量状态"):
		return "host.batch.status.update"
	case strings.Contains(v, "ops.aggregate.status") || strings.Contains(v, "聚合") || strings.Contains(v, "汇总"):
		return "ops.aggregate.status"
	case strings.Contains(v, "deployment.rollback") || strings.Contains(v, "回滚"):
		return "deployment.rollback"
	case strings.Contains(v, "cicd.release.approve") || strings.Contains(v, "审批"):
		return "cicd.release.approve"
	case strings.Contains(v, "deployment.release") || strings.Contains(v, "发布") || strings.Contains(v, "部署"):
		return "deployment.release"
	case strings.Contains(v, "monitor.alerts") || strings.Contains(v, "告警"):
		return "monitor.alerts"
	case strings.Contains(v, "cmdb") || strings.Contains(v, "资产"):
		return "cmdb.asset.search"
	case strings.Contains(v, "service.status") || strings.Contains(v, "服务"):
		return "service.status"
	default:
		return "ops.aggregate.status"
	}
}

func parseCommandParams(command string) map[string]any {
	out := map[string]any{}
	parts := strings.Fields(command)
	for _, item := range parts {
		idx := strings.Index(item, "=")
		if idx <= 0 || idx >= len(item)-1 {
			continue
		}
		k := strings.TrimSpace(item[:idx])
		v := strings.TrimSpace(item[idx+1:])
		if k == "" || v == "" {
			continue
		}
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			out[k] = n
			continue
		}
		if b, err := strconv.ParseBool(v); err == nil {
			out[k] = b
			continue
		}
		out[k] = v
	}
	return out
}

func mergeParams(base, extra map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func missingParams(required []string, params map[string]any) []string {
	miss := make([]string, 0)
	for _, key := range required {
		v, ok := params[key]
		if !ok {
			miss = append(miss, key)
			continue
		}
		switch x := v.(type) {
		case string:
			if strings.TrimSpace(x) == "" {
				miss = append(miss, key)
			}
		}
	}
	return miss
}

func paramPrompts(fields []string) map[string]string {
	if len(fields) == 0 {
		return nil
	}
	out := make(map[string]string, len(fields))
	for _, f := range fields {
		out[f] = fmt.Sprintf("请补充参数 `%s`", f)
	}
	return out
}

func (h *handler) buildCommandContext(command, scene string, params map[string]any) (commandContext, error) {
	routes := h.buildCommandRoutes()
	intent := detectIntent(command)
	action, ok := routes[intent]
	if !ok {
		return commandContext{}, fmt.Errorf("unsupported intent: %s", intent)
	}
	parsed := parseCommandParams(command)
	merged := mergeParams(parsed, params)
	miss := missingParams(action.Required, merged)
	traceID := fmt.Sprintf("trace-%d", time.Now().UnixNano())
	plan := map[string]any{
		"target": map[string]any{"domain": action.Domain, "intent": action.Intent},
		"steps":  []map[string]any{{"name": "validate_params", "status": "pending"}, {"name": "authorize", "status": "pending"}, {"name": "execute", "status": "pending"}},
		"params": merged,
		"risk":   classifyRisk(action),
	}
	hashInput := fmt.Sprintf("%s|%s|%s", command, intent, mustJSON(merged))
	sum := sha256.Sum256([]byte(hashInput))
	return commandContext{
		CommandID: fmt.Sprintf("cmd-%d", time.Now().UnixNano()),
		TraceID:   traceID,
		Scene:     normalizeScene(scene),
		Text:      strings.TrimSpace(command),
		Intent:    intent,
		Params:    merged,
		Missing:   miss,
		Prompts:   paramPrompts(miss),
		PlanHash:  hex.EncodeToString(sum[:]),
		Risk:      classifyRisk(action),
		Action:    action,
		Plan:      plan,
	}, nil
}

func (h *handler) previewCommand(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	var req commandPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	cc, err := h.buildCommandContext(req.Command, req.Scene, req.Params)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, err.Error())
		return
	}
	if !h.hasPermission(uid, cc.Action.Permission) {
		httpx.Fail(c, xcode.Forbidden, "permission denied")
		return
	}
	if err := h.enforceHostOperationPolicy(c.Request.Context(), cc); err != nil {
		httpx.Fail(c, xcode.Forbidden, err.Error())
		return
	}
	if err := h.store.saveCommandRecord(uid, cc, "previewed", nil, nil, ""); err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	result := commandResult{
		Status:      "previewed",
		Summary:     fmt.Sprintf("命令 `%s` 预览已生成", cc.Intent),
		Artifacts:   map[string]any{"target": cc.Action.Domain, "params": cc.Params},
		TraceID:     cc.TraceID,
		NextActions: cc.Action.NextActions,
		Plan:        cc.Plan,
		Risk:        cc.Risk,
	}
	if cc.Action.BuildPreview != nil && len(cc.Missing) == 0 {
		previewArtifacts, previewErr := cc.Action.BuildPreview(c.Request.Context(), h, uid, cc)
		if previewErr != nil {
			result.Artifacts["preview_error"] = previewErr.Error()
		} else if len(previewArtifacts) > 0 {
			result.Artifacts["preview"] = previewArtifacts
		}
	}
	if isHostExecutionIntent(cc.Intent) {
		result.Artifacts["host_execution_plan"] = h.buildHostExecutionPlan(cc)
	}
	if len(cc.Missing) > 0 {
		result.Status = "blocked"
		result.Summary = "参数未补全，无法执行"
		result.Missing = cc.Missing
		result.Prompts = cc.Prompts
	}
	if cc.Risk == commandRiskHigh || isHostMutatingIntent(cc.Intent) {
		ticket := h.store.newApproval(uid, approvalTicket{
			Tool:   cc.Intent,
			Params: cc.Params,
			Risk:   cc.Action.Risk,
			Mode:   tools.ToolModeMutating,
		})
		result.Artifacts["approval_required"] = true
		result.Artifacts["approval_token"] = ticket.ID
		result.Artifacts["approval_expires_at"] = ticket.ExpiresAt
		result.Artifacts["can_review"] = h.hasPermission(uid, "ai:approval:review")
	}
	result.Artifacts["command_id"] = cc.CommandID
	result.Artifacts["intent"] = cc.Intent
	result.Artifacts["plan_hash"] = cc.PlanHash
	httpx.OK(c, result)
}

func (h *handler) executeCommand(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	var req commandExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	if !req.Confirm {
		httpx.Fail(c, xcode.ParamError, "confirm=true is required")
		return
	}
	cc, err := h.loadOrBuildCommandContext(uid, req)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, err.Error())
		return
	}
	if !h.hasPermission(uid, cc.Action.Permission) {
		httpx.Fail(c, xcode.Forbidden, "permission denied")
		return
	}
	if len(cc.Missing) > 0 {
		httpx.Fail(c, xcode.ParamError, "missing required params")
		return
	}
	if err := h.enforceHostOperationPolicy(c.Request.Context(), cc); err != nil {
		httpx.Fail(c, xcode.Forbidden, err.Error())
		return
	}
	approvalContext := map[string]any{}
	if cc.Risk == commandRiskHigh || isHostMutatingIntent(cc.Intent) {
		token := strings.TrimSpace(req.ApprovalToken)
		if token == "" {
			httpx.Fail(c, xcode.Forbidden, "approval token required")
			return
		}
		ticket, ok := h.store.getApproval(token)
		if !ok || ticket.Status != "approved" {
			httpx.Fail(c, xcode.Forbidden, "approval not approved")
			return
		}
		approvalContext["approval_token"] = token
		approvalContext["approved_by"] = ticket.ReviewUID
	}

	var artifacts map[string]any
	if cc.Action.Exec != nil {
		artifacts, err = cc.Action.Exec(c.Request.Context(), h, uid, cc, strings.TrimSpace(req.ApprovalToken))
	} else {
		artifacts, err = h.executeWithTool(c.Request.Context(), uid, cc, strings.TrimSpace(req.ApprovalToken))
	}
	artifacts = redactArtifactsMap(artifacts)
	result := commandResult{
		Status:      "succeeded",
		Summary:     fmt.Sprintf("命令 `%s` 执行成功", cc.Intent),
		Artifacts:   artifacts,
		TraceID:     cc.TraceID,
		NextActions: cc.Action.NextActions,
		Risk:        cc.Risk,
	}
	if err != nil {
		result.Status = "failed"
		result.Summary = err.Error()
		if result.Artifacts == nil {
			result.Artifacts = map[string]any{}
		}
		result.Artifacts["error"] = err.Error()
	}
	if isHostExecutionIntent(cc.Intent) {
		records := h.persistHostExecutionRecords(c.Request.Context(), uid, cc, result.Artifacts)
		if result.Artifacts == nil {
			result.Artifacts = map[string]any{}
		}
		result.Artifacts["host_execution_records"] = records
	}
	_ = h.store.saveCommandRecord(uid, cc, result.Status, result.Artifacts, approvalContext, result.Summary)
	httpx.OK(c, result)
}

func (h *handler) loadOrBuildCommandContext(uid uint64, req commandExecuteRequest) (commandContext, error) {
	id := strings.TrimSpace(req.CommandID)
	if id != "" {
		rec, err := h.store.getCommandRecord(uid, id)
		if err == nil {
			return h.commandContextFromRecord(rec)
		}
	}
	if strings.TrimSpace(req.Command) == "" {
		return commandContext{}, fmt.Errorf("command is required")
	}
	return h.buildCommandContext(req.Command, req.Scene, req.Params)
}

func (h *handler) commandContextFromRecord(rec *model.AICommandExecution) (commandContext, error) {
	if rec == nil {
		return commandContext{}, fmt.Errorf("record not found")
	}
	routes := h.buildCommandRoutes()
	action, ok := routes[strings.TrimSpace(rec.Intent)]
	if !ok {
		return commandContext{}, fmt.Errorf("unsupported intent: %s", rec.Intent)
	}
	params := map[string]any{}
	_ = json.Unmarshal([]byte(rec.ParamsJSON), &params)
	plan := map[string]any{}
	_ = json.Unmarshal([]byte(rec.PlanJSON), &plan)
	missing := []string{}
	_ = json.Unmarshal([]byte(rec.MissingJSON), &missing)
	return commandContext{
		CommandID: rec.ID,
		TraceID:   rec.TraceID,
		Scene:     rec.Scene,
		Text:      rec.CommandText,
		Intent:    rec.Intent,
		Params:    params,
		Missing:   missing,
		Prompts:   paramPrompts(missing),
		PlanHash:  rec.PlanHash,
		Risk:      commandRisk(strings.TrimSpace(rec.Risk)),
		Action:    action,
		Plan:      plan,
	}, nil
}

func (h *handler) executeWithTool(ctx context.Context, uid uint64, cc commandContext, approvalToken string) (map[string]any, error) {
	if h.svcCtx.AI == nil {
		return nil, fmt.Errorf("ai runtime not initialized")
	}
	if strings.TrimSpace(cc.Action.Tool) == "" {
		return nil, fmt.Errorf("action %s has no executor", cc.Intent)
	}
	runCtx := tools.WithToolUser(ctx, uid, approvalToken)
	runCtx = tools.WithToolPolicyChecker(runCtx, h.toolPolicy)
	result, err := h.svcCtx.AI.RunTool(runCtx, cc.Action.Tool, cc.Params)
	artifacts := map[string]any{
		"tool":       cc.Action.Tool,
		"ok":         result.OK,
		"source":     result.Source,
		"latency_ms": result.LatencyMS,
		"data":       result.Data,
	}
	if err != nil {
		artifacts["error"] = err.Error()
		return artifacts, err
	}
	if !result.OK {
		return artifacts, fmt.Errorf("%s", result.Error)
	}
	return artifacts, nil
}

func (h *handler) executeWithSpecificTool(ctx context.Context, uid uint64, cc commandContext, approvalToken, toolName string) (map[string]any, error) {
	if h.svcCtx.AI == nil {
		return nil, fmt.Errorf("ai runtime not initialized")
	}
	runCtx := tools.WithToolUser(ctx, uid, approvalToken)
	runCtx = tools.WithToolPolicyChecker(runCtx, h.toolPolicy)
	result, err := h.svcCtx.AI.RunTool(runCtx, strings.TrimSpace(toolName), cc.Params)
	artifacts := map[string]any{
		"tool":       strings.TrimSpace(toolName),
		"ok":         result.OK,
		"source":     result.Source,
		"latency_ms": result.LatencyMS,
		"data":       result.Data,
	}
	if err != nil {
		artifacts["error"] = err.Error()
		return artifacts, err
	}
	if !result.OK {
		return artifacts, fmt.Errorf("%s", result.Error)
	}
	return artifacts, nil
}

func executeTriggerRelease(ctx context.Context, h *handler, uid uint64, cc commandContext, _ string) (map[string]any, error) {
	logic := cicd.NewLogic(h.svcCtx)
	req := cicd.TriggerReleaseReq{
		ServiceID:    uint(toInt64(cc.Params["service_id"])),
		DeploymentID: uint(toInt64(cc.Params["deployment_id"])),
		Env:          toString(cc.Params["env"]),
		RuntimeType:  toString(cc.Params["runtime_type"]),
		Version:      toString(cc.Params["version"]),
	}
	ctx = cicd.WithCommandAuditContext(ctx, cicd.CommandAuditContext{CommandID: cc.CommandID, Intent: cc.Intent, PlanHash: cc.PlanHash, TraceID: cc.TraceID, Summary: "release triggered"})
	resp, err := logic.TriggerRelease(ctx, uint(uid), req)
	if err != nil {
		return nil, err
	}
	return map[string]any{"release": resp}, nil
}

func executeRollbackRelease(ctx context.Context, h *handler, uid uint64, cc commandContext, _ string) (map[string]any, error) {
	logic := cicd.NewLogic(h.svcCtx)
	releaseID := uint(toInt64(cc.Params["release_id"]))
	targetVersion := toString(cc.Params["target_version"])
	comment := toString(cc.Params["comment"])
	ctx = cicd.WithCommandAuditContext(ctx, cicd.CommandAuditContext{CommandID: cc.CommandID, Intent: cc.Intent, PlanHash: cc.PlanHash, TraceID: cc.TraceID, Summary: "rollback executed"})
	resp, err := logic.RollbackRelease(ctx, uint(uid), releaseID, targetVersion, comment)
	if err != nil {
		return nil, err
	}
	return map[string]any{"rollback": resp}, nil
}

func executeReleaseApproval(ctx context.Context, h *handler, uid uint64, cc commandContext, _ string) (map[string]any, error) {
	logic := cicd.NewLogic(h.svcCtx)
	releaseID := uint(toInt64(cc.Params["release_id"]))
	approve := toBool(cc.Params["approve"])
	comment := toString(cc.Params["comment"])
	ctx = cicd.WithCommandAuditContext(ctx, cicd.CommandAuditContext{CommandID: cc.CommandID, Intent: cc.Intent, PlanHash: cc.PlanHash, TraceID: cc.TraceID, Summary: "release approval processed"})
	if approve {
		resp, err := logic.ApproveRelease(ctx, uint(uid), releaseID, comment)
		if err != nil {
			return nil, err
		}
		return map[string]any{"release": resp, "decision": "approved"}, nil
	}
	resp, err := logic.RejectRelease(ctx, uint(uid), releaseID, comment)
	if err != nil {
		return nil, err
	}
	return map[string]any{"release": resp, "decision": "rejected"}, nil
}

func executeCMDBSearch(ctx context.Context, h *handler, _ uint64, cc commandContext, _ string) (map[string]any, error) {
	limit := 20
	if v := toInt64(cc.Params["limit"]); v > 0 {
		limit = int(v)
	}
	keyword := strings.TrimSpace(toString(cc.Params["keyword"]))
	q := h.svcCtx.DB.WithContext(ctx).Model(&model.CMDBCI{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR ci_uid LIKE ?", like, like)
	}
	rows := make([]model.CMDBCI, 0)
	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	relCount := int64(0)
	_ = h.svcCtx.DB.WithContext(ctx).Model(&model.CMDBRelation{}).Count(&relCount).Error
	return map[string]any{"assets": rows, "relation_count": relCount}, nil
}

func executeAlertSearch(ctx context.Context, h *handler, _ uint64, cc commandContext, _ string) (map[string]any, error) {
	limit := 20
	if v := toInt64(cc.Params["limit"]); v > 0 {
		limit = int(v)
	}
	rows := make([]model.AlertEvent, 0)
	q := h.svcCtx.DB.WithContext(ctx).Model(&model.AlertEvent{})
	if status := strings.TrimSpace(toString(cc.Params["status"])); status != "" {
		q = q.Where("status = ?", status)
	} else {
		q = q.Where("status = ?", "firing")
	}
	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return map[string]any{"alerts": rows, "count": len(rows)}, nil
}

func executeAggregate(ctx context.Context, h *handler, _ uint64, cc commandContext, _ string) (map[string]any, error) {
	timeout := 5 * time.Second
	if v := toInt64(cc.Params["timeout_sec"]); v > 0 && v <= 30 {
		timeout = time.Duration(v) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	limit := 10
	if v := toInt64(cc.Params["limit"]); v > 0 {
		limit = int(v)
	}

	type part struct {
		name string
		data map[string]any
		err  error
	}
	jobs := []func(context.Context) part{
		func(c context.Context) part {
			rows := make([]model.Service, 0)
			err := h.svcCtx.DB.WithContext(c).Model(&model.Service{}).Order("id DESC").Limit(limit).Find(&rows).Error
			return part{name: "services", data: map[string]any{"items": rows, "count": len(rows)}, err: err}
		},
		func(c context.Context) part {
			rows := make([]model.CICDRelease, 0)
			err := h.svcCtx.DB.WithContext(c).Model(&model.CICDRelease{}).Order("id DESC").Limit(limit).Find(&rows).Error
			return part{name: "releases", data: map[string]any{"items": rows, "count": len(rows)}, err: err}
		},
		func(c context.Context) part {
			rows := make([]model.AlertEvent, 0)
			err := h.svcCtx.DB.WithContext(c).Model(&model.AlertEvent{}).Where("status = ?", "firing").Order("id DESC").Limit(limit).Find(&rows).Error
			return part{name: "alerts", data: map[string]any{"items": rows, "count": len(rows)}, err: err}
		},
		func(c context.Context) part {
			var relCount int64
			err := h.svcCtx.DB.WithContext(c).Model(&model.CMDBRelation{}).Count(&relCount).Error
			return part{name: "cmdb", data: map[string]any{"relation_count": relCount}, err: err}
		},
	}
	maxParallel := 2
	if v := toInt64(cc.Params["max_parallel"]); v > 0 && v <= int64(len(jobs)) {
		maxParallel = int(v)
	}
	sem := make(chan struct{}, maxParallel)
	outCh := make(chan part, len(jobs))
	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go func(fn func(context.Context) part) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				outCh <- part{name: "aggregate", err: ctx.Err()}
				return
			}
			defer func() { <-sem }()
			outCh <- fn(ctx)
		}(job)
	}
	wg.Wait()
	close(outCh)
	details := map[string]any{}
	errorsArr := make([]string, 0)
	for it := range outCh {
		if it.err != nil {
			errorsArr = append(errorsArr, fmt.Sprintf("%s: %v", it.name, it.err))
			continue
		}
		details[it.name] = it.data
	}
	keys := make([]string, 0, len(details))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	summary := fmt.Sprintf("聚合完成：%s", strings.Join(keys, ", "))
	if len(errorsArr) > 0 {
		summary = fmt.Sprintf("聚合部分完成，失败 %d 项", len(errorsArr))
		details["errors"] = errorsArr
	}
	return map[string]any{"summary": summary, "details": details}, nil
}

func (h *handler) listCommandHistory(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	limit := 20
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	list, err := h.store.listCommandRecords(uid, limit)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	out := make([]commandHistoryItem, 0, len(list))
	for i := range list {
		out = append(out, toCommandHistoryItem(&list[i]))
	}
	httpx.OK(c, gin.H{"list": out, "total": len(out)})
}

func (h *handler) getCommandHistory(c *gin.Context) {
	uid, ok := uidFromContext(c)
	if !ok {
		httpx.Fail(c, xcode.Unauthorized, "unauthorized")
		return
	}
	rec, err := h.store.getCommandRecord(uid, c.Param("id"))
	if err != nil {
		httpx.Fail(c, xcode.NotFound, "record not found")
		return
	}
	item := toCommandHistoryItem(rec)
	audits := make([]model.CICDAuditEvent, 0)
	if h.svcCtx.DB != nil {
		_ = h.svcCtx.DB.WithContext(c.Request.Context()).Where("trace_id = ? OR command_id = ?", rec.TraceID, rec.ID).Order("id DESC").Limit(50).Find(&audits).Error
	}
	hostResults := make([]model.AIHostExecutionRecord, 0)
	if h.svcCtx.DB != nil {
		_ = h.svcCtx.DB.WithContext(c.Request.Context()).Where("command_id = ?", rec.ID).Order("id ASC").Find(&hostResults).Error
	}
	httpx.OK(c, gin.H{"record": item, "audit_events": audits, "host_execution_records": hostResults})
}

func (h *handler) commandSuggestions(c *gin.Context) {
	examples := []map[string]any{
		{"command": "ops.aggregate.status limit=5", "hint": "一条命令汇总服务/发布/告警/资产关系"},
		{"command": "host.batch.exec.preview host_ids=[1,2] command='systemctl status kubelet'", "hint": "预览主机批量命令执行"},
		{"command": "host.script.exec.apply host_ids=[1,2]", "hint": "脚本上传执行（需二次确认与审批）"},
		{"command": "deployment.release service_id=1 deployment_id=2 env=prod runtime_type=k8s version=v1.2.3", "hint": "触发发布（需要确认）"},
		{"command": "deployment.rollback release_id=18 target_version=v1.2.2", "hint": "高风险回滚，需审批"},
		{"command": "monitor.alerts status=firing limit=20", "hint": "查看当前告警"},
	}
	httpx.OK(c, examples)
}

func toCommandHistoryItem(rec *model.AICommandExecution) commandHistoryItem {
	item := commandHistoryItem{
		ID:               rec.ID,
		Command:          rec.CommandText,
		Intent:           rec.Intent,
		Status:           rec.Status,
		Risk:             rec.Risk,
		TraceID:          rec.TraceID,
		PlanHash:         rec.PlanHash,
		CreatedAt:        rec.CreatedAt,
		ExecutionSummary: rec.ExecutionSummary,
	}
	_ = json.Unmarshal([]byte(rec.PlanJSON), &item.Plan)
	_ = json.Unmarshal([]byte(rec.ResultJSON), &item.Result)
	return item
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case uint:
		return int64(x)
	case uint64:
		return int64(x)
	case float64:
		return int64(x)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		return n
	default:
		return 0
	}
}

func toBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		b, _ := strconv.ParseBool(strings.TrimSpace(x))
		return b
	case int:
		return x != 0
	case int64:
		return x != 0
	default:
		return false
	}
}

func isHostExecutionIntent(intent string) bool {
	trimmed := strings.TrimSpace(intent)
	return strings.HasPrefix(trimmed, "host.batch.exec") || trimmed == "host.script.exec.apply"
}

func isHostMutatingIntent(intent string) bool {
	trimmed := strings.TrimSpace(intent)
	return trimmed == "host.batch.exec.apply" || trimmed == "host.batch.status.update" || trimmed == "host.script.exec.apply"
}

func (h *handler) buildHostExecutionPlan(cc commandContext) map[string]any {
	hostIDs := hostIDsFromParams(cc.Params)
	mode := "command"
	command := strings.TrimSpace(toString(cc.Params["command"]))
	scriptPath := ""
	if strings.TrimSpace(cc.Intent) == "host.script.exec.apply" {
		mode = "script"
		command = ""
		scriptPath = fmt.Sprintf("/tmp/opsx/%s/<host_id>/run.sh", cc.CommandID)
	}
	return map[string]any{
		"command_id":   cc.CommandID,
		"intent":       cc.Intent,
		"host_ids":     hostIDs,
		"target_count": len(hostIDs),
		"mode":         mode,
		"command":      command,
		"script_path":  scriptPath,
		"risk":         cc.Risk,
		"limits": map[string]any{
			"timeout_sec":  scriptTimeoutSec(cc.Params),
			"concurrency":  10,
			"approval_ttl": "10m",
		},
	}
}

func (h *handler) persistHostExecutionRecords(ctx context.Context, uid uint64, cc commandContext, artifacts map[string]any) []map[string]any {
	if h.svcCtx.DB == nil || len(artifacts) == 0 {
		return nil
	}
	hostIDs := hostIDsFromParams(cc.Params)
	if len(hostIDs) == 0 {
		return nil
	}
	execID := fmt.Sprintf("exec-%s", cc.CommandID)
	commandText := strings.TrimSpace(toString(cc.Params["command"]))
	now := time.Now()

	nodes := make(map[uint64]model.Node, len(hostIDs))
	var rows []model.Node
	if err := h.svcCtx.DB.WithContext(ctx).Where("id IN ?", hostIDs).Find(&rows).Error; err == nil {
		for i := range rows {
			nodes[uint64(rows[i].ID)] = rows[i]
		}
	}
	data, _ := artifacts["data"].(map[string]any)
	resultMap, _ := data["results"].(map[string]any)
	out := make([]map[string]any, 0, len(hostIDs))
	for _, hostID := range hostIDs {
		node := nodes[hostID]
		status := "succeeded"
		exitCode := 0
		stdout := ""
		stderr := ""
		if one, ok := resultMap[strconv.FormatUint(hostID, 10)].(map[string]any); ok {
			stdout = redactSensitiveText(toString(one["stdout"]))
			stderr = redactSensitiveText(toString(one["stderr"]))
			exitCode = int(toInt64(one["exit_code"]))
		}
		if exitCode != 0 || strings.TrimSpace(stderr) != "" {
			status = "failed"
		}
		record := model.AIHostExecutionRecord{
			ExecutionID: execID,
			CommandID:   cc.CommandID,
			HostID:      hostID,
			HostIP:      node.IP,
			HostName:    node.Name,
			CommandText: commandText,
			Status:      status,
			StdoutText:  redactSensitiveText(stdout),
			StderrText:  redactSensitiveText(stderr),
			ExitCode:    exitCode,
			StartedAt:   &now,
			FinishedAt:  &now,
			CreatedBy:   uid,
		}
		_ = h.svcCtx.DB.WithContext(ctx).Create(&record).Error
		out = append(out, map[string]any{
			"host_id":   hostID,
			"host_name": node.Name,
			"host_ip":   node.IP,
			"status":    status,
			"exit_code": exitCode,
		})
	}
	return out
}

func hostIDsFromParams(params map[string]any) []uint64 {
	raw, ok := params["host_ids"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []uint64:
		return append([]uint64{}, v...)
	case []int:
		out := make([]uint64, 0, len(v))
		for _, id := range v {
			if id > 0 {
				out = append(out, uint64(id))
			}
		}
		return out
	case []any:
		out := make([]uint64, 0, len(v))
		for _, it := range v {
			switch x := it.(type) {
			case int:
				if x > 0 {
					out = append(out, uint64(x))
				}
			case int64:
				if x > 0 {
					out = append(out, uint64(x))
				}
			case uint64:
				if x > 0 {
					out = append(out, x)
				}
			case float64:
				if x > 0 {
					out = append(out, uint64(x))
				}
			case string:
				if n, err := strconv.ParseUint(strings.TrimSpace(x), 10, 64); err == nil && n > 0 {
					out = append(out, n)
				}
			}
		}
		return out
	default:
		return nil
	}
}

func (h *handler) enforceHostOperationPolicy(ctx context.Context, cc commandContext) error {
	if !isHostExecutionIntent(cc.Intent) && !isHostMutatingIntent(cc.Intent) {
		return nil
	}
	if !config.AIGovernedHostExecutionEnabled() {
		return fmt.Errorf("ai governed host execution is disabled")
	}
	hostIDs := hostIDsFromParams(cc.Params)
	if len(hostIDs) == 0 {
		return fmt.Errorf("host_ids is required")
	}
	if len(hostIDs) > 50 {
		return fmt.Errorf("host scope exceeds policy limit: 50")
	}
	if timeout := scriptTimeoutSec(cc.Params); timeout > 600 {
		return fmt.Errorf("timeout exceeds policy limit: 600s")
	}
	for _, hostID := range hostIDs {
		var host model.Node
		if err := h.svcCtx.DB.WithContext(ctx).First(&host, hostID).Error; err != nil {
			return fmt.Errorf("host %d not found", hostID)
		}
		if ok, reason := hostlogic.EvaluateOperationalEligibility(&host); !ok {
			return fmt.Errorf("host %d is not eligible: %s", hostID, reason)
		}
	}
	command := strings.ToLower(strings.TrimSpace(toString(cc.Params["command"])))
	if command != "" {
		denylist := []string{"rm -rf /", "shutdown", "reboot", "mkfs", "dd if=", "iptables -f"}
		for _, keyword := range denylist {
			if strings.Contains(command, keyword) {
				return fmt.Errorf("command blocked by policy: %s", keyword)
			}
		}
	}
	return nil
}

func redactArtifactsMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return in
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch x := v.(type) {
		case string:
			out[k] = redactSensitiveText(x)
		case map[string]any:
			out[k] = redactArtifactsMap(x)
		case []any:
			arr := make([]any, 0, len(x))
			for _, item := range x {
				if m, ok := item.(map[string]any); ok {
					arr = append(arr, redactArtifactsMap(m))
				} else if s, ok := item.(string); ok {
					arr = append(arr, redactSensitiveText(s))
				} else {
					arr = append(arr, item)
				}
			}
			out[k] = arr
		default:
			out[k] = v
		}
	}
	return out
}

func redactSensitiveText(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return s
	}
	replacements := map[string]string{
		"password=":      "password=[REDACTED]",
		"token=":         "token=[REDACTED]",
		"authorization:": "authorization: [REDACTED]",
		"private_key":    "private_key:[REDACTED]",
	}
	out := s
	lower := strings.ToLower(out)
	for key, replacement := range replacements {
		idx := strings.Index(lower, key)
		if idx < 0 {
			continue
		}
		end := strings.Index(out[idx:], "\n")
		if end < 0 {
			out = out[:idx] + replacement
		} else {
			out = out[:idx] + replacement + out[idx+end:]
		}
		lower = strings.ToLower(out)
	}
	return out
}

func resolveScriptContent(params map[string]any) string {
	content := strings.TrimSpace(toString(params["script_content"]))
	if content == "" {
		content = strings.TrimSpace(toString(params["script"]))
	}
	return content
}

func scriptTimeoutSec(params map[string]any) int64 {
	timeout := toInt64(params["timeout_sec"])
	if timeout <= 0 {
		timeout = 60
	}
	if timeout > 1800 {
		timeout = 1800
	}
	return timeout
}

func executeHostScriptApply(ctx context.Context, h *handler, uid uint64, cc commandContext, _ string) (map[string]any, error) {
	hostIDs := hostIDsFromParams(cc.Params)
	if len(hostIDs) == 0 {
		return nil, fmt.Errorf("host_ids is required")
	}
	script := resolveScriptContent(cc.Params)
	if script == "" {
		return nil, fmt.Errorf("script_content is required")
	}
	timeout := scriptTimeoutSec(cc.Params)
	encoded := base64.StdEncoding.EncodeToString([]byte(script))
	results := map[string]any{}
	succeeded := 0
	failed := 0

	for _, hostID := range hostIDs {
		key := strconv.FormatUint(hostID, 10)
		var node model.Node
		if err := h.svcCtx.DB.WithContext(ctx).First(&node, hostID).Error; err != nil {
			results[key] = map[string]any{"stdout": "", "stderr": "host not found", "exit_code": 1}
			failed++
			continue
		}
		privateKey, passphrase, err := h.loadNodePrivateKey(ctx, &node)
		if err != nil {
			results[key] = map[string]any{"stdout": "", "stderr": err.Error(), "exit_code": 1}
			failed++
			continue
		}
		password := strings.TrimSpace(node.SSHPassword)
		if strings.TrimSpace(privateKey) != "" {
			password = ""
		}
		cli, err := sshclient.NewSSHClient(node.SSHUser, password, node.IP, node.Port, privateKey, passphrase)
		if err != nil {
			results[key] = map[string]any{"stdout": "", "stderr": err.Error(), "exit_code": 1}
			failed++
			continue
		}
		scriptDir := fmt.Sprintf("/tmp/opsx/%s/%d", cc.CommandID, hostID)
		scriptPath := fmt.Sprintf("%s/run.sh", scriptDir)
		cmd := fmt.Sprintf("mkdir -p %s && echo '%s' | base64 -d > %s && chmod 700 %s && timeout %d bash %s", scriptDir, encoded, scriptPath, scriptPath, timeout, scriptPath)
		out, runErr := sshclient.RunCommand(cli, cmd)
		_ = cli.Close()
		if runErr != nil {
			results[key] = map[string]any{"stdout": out, "stderr": runErr.Error(), "exit_code": 1, "script_path": scriptPath}
			failed++
			continue
		}
		results[key] = map[string]any{"stdout": out, "stderr": "", "exit_code": 0, "script_path": scriptPath}
		succeeded++
	}
	return map[string]any{
		"mode":         "script",
		"script_path":  fmt.Sprintf("/tmp/opsx/%s/<host_id>/run.sh", cc.CommandID),
		"target_count": len(hostIDs),
		"succeeded":    succeeded,
		"failed":       failed,
		"results":      results,
	}, nil
}

func (h *handler) loadNodePrivateKey(ctx context.Context, node *model.Node) (string, string, error) {
	if node == nil || node.SSHKeyID == nil {
		return "", "", nil
	}
	var key model.SSHKey
	if err := h.svcCtx.DB.WithContext(ctx).
		Select("id", "private_key", "passphrase", "encrypted").
		Where("id = ?", uint64(*node.SSHKeyID)).
		First(&key).Error; err != nil {
		return "", "", err
	}
	passphrase := strings.TrimSpace(key.Passphrase)
	if !key.Encrypted {
		return strings.TrimSpace(key.PrivateKey), passphrase, nil
	}
	privateKey, err := utils.DecryptText(strings.TrimSpace(key.PrivateKey), config.CFG.Security.EncryptionKey)
	if err != nil {
		return "", "", fmt.Errorf("decrypt private key: %w", err)
	}
	return privateKey, passphrase, nil
}
