package deployment

import (
	"context"
	"strconv"
	"time"

	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type PolicyHandler struct {
	svcCtx *svc.ServiceContext
}

func NewPolicyHandler(svcCtx *svc.ServiceContext) *PolicyHandler {
	return &PolicyHandler{svcCtx: svcCtx}
}

type listPoliciesReq struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Type     string `form:"type"`
	TargetID uint   `form:"target_id"`
}

type createPolicyReq struct {
	Name     string                 `json:"name" binding:"required"`
	Type     string                 `json:"type" binding:"required,oneof=traffic resilience access slo"`
	TargetID uint                   `json:"target_id"`
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
}

type updatePolicyReq struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Config  map[string]interface{} `json:"config"`
	Enabled *bool                  `json:"enabled"`
}

// ListPolicies 获取策略列表
func (h *PolicyHandler) ListPolicies(c *gin.Context) {
	var req listPoliciesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	ctx := c.Request.Context()
	policies, total, err := h.listPolicies(ctx, req)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, gin.H{"list": policies, "total": total})
}

// GetPolicy 获取策略详情
func (h *PolicyHandler) GetPolicy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}

	ctx := c.Request.Context()
	policy, err := h.getPolicy(ctx, uint(id))
	if err != nil {
		httpx.Fail(c, xcode.NotFound, "policy not found")
		return
	}

	httpx.OK(c, policy)
}

// CreatePolicy 创建策略
func (h *PolicyHandler) CreatePolicy(c *gin.Context) {
	var req createPolicyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}

	ctx := c.Request.Context()
	policy, err := h.createPolicy(ctx, req)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, policy)
}

// UpdatePolicy 更新策略
func (h *PolicyHandler) UpdatePolicy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}

	var req updatePolicyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}

	ctx := c.Request.Context()
	policy, err := h.updatePolicy(ctx, uint(id), req)
	if err != nil {
		httpx.Fail(c, xcode.NotFound, "policy not found")
		return
	}

	httpx.OK(c, policy)
}

// DeletePolicy 删除策略
func (h *PolicyHandler) DeletePolicy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}

	ctx := c.Request.Context()
	if err := h.deletePolicy(ctx, uint(id)); err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, gin.H{"message": "deleted"})
}

func (h *PolicyHandler) listPolicies(ctx context.Context, req listPoliciesReq) ([]model.Policy, int64, error) {
	query := h.svcCtx.DB.WithContext(ctx).Model(&model.Policy{})

	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.TargetID > 0 {
		query = query.Where("target_id = ?", req.TargetID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var policies []model.Policy
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("id desc").Offset(offset).Limit(req.PageSize).Find(&policies).Error; err != nil {
		return nil, 0, err
	}

	return policies, total, nil
}

func (h *PolicyHandler) getPolicy(ctx context.Context, id uint) (*model.Policy, error) {
	var policy model.Policy
	if err := h.svcCtx.DB.WithContext(ctx).First(&policy, id).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}

func (h *PolicyHandler) createPolicy(ctx context.Context, req createPolicyReq) (*model.Policy, error) {
	policy := model.Policy{
		Name:     req.Name,
		Type:     req.Type,
		TargetID: req.TargetID,
		Config:   req.Config,
		Enabled:  req.Enabled,
	}

	if err := h.svcCtx.DB.WithContext(ctx).Create(&policy).Error; err != nil {
		return nil, err
	}

	return &policy, nil
}

func (h *PolicyHandler) updatePolicy(ctx context.Context, id uint, req updatePolicyReq) (*model.Policy, error) {
	var policy model.Policy
	if err := h.svcCtx.DB.WithContext(ctx).First(&policy, id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Config != nil {
		updates["config"] = req.Config
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := h.svcCtx.DB.WithContext(ctx).Model(&policy).Updates(updates).Error; err != nil {
		return nil, err
	}

	return h.getPolicy(ctx, id)
}

func (h *PolicyHandler) deletePolicy(ctx context.Context, id uint) error {
	return h.svcCtx.DB.WithContext(ctx).Delete(&model.Policy{}, id).Error
}
