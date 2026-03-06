package tools

type ServiceCatalogListInput struct {
	Keyword    string `json:"keyword,omitempty" jsonschema:"description=optional keyword on service name/owner"`
	CategoryID int    `json:"category_id,omitempty" jsonschema:"description=optional category id: 1 middleware, 2 business"`
	Limit      int    `json:"limit,omitempty" jsonschema:"description=max services,default=50"`
}

type ServiceVisibilityCheckInput struct {
	ServiceID int `json:"service_id" jsonschema:"required,description=service id"`
}

type DeploymentTargetListInput struct {
	Env     string `json:"env,omitempty" jsonschema:"description=optional environment filter"`
	Status  string `json:"status,omitempty" jsonschema:"description=optional target status filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional target keyword filter"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max targets,default=50"`
}

type DeploymentTargetDetailInput struct {
	TargetID int `json:"target_id" jsonschema:"required,description=deployment target id"`
}

type DeploymentBootstrapStatusInput struct {
	TargetID int `json:"target_id" jsonschema:"required,description=deployment target id"`
}

type CredentialListInput struct {
	Type    string `json:"type,omitempty" jsonschema:"description=credential type or runtime type"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on name/endpoint"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max credentials,default=50"`
}

type CredentialTestInput struct {
	CredentialID int `json:"credential_id" jsonschema:"required,description=credential id"`
}

type CICDPipelineListInput struct {
	Status  string `json:"status,omitempty" jsonschema:"description=optional status filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on repo/branch"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max pipelines,default=50"`
}

type CICDPipelineStatusInput struct {
	PipelineID int `json:"pipeline_id" jsonschema:"required,description=pipeline config id"`
}

type CICDPipelineTriggerInput struct {
	PipelineID int               `json:"pipeline_id" jsonschema:"required,description=pipeline config id"`
	Branch     string            `json:"branch" jsonschema:"required,description=branch to build"`
	Params     map[string]string `json:"params,omitempty" jsonschema:"description=optional trigger params"`
}

type JobListInput struct {
	Status  string `json:"status,omitempty" jsonschema:"description=optional status filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on name/type"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max jobs,default=50"`
}

type JobExecutionStatusInput struct {
	JobID       int `json:"job_id" jsonschema:"required,description=job id"`
	ExecutionID int `json:"execution_id,omitempty" jsonschema:"description=optional execution id"`
}

type JobRunInput struct {
	JobID  int            `json:"job_id" jsonschema:"required,description=job id"`
	Params map[string]any `json:"params,omitempty" jsonschema:"description=optional run params"`
}

type ConfigAppListInput struct {
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on service name"`
	Env     string `json:"env,omitempty" jsonschema:"description=optional env filter"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max apps,default=50"`
}

type ConfigItemGetInput struct {
	AppID int    `json:"app_id" jsonschema:"required,description=service id as config app id"`
	Key   string `json:"key" jsonschema:"required,description=config key"`
	Env   string `json:"env,omitempty" jsonschema:"description=optional env"`
}

type ConfigDiffInput struct {
	AppID int    `json:"app_id" jsonschema:"required,description=service id as config app id"`
	EnvA  string `json:"env_a" jsonschema:"required,description=compare env a"`
	EnvB  string `json:"env_b" jsonschema:"required,description=compare env b"`
}

type MonitorAlertRuleListInput struct {
	Status  string `json:"status,omitempty" jsonschema:"description=optional rule state filter"`
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional keyword on name/metric"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max rules,default=50"`
}

type MonitorAlertActiveInput struct {
	Severity  string `json:"severity,omitempty" jsonschema:"description=optional severity filter"`
	ServiceID int    `json:"service_id,omitempty" jsonschema:"description=optional service id filter"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max alerts,default=50"`
}

type MonitorAlertInput struct {
	Severity  string `json:"severity,omitempty" jsonschema:"description=optional severity filter"`
	ServiceID int    `json:"service_id,omitempty" jsonschema:"description=optional service id filter"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=max alerts,default=50"`
}

type MonitorMetricQueryInput struct {
	Query     string `json:"query" jsonschema:"required,description=metric query or metric name"`
	TimeRange string `json:"time_range,omitempty" jsonschema:"description=time range,default=1h"`
	Step      int    `json:"step,omitempty" jsonschema:"description=step seconds,default=60"`
}

type MonitorMetricInput struct {
	Query     string `json:"query" jsonschema:"required,description=metric query or metric name"`
	TimeRange string `json:"time_range,omitempty" jsonschema:"description=time range,default=1h"`
	Step      int    `json:"step,omitempty" jsonschema:"description=step seconds,default=60"`
}

type TopologyGetInput struct {
	ServiceID int `json:"service_id,omitempty" jsonschema:"description=optional service id"`
	Depth     int `json:"depth,omitempty" jsonschema:"description=max depth,default=2"`
}

type AuditLogSearchInput struct {
	TimeRange    string `json:"time_range,omitempty" jsonschema:"description=time range,default=24h"`
	ResourceType string `json:"resource_type,omitempty" jsonschema:"description=optional resource type"`
	Action       string `json:"action,omitempty" jsonschema:"description=optional action type"`
	UserID       int    `json:"user_id,omitempty" jsonschema:"description=optional actor user id"`
	Limit        int    `json:"limit,omitempty" jsonschema:"description=max logs,default=50"`
}

type UserListInput struct {
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional username/email keyword"`
	Status  int    `json:"status,omitempty" jsonschema:"description=optional status filter"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max users,default=50"`
}

type RoleListInput struct {
	Keyword string `json:"keyword,omitempty" jsonschema:"description=optional role keyword"`
	Limit   int    `json:"limit,omitempty" jsonschema:"description=max roles,default=50"`
}

type PermissionCheckInput struct {
	UserID   int    `json:"user_id" jsonschema:"required,description=user id"`
	Resource string `json:"resource" jsonschema:"required,description=resource name"`
	Action   string `json:"action" jsonschema:"required,description=action name"`
}
