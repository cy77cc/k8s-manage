package catalog

import (
	"strconv"

	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	logic  *Logic
	svcCtx *svc.ServiceContext
}

func NewHandler(svcCtx *svc.ServiceContext) *Handler {
	return &Handler{logic: NewLogic(svcCtx), svcCtx: svcCtx}
}

func (h *Handler) ListCategories(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:read") {
		return
	}
	rows, err := h.logic.ListCategories(c.Request.Context())
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, gin.H{"list": rows, "total": len(rows)})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:manage") {
		return
	}
	var req CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	resp, err := h.logic.CreateCategory(c.Request.Context(), httpx.UIDFromCtx(c), req)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:manage") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}
	var req CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	resp, err := h.logic.UpdateCategory(c.Request.Context(), uint(id), req)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:manage") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}
	if err := h.logic.DeleteCategory(c.Request.Context(), uint(id)); err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, nil)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:read") {
		return
	}
	filters := map[string]string{
		"category_id": c.Query("category_id"),
		"status":      c.Query("status"),
		"visibility":  c.Query("visibility"),
		"q":           c.Query("q"),
	}
	if c.Query("mine") == "true" {
		filters["owner_id"] = strconv.FormatUint(httpx.UIDFromCtx(c), 10)
	}
	resp, err := h.logic.ListTemplates(c.Request.Context(), filters)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, gin.H{"list": resp.List, "total": resp.Total})
}

func (h *Handler) GetTemplate(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:read") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}
	resp, err := h.logic.GetTemplate(c.Request.Context(), uint(id))
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:write") {
		return
	}
	var req TemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	resp, err := h.logic.CreateTemplate(c.Request.Context(), httpx.UIDFromCtx(c), req)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:write") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}
	var req TemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	resp, err := h.logic.UpdateTemplate(c.Request.Context(), uint(id), httpx.UIDFromCtx(c), httpx.IsAdmin(h.svcCtx.DB, httpx.UIDFromCtx(c)), req)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}

func (h *Handler) DeleteTemplate(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:write") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}
	if err := h.logic.DeleteTemplate(c.Request.Context(), uint(id), httpx.UIDFromCtx(c), httpx.IsAdmin(h.svcCtx.DB, httpx.UIDFromCtx(c))); err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, nil)
}

func (h *Handler) SubmitTemplate(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:write") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}
	resp, err := h.logic.SubmitForReview(c.Request.Context(), uint(id), httpx.UIDFromCtx(c), httpx.IsAdmin(h.svcCtx.DB, httpx.UIDFromCtx(c)))
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}

func (h *Handler) PublishTemplate(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:approve") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}
	resp, err := h.logic.PublishTemplate(c.Request.Context(), uint(id))
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}

func (h *Handler) RejectTemplate(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:approve") {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	}
	var req ReviewActionRequest
	_ = c.ShouldBindJSON(&req)
	resp, err := h.logic.RejectTemplate(c.Request.Context(), uint(id), req.Reason)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}

func (h *Handler) Preview(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:read") {
		return
	}
	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	resp, err := h.logic.PreviewYAML(c.Request.Context(), req)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}

func (h *Handler) Deploy(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "catalog:write") {
		return
	}
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}
	resp, err := h.logic.DeployFromTemplate(c.Request.Context(), httpx.UIDFromCtx(c), req)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}
