package repo

import (
	"context"

	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetServiceCIConfig(ctx context.Context, serviceID uint) (*model.CICDServiceCIConfig, error) {
	var row model.CICDServiceCIConfig
	if err := r.db.WithContext(ctx).Where("service_id = ?", serviceID).Order("id DESC").First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) UpsertServiceCIConfig(ctx context.Context, in model.CICDServiceCIConfig) (*model.CICDServiceCIConfig, error) {
	var existing model.CICDServiceCIConfig
	err := r.db.WithContext(ctx).Where("service_id = ?", in.ServiceID).First(&existing).Error
	if err == nil {
		existing.RepoURL = in.RepoURL
		existing.Branch = in.Branch
		existing.BuildStepsJSON = in.BuildStepsJSON
		existing.ArtifactTarget = in.ArtifactTarget
		existing.TriggerMode = in.TriggerMode
		existing.Status = in.Status
		existing.UpdatedBy = in.UpdatedBy
		if uerr := r.db.WithContext(ctx).Save(&existing).Error; uerr != nil {
			return nil, uerr
		}
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if in.Status == "" {
		in.Status = "active"
	}
	if cerr := r.db.WithContext(ctx).Create(&in).Error; cerr != nil {
		return nil, cerr
	}
	return &in, nil
}

func (r *Repository) DeleteServiceCIConfig(ctx context.Context, serviceID uint) error {
	return r.db.WithContext(ctx).Where("service_id = ?", serviceID).Delete(&model.CICDServiceCIConfig{}).Error
}

func (r *Repository) CreateCIRun(ctx context.Context, in model.CICDServiceCIRun) (*model.CICDServiceCIRun, error) {
	if err := r.db.WithContext(ctx).Create(&in).Error; err != nil {
		return nil, err
	}
	return &in, nil
}

func (r *Repository) ListCIRuns(ctx context.Context, serviceID uint) ([]model.CICDServiceCIRun, error) {
	rows := make([]model.CICDServiceCIRun, 0)
	if err := r.db.WithContext(ctx).Where("service_id = ?", serviceID).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) GetDeploymentCDConfig(ctx context.Context, deploymentID uint, env string) (*model.CICDDeploymentCDConfig, error) {
	var row model.CICDDeploymentCDConfig
	q := r.db.WithContext(ctx).Where("deployment_id = ?", deploymentID)
	if env != "" {
		q = q.Where("env = ?", env)
	}
	if err := q.Order("id DESC").First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) UpsertDeploymentCDConfig(ctx context.Context, in model.CICDDeploymentCDConfig) (*model.CICDDeploymentCDConfig, error) {
	var existing model.CICDDeploymentCDConfig
	err := r.db.WithContext(ctx).Where("deployment_id = ? AND env = ?", in.DeploymentID, in.Env).First(&existing).Error
	if err == nil {
		existing.Strategy = in.Strategy
		existing.StrategyConfigJSON = in.StrategyConfigJSON
		existing.ApprovalRequired = in.ApprovalRequired
		existing.UpdatedBy = in.UpdatedBy
		if uerr := r.db.WithContext(ctx).Save(&existing).Error; uerr != nil {
			return nil, uerr
		}
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if cerr := r.db.WithContext(ctx).Create(&in).Error; cerr != nil {
		return nil, cerr
	}
	return &in, nil
}

func (r *Repository) CreateRelease(ctx context.Context, in model.CICDRelease) (*model.CICDRelease, error) {
	if err := r.db.WithContext(ctx).Create(&in).Error; err != nil {
		return nil, err
	}
	return &in, nil
}

func (r *Repository) GetRelease(ctx context.Context, id uint) (*model.CICDRelease, error) {
	var row model.CICDRelease
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) SaveRelease(ctx context.Context, row *model.CICDRelease) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *Repository) ListReleases(ctx context.Context, serviceID, deploymentID uint) ([]model.CICDRelease, error) {
	q := r.db.WithContext(ctx).Model(&model.CICDRelease{})
	if serviceID > 0 {
		q = q.Where("service_id = ?", serviceID)
	}
	if deploymentID > 0 {
		q = q.Where("deployment_id = ?", deploymentID)
	}
	rows := make([]model.CICDRelease, 0)
	if err := q.Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) CreateApproval(ctx context.Context, in model.CICDReleaseApproval) (*model.CICDReleaseApproval, error) {
	if err := r.db.WithContext(ctx).Create(&in).Error; err != nil {
		return nil, err
	}
	return &in, nil
}

func (r *Repository) ListApprovals(ctx context.Context, releaseID uint) ([]model.CICDReleaseApproval, error) {
	rows := make([]model.CICDReleaseApproval, 0)
	if err := r.db.WithContext(ctx).Where("release_id = ?", releaseID).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) CreateAuditEvent(ctx context.Context, in model.CICDAuditEvent) (*model.CICDAuditEvent, error) {
	if err := r.db.WithContext(ctx).Create(&in).Error; err != nil {
		return nil, err
	}
	return &in, nil
}

func (r *Repository) ListAuditEventsByService(ctx context.Context, serviceID uint, limit int) ([]model.CICDAuditEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows := make([]model.CICDAuditEvent, 0)
	if err := r.db.WithContext(ctx).Where("service_id = ?", serviceID).Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
