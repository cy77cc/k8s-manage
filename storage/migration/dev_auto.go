package migration

import (
	"github.com/cy77cc/k8s-manage/internal/model"
	"gorm.io/gorm"
)

// RunDevAutoMigrate is only for local development convenience.
func RunDevAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.Node{},
		&model.NodeEvent{},
		&model.SSHKey{},
		&model.HostCloudAccount{},
		&model.HostImportTask{},
		&model.HostVirtualizationTask{},
		&model.AIChatSession{},
		&model.AIChatMessage{},
		&model.HostProbeSession{},
		&model.Project{},
		&model.Service{},
		&model.ServiceHelmRelease{},
		&model.ServiceRenderSnapshot{},
	)
}
