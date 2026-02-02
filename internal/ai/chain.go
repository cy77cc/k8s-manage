package ai

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"k8s.io/client-go/kubernetes"
)

type K8sCopilot struct {
	Runnable *react.Agent
}

func NewK8sCopilot(ctx context.Context, cm model.ToolCallingChatModel, clientset *kubernetes.Clientset) (*K8sCopilot, error) {
	if cm == nil {
		return nil, nil
	}

	// 1. Initialize Tools
	tools, err := NewK8sTools(clientset)
	if err != nil {
		return nil, err
	}

	// 2. Create ReAct Agent
	// The ReAct agent will use the ChatModel and Tools to reason and act.
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools,
		},
		MaxStep: 10, // Limit the number of reasoning steps
	})
	if err != nil {
		return nil, err
	}

	return &K8sCopilot{Runnable: agent}, nil
}
