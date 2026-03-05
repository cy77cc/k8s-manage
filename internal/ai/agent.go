package ai

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

func newADKPlanExecuteAgent(ctx context.Context, chatModel model.ToolCallingChatModel, allTools []tool.BaseTool) (adk.Agent, error) {
	if chatModel == nil {
		return nil, fmt.Errorf("chat model is nil")
	}

	planner, err := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: chatModel,
	})
	if err != nil {
		return nil, fmt.Errorf("create planner: %w", err)
	}

	executor, err := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model: chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: allTools,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create executor: %w", err)
	}

	replanner, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: chatModel,
	})
	if err != nil {
		return nil, fmt.Errorf("create replanner: %w", err)
	}

	agent, err := planexecute.New(ctx, &planexecute.Config{
		Planner:       planner,
		Executor:      executor,
		Replanner:     replanner,
		MaxIterations: 20,
	})
	if err != nil {
		return nil, fmt.Errorf("create planexecute agent: %w", err)
	}
	return agent, nil
}
