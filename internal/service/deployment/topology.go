package deployment

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/httpx"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/gin-gonic/gin"
)

type TopologyHandler struct {
	svcCtx *svc.ServiceContext
}

func NewTopologyHandler(svcCtx *svc.ServiceContext) *TopologyHandler {
	return &TopologyHandler{svcCtx: svcCtx}
}

// TopologyService 拓扑服务节点
type TopologyService struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Environment    string `json:"environment"`
	Status         string `json:"status"`
	LastDeployment string `json:"last_deployment,omitempty"`
	TargetID       uint   `json:"target_id"`
	TargetName     string `json:"target_name,omitempty"`
	RuntimeType    string `json:"runtime_type,omitempty"`
}

// TopologyConnection 拓扑连接
type TopologyConnection struct {
	SourceID uint   `json:"source_id"`
	TargetID uint   `json:"target_id"`
	Type     string `json:"type"`
}

// DeploymentTopology 部署拓扑
type DeploymentTopology struct {
	Services    []TopologyService    `json:"services"`
	Connections []TopologyConnection `json:"connections"`
}

// GetTopology 获取部署拓扑
func (h *TopologyHandler) GetTopology(c *gin.Context) {
	ctx := c.Request.Context()
	env := c.Query("environment")

	topology, err := h.getTopology(ctx, env)
	if err != nil {
		httpx.ServerErr(c, err)
		return
	}

	httpx.OK(c, topology)
}

func (h *TopologyHandler) getTopology(ctx context.Context, envFilter string) (*DeploymentTopology, error) {
	topology := &DeploymentTopology{
		Services:    []TopologyService{},
		Connections: []TopologyConnection{},
	}

	// 查询部署目标
	query := h.svcCtx.DB.WithContext(ctx).Model(&model.DeploymentTarget{})
	if envFilter != "" && envFilter != "all" {
		query = query.Where("env = ?", envFilter)
	}

	var targets []model.DeploymentTarget
	if err := query.Find(&targets).Error; err != nil {
		return nil, err
	}

	// 获取每个目标的最新发布
	for _, target := range targets {
		var latestRelease model.DeploymentRelease
		err := h.svcCtx.DB.WithContext(ctx).
			Model(&model.DeploymentRelease{}).
			Where("target_id = ?", target.ID).
			Order("created_at desc").
			First(&latestRelease).Error

		status := "unknown"
		lastDeployment := ""

		if err == nil {
			status = latestRelease.Status
			lastDeployment = latestRelease.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		} else {
			status = "no_deployments"
		}

		// 根据目标状态判断健康状态
		if target.ReadinessStatus != "" {
			status = target.ReadinessStatus
		}

		service := TopologyService{
			ID:             target.ID,
			Name:           target.Name,
			Environment:    target.Env,
			Status:         status,
			LastDeployment: lastDeployment,
			TargetID:       target.ID,
			TargetName:     target.Name,
			RuntimeType:    target.RuntimeType,
		}

		topology.Services = append(topology.Services, service)
	}

	return topology, nil
}
