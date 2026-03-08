package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	einomodel "github.com/cloudwego/eino/components/model"
)

func NewPlanner(ctx context.Context, chatModel einomodel.ToolCallingChatModel) (adk.Agent, error) {
	if chatModel == nil {
		return nil, fmt.Errorf("chat model is nil")
	}

	return planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: chatModel,
	})
}

func NewReplanAgent(ctx context.Context, chatModel einomodel.ToolCallingChatModel) (adk.Agent, error) {
	if chatModel == nil {
		return nil, fmt.Errorf("chat model is nil")
	}
	return planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: chatModel,
	})
}
