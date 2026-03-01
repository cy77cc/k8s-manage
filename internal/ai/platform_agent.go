package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

const toolCallGuide = `Tool calling rules:
1) NEVER call a tool with empty {} arguments when the tool has required fields.
2) Prefer using runtime context (scene/page/host_id/cluster_id/service_id/namespace) as arguments.
3) If any required field is missing, ask for it or choose a safe readonly tool first.
4) For mutating tools, require approval token before execution.
5) For inventory/list/assets/清单/已添加资源 requests, call inventory tools first (host_list_inventory/cluster_list_inventory/service_list_inventory).
6) Return concise explanation after each tool result.`

type PlatformAgent struct {
	Runnable *react.Agent
	Model    model.ToolCallingChatModel
	experts  map[string]*react.Agent
	tools    map[string]tool.InvokableTool
	metas    map[string]tools.ToolMeta
	mcp      *tools.MCPClientManager
}

func NewPlatformAgent(ctx context.Context, chatModel model.ToolCallingChatModel, deps tools.PlatformDeps) (*PlatformAgent, error) {
	if chatModel == nil {
		return nil, nil
	}

	localTools, err := tools.BuildLocalTools(deps)
	if err != nil {
		return nil, err
	}
	mcpManager, err := tools.NewMCPClientManager(ctx, tools.MCPConfigFromEnv())
	if err != nil {
		return nil, err
	}
	mcpTools, err := tools.BuildMCPProxyTools(mcpManager)
	if err != nil {
		return nil, err
	}
	registered := append(localTools, mcpTools...)
	baseTools := make([]tool.BaseTool, 0, len(registered))
	toolMap := make(map[string]tool.InvokableTool, len(registered))
	metaMap := make(map[string]tools.ToolMeta, len(registered))
	for _, item := range registered {
		baseTools = append(baseTools, item.Tool)
		toolMap[item.Meta.Name] = item.Tool
		metaMap[item.Meta.Name] = item.Meta
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: baseTools,
		},
		MaxStep:         20,
		MessageModifier: react.NewPersonaModifier("You are Platform Ops Agent. Use tools safely and with complete parameters.\n" + toolCallGuide),
	})
	if err != nil {
		return nil, err
	}
	opsExpert, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: baseTools,
		},
		MaxStep:         20,
		MessageModifier: react.NewPersonaModifier("You are Ops Expert. Focus on host/os diagnostics, stability and safe operations.\n" + toolCallGuide),
	})
	if err != nil {
		return nil, err
	}
	k8sExpert, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: baseTools,
		},
		MaxStep:         20,
		MessageModifier: react.NewPersonaModifier("You are Kubernetes Expert. Focus on cluster health, events, pods and rollout troubleshooting.\n" + toolCallGuide),
	})
	if err != nil {
		return nil, err
	}
	securityExpert, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: baseTools,
		},
		MaxStep:         20,
		MessageModifier: react.NewPersonaModifier("You are Security/RBAC Expert. Focus on permissions, roles, least privilege and access diagnostics.\n" + toolCallGuide),
	})
	if err != nil {
		return nil, err
	}

	return &PlatformAgent{
		Runnable: agent,
		Model:    chatModel,
		experts: map[string]*react.Agent{
			"default":  agent,
			"ops":      opsExpert,
			"k8s":      k8sExpert,
			"security": securityExpert,
		},
		tools: toolMap,
		metas: metaMap,
		mcp:   mcpManager,
	}, nil
}

func (p *PlatformAgent) ToolMetas() []tools.ToolMeta {
	if p == nil {
		return nil
	}
	out := make([]tools.ToolMeta, 0, len(p.metas))
	for _, m := range p.metas {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (p *PlatformAgent) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	if p == nil {
		return nil, fmt.Errorf("agent not initialized")
	}
	a := p.selectAgent(messages)
	return a.Stream(ctx, messages)
}

func (p *PlatformAgent) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	if p == nil {
		return nil, fmt.Errorf("agent not initialized")
	}
	a := p.selectAgent(messages)
	return a.Generate(ctx, messages)
}

func (p *PlatformAgent) selectAgent(messages []*schema.Message) *react.Agent {
	if p == nil || p.experts == nil {
		return p.Runnable
	}
	content := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i] != nil && messages[i].Role == schema.User {
			content = strings.ToLower(strings.TrimSpace(messages[i].Content))
			break
		}
	}
	switch {
	case strings.Contains(content, "k8s") || strings.Contains(content, "kubernetes") || strings.Contains(content, "pod") || strings.Contains(content, "deployment") || strings.Contains(content, "cluster"):
		return p.experts["k8s"]
	case strings.Contains(content, "rbac") || strings.Contains(content, "permission") || strings.Contains(content, "role") || strings.Contains(content, "权限"):
		return p.experts["security"]
	case strings.Contains(content, "ssh") || strings.Contains(content, "cpu") || strings.Contains(content, "memory") || strings.Contains(content, "disk") || strings.Contains(content, "host") || strings.Contains(content, "系统"):
		return p.experts["ops"]
	default:
		return p.experts["default"]
	}
}

func (p *PlatformAgent) RunTool(ctx context.Context, toolName string, params map[string]any) (tools.ToolResult, error) {
	if p == nil {
		return tools.ToolResult{
				OK:     false,
				Error:  "agent not initialized",
				Source: "platform",
			},
			fmt.Errorf("agent not initialized")
	}
	normalizedName := tools.NormalizeToolName(toolName)
	t, ok := p.tools[normalizedName]
	if !ok {
		return tools.ToolResult{
				OK:     false,
				Error:  "tool not found",
				Source: "platform",
			},
			fmt.Errorf("tool not found")
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return tools.ToolResult{
				OK:     false,
				Error:  err.Error(),
				Source: "platform",
			},
			err
	}
	out, err := t.InvokableRun(ctx, string(raw))
	if err != nil {
		return tools.ToolResult{
				OK:     false,
				Error:  err.Error(),
				Source: "platform",
			},
			nil
	}
	var result tools.ToolResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return tools.ToolResult{
				OK:     true,
				Data:   out,
				Source: "platform",
			},
			nil
	}
	return result, nil
}

func (p *PlatformAgent) Close() error {
	if p == nil || p.mcp == nil {
		return nil
	}
	return p.mcp.Close()
}
