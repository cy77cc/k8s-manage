package cluster

import (
	"time"

	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler handles cluster-related HTTP requests
type Handler struct {
	svcCtx *svc.ServiceContext
}

// NewHandler creates a new cluster handler
func NewHandler(svcCtx *svc.ServiceContext) *Handler {
	return &Handler{svcCtx: svcCtx}
}

// GetClusters returns list of clusters
func (h *Handler) GetClusters(c *gin.Context) {
	var clusters []model.Cluster
	query := h.svcCtx.DB.WithContext(c.Request.Context()).Model(&model.Cluster{})

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by source
	if source := c.Query("source"); source != "" {
		query = query.Where("source = ?", source)
	}

	if err := query.Order("id DESC").Find(&clusters).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	// Build response with node counts
	items := make([]ClusterListItem, 0, len(clusters))
	for _, cl := range clusters {
		var nodeCount int64
		h.svcCtx.DB.Model(&model.ClusterNode{}).Where("cluster_id = ?", cl.ID).Count(&nodeCount)

		items = append(items, ClusterListItem{
			ID:          cl.ID,
			Name:        cl.Name,
			Version:     cl.Version,
			K8sVersion:  cl.K8sVersion,
			Status:      cl.Status,
			Source:      cl.Source,
			NodeCount:   int(nodeCount),
			Endpoint:    cl.Endpoint,
			Description: cl.Description,
			LastSyncAt:  cl.LastSyncAt,
			CreatedAt:   cl.CreatedAt,
		})
	}

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

	var cluster model.Cluster
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).First(&cluster, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			httpx.NotFound(c, "cluster not found")
			return
		}
		httpx.ServerErr(c, err)
		return
	}

	// Get node count
	var nodeCount int64
	h.svcCtx.DB.Model(&model.ClusterNode{}).Where("cluster_id = ?", cluster.ID).Count(&nodeCount)

	detail := ClusterDetail{
		ID:             cluster.ID,
		Name:           cluster.Name,
		Description:    cluster.Description,
		Version:        cluster.Version,
		K8sVersion:     cluster.K8sVersion,
		Status:         cluster.Status,
		Source:         cluster.Source,
		Type:           cluster.Type,
		NodeCount:      int(nodeCount),
		Endpoint:       cluster.Endpoint,
		PodCIDR:        cluster.PodCIDR,
		ServiceCIDR:    cluster.ServiceCIDR,
		ManagementMode: cluster.ManagementMode,
		CredentialID:   cluster.CredentialID,
		LastSyncAt:     cluster.LastSyncAt,
		CreatedAt:      cluster.CreatedAt,
		UpdatedAt:      cluster.UpdatedAt,
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

	var nodes []model.ClusterNode
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).
		Where("cluster_id = ?", id).
		Order("role DESC, name ASC").
		Find(&nodes).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	items := make([]ClusterNode, 0, len(nodes))
	for _, n := range nodes {
		items = append(items, ClusterNode{
			ID:               n.ID,
			ClusterID:        n.ClusterID,
			HostID:           n.HostID,
			Name:             n.Name,
			IP:               n.IP,
			Role:             n.Role,
			Status:           n.Status,
			KubeletVersion:   n.KubeletVersion,
			ContainerRuntime: n.ContainerRuntime,
			OSImage:          n.OSImage,
			KernelVersion:    n.KernelVersion,
			AllocatableCPU:   n.AllocatableCPU,
			AllocatableMem:   n.AllocatableMem,
			Labels:           n.Labels,
			CreatedAt:        n.CreatedAt,
			UpdatedAt:        n.UpdatedAt,
		})
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

	var cluster model.Cluster
	if err := h.svcCtx.DB.First(&cluster, id).Error; err != nil {
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

	if err := h.svcCtx.DB.Model(&cluster).Updates(updates).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

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

	var cluster model.Cluster
	if err := h.svcCtx.DB.First(&cluster, id).Error; err != nil {
		httpx.NotFound(c, "cluster not found")
		return
	}

	// Delete in transaction
	err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// Delete cluster nodes
		if err := tx.Where("cluster_id = ?", cluster.ID).Delete(&model.ClusterNode{}).Error; err != nil {
			return err
		}
		// Delete credentials
		if err := tx.Where("cluster_id = ?", cluster.ID).Delete(&model.ClusterCredential{}).Error; err != nil {
			return err
		}
		// Delete cluster
		if err := tx.Delete(&cluster).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, gin.H{"id": cluster.ID, "message": "deleted"})
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
