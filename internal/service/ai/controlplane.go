package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "github.com/cy77cc/k8s-manage/api/project/v1"
	sshclient "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/model"
	projectlogic "github.com/cy77cc/k8s-manage/internal/service/project/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type riskLevel string

const (
	riskLow    riskLevel = "low"
	riskMedium riskLevel = "medium"
	riskHigh   riskLevel = "high"
)

type toolMode string

const (
	modeReadonly toolMode = "readonly"
	modeMutating toolMode = "mutating"
)

type toolResult struct {
	OK        bool   `json:"ok"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
	Source    string `json:"source"`
	LatencyMS int64  `json:"latency_ms"`
}

type toolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Mode        toolMode       `json:"mode"`
	Risk        riskLevel      `json:"risk"`
	Schema      map[string]any `json:"schema"`
	Permission  string         `json:"permission"`
	Execute     toolExecutor   `json:"-"`
}

type toolExecutor func(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error)

type approvalTicket struct {
	ID         string         `json:"id"`
	Tool       string         `json:"tool"`
	Params     map[string]any `json:"params"`
	Risk       riskLevel      `json:"risk"`
	Mode       toolMode       `json:"mode"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"createdAt"`
	ExpiresAt  time.Time      `json:"expiresAt"`
	RequestUID uint64         `json:"requestUid"`
	ReviewUID  uint64         `json:"reviewUid,omitempty"`
}

type executionRecord struct {
	ID         string         `json:"id"`
	Tool       string         `json:"tool"`
	Params     map[string]any `json:"params"`
	Mode       toolMode       `json:"mode"`
	Status     string         `json:"status"`
	Result     *toolResult    `json:"result,omitempty"`
	ApprovalID string         `json:"approvalId,omitempty"`
	RequestUID uint64         `json:"requestUid"`
	CreatedAt  time.Time      `json:"createdAt"`
	FinishedAt *time.Time     `json:"finishedAt,omitempty"`
	Error      string         `json:"error,omitempty"`
}

type aiControlPlane struct {
	svcCtx *svc.ServiceContext
	mu     sync.RWMutex

	tools      map[string]toolSpec
	approvals  map[string]*approvalTicket
	executions map[string]*executionRecord
}

var (
	controlPlaneOnce sync.Once
	controlPlaneInst *aiControlPlane

	serviceUnitRegexp = regexp.MustCompile(`^[a-zA-Z0-9_.@-]+$`)
)

func ensureControlPlane(svcCtx *svc.ServiceContext) *aiControlPlane {
	controlPlaneOnce.Do(func() {
		cp := &aiControlPlane{
			svcCtx:     svcCtx,
			tools:      make(map[string]toolSpec),
			approvals:  make(map[string]*approvalTicket),
			executions: make(map[string]*executionRecord),
		}
		cp.registerDefaultTools()
		controlPlaneInst = cp
	})
	return controlPlaneInst
}

func (cp *aiControlPlane) registerDefaultTools() {
	cp.tools["os.get_cpu_mem"] = toolSpec{
		Name:        "os.get_cpu_mem",
		Description: "读取 CPU/内存/负载概览",
		Mode:        modeReadonly,
		Risk:        riskLow,
		Permission:  "ai:tool:read",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"target": map[string]any{"type": "string", "description": "localhost 或节点ID/IP/Name"},
			},
		},
		Execute: execOSCPUMem,
	}
	cp.tools["os.get_disk_fs"] = toolSpec{
		Name:        "os.get_disk_fs",
		Description: "读取磁盘与文件系统占用",
		Mode:        modeReadonly,
		Risk:        riskLow,
		Permission:  "ai:tool:read",
		Schema:      map[string]any{"type": "object", "properties": map[string]any{"target": map[string]any{"type": "string"}}},
		Execute:     execOSDiskFS,
	}
	cp.tools["os.get_net_stat"] = toolSpec{
		Name:        "os.get_net_stat",
		Description: "读取网络连接与监听端口摘要",
		Mode:        modeReadonly,
		Risk:        riskLow,
		Permission:  "ai:tool:read",
		Schema:      map[string]any{"type": "object", "properties": map[string]any{"target": map[string]any{"type": "string"}}},
		Execute:     execOSNetStat,
	}
	cp.tools["os.get_process_top"] = toolSpec{
		Name:        "os.get_process_top",
		Description: "读取高占用进程列表",
		Mode:        modeReadonly,
		Risk:        riskLow,
		Permission:  "ai:tool:read",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"target": map[string]any{"type": "string"},
				"limit":  map[string]any{"type": "integer"},
			},
		},
		Execute: execOSProcessTop,
	}
	cp.tools["os.get_journal_tail"] = toolSpec{
		Name:        "os.get_journal_tail",
		Description: "按服务名读取系统日志窗口",
		Mode:        modeReadonly,
		Risk:        riskMedium,
		Permission:  "ai:tool:read",
		Schema: map[string]any{
			"type":     "object",
			"required": []string{"service"},
			"properties": map[string]any{
				"target":  map[string]any{"type": "string"},
				"service": map[string]any{"type": "string"},
				"lines":   map[string]any{"type": "integer"},
			},
		},
		Execute: execOSJournalTail,
	}
	cp.tools["os.get_container_runtime"] = toolSpec{
		Name:        "os.get_container_runtime",
		Description: "读取容器运行时摘要（docker/containerd）",
		Mode:        modeReadonly,
		Risk:        riskLow,
		Permission:  "ai:tool:read",
		Schema:      map[string]any{"type": "object", "properties": map[string]any{"target": map[string]any{"type": "string"}}},
		Execute:     execOSContainerRuntime,
	}
	cp.tools["k8s.list_resources"] = toolSpec{
		Name:        "k8s.list_resources",
		Description: "按类型列出 K8s 资源",
		Mode:        modeReadonly,
		Risk:        riskLow,
		Permission:  "ai:tool:read",
		Schema: map[string]any{
			"type":     "object",
			"required": []string{"resource"},
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "integer"},
				"namespace":  map[string]any{"type": "string"},
				"resource":   map[string]any{"type": "string", "enum": []string{"pods", "services", "deployments", "nodes"}},
				"limit":      map[string]any{"type": "integer"},
			},
		},
		Execute: execK8sListResources,
	}
	cp.tools["k8s.get_events"] = toolSpec{
		Name:        "k8s.get_events",
		Description: "读取 K8s 事件",
		Mode:        modeReadonly,
		Risk:        riskLow,
		Permission:  "ai:tool:read",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "integer"},
				"namespace":  map[string]any{"type": "string"},
				"limit":      map[string]any{"type": "integer"},
			},
		},
		Execute: execK8sGetEvents,
	}
	cp.tools["k8s.get_pod_logs"] = toolSpec{
		Name:        "k8s.get_pod_logs",
		Description: "读取 Pod 日志",
		Mode:        modeReadonly,
		Risk:        riskMedium,
		Permission:  "ai:tool:read",
		Schema: map[string]any{
			"type":     "object",
			"required": []string{"pod"},
			"properties": map[string]any{
				"cluster_id": map[string]any{"type": "integer"},
				"namespace":  map[string]any{"type": "string"},
				"pod":        map[string]any{"type": "string"},
				"container":  map[string]any{"type": "string"},
				"tail_lines": map[string]any{"type": "integer"},
			},
		},
		Execute: execK8sGetPodLogs,
	}
	cp.tools["service.get_detail"] = toolSpec{
		Name:        "service.get_detail",
		Description: "查询服务详情",
		Mode:        modeReadonly,
		Risk:        riskLow,
		Permission:  "ai:tool:read",
		Schema: map[string]any{
			"type":     "object",
			"required": []string{"service_id"},
			"properties": map[string]any{
				"service_id": map[string]any{"type": "integer"},
			},
		},
		Execute: execServiceGetDetail,
	}
	cp.tools["service.deploy_preview"] = toolSpec{
		Name:        "service.deploy_preview",
		Description: "预览服务部署动作",
		Mode:        modeReadonly,
		Risk:        riskMedium,
		Permission:  "ai:tool:read",
		Schema: map[string]any{
			"type":     "object",
			"required": []string{"service_id", "cluster_id"},
			"properties": map[string]any{
				"service_id": map[string]any{"type": "integer"},
				"cluster_id": map[string]any{"type": "integer"},
			},
		},
		Execute: execServiceDeployPreview,
	}
	cp.tools["service.deploy_apply"] = toolSpec{
		Name:        "service.deploy_apply",
		Description: "执行服务部署（需审批）",
		Mode:        modeMutating,
		Risk:        riskHigh,
		Permission:  "ai:tool:execute",
		Schema: map[string]any{
			"type":     "object",
			"required": []string{"service_id", "cluster_id"},
			"properties": map[string]any{
				"service_id": map[string]any{"type": "integer"},
				"cluster_id": map[string]any{"type": "integer"},
			},
		},
		Execute: execServiceDeployApply,
	}
	cp.tools["host.ssh_exec_readonly"] = toolSpec{
		Name:        "host.ssh_exec_readonly",
		Description: "远程只读命令执行（白名单）",
		Mode:        modeReadonly,
		Risk:        riskMedium,
		Permission:  "ai:tool:read",
		Schema: map[string]any{
			"type":     "object",
			"required": []string{"host_id", "command"},
			"properties": map[string]any{
				"host_id": map[string]any{"type": "integer"},
				"command": map[string]any{"type": "string", "enum": []string{"hostname", "uptime", "df -h", "free -m", "ps aux --sort=-%cpu"}},
			},
		},
		Execute: execHostSSHReadonly,
	}
}

func (cp *aiControlPlane) listCapabilities(uid uint64) []toolSpec {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	out := make([]toolSpec, 0, len(cp.tools))
	for _, t := range cp.tools {
		if !hasAIPermission(cp.svcCtx, uid, t.Permission) {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (cp *aiControlPlane) previewTool(uid uint64, toolName string, params map[string]any) (map[string]any, error) {
	cp.mu.RLock()
	spec, ok := cp.tools[toolName]
	cp.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
	if !hasAIPermission(cp.svcCtx, uid, spec.Permission) {
		return nil, errors.New("permission denied")
	}
	if params == nil {
		params = map[string]any{}
	}
	if err := validateParams(spec.Name, params); err != nil {
		return nil, err
	}

	preview := map[string]any{
		"tool":   spec.Name,
		"mode":   spec.Mode,
		"risk":   spec.Risk,
		"params": params,
	}
	if spec.Mode == modeMutating {
		ticket := cp.createApproval(uid, spec, params)
		preview["approval_required"] = true
		preview["approval_token"] = ticket.ID
		preview["expiresAt"] = ticket.ExpiresAt
		preview["previewDiff"] = "Mutating operation requires approval."
	} else {
		preview["approval_required"] = false
		preview["previewDiff"] = "Readonly operation."
	}
	return preview, nil
}

func (cp *aiControlPlane) executeTool(ctx context.Context, uid uint64, toolName string, params map[string]any, approvalToken string) (*executionRecord, error) {
	cp.mu.RLock()
	spec, ok := cp.tools[toolName]
	cp.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
	if !hasAIPermission(cp.svcCtx, uid, spec.Permission) {
		return nil, errors.New("permission denied")
	}
	if params == nil {
		params = map[string]any{}
	}
	if err := validateParams(spec.Name, params); err != nil {
		return nil, err
	}

	rec := &executionRecord{
		ID:         fmt.Sprintf("exe-%d", time.Now().UnixNano()),
		Tool:       spec.Name,
		Params:     params,
		Mode:       spec.Mode,
		Status:     "running",
		RequestUID: uid,
		CreatedAt:  time.Now(),
	}
	if spec.Mode == modeMutating {
		if approvalToken == "" {
			rec.Status = "failed"
			rec.Error = "approval_token required for mutating tool"
			now := time.Now()
			rec.FinishedAt = &now
			cp.saveExecution(rec)
			return rec, errors.New(rec.Error)
		}
		ticket, err := cp.consumeApproval(uid, approvalToken, spec.Name)
		if err != nil {
			rec.Status = "failed"
			rec.Error = err.Error()
			now := time.Now()
			rec.FinishedAt = &now
			cp.saveExecution(rec)
			return rec, err
		}
		rec.ApprovalID = ticket.ID
	}

	start := time.Now()
	data, source, err := spec.Execute(ctx, cp.svcCtx, params)
	finish := time.Now()
	if err != nil {
		rec.Status = "failed"
		rec.Error = err.Error()
		rec.Result = &toolResult{OK: false, Error: err.Error(), Source: source, LatencyMS: time.Since(start).Milliseconds()}
		rec.FinishedAt = &finish
		cp.saveExecution(rec)
		return rec, err
	}
	rec.Status = "succeeded"
	rec.Result = &toolResult{OK: true, Data: data, Source: source, LatencyMS: time.Since(start).Milliseconds()}
	rec.FinishedAt = &finish
	cp.saveExecution(rec)
	return rec, nil
}

func (cp *aiControlPlane) saveExecution(rec *executionRecord) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.executions[rec.ID] = rec
}

func (cp *aiControlPlane) getExecution(id string) (*executionRecord, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	rec, ok := cp.executions[id]
	return rec, ok
}

func (cp *aiControlPlane) createApproval(uid uint64, spec toolSpec, params map[string]any) *approvalTicket {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	t := &approvalTicket{
		ID:         fmt.Sprintf("apv-%d", time.Now().UnixNano()),
		Tool:       spec.Name,
		Params:     params,
		Risk:       spec.Risk,
		Mode:       spec.Mode,
		Status:     "pending",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		RequestUID: uid,
	}
	cp.approvals[t.ID] = t
	return t
}

func (cp *aiControlPlane) getApproval(id string) (*approvalTicket, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	t, ok := cp.approvals[id]
	return t, ok
}

func (cp *aiControlPlane) confirmApproval(reviewUID uint64, id string, approve bool) (*approvalTicket, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	t, ok := cp.approvals[id]
	if !ok {
		return nil, errors.New("approval not found")
	}
	if time.Now().After(t.ExpiresAt) {
		t.Status = "expired"
		return nil, errors.New("approval expired")
	}
	if t.Status != "pending" {
		return t, nil
	}
	if approve {
		t.Status = "approved"
	} else {
		t.Status = "rejected"
	}
	t.ReviewUID = reviewUID
	return t, nil
}

func (cp *aiControlPlane) consumeApproval(uid uint64, id string, toolName string) (*approvalTicket, error) {
	cp.mu.RLock()
	t, ok := cp.approvals[id]
	cp.mu.RUnlock()
	if !ok {
		return nil, errors.New("approval not found")
	}
	if t.Tool != toolName {
		return nil, errors.New("approval tool mismatch")
	}
	if t.RequestUID != uid && !isAdminUser(cp.svcCtx, uid) {
		return nil, errors.New("approval owner mismatch")
	}
	if time.Now().After(t.ExpiresAt) {
		return nil, errors.New("approval expired")
	}
	if t.Status != "approved" {
		return nil, errors.New("approval not approved")
	}
	return t, nil
}

func capabilitiesHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := getUIDFromContext(c)
		if !ok {
			c.JSON(401, gin.H{"code": 401, "msg": "unauthorized"})
			return
		}
		cp := ensureControlPlane(svcCtx)
		specs := cp.listCapabilities(uid)
		c.JSON(200, gin.H{"code": 1000, "msg": "ok", "data": specs})
	}
}

func previewToolHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := getUIDFromContext(c)
		if !ok {
			c.JSON(401, gin.H{"code": 401, "msg": "unauthorized"})
			return
		}
		var req struct {
			Tool   string         `json:"tool" binding:"required"`
			Params map[string]any `json:"params"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		cp := ensureControlPlane(svcCtx)
		out, err := cp.previewTool(uid, req.Tool, req.Params)
		if err != nil {
			c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.JSON(200, gin.H{"code": 1000, "msg": "ok", "data": out})
	}
}

func executeToolHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := getUIDFromContext(c)
		if !ok {
			c.JSON(401, gin.H{"code": 401, "msg": "unauthorized"})
			return
		}
		var req struct {
			Tool          string         `json:"tool" binding:"required"`
			Params        map[string]any `json:"params"`
			ApprovalToken string         `json:"approval_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		cp := ensureControlPlane(svcCtx)
		rec, err := cp.executeTool(c.Request.Context(), uid, req.Tool, req.Params, req.ApprovalToken)
		if err != nil {
			c.JSON(200, gin.H{"code": 1000, "msg": "execution failed", "data": rec})
			return
		}
		c.JSON(200, gin.H{"code": 1000, "msg": "ok", "data": rec})
	}
}

func executionHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := getUIDFromContext(c); !ok {
			c.JSON(401, gin.H{"code": 401, "msg": "unauthorized"})
			return
		}
		cp := ensureControlPlane(svcCtx)
		rec, ok := cp.getExecution(c.Param("id"))
		if !ok {
			c.JSON(404, gin.H{"code": 404, "msg": "execution not found"})
			return
		}
		c.JSON(200, gin.H{"code": 1000, "msg": "ok", "data": rec})
	}
}

func createApprovalHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := getUIDFromContext(c)
		if !ok {
			c.JSON(401, gin.H{"code": 401, "msg": "unauthorized"})
			return
		}
		var req struct {
			Tool   string         `json:"tool" binding:"required"`
			Params map[string]any `json:"params"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		cp := ensureControlPlane(svcCtx)
		cp.mu.RLock()
		spec, ok := cp.tools[req.Tool]
		cp.mu.RUnlock()
		if !ok {
			c.JSON(404, gin.H{"code": 404, "msg": "tool not found"})
			return
		}
		if spec.Mode == modeReadonly {
			c.JSON(400, gin.H{"code": 400, "msg": "readonly tool does not require approval"})
			return
		}
		ticket := cp.createApproval(uid, spec, req.Params)
		c.JSON(200, gin.H{"code": 1000, "msg": "ok", "data": ticket})
	}
}

func confirmApprovalHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := getUIDFromContext(c)
		if !ok {
			c.JSON(401, gin.H{"code": 401, "msg": "unauthorized"})
			return
		}
		if !hasAIPermission(svcCtx, uid, "ai:approval:review") {
			c.JSON(403, gin.H{"code": 403, "msg": "permission denied"})
			return
		}
		var req struct {
			Approve bool `json:"approve"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		cp := ensureControlPlane(svcCtx)
		t, err := cp.confirmApproval(uid, c.Param("id"), req.Approve)
		if err != nil {
			c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.JSON(200, gin.H{"code": 1000, "msg": "ok", "data": t})
	}
}

func validateParams(tool string, params map[string]any) error {
	switch tool {
	case "k8s.get_pod_logs":
		if strings.TrimSpace(toString(params["pod"])) == "" {
			return errors.New("pod is required")
		}
	case "k8s.list_resources":
		r := strings.ToLower(strings.TrimSpace(toString(params["resource"])))
		if r == "" {
			return errors.New("resource is required")
		}
		if r != "pods" && r != "services" && r != "deployments" && r != "nodes" {
			return errors.New("resource must be one of pods/services/deployments/nodes")
		}
	case "service.get_detail", "service.deploy_preview", "service.deploy_apply":
		if toInt(params["service_id"]) <= 0 {
			return errors.New("service_id is required")
		}
	case "host.ssh_exec_readonly":
		if toInt(params["host_id"]) <= 0 {
			return errors.New("host_id is required")
		}
		cmd := strings.TrimSpace(toString(params["command"]))
		if !isReadonlyHostCommand(cmd) {
			return errors.New("command is not in readonly whitelist")
		}
	case "os.get_journal_tail":
		svc := strings.TrimSpace(toString(params["service"]))
		if svc == "" {
			return errors.New("service is required")
		}
		if !serviceUnitRegexp.MatchString(svc) {
			return errors.New("service contains illegal characters")
		}
	}
	return nil
}

func getUIDFromContext(c *gin.Context) (uint64, bool) {
	uid, ok := c.Get("uid")
	if !ok {
		return 0, false
	}
	return toUint64(uid), true
}

func toUint64(v any) uint64 {
	switch x := v.(type) {
	case uint:
		return uint64(x)
	case uint64:
		return x
	case int:
		return uint64(x)
	case int64:
		return uint64(x)
	case float64:
		return uint64(x)
	default:
		return 0
	}
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	default:
		return fmt.Sprintf("%v", x)
	}
}

func toInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case uint64:
		return int(x)
	case json.Number:
		n, _ := strconv.Atoi(x.String())
		return n
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(x))
		return n
	default:
		return 0
	}
}

func hasAIPermission(svcCtx *svc.ServiceContext, uid uint64, required string) bool {
	if uid == 0 {
		return false
	}
	if isAdminUser(svcCtx, uid) {
		return true
	}
	perms, err := fetchPermissionsByUserID(svcCtx, uid)
	if err != nil {
		return false
	}
	if required == "" {
		return true
	}
	parts := strings.Split(required, ":")
	resource := required
	if len(parts) >= 1 {
		resource = parts[0]
	}
	for _, p := range perms {
		if p == required || p == resource+":*" || p == "*:*" {
			return true
		}
	}
	return false
}

func isAdminUser(svcCtx *svc.ServiceContext, userID uint64) bool {
	if svcCtx == nil || svcCtx.DB == nil || userID == 0 {
		return false
	}
	var u model.User
	if err := svcCtx.DB.Select("id", "username").Where("id = ?", userID).First(&u).Error; err == nil {
		if strings.EqualFold(strings.TrimSpace(u.Username), "admin") {
			return true
		}
	}
	type roleRow struct {
		Code string `gorm:"column:code"`
	}
	var rows []roleRow
	err := svcCtx.DB.Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return false
	}
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.Code), "admin") {
			return true
		}
	}
	return false
}

func fetchPermissionsByUserID(svcCtx *svc.ServiceContext, userID uint64) ([]string, error) {
	type row struct {
		Code string `gorm:"column:code"`
	}
	var rows []row
	err := svcCtx.DB.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Code)
	}
	return out, nil
}

func resolveK8sClient(svcCtx *svc.ServiceContext, params map[string]any) (*kubernetes.Clientset, string, error) {
	clusterID := toInt(params["cluster_id"])
	if clusterID > 0 {
		var cluster model.Cluster
		if err := svcCtx.DB.First(&cluster, clusterID).Error; err == nil && strings.TrimSpace(cluster.KubeConfig) != "" {
			cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(cluster.KubeConfig))
			if err != nil {
				return nil, "cluster_kubeconfig", err
			}
			cli, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return nil, "cluster_kubeconfig", err
			}
			return cli, "cluster_kubeconfig", nil
		}
	}
	if svcCtx.Clientset != nil {
		return svcCtx.Clientset, "default_clientset", nil
	}
	return nil, "fallback", errors.New("k8s client unavailable")
}

func runLocalCommand(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, name, args...)
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if cctx.Err() == context.DeadlineExceeded {
		return text, errors.New("command timeout")
	}
	return text, err
}

func isReadonlyHostCommand(cmd string) bool {
	switch strings.TrimSpace(cmd) {
	case "hostname", "uptime", "df -h", "free -m", "ps aux --sort=-%cpu":
		return true
	default:
		return false
	}
}

func resolveNodeByTarget(svcCtx *svc.ServiceContext, target string) (*model.Node, error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" || trimmed == "localhost" {
		return nil, nil
	}
	var node model.Node
	if id, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
		if err := svcCtx.DB.First(&node, id).Error; err == nil {
			return &node, nil
		}
	}
	if err := svcCtx.DB.Where("ip = ? OR name = ? OR hostname = ?", trimmed, trimmed, trimmed).First(&node).Error; err != nil {
		return nil, errors.New("target not in host whitelist")
	}
	return &node, nil
}

func runOnTarget(ctx context.Context, svcCtx *svc.ServiceContext, target string, localName string, localArgs []string, remoteCmd string) (string, string, error) {
	node, err := resolveNodeByTarget(svcCtx, target)
	if err != nil {
		return "", "target_check", err
	}
	if node == nil {
		out, err := runLocalCommand(ctx, 6*time.Second, localName, localArgs...)
		return out, "local", err
	}
	cli, err := sshclient.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, "")
	if err != nil {
		return "", "remote_ssh", err
	}
	defer cli.Close()
	out, err := sshclient.RunCommand(cli, remoteCmd)
	return out, "remote_ssh", err
}

func execOSCPUMem(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	target := toString(params["target"])
	loadavg := "n/a"
	if raw, src, err := runOnTarget(ctx, svcCtx, target, "cat", []string{"/proc/loadavg"}, "cat /proc/loadavg"); err == nil {
		loadavg = raw
		_ = src
	}
	meminfo, source, err := runOnTarget(ctx, svcCtx, target, "cat", []string{"/proc/meminfo"}, "cat /proc/meminfo")
	if err != nil {
		return nil, source, err
	}
	uptime, _, _ := runOnTarget(ctx, svcCtx, target, "uptime", nil, "uptime")
	return gin.H{"loadavg": loadavg, "meminfo": meminfo, "uptime": uptime}, source, nil
}

func execOSDiskFS(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	target := toString(params["target"])
	out, source, err := runOnTarget(ctx, svcCtx, target, "df", []string{"-h"}, "df -h")
	if err != nil {
		return nil, source, err
	}
	return gin.H{"filesystem": out}, source, nil
}

func execOSNetStat(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	target := toString(params["target"])
	dev, source, err := runOnTarget(ctx, svcCtx, target, "cat", []string{"/proc/net/dev"}, "cat /proc/net/dev")
	if err != nil {
		return nil, source, err
	}
	listen, _, _ := runOnTarget(ctx, svcCtx, target, "ss", []string{"-ltn"}, "ss -ltn")
	return gin.H{"net_dev": dev, "listening_ports": listen}, source, nil
}

func execOSProcessTop(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	target := toString(params["target"])
	limit := toInt(params["limit"])
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	cmd := fmt.Sprintf("ps aux --sort=-%%cpu | head -n %d", limit+1)
	out, source, err := runOnTarget(ctx, svcCtx, target, "ps", []string{"aux", "--sort=-%cpu"}, cmd)
	if err != nil {
		return nil, source, err
	}
	lines := strings.Split(out, "\n")
	if len(lines) > limit+1 {
		lines = lines[:limit+1]
	}
	return gin.H{"top_processes": strings.Join(lines, "\n"), "limit": limit}, source, nil
}

func execOSJournalTail(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	target := toString(params["target"])
	service := strings.TrimSpace(toString(params["service"]))
	lines := toInt(params["lines"])
	if lines <= 0 {
		lines = 200
	}
	if lines > 500 {
		lines = 500
	}
	if !serviceUnitRegexp.MatchString(service) {
		return nil, "validation", errors.New("invalid service name")
	}
	localArgs := []string{"-u", service, "-n", strconv.Itoa(lines), "--no-pager"}
	remoteCmd := fmt.Sprintf("journalctl -u %s -n %d --no-pager", service, lines)
	out, source, err := runOnTarget(ctx, svcCtx, target, "journalctl", localArgs, remoteCmd)
	if err != nil {
		return nil, source, err
	}
	return gin.H{"service": service, "logs": out, "lines": lines}, source, nil
}

func execOSContainerRuntime(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	target := toString(params["target"])
	out, source, err := runOnTarget(ctx, svcCtx, target, "docker", []string{"ps", "--format", "{{.ID}} {{.Image}} {{.Status}}"}, "docker ps --format '{{.ID}} {{.Image}} {{.Status}}'")
	if err == nil {
		return gin.H{"runtime": "docker", "containers": out}, source, nil
	}
	out2, source2, err2 := runOnTarget(ctx, svcCtx, target, "ctr", []string{"-n", "k8s.io", "containers", "list"}, "ctr -n k8s.io containers list")
	if err2 == nil {
		return gin.H{"runtime": "containerd", "containers": out2}, source2, nil
	}
	return nil, source2, fmt.Errorf("docker and containerd unavailable: %v / %v", err, err2)
}

func execK8sListResources(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	cli, source, err := resolveK8sClient(svcCtx, params)
	if err != nil {
		return nil, source, err
	}
	ns := strings.TrimSpace(toString(params["namespace"]))
	if ns == "" {
		ns = corev1.NamespaceAll
	}
	resource := strings.ToLower(strings.TrimSpace(toString(params["resource"])))
	limit := toInt(params["limit"])
	if limit <= 0 {
		limit = 50
	}
	switch resource {
	case "pods":
		list, err := cli.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, source, err
		}
		out := make([]gin.H, 0, len(list.Items))
		for i, p := range list.Items {
			if i >= limit {
				break
			}
			out = append(out, gin.H{"name": p.Name, "namespace": p.Namespace, "phase": p.Status.Phase, "node": p.Spec.NodeName})
		}
		return out, source, nil
	case "services":
		list, err := cli.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, source, err
		}
		out := make([]gin.H, 0, len(list.Items))
		for i, s := range list.Items {
			if i >= limit {
				break
			}
			out = append(out, gin.H{"name": s.Name, "namespace": s.Namespace, "type": s.Spec.Type, "clusterIP": s.Spec.ClusterIP})
		}
		return out, source, nil
	case "deployments":
		list, err := cli.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, source, err
		}
		out := make([]gin.H, 0, len(list.Items))
		for i, d := range list.Items {
			if i >= limit {
				break
			}
			out = append(out, gin.H{"name": d.Name, "namespace": d.Namespace, "ready": d.Status.ReadyReplicas, "replicas": d.Status.Replicas})
		}
		return out, source, nil
	case "nodes":
		list, err := cli.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, source, err
		}
		out := make([]gin.H, 0, len(list.Items))
		for i, n := range list.Items {
			if i >= limit {
				break
			}
			out = append(out, gin.H{"name": n.Name, "labels": n.Labels})
		}
		return out, source, nil
	default:
		return nil, source, errors.New("unsupported resource")
	}
}

func execK8sGetEvents(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	cli, source, err := resolveK8sClient(svcCtx, params)
	if err != nil {
		return nil, source, err
	}
	ns := strings.TrimSpace(toString(params["namespace"]))
	if ns == "" {
		ns = corev1.NamespaceAll
	}
	limit := toInt(params["limit"])
	if limit <= 0 {
		limit = 50
	}
	list, err := cli.CoreV1().Events(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, source, err
	}
	out := make([]gin.H, 0, len(list.Items))
	for i, e := range list.Items {
		if i >= limit {
			break
		}
		out = append(out, gin.H{"type": e.Type, "reason": e.Reason, "message": e.Message, "time": e.LastTimestamp})
	}
	return out, source, nil
}

func execK8sGetPodLogs(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	cli, source, err := resolveK8sClient(svcCtx, params)
	if err != nil {
		return nil, source, err
	}
	ns := strings.TrimSpace(toString(params["namespace"]))
	if ns == "" {
		ns = "default"
	}
	pod := strings.TrimSpace(toString(params["pod"]))
	if pod == "" {
		return nil, source, errors.New("pod is required")
	}
	tailLines := int64(toInt(params["tail_lines"]))
	if tailLines <= 0 {
		tailLines = 200
	}
	opt := &corev1.PodLogOptions{
		Container: strings.TrimSpace(toString(params["container"])),
		TailLines: &tailLines,
	}
	raw, err := cli.CoreV1().Pods(ns).GetLogs(pod, opt).DoRaw(ctx)
	if err != nil {
		return nil, source, err
	}
	return gin.H{"namespace": ns, "pod": pod, "logs": string(raw)}, source, nil
}

func execServiceGetDetail(_ context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	sid := toInt(params["service_id"])
	if sid <= 0 {
		return nil, "db", errors.New("service_id is required")
	}
	var s model.Service
	if err := svcCtx.DB.First(&s, sid).Error; err != nil {
		return nil, "db", err
	}
	return s, "db", nil
}

func execServiceDeployPreview(_ context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	sid := toInt(params["service_id"])
	cid := toInt(params["cluster_id"])
	if sid <= 0 || cid <= 0 {
		return nil, "preview", errors.New("service_id and cluster_id are required")
	}
	var s model.Service
	if err := svcCtx.DB.First(&s, sid).Error; err != nil {
		return nil, "db", err
	}
	return gin.H{
		"service_id": sid,
		"cluster_id": cid,
		"summary":    "deploy preview generated",
		"manifest": gin.H{
			"name":     s.Name,
			"image":    s.Image,
			"replicas": s.Replicas,
		},
	}, "preview", nil
}

func execServiceDeployApply(ctx context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	sid := toInt(params["service_id"])
	cid := toInt(params["cluster_id"])
	if sid <= 0 || cid <= 0 {
		return nil, "deploy", errors.New("service_id and cluster_id are required")
	}
	logic := projectlogic.NewServiceLogic(svcCtx)
	if err := logic.DeployService(ctx, v1.DeployServiceReq{
		ServiceID: uint(sid),
		ClusterID: uint(cid),
	}); err != nil {
		return nil, "deploy", err
	}
	return gin.H{"applied": true, "service_id": sid, "cluster_id": cid}, "deploy", nil
}

func execHostSSHReadonly(_ context.Context, svcCtx *svc.ServiceContext, params map[string]any) (any, string, error) {
	hostID := toInt(params["host_id"])
	cmd := strings.TrimSpace(toString(params["command"]))
	if hostID <= 0 {
		return nil, "host_ssh", errors.New("host_id is required")
	}
	if !isReadonlyHostCommand(cmd) {
		return nil, "host_ssh", errors.New("command not allowed")
	}
	var node model.Node
	if err := svcCtx.DB.First(&node, hostID).Error; err != nil {
		return nil, "db", err
	}
	cli, err := sshclient.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, "")
	if err != nil {
		return nil, "host_ssh", err
	}
	defer cli.Close()
	out, err := sshclient.RunCommand(cli, cmd)
	if err != nil {
		return gin.H{"stdout": out, "stderr": err.Error(), "exit_code": 1}, "host_ssh", nil
	}
	return gin.H{"stdout": out, "stderr": "", "exit_code": 0}, "host_ssh", nil
}

func maybeAddAIPermissions(svcCtx *svc.ServiceContext) {
	if svcCtx == nil || svcCtx.DB == nil {
		return
	}
	type permSeed struct {
		Name, Code, Resource, Action, Desc string
	}
	seeds := []permSeed{
		{Name: "AI Chat", Code: "ai:chat", Resource: "ai", Action: "chat", Desc: "Use AI chat"},
		{Name: "AI Tool Read", Code: "ai:tool:read", Resource: "ai", Action: "tool:read", Desc: "Execute readonly AI tools"},
		{Name: "AI Tool Execute", Code: "ai:tool:execute", Resource: "ai", Action: "tool:execute", Desc: "Execute mutating AI tools"},
		{Name: "AI Approval Review", Code: "ai:approval:review", Resource: "ai", Action: "approval:review", Desc: "Review approval tickets"},
		{Name: "AI Admin", Code: "ai:admin", Resource: "ai", Action: "admin", Desc: "AI admin permissions"},
	}
	now := time.Now().Unix()
	for _, s := range seeds {
		var existing model.Permission
		if err := svcCtx.DB.Where("code = ?", s.Code).First(&existing).Error; err == nil {
			continue
		}
		_ = svcCtx.DB.Create(&model.Permission{
			Name:        s.Name,
			Code:        s.Code,
			Resource:    s.Resource,
			Action:      s.Action,
			Description: s.Desc,
			Status:      1,
			CreateTime:  now,
			UpdateTime:  now,
		}).Error
	}
}

func cleanupSensitiveText(input string) string {
	if input == "" {
		return input
	}
	lines := strings.Split(input, "\n")
	for i, ln := range lines {
		low := strings.ToLower(ln)
		if strings.Contains(low, "password") || strings.Contains(low, "token") || strings.Contains(low, "secret") {
			lines[i] = "[REDACTED]"
		}
	}
	return strings.Join(lines, "\n")
}

func readFileTrim(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
