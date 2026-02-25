package service

import "time"

type LabelKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type EnvKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PortConfig struct {
	Name          string `json:"name"`
	Protocol      string `json:"protocol"`
	ContainerPort int32  `json:"container_port"`
	ServicePort   int32  `json:"service_port"`
}

type HealthCheckConfig struct {
	Type             string `json:"type"` // http/tcp/cmd
	Path             string `json:"path"`
	Port             int32  `json:"port"`
	Command          string `json:"command"`
	InitialDelaySec  int32  `json:"initial_delay_sec"`
	PeriodSec        int32  `json:"period_sec"`
	FailureThreshold int32  `json:"failure_threshold"`
}

type VolumeConfig struct {
	Name      string `json:"name"`
	MountPath string `json:"mount_path"`
	HostPath  string `json:"host_path"`
}

type StandardServiceConfig struct {
	Image       string             `json:"image"`
	Replicas    int32              `json:"replicas"`
	Ports       []PortConfig       `json:"ports"`
	Envs        []EnvKV            `json:"envs"`
	Resources   map[string]string  `json:"resources"` // cpu,memory
	Volumes     []VolumeConfig     `json:"volumes"`
	HealthCheck *HealthCheckConfig `json:"health_check"`
}

type RenderPreviewReq struct {
	Mode           string                 `json:"mode"`   // standard/custom
	Target         string                 `json:"target"` // k8s/compose
	StandardConfig *StandardServiceConfig `json:"standard_config"`
	CustomYAML     string                 `json:"custom_yaml"`
	ServiceName    string                 `json:"service_name"`
	ServiceType    string                 `json:"service_type"` // stateless/stateful
}

type RenderDiagnostic struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RenderPreviewResp struct {
	RenderedYAML     string             `json:"rendered_yaml"`
	Diagnostics      []RenderDiagnostic `json:"diagnostics"`
	NormalizedConfig any                `json:"normalized_config,omitempty"`
}

type TransformReq struct {
	StandardConfig *StandardServiceConfig `json:"standard_config"`
	Target         string                 `json:"target"`
	ServiceName    string                 `json:"service_name"`
	ServiceType    string                 `json:"service_type"`
}

type TransformResp struct {
	CustomYAML string `json:"custom_yaml"`
	SourceHash string `json:"source_hash"`
}

type ServiceCreateReq struct {
	ProjectID       uint                   `json:"project_id"`
	TeamID          uint                   `json:"team_id"`
	Name            string                 `json:"name"`
	Env             string                 `json:"env"`
	Owner           string                 `json:"owner"`
	ServiceKind     string                 `json:"service_kind"`
	ServiceType     string                 `json:"service_type"`
	RuntimeType     string                 `json:"runtime_type"`
	ConfigMode      string                 `json:"config_mode"`
	RenderTarget    string                 `json:"render_target"`
	Labels          []LabelKV              `json:"labels"`
	StandardConfig  *StandardServiceConfig `json:"standard_config"`
	CustomYAML      string                 `json:"custom_yaml"`
	SourceTemplateV string                 `json:"source_template_version"`
	Status          string                 `json:"status"`

	// legacy compat
	Image         string            `json:"image"`
	Replicas      int32             `json:"replicas"`
	ServicePort   int32             `json:"service_port"`
	ContainerPort int32             `json:"container_port"`
	NodePort      int32             `json:"node_port"`
	EnvVars       []EnvKV           `json:"env_vars"`
	Resources     map[string]string `json:"resources"`
	YamlContent   string            `json:"yaml_content"`
}

type ServiceListItem struct {
	ID             uint                   `json:"id"`
	ProjectID      uint                   `json:"project_id"`
	TeamID         uint                   `json:"team_id"`
	Name           string                 `json:"name"`
	Env            string                 `json:"env"`
	Owner          string                 `json:"owner"`
	RuntimeType    string                 `json:"runtime_type"`
	ConfigMode     string                 `json:"config_mode"`
	ServiceKind    string                 `json:"service_kind"`
	Status         string                 `json:"status"`
	Labels         []LabelKV              `json:"labels"`
	StandardConfig *StandardServiceConfig `json:"standard_config,omitempty"`
	CustomYAML     string                 `json:"custom_yaml,omitempty"`
	RenderedYAML   string                 `json:"rendered_yaml,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type HelmImportReq struct {
	ServiceID    uint   `json:"service_id"`
	ChartName    string `json:"chart_name"`
	ChartVersion string `json:"chart_version"`
	ChartRef     string `json:"chart_ref"`
	ValuesYAML   string `json:"values_yaml"`
	RenderedYAML string `json:"rendered_yaml"`
}

type HelmRenderReq struct {
	ReleaseID    uint   `json:"release_id"`
	ChartRef     string `json:"chart_ref"`
	ChartName    string `json:"chart_name"`
	ValuesYAML   string `json:"values_yaml"`
	RenderedYAML string `json:"rendered_yaml"`
}

type DeployReq struct {
	ClusterID     uint   `json:"cluster_id"`
	DeployTarget  string `json:"deploy_target"` // k8s/compose/helm
	ApprovalToken string `json:"approval_token"`
}
