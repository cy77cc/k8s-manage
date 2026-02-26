package deployment

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) StartEnvironmentBootstrap(c *gin.Context) {
	if !h.authorize(c, "deploy:target:write") {
		return
	}
	var req EnvironmentBootstrapReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	if !h.authorizeRuntime(c, req.RuntimeType, "apply") {
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.StartEnvironmentBootstrap(c.Request.Context(), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) GetEnvironmentBootstrapJob(c *gin.Context) {
	if !h.authorize(c, "deploy:target:read") {
		return
	}
	job, err := h.logic.GetEnvironmentBootstrapJob(c.Request.Context(), strings.TrimSpace(c.Param("job_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": job})
}

func (h *Handler) RegisterPlatformCredential(c *gin.Context) {
	if !h.authorize(c, "deploy:credential:write", "deploy:target:write") {
		return
	}
	var req PlatformCredentialRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.RegisterPlatformCredential(c.Request.Context(), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) ImportExternalCredential(c *gin.Context) {
	if !h.authorize(c, "deploy:credential:write", "deploy:target:write") {
		return
	}
	var req ClusterCredentialImportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 2000, "msg": err.Error()})
		return
	}
	uid, _ := c.Get("uid")
	resp, err := h.logic.ImportExternalCredential(c.Request.Context(), toUint(uid), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) TestCredential(c *gin.Context) {
	if !h.authorize(c, "deploy:credential:read", "deploy:target:read") {
		return
	}
	resp, err := h.logic.TestCredentialConnectivity(c.Request.Context(), uintFromParam(c, "id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": resp})
}

func (h *Handler) ListCredentials(c *gin.Context) {
	if !h.authorize(c, "deploy:credential:read", "deploy:target:read") {
		return
	}
	list, err := h.logic.ListCredentials(c.Request.Context(), strings.TrimSpace(c.Query("runtime_type")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3000, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1000, "msg": "ok", "data": gin.H{"list": list, "total": len(list)}})
}
