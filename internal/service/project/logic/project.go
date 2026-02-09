package logic

import (
	"context"

	v1 "github.com/cy77cc/k8s-manage/api/project/v1"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type ProjectLogic struct {
	svcCtx *svc.ServiceContext
}

func NewProjectLogic(svcCtx *svc.ServiceContext) *ProjectLogic {
	return &ProjectLogic{
		svcCtx: svcCtx,
	}
}

func (l *ProjectLogic) CreateProject(ctx context.Context, req v1.CreateProjectReq) (v1.ProjectResp, error) {
	project := &model.Project{
		Name:        req.Name,
		Description: req.Description,
		// OwnerID:     ctx.Value("uid").(int64), // TODO: Get from context
	}

	if err := l.svcCtx.DB.Create(project).Error; err != nil {
		return v1.ProjectResp{}, err
	}

	return v1.ProjectResp{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
	}, nil
}

func (l *ProjectLogic) ListProjects(ctx context.Context) ([]v1.ProjectResp, error) {
	var projects []model.Project
	if err := l.svcCtx.DB.Find(&projects).Error; err != nil {
		return nil, err
	}

	var res []v1.ProjectResp
	for _, p := range projects {
		res = append(res, v1.ProjectResp{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			OwnerID:     p.OwnerID,
			CreatedAt:   p.CreatedAt,
		})
	}
	return res, nil
}

func (l *ProjectLogic) DeployProject(ctx context.Context, req v1.DeployProjectReq) error {
	// 1. Get Project with Services
	var project model.Project
	if err := l.svcCtx.DB.Preload("Services").First(&project, req.ProjectID).Error; err != nil {
		return err
	}

	// 2. Get Cluster
	var cluster model.Cluster
	if err := l.svcCtx.DB.First(&cluster, req.ClusterID).Error; err != nil {
		return err
	}

	// 3. Deploy each service
	for _, svc := range project.Services {
		if err := DeployToCluster(ctx, &cluster, svc.YamlContent); err != nil {
			return err
		}
	}

	return nil
}
