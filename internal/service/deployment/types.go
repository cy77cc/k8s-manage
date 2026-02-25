package deployment

import "time"

type TargetNodeReq struct {
	HostID uint   `json:"host_id"`
	Role   string `json:"role"`
	Weight int    `json:"weight"`
}

type TargetUpsertReq struct {
	Name       string          `json:"name" binding:"required"`
	TargetType string          `json:"target_type" binding:"required"` // k8s|compose
	ClusterID  uint            `json:"cluster_id"`
	ProjectID  uint            `json:"project_id"`
	TeamID     uint            `json:"team_id"`
	Env        string          `json:"env"`
	Nodes      []TargetNodeReq `json:"nodes"`
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
	ID         uint             `json:"id"`
	Name       string           `json:"name"`
	TargetType string           `json:"target_type"`
	ClusterID  uint             `json:"cluster_id"`
	ProjectID  uint             `json:"project_id"`
	TeamID     uint             `json:"team_id"`
	Env        string           `json:"env"`
	Status     string           `json:"status"`
	Nodes      []TargetNodeResp `json:"nodes,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type ReleasePreviewReq struct {
	ServiceID     uint              `json:"service_id" binding:"required"`
	TargetID      uint              `json:"target_id" binding:"required"`
	Env           string            `json:"env"`
	Strategy      string            `json:"strategy"`
	Variables     map[string]string `json:"variables"`
	ApprovalToken string            `json:"approval_token"`
}

type ReleasePreviewResp struct {
	ResolvedManifest string              `json:"resolved_manifest"`
	Checks           []map[string]string `json:"checks"`
	Warnings         []map[string]string `json:"warnings"`
	Runtime          string              `json:"runtime"`
}

type ReleaseApplyResp struct {
	ReleaseID uint   `json:"release_id"`
	Status    string `json:"status"`
}

type GovernanceReq struct {
	Env              string         `json:"env"`
	TrafficPolicy    map[string]any `json:"traffic_policy"`
	ResiliencePolicy map[string]any `json:"resilience_policy"`
	AccessPolicy     map[string]any `json:"access_policy"`
	SLOPolicy        map[string]any `json:"slo_policy"`
}

