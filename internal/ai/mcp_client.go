package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPConfig struct {
	Enable    bool
	Transport string
	Endpoint  string
	Command   string
}

type MCPToolInfo struct {
	Name        string
	Description string
}

type MCPClientManager struct {
	mu    sync.RWMutex
	cli   *client.Client
	tools []MCPToolInfo
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
	return MCPConfig{
		Enable:    enable,
		Transport: transportType,
		Endpoint:  endpoint,
		Command:   command,
	}
}

func NewMCPClientManager(ctx context.Context, cfg MCPConfig) (*MCPClientManager, error) {
	if !cfg.Enable {
		return nil, nil
	}

	manager := &MCPClientManager{}
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
	if err := manager.refreshTools(startCtx); err != nil {
		_ = cli.Close()
		return nil, err
	}
	return manager, nil
}

func (m *MCPClientManager) refreshTools(ctx context.Context) error {
	if m == nil || m.cli == nil {
		return nil
	}
	res, err := m.cli.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("mcp list tools failed: %w", err)
	}
	out := make([]MCPToolInfo, 0, len(res.Tools))
	for _, t := range res.Tools {
		out = append(out, MCPToolInfo{
			Name:        t.Name,
			Description: t.Description,
		})
	}
	m.mu.Lock()
	m.tools = out
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
		return nil, fmt.Errorf("mcp client not enabled")
	}
	return m.cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
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
