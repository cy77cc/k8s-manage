package summarizer

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type decisionTool struct {
	info *schema.ToolInfo
}

func (t decisionTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return t.info, nil
}

func (t decisionTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	return argumentsInJSON, nil
}

func summaryDecisionTool() tool.BaseTool {
	return decisionTool{
		info: &schema.ToolInfo{
			Name: "emit_summary",
			Desc: "Emit the final structured summary output for the current execution iteration.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"summary": {
					Type:     schema.String,
					Required: true,
					Desc:     "short structured summary for the frontend summary stage",
				},
				"conclusion": {
					Type: schema.String,
					Desc: "high-level conclusion for the user",
				},
				"next_actions": {
					Type: schema.Array,
					ElemInfo: &schema.ParameterInfo{
						Type: schema.String,
					},
				},
				"need_more_investigation": {
					Type:     schema.Boolean,
					Required: true,
				},
				"narrative": {
					Type:     schema.String,
					Required: true,
				},
				"replan_hint": {
					Type: schema.Object,
				},
			}),
		},
	}
}
