package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/model"
	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
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

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid id"}})
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

func isAdminByUserID(db *svc.ServiceContext, uid uint64) bool {
	if uid == 0 {
		return false
	}
	var u model.User
	if err := db.DB.Select("id", "username").Where("id = ?", uid).First(&u).Error; err == nil {
		if strings.EqualFold(strings.TrimSpace(u.Username), "admin") {
			return true
		}
	}
	type roleRow struct {
		Code string `gorm:"column:code"`
	}
	var rows []roleRow
	err := db.DB.Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", uid).
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
