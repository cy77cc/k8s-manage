package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/cy77cc/OpsPilot/internal/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListClusters(ctx context.Context, status, source string) ([]ClusterListItem, error) {
	var clusters []model.Cluster
	q := r.db.WithContext(ctx).Model(&model.Cluster{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if source != "" {
		q = q.Where("source = ?", source)
	}
	if err := q.Order("id DESC").Find(&clusters).Error; err != nil {
		return nil, err
	}

	type nodeCountRow struct {
		ClusterID uint
		Count     int64
	}
	counts := []nodeCountRow{}
	if len(clusters) > 0 {
		ids := make([]uint, 0, len(clusters))
		for _, c := range clusters {
			ids = append(ids, c.ID)
		}
		if err := r.db.WithContext(ctx).Model(&model.ClusterNode{}).
			Select("cluster_id, COUNT(1) as count").
			Where("cluster_id IN ?", ids).
			Group("cluster_id").
			Find(&counts).Error; err != nil {
			return nil, err
		}
	}
	countMap := map[uint]int64{}
	for _, row := range counts {
		countMap[row.ClusterID] = row.Count
	}

	items := make([]ClusterListItem, 0, len(clusters))
	for _, cl := range clusters {
		items = append(items, ClusterListItem{
			ID:          cl.ID,
			Name:        cl.Name,
			Version:     cl.Version,
			K8sVersion:  cl.K8sVersion,
			Status:      cl.Status,
			Source:      cl.Source,
			NodeCount:   int(countMap[cl.ID]),
			Endpoint:    cl.Endpoint,
			Description: cl.Description,
			LastSyncAt:  cl.LastSyncAt,
			CreatedAt:   cl.CreatedAt,
		})
	}
	return items, nil
}

func (r *Repository) GetClusterModel(ctx context.Context, id uint) (*model.Cluster, error) {
	var row model.Cluster
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) GetClusterDetail(ctx context.Context, id uint) (ClusterDetail, error) {
	var cluster model.Cluster
	if err := r.db.WithContext(ctx).First(&cluster, id).Error; err != nil {
		return ClusterDetail{}, err
	}

	var nodeCount int64
	if err := r.db.WithContext(ctx).Model(&model.ClusterNode{}).Where("cluster_id = ?", cluster.ID).Count(&nodeCount).Error; err != nil {
		return ClusterDetail{}, err
	}

	return ClusterDetail{
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
	}, nil
}

func (r *Repository) ListClusterNodes(ctx context.Context, clusterID uint) ([]ClusterNode, error) {
	var nodes []model.ClusterNode
	if err := r.db.WithContext(ctx).
		Where("cluster_id = ?", clusterID).
		Order("role DESC, name ASC").
		Find(&nodes).Error; err != nil {
		return nil, err
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
	return items, nil
}

func (r *Repository) ListBootstrapProfiles(ctx context.Context) ([]BootstrapProfileItem, error) {
	var rows []model.ClusterBootstrapProfile
	if err := r.db.WithContext(ctx).Order("id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]BootstrapProfileItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toBootstrapProfileItem(row))
	}
	return items, nil
}

func (r *Repository) CreateCluster(ctx context.Context, in *model.Cluster) error {
	return r.db.WithContext(ctx).Create(in).Error
}

func (r *Repository) CreateClusterCredential(ctx context.Context, in *model.ClusterCredential) error {
	return r.db.WithContext(ctx).Create(in).Error
}

func (r *Repository) UpdateClusterCredentialID(ctx context.Context, clusterID, credentialID uint) error {
	return r.db.WithContext(ctx).Model(&model.Cluster{}).Where("id = ?", clusterID).Update("credential_id", credentialID).Error
}

func (r *Repository) UpdateCluster(ctx context.Context, id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&model.Cluster{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) DeleteClusterWithRelations(ctx context.Context, clusterID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cluster_id = ?", clusterID).Delete(&model.ClusterNode{}).Error; err != nil {
			return err
		}
		if err := tx.Where("cluster_id = ?", clusterID).Delete(&model.ClusterCredential{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ?", clusterID).Delete(&model.Cluster{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *Repository) FindClusterCredentialByClusterID(ctx context.Context, clusterID uint) (*model.ClusterCredential, error) {
	var cred model.ClusterCredential
	if err := r.db.WithContext(ctx).Where("cluster_id = ?", clusterID).First(&cred).Error; err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *Repository) UpsertClusterNode(ctx context.Context, clusterID uint, nodeName string, row model.ClusterNode, updates map[string]interface{}) error {
	var existing model.ClusterNode
	res := r.db.WithContext(ctx).Where("cluster_id = ? AND name = ?", clusterID, nodeName).First(&existing)
	if res.Error == nil {
		return r.db.WithContext(ctx).Model(&existing).Updates(updates).Error
	}
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		return res.Error
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *Repository) UpdateClusterLastSync(ctx context.Context, clusterID uint, ts *time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Cluster{}).Where("id = ?", clusterID).Update("last_sync_at", ts).Error
}

func (r *Repository) MustNotBeNil() error {
	if r == nil || r.db == nil {
		return fmt.Errorf("cluster repository is not initialized")
	}
	return nil
}
