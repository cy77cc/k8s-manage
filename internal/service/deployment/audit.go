package deployment

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	svcCtx *svc.ServiceContext
}

func NewAuditHandler(svcCtx *svc.ServiceContext) *AuditHandler {
	return &AuditHandler{svcCtx: svcCtx}
}

type listAuditLogsReq struct {
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	ActionType   string `form:"action_type"`
	ResourceType string `form:"resource_type"`
}

// ListAuditLogs 获取审计日志列表
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	var req listAuditLogsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		h.respondBadRequest(c, err)
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	ctx := c.Request.Context()
	logs, total, err := h.listAuditLogs(ctx, req)
	if err != nil {
		h.respondInternalError(c, err)
		return
	}

	h.respondOK(c, gin.H{"list": logs, "total": total})
}

func (h *AuditHandler) listAuditLogs(ctx context.Context, req listAuditLogsReq) ([]model.AuditLog, int64, error) {
	query := h.svcCtx.DB.WithContext(ctx).Model(&model.AuditLog{})

	if req.ActionType != "" {
		query = query.Where("action_type = ?", req.ActionType)
	}
	if req.ResourceType != "" {
		query = query.Where("resource_type = ?", req.ResourceType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.AuditLog
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("id desc").Offset(offset).Limit(req.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// CreateAuditLog 创建审计日志（内部方法）
func CreateAuditLog(ctx context.Context, db *svc.ServiceContext, actionType, resourceType string, resourceID, actorID uint, actorName string, detail map[string]interface{}) error {
	log := model.AuditLog{
		ActionType:   actionType,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ActorID:      actorID,
		ActorName:    actorName,
		Detail:       detail,
	}
	return db.DB.WithContext(ctx).Create(&log).Error
}

// Response helpers
func (h *AuditHandler) respondOK(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{"code": 1000, "msg": "success", "data": data})
}

func (h *AuditHandler) respondBadRequest(c *gin.Context, err error) {
	c.JSON(400, gin.H{"code": xcode.ParamError, "msg": err.Error()})
}

func (h *AuditHandler) respondInternalError(c *gin.Context, err error) {
	c.JSON(500, gin.H{"code": xcode.ServerError, "msg": err.Error()})
}
