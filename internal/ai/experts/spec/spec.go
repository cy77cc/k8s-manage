package spec

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
)

type ToolCapability struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Mode        string `json:"mode"`
	Risk        string `json:"risk"`
}

type ToolExport struct {
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Capabilities []ToolCapability `json:"capabilities"`
}

type Expert interface {
	Name() string
	Description() string
	Capabilities() []ToolCapability
	Tools(ctx context.Context) []tool.InvokableTool
	AsTool() ToolExport
}

func FilterToolsByName(ctx context.Context, tools []tool.InvokableTool, excluded ...string) []tool.InvokableTool {
	if len(excluded) == 0 {
		return tools
	}
	blocked := make(map[string]struct{}, len(excluded))
	for _, name := range excluded {
		if name != "" {
			blocked[name] = struct{}{}
		}
	}
	out := make([]tool.InvokableTool, 0, len(tools))
	for _, invokable := range tools {
		if invokable == nil {
			continue
		}
		info, err := invokable.Info(ctx)
		if err != nil || info == nil {
			out = append(out, invokable)
			continue
		}
		if _, skip := blocked[info.Name]; skip {
			continue
		}
		out = append(out, invokable)
	}
	return out
}
