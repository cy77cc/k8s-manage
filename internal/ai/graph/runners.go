package graph

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/experts"
)

type registryPrimaryRunner struct {
	registry experts.ExpertRegistry
}

func NewRegistryPrimaryRunner(registry experts.ExpertRegistry) PrimaryRunner {
	return &registryPrimaryRunner{registry: registry}
}

func (r *registryPrimaryRunner) RunPrimary(ctx context.Context, req *experts.ExecuteRequest) (string, error) {
	if req == nil || req.Decision == nil {
		return "", fmt.Errorf("request decision is empty")
	}
	if r == nil || r.registry == nil {
		return "", fmt.Errorf("expert registry is nil")
	}
	exp, ok := r.registry.GetExpert(req.Decision.PrimaryExpert)
	if !ok || exp == nil {
		return "", fmt.Errorf("primary expert not found: %s", req.Decision.PrimaryExpert)
	}
	if exp.Agent == nil {
		return fmt.Sprintf("[%s] %s", req.Decision.PrimaryExpert, req.Message), nil
	}
	resp, err := exp.Agent.Generate(ctx, buildMessagesWithHistory(req.History, req.Message), exp.AgentOptions...)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", nil
	}
	return resp.Content, nil
}

type registryHelperRunner struct {
	registry experts.ExpertRegistry
}

func NewRegistryHelperRunner(registry experts.ExpertRegistry) HelperRunner {
	return &registryHelperRunner{registry: registry}
}

func (r *registryHelperRunner) RunHelper(ctx context.Context, req *experts.ExecuteRequest, helper experts.HelperRequest) (experts.ExpertResult, error) {
	if r == nil || r.registry == nil {
		return experts.ExpertResult{}, fmt.Errorf("expert registry is nil")
	}
	exp, ok := r.registry.GetExpert(helper.ExpertName)
	if !ok || exp == nil {
		return experts.ExpertResult{}, fmt.Errorf("helper expert not found: %s", helper.ExpertName)
	}
	if exp.Agent == nil {
		return experts.ExpertResult{ExpertName: helper.ExpertName, Output: fmt.Sprintf("[%s] %s", helper.ExpertName, helper.Task)}, nil
	}
	taskPrompt := fmt.Sprintf("用户原始请求: %s\n\n你的任务: %s\n\n请执行分析，输出结果供主专家汇总。", req.Message, helper.Task)
	resp, err := exp.Agent.Generate(ctx, buildMessagesWithHistory(req.History, taskPrompt), exp.AgentOptions...)
	if err != nil {
		return experts.ExpertResult{ExpertName: helper.ExpertName, Error: err}, err
	}
	out := ""
	if resp != nil {
		out = resp.Content
	}
	return experts.ExpertResult{ExpertName: helper.ExpertName, Output: out}, nil
}

type registryStreamPrimaryRunner struct {
	registry experts.ExpertRegistry
}

func NewRegistryStreamPrimaryRunner(registry experts.ExpertRegistry) StreamPrimaryRunner {
	return &registryStreamPrimaryRunner{registry: registry}
}

func (r *registryStreamPrimaryRunner) RunPrimaryStream(ctx context.Context, req *experts.ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
	if req == nil || req.Decision == nil {
		return nil, fmt.Errorf("request decision is empty")
	}
	if r == nil || r.registry == nil {
		return nil, fmt.Errorf("expert registry is nil")
	}
	exp, ok := r.registry.GetExpert(req.Decision.PrimaryExpert)
	if !ok || exp == nil {
		return nil, fmt.Errorf("primary expert not found: %s", req.Decision.PrimaryExpert)
	}
	if exp.Agent == nil {
		sr, sw := schema.Pipe[*schema.Message](1)
		sw.Send(schema.AssistantMessage(fmt.Sprintf("[%s] %s", req.Decision.PrimaryExpert, req.Message), nil), nil)
		sw.Close()
		return sr, nil
	}
	return exp.Agent.Stream(ctx, buildMessagesWithHistory(req.History, req.Message), exp.AgentOptions...)
}

type registryStreamHelperRunner struct {
	registry experts.ExpertRegistry
}

func NewRegistryStreamHelperRunner(registry experts.ExpertRegistry) StreamHelperRunner {
	return &registryStreamHelperRunner{registry: registry}
}

func (r *registryStreamHelperRunner) RunHelperStream(ctx context.Context, req *experts.ExecuteRequest, helper experts.HelperRequest) (*schema.StreamReader[*schema.Message], error) {
	if r == nil || r.registry == nil {
		return nil, fmt.Errorf("expert registry is nil")
	}
	exp, ok := r.registry.GetExpert(helper.ExpertName)
	if !ok || exp == nil {
		return nil, fmt.Errorf("helper expert not found: %s", helper.ExpertName)
	}
	if exp.Agent == nil {
		sr, sw := schema.Pipe[*schema.Message](1)
		sw.Send(schema.AssistantMessage(fmt.Sprintf("[%s] %s", helper.ExpertName, helper.Task), nil), nil)
		sw.Close()
		return sr, nil
	}
	taskPrompt := fmt.Sprintf("用户原始请求: %s\n\n你的任务: %s\n\n请执行分析，输出结果供主专家汇总。", req.Message, helper.Task)
	return exp.Agent.Stream(ctx, buildMessagesWithHistory(req.History, taskPrompt), exp.AgentOptions...)
}
