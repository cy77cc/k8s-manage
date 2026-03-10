package handler

import (
	"strconv"

	"github.com/cy77cc/OpsPilot/internal/config"
	"github.com/cy77cc/OpsPilot/internal/httpx"
	hostlogic "github.com/cy77cc/OpsPilot/internal/service/host/logic"
	"github.com/cy77cc/OpsPilot/internal/svc"
	"github.com/cy77cc/OpsPilot/internal/xcode"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svcCtx      *svc.ServiceContext
	hostService *hostlogic.HostService
}

func NewHandler(svcCtx *svc.ServiceContext) *Handler {
	return &Handler{
		svcCtx:      svcCtx,
		hostService: hostlogic.NewHostService(svcCtx),
	}
}

func (h *Handler) StartHealthCollector() {
	if !config.HostHealthDiagnosticsEnabled() {
		return
	}
	h.hostService.StartHealthSnapshotCollector()
}

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return 0, false
	}
	return id, true
}

func getUID(c *gin.Context) uint64 {
	uid, ok := c.Get("uid")
	if !ok {
		return 0
	}
	switch v := uid.(type) {
	case uint:
		return uint64(v)
	case uint64:
		return v
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	case float64:
		return uint64(v)
	default:
		return 0
	}
}
