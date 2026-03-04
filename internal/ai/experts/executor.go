package experts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

type ExpertExecutor struct {
	registry ExpertRegistry
}

func NewExpertExecutor(registry ExpertRegistry) *ExpertExecutor {
	return &ExpertExecutor{registry: registry}
}

func (e *ExpertExecutor) ExecuteStep(ctx context.Context, step *ExecutionStep, req *ExecuteRequest, priorResults []ExpertResult) (*ExpertResult, error) {
	if step == nil {
		return nil, fmt.Errorf("execution step is nil")
	}
	if e == nil || e.registry == nil {
		return nil, fmt.Errorf("expert registry is nil")
	}
	exp, ok := e.registry.GetExpert(step.ExpertName)
	if !ok || exp == nil {
		return nil, fmt.Errorf("expert not found: %s", step.ExpertName)
	}
	var history []*schema.Message
	baseMessage := ""
	if req != nil {
		history = req.History
		baseMessage = req.Message
	}
	messages := e.buildExpertMessages(history, step, baseMessage, priorResults)
	start := time.Now()
	if exp.Agent == nil {
		fallback := e.fallbackMessage(messages)
		return &ExpertResult{
			ExpertName: exp.Name,
			Output:     "专家模型未初始化，返回静态诊断建议：" + fallback,
			Duration:   time.Since(start),
		}, nil
	}
	resp, err := exp.Agent.Generate(ctx, messages)
	if err != nil {
		return &ExpertResult{
			ExpertName: exp.Name,
			Error:      err,
			Duration:   time.Since(start),
		}, err
	}
	output := ""
	if resp != nil {
		output = strings.TrimSpace(resp.Content)
	}
	return &ExpertResult{
		ExpertName: exp.Name,
		Output:     output,
		Duration:   time.Since(start),
	}, nil
}

func (e *ExpertExecutor) StreamStep(ctx context.Context, step *ExecutionStep, req *ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
	if step == nil {
		return nil, fmt.Errorf("execution step is nil")
	}
	if e == nil || e.registry == nil {
		return nil, fmt.Errorf("expert registry is nil")
	}
	exp, ok := e.registry.GetExpert(step.ExpertName)
	if !ok || exp == nil {
		return nil, fmt.Errorf("expert not found: %s", step.ExpertName)
	}
	var history []*schema.Message
	baseMessage := ""
	if req != nil {
		history = req.History
		baseMessage = req.Message
	}
	messages := e.buildExpertMessages(history, step, baseMessage, nil)
	if exp.Agent == nil {
		return schema.StreamReaderFromArray([]*schema.Message{
			schema.AssistantMessage("专家模型未初始化，返回静态诊断建议："+e.fallbackMessage(messages), nil),
		}), nil
	}
	return exp.Agent.Stream(ctx, messages)
}

func (e *ExpertExecutor) buildExpertMessages(history []*schema.Message, step *ExecutionStep, baseMessage string, priorResults []ExpertResult) []*schema.Message {
	messages := make([]*schema.Message, 0, len(history)+1)
	maxHistory := 10
	start := 0
	if len(history) > maxHistory {
		start = len(history) - maxHistory
	}
	for i := start; i < len(history); i++ {
		if history[i] != nil {
			messages = append(messages, history[i])
		}
	}

	var taskMsg strings.Builder
	base := strings.TrimSpace(baseMessage)
	if base != "" {
		taskMsg.WriteString("用户请求:\n")
		taskMsg.WriteString(base)
		taskMsg.WriteString("\n")
	}
	task := strings.TrimSpace(step.Task)
	if task != "" {
		taskMsg.WriteString("\n当前任务:\n")
		taskMsg.WriteString(task)
		taskMsg.WriteString("\n")
	}
	if len(step.ContextFrom) > 0 && len(priorResults) > 0 {
		taskMsg.WriteString("\n上游专家结果:\n")
		for _, idx := range step.ContextFrom {
			if idx < 0 || idx >= len(priorResults) {
				continue
			}
			item := priorResults[idx]
			taskMsg.WriteString("- ")
			taskMsg.WriteString(item.ExpertName)
			taskMsg.WriteString(": ")
			taskMsg.WriteString(strings.TrimSpace(item.Output))
			taskMsg.WriteString("\n")
		}
	}
	messages = append(messages, schema.UserMessage(strings.TrimSpace(taskMsg.String())))
	return messages
}

func (e *ExpertExecutor) fallbackMessage(messages []*schema.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i] != nil && strings.TrimSpace(messages[i].Content) != "" {
			return messages[i].Content
		}
	}
	return ""
}
