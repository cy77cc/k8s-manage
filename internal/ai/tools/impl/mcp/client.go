package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	. "github.com/cy77cc/k8s-manage/internal/ai/tools/core"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPConfig struct {
	Enable    bool
	Transport string
	Endpoint  string
	Command   string
	Prefix    string
}

type MCPToolInfo struct {
	Name        string
	RemoteName  string
	Description string
	Schema      map[string]any
	Required    []string
}

type MCPClientManager struct {
	mu          sync.RWMutex
	cli         *client.Client
	tools       []MCPToolInfo
	prefix      string
	remoteIndex map[string]string
	callToolFn  func(ctx context.Context, toolName string, args map[string]any) (*mcp.CallToolResult, error)
}

func MCPConfigFromEnv() MCPConfig {
	enable := strings.EqualFold(strings.TrimSpace(os.Getenv("AI_MCP_ENABLE")), "true")
	transportType := strings.TrimSpace(os.Getenv("AI_MCP_TRANSPORT"))
	if transportType == "" {
		transportType = "sse"
	}
	endpoint := strings.TrimSpace(os.Getenv("AI_MCP_ENDPOINT"))
	if endpoint == "" {
		endpoint = "http://localhost:12345/sse"
	}
	command := strings.TrimSpace(os.Getenv("AI_MCP_COMMAND"))
	prefix := NormalizeToolName(strings.TrimSpace(os.Getenv("AI_MCP_PREFIX")))
	if prefix == "" {
		prefix = "mcp_default"
	}
	return MCPConfig{
		Enable:    enable,
		Transport: transportType,
		Endpoint:  endpoint,
		Command:   command,
		Prefix:    prefix,
	}
}

func NewMCPClientManager(ctx context.Context, cfg MCPConfig) (*MCPClientManager, error) {
	if !cfg.Enable {
		return nil, nil
	}

	manager := &MCPClientManager{
		prefix:      sanitizeMCPPrefix(cfg.Prefix),
		remoteIndex: map[string]string{},
	}
	var cli *client.Client
	var err error
	switch strings.ToLower(strings.TrimSpace(cfg.Transport)) {
	case "stdio":
		if cfg.Command == "" {
			return nil, fmt.Errorf("AI_MCP_COMMAND is required for stdio transport")
		}
		tr := transport.NewStdio("sh", nil, "-lc", cfg.Command)
		cli = client.NewClient(tr)
	case "sse", "":
		cli, err = client.NewSSEMCPClient(cfg.Endpoint)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported mcp transport: %s", cfg.Transport)
	}
	if err != nil {
		return nil, err
	}

	startCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := cli.Start(startCtx); err != nil {
		return nil, fmt.Errorf("mcp start failed: %w", err)
	}
	_, err = cli.Initialize(startCtx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "k8s-manage-ai",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("mcp initialize failed: %w", err)
	}

	manager.cli = cli
	if err := manager.RefreshTools(startCtx); err != nil {
		_ = cli.Close()
		return nil, err
	}
	return manager, nil
}

func sanitizeMCPPrefix(prefix string) string {
	normalized := NormalizeToolName(prefix)
	if normalized == "" {
		return "mcp_default"
	}
	return normalized
}

func (m *MCPClientManager) Prefix() string {
	if m == nil {
		return "mcp_default"
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if strings.TrimSpace(m.prefix) == "" {
		return "mcp_default"
	}
	return m.prefix
}

func (m *MCPClientManager) ToolNameForRemote(remoteName string) string {
	prefix := m.Prefix()
	if strings.TrimSpace(remoteName) == "" {
		return prefix
	}
	return prefix + "_" + NormalizeToolName(remoteName)
}

func (m *MCPClientManager) RemoteNameForTool(toolName string) string {
	if m == nil {
		return ""
	}
	normalized := NormalizeToolName(toolName)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.remoteIndex[normalized]
}

func (m *MCPClientManager) RefreshTools(ctx context.Context) error {
	if m == nil || m.cli == nil {
		return nil
	}
	res, err := m.cli.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("mcp list tools failed: %w", err)
	}
	out := make([]MCPToolInfo, 0, len(res.Tools))
	index := make(map[string]string, len(res.Tools))
	for _, t := range res.Tools {
		schema := map[string]any{}
		required := []string{}
		if t.InputSchema.Type != "" {
			schema["type"] = t.InputSchema.Type
			schema["properties"] = t.InputSchema.Properties
			required = append(required, t.InputSchema.Required...)
			if len(required) > 0 {
				schema["required"] = required
			}
		}
		prefixedName := m.ToolNameForRemote(t.Name)
		out = append(out, MCPToolInfo{
			Name:        prefixedName,
			RemoteName:  t.Name,
			Description: t.Description,
			Schema:      schema,
			Required:    required,
		})
		index[prefixedName] = t.Name
	}
	m.mu.Lock()
	m.tools = out
	m.remoteIndex = index
	m.mu.Unlock()
	return nil
}

func (m *MCPClientManager) ListTools() []MCPToolInfo {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]MCPToolInfo, 0, len(m.tools))
	out = append(out, m.tools...)
	return out
}

func (m *MCPClientManager) CallTool(ctx context.Context, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	if m == nil || m.cli == nil {
		if m == nil || m.callToolFn == nil {
			return nil, fmt.Errorf("mcp client not enabled")
		}
	}
	remoteName := strings.TrimSpace(toolName)
	if mapped := strings.TrimSpace(m.RemoteNameForTool(toolName)); mapped != "" {
		remoteName = mapped
	}
	if m.callToolFn != nil {
		return m.callToolFn(ctx, remoteName, args)
	}
	return m.cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      remoteName,
			Arguments: args,
		},
	})
}

func (m *MCPClientManager) Close() error {
	if m == nil || m.cli == nil {
		return nil
	}
	return m.cli.Close()
}
