package cluster

import (
	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/gin-gonic/gin"
)

// ClusterServiceInfo represents a service deployed to the cluster
type ClusterServiceInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	ProjectName  string `json:"project_name"`
	TeamName     string `json:"team_name"`
	Env          string `json:"env"`
	LastDeployAt string `json:"last_deploy_at"`
	Status       string `json:"status"`
}

// GetClusterServices returns services deployed to the cluster
func (h *Handler) GetClusterServices(c *gin.Context) {
	id := httpx.UintFromParam(c, "id")
	if id == 0 {
		httpx.BindErr(c, nil)
		return
	}

	// Query deployment targets linked to this cluster
	var targets []model.DeploymentTarget
	if err := h.svcCtx.DB.WithContext(c.Request.Context()).
		Where("cluster_id = ?", id).
		Find(&targets).Error; err != nil {
		httpx.ServerErr(c, err)
		return
	}

	// Query services that have releases to this cluster
	items := make([]ClusterServiceInfo, 0)
	for _, target := range targets {
		// Get latest release for this target
		var release model.DeploymentRelease
		if err := h.svcCtx.DB.WithContext(c.Request.Context()).
			Where("target_id = ?", target.ID).
			Order("id DESC").
			First(&release).Error; err == nil {

			// Get service info
			var service model.Service
			if err := h.svcCtx.DB.WithContext(c.Request.Context()).
				First(&service, release.ServiceID).Error; err == nil {

				// Get project info
				projectName := ""
				if target.ProjectID > 0 {
					var project model.Project
					if err := h.svcCtx.DB.WithContext(c.Request.Context()).
						First(&project, target.ProjectID).Error; err == nil {
						projectName = project.Name
					}
				}

				items = append(items, ClusterServiceInfo{
					ID:           service.ID,
					Name:         service.Name,
					ProjectName:  projectName,
					TeamName:     "",
					Env:          target.Env,
					LastDeployAt: release.CreatedAt.Format("2006-01-02 15:04:05"),
					Status:       release.Status,
				})
			}
		}
	}

	httpx.OK(c, gin.H{"list": items, "total": len(items)})
}
