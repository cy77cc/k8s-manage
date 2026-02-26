package deployment

import "time"

type TargetNodeReq struct {
	HostID uint   `json:"host_id"`
	Role   string `json:"role"`
	Weight int    `json:"weight"`
}

type TargetUpsertReq struct {
	Name        string          `json:"name" binding:"required"`
	TargetType  string          `json:"target_type" binding:"required"` // k8s|compose
	RuntimeType string          `json:"runtime_type"`
	ClusterID   uint            `json:"cluster_id"`
	ProjectID   uint            `json:"project_id"`
	TeamID      uint            `json:"team_id"`
	Env         string          `json:"env"`
	Nodes       []TargetNodeReq `json:"nodes"`
}

type TargetNodeResp struct {
	HostID uint   `json:"host_id"`
	Name   string `json:"name"`
	IP     string `json:"ip"`
	Status string `json:"status"`
	Role   string `json:"role"`
	Weight int    `json:"weight"`
}

type TargetResp struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	TargetType  string           `json:"target_type"`
	RuntimeType string           `json:"runtime_type"`
	ClusterID   uint             `json:"cluster_id"`
	ProjectID   uint             `json:"project_id"`
	TeamID      uint             `json:"team_id"`
	Env         string           `json:"env"`
	Status      string           `json:"status"`
	Nodes       []TargetNodeResp `json:"nodes,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type ReleasePreviewReq struct {
	ServiceID     uint              `json:"service_id" binding:"required"`
	TargetID      uint              `json:"target_id" binding:"required"`
	Env           string            `json:"env"`
	Strategy      string            `json:"strategy"`
	Variables     map[string]string `json:"variables"`
	ApprovalToken string            `json:"approval_token"` // backward compatibility
	PreviewToken  string            `json:"preview_token"`
}

type ReleasePreviewResp struct {
	ResolvedManifest string              `json:"resolved_manifest"`
	Checks           []map[string]string `json:"checks"`
	Warnings         []map[string]string `json:"warnings"`
	Runtime          string              `json:"runtime"`
	PreviewToken     string              `json:"preview_token,omitempty"`
	PreviewExpiresAt *time.Time          `json:"preview_expires_at,omitempty"`
}

type ReleaseApplyResp struct {
	ReleaseID        uint   `json:"release_id"`
	Status           string `json:"status"`
	RuntimeType      string `json:"runtime_type"`
	ApprovalRequired bool   `json:"approval_required,omitempty"`
	ApprovalTicket   string `json:"approval_ticket,omitempty"`
	LifecycleState   string `json:"lifecycle_state,omitempty"`
	ReasonCode       string `json:"reason_code,omitempty"`
}

type ReleaseSummaryResp struct {
	ID                 uint       `json:"id"`
	ServiceID          uint       `json:"service_id"`
	TargetID           uint       `json:"target_id"`
	NamespaceOrProject string     `json:"namespace_or_project"`
	RuntimeType        string     `json:"runtime_type"`
	Strategy           string     `json:"strategy"`
	RevisionID         uint       `json:"revision_id"`
	SourceReleaseID    uint       `json:"source_release_id"`
	TargetRevision     string     `json:"target_revision"`
	Status             string     `json:"status"`
	LifecycleState     string     `json:"lifecycle_state"`
	DiagnosticsJSON    string     `json:"diagnostics_json"`
	VerificationJSON   string     `json:"verification_json"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	PreviewExpiresAt   *time.Time `json:"preview_expires_at,omitempty"`
}

type ReleaseDecisionReq struct {
	Comment string `json:"comment"`
}

type ReleaseTimelineEventResp struct {
	ID            uint      `json:"id"`
	ReleaseID     uint      `json:"release_id"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	TraceID       string    `json:"trace_id,omitempty"`
	Action        string    `json:"action"`
	Actor         uint      `json:"actor"`
	Detail        any       `json:"detail"`
	CreatedAt     time.Time `json:"created_at"`
}

type GovernanceReq struct {
	Env              string         `json:"env"`
	TrafficPolicy    map[string]any `json:"traffic_policy"`
	ResiliencePolicy map[string]any `json:"resilience_policy"`
	AccessPolicy     map[string]any `json:"access_policy"`
	SLOPolicy        map[string]any `json:"slo_policy"`
}

type ClusterBootstrapPreviewReq struct {
	Name           string `json:"name" binding:"required"`
	ControlPlaneID uint   `json:"control_plane_host_id" binding:"required"`
	WorkerIDs      []uint `json:"worker_host_ids"`
	CNI            string `json:"cni"`
}

type ClusterBootstrapPreviewResp struct {
	Name             string   `json:"name"`
	ControlPlaneID   uint     `json:"control_plane_host_id"`
	WorkerHostIDs    []uint   `json:"worker_host_ids"`
	CNI              string   `json:"cni"`
	Steps            []string `json:"steps"`
	ExpectedEndpoint string   `json:"expected_endpoint"`
}

type ClusterBootstrapApplyResp struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	ClusterID uint   `json:"cluster_id,omitempty"`
	TargetID  uint   `json:"target_id,omitempty"`
}
