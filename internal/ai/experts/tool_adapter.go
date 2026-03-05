package experts

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// ExpertToolInput is the structured input for expert-as-tool calls.
type ExpertToolInput struct {
	Task string `json:"task" jsonschema:"required,description=helper expert task"`
}

// BuildExpertTool exposes an expert as an invokable tool.
func BuildExpertTool(expert *Expert) (tool.InvokableTool, error) {
	if expert == nil || strings.TrimSpace(expert.Name) == "" {
		return nil, fmt.Errorf("expert is nil or name is empty")
	}
	name := strings.TrimSpace(expert.Name)
	desc := strings.TrimSpace(expert.Persona)
	if desc == "" {
		desc = "专家协作工具"
	}
	info, err := toolutils.GoStruct2ToolInfo[ExpertToolInput](name, desc)
	if err != nil {
		return nil, err
	}
	t := toolutils.NewTool(info, func(ctx context.Context, input ExpertToolInput) (string, error) {
		task := strings.TrimSpace(input.Task)
		if task == "" {
			return "", fmt.Errorf("task is required")
		}
		if expert.Agent == nil {
			return fmt.Sprintf("专家 %s 暂不可用，任务：%s", expert.Name, task), nil
		}
		prompt := fmt.Sprintf("请协助完成以下任务，并输出可直接汇总的结果：%s", task)
		resp, err := expert.Agent.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
		if err != nil {
			return "", err
		}
		if resp == nil {
			return "", nil
		}
		return strings.TrimSpace(resp.Content), nil
	})
	return t, nil
}

// BuildExpertTools builds helper-expert tool set, excluding self expert.
func BuildExpertTools(expertsMap map[string]*Expert, self string) map[string]tool.InvokableTool {
	out := map[string]tool.InvokableTool{}
	if len(expertsMap) == 0 {
		return out
	}
	self = strings.TrimSpace(self)
	names := make([]string, 0, len(expertsMap))
	for name := range expertsMap {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if self != "" && name == self {
			continue
		}
		t, err := BuildExpertTool(expertsMap[name])
		if err != nil {
			continue
		}
		out[name] = t
	}
	return out
}
