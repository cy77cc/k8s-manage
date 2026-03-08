package router

import "github.com/cy77cc/k8s-manage/internal/ai/tools"

// DefaultDomainRoutingConfig returns the baseline domain routing rules used by the classifier.
func DefaultDomainRoutingConfig() []DomainRouteConfig {
	return []DomainRouteConfig{
		{
			Domain: tools.DomainInfrastructure,
			Keywords: []string{
				"host", "node", "server", "ssh", "machine", "cpu", "memory", "disk", "cluster",
				"k8s", "kubernetes", "pod", "namespace", "container", "日志", "节点", "集群", "主机",
			},
		},
		{
			Domain: tools.DomainService,
			Keywords: []string{
				"service", "deploy", "deployment", "credential", "rollout",
				"服务", "部署", "发布", "应用", "凭证",
			},
		},
		{
			Domain: tools.DomainCICD,
			Keywords: []string{
				"pipeline", "build", "job", "ci", "cd", "workflow", "流水线", "构建", "任务",
			},
		},
		{
			Domain: tools.DomainMonitor,
			Keywords: []string{
				"monitor", "metric", "metrics", "alert", "alarm", "prometheus", "grafana",
				"监控", "指标", "告警",
			},
		},
		{
			Domain: tools.DomainConfig,
			Keywords: []string{
				"config", "configuration", "env", "environment", "setting", "settings",
				"配置", "环境", "参数",
			},
		},
		{
			Domain: tools.DomainGeneral,
			Keywords: []string{
				"help", "explain", "summary", "summarize", "assist", "帮助", "解释", "总结",
			},
		},
	}
}
