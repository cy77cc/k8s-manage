package v1

import "time"

type CatalogVariableSchema struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Default     any      `json:"default,omitempty"`
	Required    bool     `json:"required"`
	Description string   `json:"description,omitempty"`
	Options     []string `json:"options,omitempty"`
}

type CategoryCreateRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type CategoryUpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

type CategoryResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TemplateCreateRequest struct {
	Name            string                  `json:"name"`
	DisplayName     string                  `json:"display_name"`
	Description     string                  `json:"description"`
	Icon            string                  `json:"icon"`
	CategoryID      uint                    `json:"category_id"`
	Version         string                  `json:"version"`
	Visibility      string                  `json:"visibility"`
	K8sTemplate     string                  `json:"k8s_template"`
	ComposeTemplate string                  `json:"compose_template"`
	VariablesSchema []CatalogVariableSchema `json:"variables_schema"`
	Readme          string                  `json:"readme"`
	Tags            []string                `json:"tags"`
}

type TemplateUpdateRequest struct {
	DisplayName     *string                  `json:"display_name,omitempty"`
	Description     *string                  `json:"description,omitempty"`
	Icon            *string                  `json:"icon,omitempty"`
	CategoryID      *uint                    `json:"category_id,omitempty"`
	Version         *string                  `json:"version,omitempty"`
	Visibility      *string                  `json:"visibility,omitempty"`
	K8sTemplate     *string                  `json:"k8s_template,omitempty"`
	ComposeTemplate *string                  `json:"compose_template,omitempty"`
	VariablesSchema *[]CatalogVariableSchema `json:"variables_schema,omitempty"`
	Readme          *string                  `json:"readme,omitempty"`
	Tags            *[]string                `json:"tags,omitempty"`
}

type TemplateResponse struct {
	ID              uint                    `json:"id"`
	Name            string                  `json:"name"`
	DisplayName     string                  `json:"display_name"`
	Description     string                  `json:"description"`
	Icon            string                  `json:"icon"`
	CategoryID      uint                    `json:"category_id"`
	Version         string                  `json:"version"`
	OwnerID         uint64                  `json:"owner_id"`
	Visibility      string                  `json:"visibility"`
	Status          string                  `json:"status"`
	K8sTemplate     string                  `json:"k8s_template"`
	ComposeTemplate string                  `json:"compose_template"`
	VariablesSchema []CatalogVariableSchema `json:"variables_schema"`
	Readme          string                  `json:"readme"`
	Tags            []string                `json:"tags"`
	DeployCount     int                     `json:"deploy_count"`
	ReviewNote      string                  `json:"review_note"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

type TemplateListResponse struct {
	List  []TemplateResponse `json:"list"`
	Total int64              `json:"total"`
}

type ReviewActionRequest struct {
	Reason string `json:"reason"`
}

type PreviewRequest struct {
	TemplateID uint           `json:"template_id"`
	Target     string         `json:"target"`
	Variables  map[string]any `json:"variables"`
}

type PreviewResponse struct {
	RenderedYAML   string   `json:"rendered_yaml"`
	UnresolvedVars []string `json:"unresolved_vars"`
}

type DeployRequest struct {
	TemplateID  uint           `json:"template_id"`
	Target      string         `json:"target"`
	ProjectID   uint           `json:"project_id"`
	TeamID      uint           `json:"team_id"`
	ServiceName string         `json:"service_name"`
	Namespace   string         `json:"namespace"`
	ClusterID   uint           `json:"cluster_id"`
	Environment string         `json:"environment"`
	Variables   map[string]any `json:"variables"`
	DeployNow   bool           `json:"deploy_now"`
}

type DeployResponse struct {
	ServiceID   uint `json:"service_id"`
	TemplateID  uint `json:"template_id"`
	DeployCount int  `json:"deploy_count"`
}
