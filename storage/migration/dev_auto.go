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
		&model.AICommandExecution{},
		&model.HostProbeSession{},
		&model.Project{},
		&model.Service{},
		&model.ServiceHelmRelease{},
		&model.ServiceRenderSnapshot{},
		&model.ServiceRevision{},
		&model.ServiceVariableSet{},
		&model.ServiceDeployTarget{},
		&model.ServiceReleaseRecord{},
		&model.DeploymentTarget{},
		&model.DeploymentTargetNode{},
		&model.DeploymentRelease{},
		&model.ServiceGovernancePolicy{},
		&model.AIOPSInspection{},
		&model.AlertEvent{},
		&model.AlertRule{},
		&model.MetricPoint{},
		&model.ClusterBootstrapTask{},
		&model.ClusterNamespaceBinding{},
		&model.ClusterReleaseRecord{},
		&model.ClusterHPAPolicy{},
		&model.ClusterQuotaPolicy{},
		&model.ClusterDeployApproval{},
		&model.ClusterOperationAudit{},
		&model.CMDBCI{},
		&model.CMDBRelation{},
		&model.CMDBSyncJob{},
		&model.CMDBSyncRecord{},
		&model.CMDBAudit{},
		&model.CICDServiceCIConfig{},
		&model.CICDServiceCIRun{},
		&model.CICDDeploymentCDConfig{},
		&model.CICDRelease{},
		&model.CICDReleaseApproval{},
		&model.CICDAuditEvent{},
	)
}
