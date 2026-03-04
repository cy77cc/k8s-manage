package handler

import (
	"encoding/json"
	"math"
	"time"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/httpx"
	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

func (h *Handler) List(c *gin.Context) {
	list, err := h.hostService.List(c.Request.Context())
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, gin.H{"list": list, "total": len(list)})
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	node, err := h.hostService.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, xcode.NotFound, "host not found")
		return
	}
	httpx.OK(c, node)
}

func (h *Handler) Facts(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	node, err := h.hostService.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, xcode.NotFound, "host not found")
		return
	}
	httpx.OK(c, gin.H{"os": node.OS, "arch": node.Arch, "kernel": node.Kernel, "cpu_cores": node.CpuCores, "memory_mb": node.MemoryMB, "disk_gb": node.DiskGB, "source": "node"})
}

func (h *Handler) Tags(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	node, err := h.hostService.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, xcode.NotFound, "host not found")
		return
	}
	httpx.OK(c, hostlogic.ParseLabels(node.Labels))
}

func (h *Handler) Metrics(c *gin.Context) {
	if !config.HostHealthDiagnosticsEnabled() {
		httpx.Fail(c, xcode.Forbidden, "host health diagnostics is disabled")
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	snapshots, err := h.hostService.ListHealthSnapshots(c.Request.Context(), id, 50)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	rows := make([]gin.H, 0, len(snapshots))
	for _, s := range snapshots {
		cpuPct := math.Min(100, s.CpuLoad*20)
		memoryPct := 0.0
		if s.MemoryTotalMB > 0 {
			memoryPct = math.Min(100, float64(s.MemoryUsedMB)*100/float64(s.MemoryTotalMB))
		}
		extra := map[string]any{}
		if s.SummaryJSON != "" {
			_ = json.Unmarshal([]byte(s.SummaryJSON), &extra)
		}
		rows = append(rows, gin.H{
			"id":            s.ID,
			"cpu":           int(cpuPct),
			"memory":        int(memoryPct),
			"disk":          int(s.DiskUsedPct),
			"network":       0,
			"latency_ms":    s.LatencyMS,
			"health_state":  s.State,
			"error_message": s.ErrorMessage,
			"summary":       extra,
			"created_at":    s.CheckedAt,
		})
	}
	httpx.OK(c, rows)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	if !config.HostHealthDiagnosticsEnabled() {
		httpx.Fail(c, xcode.Forbidden, "host health diagnostics is disabled")
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	snapshot, err := h.hostService.RunHealthCheck(c.Request.Context(), id, getUID(c))
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, snapshot)
}

func (h *Handler) Audits(c *gin.Context) {
	rows := []gin.H{{"id": 1, "action": "query", "operator": "system", "detail": "host detail viewed", "created_at": time.Now()}}
	httpx.OK(c, rows)
}
