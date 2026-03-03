package cluster

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler handles cluster-related HTTP requests
type Handler struct {
	svcCtx *svc.ServiceContext
	repo   *Repository
}

// NewHandler creates a new cluster handler
func NewHandler(svcCtx *svc.ServiceContext) *Handler {
	return &Handler{
		svcCtx: svcCtx,
		repo:   NewRepository(svcCtx.DB),
	}
}

// GetClusters returns list of clusters
func (h *Handler) GetClusters(c *gin.Context) {
	status := c.Query("status")
	source := c.Query("source")
	cacheKey := CacheKeyClusterList(status, source)
	items := make([]ClusterListItem, 0)
	raw, _, err := h.svcCtx.CacheFacade.GetOrLoad(c.Request.Context(), cacheKey, ClusterPhase1CachePolicies["clusters.list"].TTL, func(ctx context.Context) (string, error) {
		rows, qerr := h.repo.ListClusters(ctx, status, source)
		if qerr != nil {
			return "", qerr
		}
		raw, merr := json.Marshal(rows)
		if merr != nil {
			return "", merr
		}
		return string(raw), nil
	})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}
	_ = json.Unmarshal([]byte(raw), &items)

	httpx.OK(c, gin.H{
		"list":  items,
		"total": len(items),
	})
}

// GetClusterDetail returns detailed cluster information
func (h *Handler) GetClusterDetail(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	cacheKey := CacheKeyClusterDetail(id)
	var detail ClusterDetail
	raw, _, err := h.svcCtx.CacheFacade.GetOrLoad(c.Request.Context(), cacheKey, ClusterPhase1CachePolicies["clusters.detail"].TTL, func(ctx context.Context) (string, error) {
		d, derr := h.repo.GetClusterDetail(ctx, id)
		if derr != nil {
			return "", derr
		}
		buf, merr := json.Marshal(d)
		if merr != nil {
			return "", merr
		}
		return string(buf), nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.NotFound(c, "cluster not found")
			return
		}
		httpx.ServerErr(c, err)
		return
	}
	if err := json.Unmarshal([]byte(raw), &detail); err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, detail)
}

// GetClusterNodes returns nodes in a cluster
func (h *Handler) GetClusterNodes(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	cacheKey := CacheKeyClusterNodes(id)
	items := make([]ClusterNode, 0)
	raw, _, err := h.svcCtx.CacheFacade.GetOrLoad(c.Request.Context(), cacheKey, ClusterPhase1CachePolicies["clusters.nodes"].TTL, func(ctx context.Context) (string, error) {
		rows, rerr := h.repo.ListClusterNodes(ctx, id)
		if rerr != nil {
			return "", rerr
		}
		buf, merr := json.Marshal(rows)
		if merr != nil {
			return "", merr
		}
		return string(buf), nil
	})
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"list":  items,
		"total": len(items),
	})
}

// CreateCluster creates a new cluster (import external)
func (h *Handler) CreateCluster(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "cluster:write") {
		return
	}

	var req ClusterCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}

	uid := httpx.UIDFromCtx(c)
	cluster, err := h.ImportCluster(c.Request.Context(), uid, req)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, cluster)
}

// UpdateCluster updates a cluster
func (h *Handler) UpdateCluster(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "cluster:write") {
		return
	}

	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	var req ClusterUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BindErr(c, err)
		return
	}

	cluster, err := h.repo.GetClusterModel(c.Request.Context(), id)
	if err != nil {
		httpx.NotFound(c, "cluster not found")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["updated_at"] = time.Now()

	if err := h.repo.UpdateCluster(c.Request.Context(), id, updates); err != nil {
		httpx.ServerErr(c, err)
		return
	}
	h.invalidateClusterCache(c.Request.Context(), id)

	httpx.OK(c, gin.H{"id": cluster.ID, "message": "updated"})
}

// DeleteCluster deletes a cluster
func (h *Handler) DeleteCluster(c *gin.Context) {
	if !httpx.Authorize(c, h.svcCtx.DB, "cluster:write") {
		return
	}

	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	cluster, err := h.repo.GetClusterModel(c.Request.Context(), id)
	if err != nil {
		httpx.NotFound(c, "cluster not found")
		return
	}

	if err := h.repo.DeleteClusterWithRelations(c.Request.Context(), cluster.ID); err != nil {
		httpx.ServerErr(c, err)
		return
	}
	h.invalidateClusterCache(c.Request.Context(), cluster.ID)

	httpx.OK(c, gin.H{"id": cluster.ID, "message": "deleted"})
}

func (h *Handler) invalidateClusterCache(ctx context.Context, clusterID uint) {
	if h.svcCtx == nil || h.svcCtx.CacheFacade == nil {
		return
	}
	h.svcCtx.CacheFacade.Delete(ctx,
		CacheKeyClusterList("", ""),
		CacheKeyClusterList("active", ""),
		CacheKeyClusterDetail(clusterID),
		CacheKeyClusterNodes(clusterID),
	)
}

// TestCluster tests cluster connectivity
func (h *Handler) TestCluster(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	result, err := h.TestConnectivity(c.Request.Context(), id)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, result)
}
