package ai

import (
	"sort"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

type sceneMeta struct {
	Scene        string   `json:"scene"`
	Description  string   `json:"description"`
	Keywords     []string `json:"keywords"`
	Tools        []string `json:"tools"`
	ContextHints []string `json:"context_hints"`
}

var sceneRegistry = map[string]sceneMeta{
	"deployment:clusters": {
		Scene:        "deployment:clusters",
		Description:  "集群管理场景",
		Keywords:     []string{"cluster", "k8s", "集群"},
		Tools:        []string{"cluster_list_inventory", "k8s_list_resources", "k8s_get_events", "deployment_bootstrap_status"},
		ContextHints: []string{"cluster_id"},
	},
	"deployment:credentials": {
		Scene:        "deployment:credentials",
		Description:  "凭证管理场景",
		Keywords:     []string{"credential", "凭证"},
		Tools:        []string{"credential_list", "credential_test"},
		ContextHints: []string{"credential_id"},
	},
	"deployment:hosts": {
		Scene:        "deployment:hosts",
		Description:  "主机管理场景",
		Keywords:     []string{"host", "ssh", "主机"},
		Tools:        []string{"host_list_inventory", "host_ssh_exec_readonly", "host_batch_exec_preview"},
		ContextHints: []string{"host_id"},
	},
	"deployment:targets": {
		Scene:        "deployment:targets",
		Description:  "部署目标场景",
		Keywords:     []string{"target", "部署目标"},
		Tools:        []string{"deployment_target_list", "deployment_target_detail", "deployment_bootstrap_status"},
		ContextHints: []string{"target_id", "env"},
	},
	"deployment:releases": {
		Scene:        "deployment:releases",
		Description:  "发布管理场景",
		Keywords:     []string{"release", "发布"},
		Tools:        []string{"service_deploy_preview", "service_deploy_apply", "cicd_pipeline_trigger"},
		ContextHints: []string{"service_id", "cluster_id", "env"},
	},
	"deployment:approvals": {
		Scene:        "deployment:approvals",
		Description:  "审批场景",
		Keywords:     []string{"approval", "审批"},
		Tools:        []string{"cicd_pipeline_status", "audit_log_search"},
		ContextHints: []string{"release_id"},
	},
	"deployment:topology": {
		Scene:        "deployment:topology",
		Description:  "拓扑场景",
		Keywords:     []string{"topology", "拓扑"},
		Tools:        []string{"topology_get", "service_get_detail"},
		ContextHints: []string{"service_id"},
	},
	"deployment:metrics": {
		Scene:        "deployment:metrics",
		Description:  "指标场景",
		Keywords:     []string{"metric", "指标"},
		Tools:        []string{"monitor_metric_query", "monitor_alert_active", "monitor_alert_rule_list"},
		ContextHints: []string{"service_id"},
	},
	"deployment:audit": {
		Scene:        "deployment:audit",
		Description:  "审计场景",
		Keywords:     []string{"audit", "审计"},
		Tools:        []string{"audit_log_search"},
		ContextHints: []string{"user_id"},
	},
	"deployment:aiops": {
		Scene:        "deployment:aiops",
		Description:  "智能运维场景",
		Keywords:     []string{"aiops", "智能运维"},
		Tools:        []string{"monitor_alert_active", "monitor_metric_query", "ops_aggregate_status"},
		ContextHints: []string{"cluster_id", "service_id"},
	},
	"services:list": {
		Scene:        "services:list",
		Description:  "服务列表场景",
		Keywords:     []string{"service", "服务"},
		Tools:        []string{"service_list_inventory", "service_catalog_list", "service_category_tree"},
		ContextHints: []string{"service_id"},
	},
	"services:detail": {
		Scene:        "services:detail",
		Description:  "服务详情场景",
		Keywords:     []string{"service", "详情"},
		Tools:        []string{"service_get_detail", "service_visibility_check", "service_deploy_preview"},
		ContextHints: []string{"service_id"},
	},
	"services:provision": {
		Scene:        "services:provision",
		Description:  "服务创建场景",
		Keywords:     []string{"provision", "创建服务"},
		Tools:        []string{"service_category_tree", "service_catalog_list"},
		ContextHints: []string{"project_id"},
	},
	"services:deploy": {
		Scene:        "services:deploy",
		Description:  "服务部署场景",
		Keywords:     []string{"deploy", "部署"},
		Tools:        []string{"service_deploy_preview", "service_deploy_apply", "deployment_target_list"},
		ContextHints: []string{"service_id", "target_id"},
	},
	"services:catalog": {
		Scene:        "services:catalog",
		Description:  "服务目录场景",
		Keywords:     []string{"catalog", "目录"},
		Tools:        []string{"service_catalog_list", "service_category_tree", "service_visibility_check"},
		ContextHints: []string{"category_id"},
	},
	"governance:users": {
		Scene:        "governance:users",
		Description:  "用户治理场景",
		Keywords:     []string{"user", "用户"},
		Tools:        []string{"user_list", "permission_check"},
		ContextHints: []string{"user_id"},
	},
	"governance:roles": {
		Scene:        "governance:roles",
		Description:  "角色治理场景",
		Keywords:     []string{"role", "角色"},
		Tools:        []string{"role_list", "permission_check"},
		ContextHints: []string{"role_id"},
	},
	"governance:permissions": {
		Scene:        "governance:permissions",
		Description:  "权限治理场景",
		Keywords:     []string{"permission", "权限"},
		Tools:        []string{"permission_check", "audit_log_search"},
		ContextHints: []string{"user_id", "resource", "action"},
	},
}

func normalizeSceneKey(scene string) string {
	v := strings.TrimSpace(scene)
	v = strings.TrimPrefix(v, "scene:")
	return v
}

func sceneMetaByKey(scene string) (sceneMeta, bool) {
	meta, ok := sceneRegistry[normalizeSceneKey(scene)]
	return meta, ok
}

func (h *handler) sceneRecommendedTools(scene string) []tools.ToolMeta {
	if h == nil || h.svcCtx == nil || h.svcCtx.AI == nil {
		return nil
	}
	meta, ok := sceneMetaByKey(scene)
	if !ok {
		return nil
	}
	all := h.svcCtx.AI.ToolMetas()
	metaByName := make(map[string]tools.ToolMeta, len(all))
	for _, item := range all {
		metaByName[item.Name] = item
	}
	out := make([]tools.ToolMeta, 0, len(meta.Tools))
	for _, name := range meta.Tools {
		if item, exists := metaByName[name]; exists {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
