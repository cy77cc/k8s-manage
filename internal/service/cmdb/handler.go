package cmdb

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	cmdbv1 "github.com/cy77cc/k8s-manage/api/cmdb/v1"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	logic  *Logic
	svcCtx *svc.ServiceContext
}

func NewHandler(svcCtx *svc.ServiceContext) *Handler {
	return &Handler{logic: NewLogic(svcCtx), svcCtx: svcCtx}
}

func (h *Handler) ListAssets(c *gin.Context) {
	if !h.authorize(c, "cmdb:read") {
		return
	}
	filter := ciFilter{
		Type:      strings.TrimSpace(c.Query("asset_type")),
		Status:    strings.TrimSpace(c.Query("status")),
		Keyword:   strings.TrimSpace(c.Query("keyword")),
		ProjectID: h.projectIDFromHeader(c),
		TeamID:    h.teamIDFromHeader(c),
		Page:      atoiDefault(c.Query("page"), 1),
		PageSize:  atoiDefault(c.Query("page_size"), 20),
	}
	rows, total, err := h.logic.listCIs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rows, "total": total})
}

func (h *Handler) CreateAsset(c *gin.Context) {
	if !h.authorize(c, "cmdb:write") {
		return
	}
	var req cmdbv1.CreateCIReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid := h.uidFromContext(c)
	created, err := h.logic.createCI(c.Request.Context(), uid, model.CMDBCI{
		CIType:     req.CIType,
		Name:       req.Name,
		Source:     req.Source,
		ExternalID: req.ExternalID,
		ProjectID:  req.ProjectID,
		TeamID:     req.TeamID,
		Owner:      req.Owner,
		Status:     req.Status,
		TagsJSON:   req.TagsJSON,
		AttrsJSON:  req.AttrsJSON,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.auditCI(c, "ci.create", uid, nil, created)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": created})
}

func (h *Handler) GetAsset(c *gin.Context) {
	if !h.authorize(c, "cmdb:read") {
		return
	}
	id := uint(atoiDefault(c.Param("id"), 0))
	row, err := h.logic.getCI(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3004, "msg": "asset not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": row})
}

func (h *Handler) UpdateAsset(c *gin.Context) {
	if !h.authorize(c, "cmdb:write") {
		return
	}
	id := uint(atoiDefault(c.Param("id"), 0))
	before, err := h.logic.getCI(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3004, "msg": "asset not found"})
		return
	}
	var req cmdbv1.UpdateCIReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Owner != nil {
		updates["owner"] = strings.TrimSpace(*req.Owner)
	}
	if req.Status != nil {
		updates["status"] = strings.TrimSpace(*req.Status)
	}
	if req.TagsJSON != nil {
		updates["tags_json"] = *req.TagsJSON
	}
	if req.AttrsJSON != nil {
		updates["attrs_json"] = *req.AttrsJSON
	}
	after, err := h.logic.updateCI(c.Request.Context(), h.uidFromContext(c), id, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.auditCI(c, "ci.update", h.uidFromContext(c), before, after)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": after})
}

func (h *Handler) DeleteAsset(c *gin.Context) {
	if !h.authorize(c, "cmdb:write") {
		return
	}
	id := uint(atoiDefault(c.Param("id"), 0))
	before, err := h.logic.getCI(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3004, "msg": "asset not found"})
		return
	}
	if err := h.logic.deleteCI(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.auditCI(c, "ci.delete", h.uidFromContext(c), before, nil)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) ListRelations(c *gin.Context) {
	if !h.authorize(c, "cmdb:read") {
		return
	}
	ciID := uint(atoiDefault(c.Query("asset_id"), 0))
	rels, err := h.logic.listRelations(c.Request.Context(), ciID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(rels))
	for _, r := range rels {
		out = append(out, gin.H{"id": r.ID, "from_asset_id": r.FromCIID, "to_asset_id": r.ToCIID, "relation_type": r.RelationType})
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": out, "total": len(out)})
}

func (h *Handler) CreateRelation(c *gin.Context) {
	if !h.authorize(c, "cmdb:write") {
		return
	}
	var req cmdbv1.CreateRelationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	rel, err := h.logic.createRelation(c.Request.Context(), h.uidFromContext(c), model.CMDBRelation{
		FromCIID:     req.FromCIID,
		ToCIID:       req.ToCIID,
		RelationType: req.RelationType,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.auditRelation(c, "relation.create", h.uidFromContext(c), nil, rel)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": rel})
}

func (h *Handler) DeleteRelation(c *gin.Context) {
	if !h.authorize(c, "cmdb:write") {
		return
	}
	id := uint(atoiDefault(c.Param("id"), 0))
	rows, _ := h.logic.listRelations(c.Request.Context(), 0)
	var before *model.CMDBRelation
	for i := range rows {
		if rows[i].ID == id {
			before = &rows[i]
			break
		}
	}
	if err := h.logic.deleteRelation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.auditRelation(c, "relation.delete", h.uidFromContext(c), before, nil)
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": nil})
}

func (h *Handler) Topology(c *gin.Context) {
	if !h.authorize(c, "cmdb:read") {
		return
	}
	graph, err := h.logic.topology(c.Request.Context(), h.projectIDFromHeader(c), h.teamIDFromHeader(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": graph})
}

func (h *Handler) TriggerSync(c *gin.Context) {
	if !h.authorize(c, "cmdb:sync") {
		return
	}
	var req cmdbv1.TriggerSyncReq
	_ = c.ShouldBindJSON(&req)
	job, err := h.logic.runSync(c.Request.Context(), h.uidFromContext(c), req.Source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	h.logic.writeAudit(c.Request.Context(), model.CMDBAudit{Action: "sync.trigger", ActorID: h.uidFromContext(c), Detail: job.ID})
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": job})
}

func (h *Handler) GetSyncJob(c *gin.Context) {
	if !h.authorize(c, "cmdb:sync") {
		return
	}
	job, err := h.logic.getSyncJob(c.Request.Context(), strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3004, "msg": "sync job not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": job})
}

func (h *Handler) RetrySyncJob(c *gin.Context) {
	if !h.authorize(c, "cmdb:sync") {
		return
	}
	job, err := h.logic.runSync(c.Request.Context(), h.uidFromContext(c), "all")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": job})
}

func (h *Handler) ListChanges(c *gin.Context) {
	if !h.authorize(c, "cmdb:audit") && !h.authorize(c, "cmdb:read") {
		return
	}
	ciID := uint(atoiDefault(c.Query("asset_id"), 0))
	items, err := h.logic.listAudits(c.Request.Context(), ciID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": items, "total": len(items)})
}

func (h *Handler) ListAudits(c *gin.Context) { h.ListChanges(c) }

func (h *Handler) uidFromContext(c *gin.Context) uint {
	v, ok := c.Get("uid")
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case uint:
		return x
	case uint64:
		return uint(x)
	case int:
		if x < 0 {
			return 0
		}
		return uint(x)
	case int64:
		if x < 0 {
			return 0
		}
		return uint(x)
	default:
		return 0
	}
}

func (h *Handler) projectIDFromHeader(c *gin.Context) uint {
	v := strings.TrimSpace(c.GetHeader("X-Project-ID"))
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseUint(v, 10, 64)
	return uint(n)
}

func (h *Handler) teamIDFromHeader(c *gin.Context) uint {
	v := strings.TrimSpace(c.GetHeader("X-Team-ID"))
	if v == "" {
		return 0
	}
	n, _ := strconv.ParseUint(v, 10, 64)
	return uint(n)
}

func (h *Handler) authorize(c *gin.Context, code string) bool {
	uid := h.uidFromContext(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return false
	}
	var user model.User
	if err := h.svcCtx.DB.Select("id,username").Where("id = ?", uid).First(&user).Error; err == nil && strings.EqualFold(user.Username, "admin") {
		return true
	}
	var rows []struct {
		Code string `gorm:"column:code"`
	}
	err := h.svcCtx.DB.Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", uid).
		Scan(&rows).Error
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden"})
		return false
	}
	for _, r := range rows {
		if r.Code == code || r.Code == "cmdb:*" || r.Code == "*:*" {
			return true
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "forbidden"})
	return false
}

func (h *Handler) auditCI(c *gin.Context, action string, actor uint, before *model.CMDBCI, after *model.CMDBCI) {
	var ciID uint
	if after != nil {
		ciID = after.ID
	} else if before != nil {
		ciID = before.ID
	}
	beforeJSON := ""
	afterJSON := ""
	if before != nil {
		buf, _ := json.Marshal(before)
		beforeJSON = string(buf)
	}
	if after != nil {
		buf, _ := json.Marshal(after)
		afterJSON = string(buf)
	}
	h.logic.writeAudit(c.Request.Context(), model.CMDBAudit{CIID: ciID, Action: action, ActorID: actor, BeforeJSON: beforeJSON, AfterJSON: afterJSON, Detail: c.Request.URL.Path})
}

func (h *Handler) auditRelation(c *gin.Context, action string, actor uint, before *model.CMDBRelation, after *model.CMDBRelation) {
	var relID uint
	if after != nil {
		relID = after.ID
	} else if before != nil {
		relID = before.ID
	}
	beforeJSON := ""
	afterJSON := ""
	if before != nil {
		buf, _ := json.Marshal(before)
		beforeJSON = string(buf)
	}
	if after != nil {
		buf, _ := json.Marshal(after)
		afterJSON = string(buf)
	}
	h.logic.writeAudit(c.Request.Context(), model.CMDBAudit{RelationID: relID, Action: action, ActorID: actor, BeforeJSON: beforeJSON, AfterJSON: afterJSON, Detail: c.Request.URL.Path})
}

func atoiDefault(v string, d int) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return d
	}
	return n
}

var _ = time.Now
