package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	modelcomponent "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

type OrchestratorPlanner struct {
	model       modelcomponent.ToolCallingChatModel
	descriptors []DomainDescriptor
}

func NewPlanner(model modelcomponent.ToolCallingChatModel) *OrchestratorPlanner {
	return &OrchestratorPlanner{
		model: model,
		descriptors: []DomainDescriptor{
			{Domain: types.DomainInfrastructure, Description: "主机、K8s、容器运行时"},
			{Domain: types.DomainService, Description: "服务、部署、凭证"},
			{Domain: types.DomainCICD, Description: "流水线、任务"},
			{Domain: types.DomainMonitor, Description: "监控、告警、拓扑"},
			{Domain: types.DomainConfig, Description: "配置管理"},
			{Domain: types.DomainUser, Description: "用户、角色、权限"},
			{Domain: types.DomainGeneral, Description: "无法归类的通用请求"},
		},
	}
}

func (p *OrchestratorPlanner) Plan(ctx context.Context, message string) ([]DomainRequest, error) {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return nil, fmt.Errorf("message is required")
	}
	if p != nil && p.model != nil {
		requests, err := p.planWithModel(ctx, trimmed)
		if err != nil {
			return nil, err
		}
		if len(requests) > 0 {
			return requests, nil
		}
	}
	return detectDomains(trimmed), nil
}

func (p *OrchestratorPlanner) planWithModel(ctx context.Context, message string) ([]DomainRequest, error) {
	prompt := BuildDomainSelectionPrompt(message, p.descriptors)
	msg, err := p.model.Generate(ctx, []*schema.Message{schema.SystemMessage(prompt), schema.UserMessage(message)})
	if err != nil {
		return nil, err
	}
	if msg == nil || strings.TrimSpace(msg.Content) == "" {
		return nil, nil
	}
	return decodeRequests(msg.Content)
}

func decodeRequests(content string) ([]DomainRequest, error) {
	raw := strings.TrimSpace(content)
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}
	var requests []DomainRequest
	if err := json.Unmarshal([]byte(raw), &requests); err != nil {
		return nil, nil
	}
	return normalizeRequests(requests), nil
}

func normalizeRequests(requests []DomainRequest) []DomainRequest {
	if len(requests) == 0 {
		return nil
	}
	seen := make(map[types.Domain]struct{}, len(requests))
	normalized := make([]DomainRequest, 0, len(requests))
	for _, item := range requests {
		domain := item.Domain
		if domain == "" {
			domain = types.DomainGeneral
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		if item.Context == nil {
			item.Context = map[string]any{}
		}
		item.Domain = domain
		normalized = append(normalized, item)
	}
	return normalized
}

func detectDomains(message string) []DomainRequest {
	msg := strings.ToLower(message)
	requests := make([]DomainRequest, 0, 2)
	add := func(domain types.Domain, focus string) {
		for _, item := range requests {
			if item.Domain == domain {
				return
			}
		}
		requests = append(requests, DomainRequest{Domain: domain, Context: map[string]any{"focus": focus}})
	}
	if strings.Contains(msg, "主机") || strings.Contains(msg, "服务器") || strings.Contains(msg, "k8s") || strings.Contains(msg, "pod") || strings.Contains(msg, "cluster") {
		add(types.DomainInfrastructure, "runtime")
	}
	if strings.Contains(msg, "服务") || strings.Contains(msg, "deploy") || strings.Contains(msg, "部署") || strings.Contains(msg, "发布") {
		add(types.DomainService, "delivery")
	}
	if strings.Contains(msg, "pipeline") || strings.Contains(msg, "流水线") || strings.Contains(msg, "job") || strings.Contains(msg, "构建") {
		add(types.DomainCICD, "pipeline")
	}
	if strings.Contains(msg, "告警") || strings.Contains(msg, "监控") || strings.Contains(msg, "拓扑") || strings.Contains(msg, "alert") {
		add(types.DomainMonitor, "alerts")
	}
	if strings.Contains(msg, "配置") || strings.Contains(msg, "config") {
		add(types.DomainConfig, "configuration")
	}
	if strings.Contains(msg, "用户") || strings.Contains(msg, "角色") || strings.Contains(msg, "权限") || strings.Contains(msg, "role") || strings.Contains(msg, "permission") {
		add(types.DomainUser, "access")
	}
	if len(requests) == 0 {
		return []DomainRequest{{Domain: types.DomainGeneral, Context: map[string]any{}}}
	}
	return requests
}
