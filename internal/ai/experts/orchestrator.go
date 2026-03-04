package experts

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

var helperRequestPattern = regexp.MustCompile(`\[REQUEST_HELPER:\s*([a-zA-Z0-9_]+):\s*([^\]]+)\]`)

type Orchestrator struct {
	registry   ExpertRegistry
	executor   *ExpertExecutor
	aggregator *ResultAggregator
}

func NewOrchestrator(registry ExpertRegistry, aggregator *ResultAggregator) *Orchestrator {
	if aggregator == nil {
		aggregator = NewResultAggregator(AggregationTemplate, nil)
	}
	return &Orchestrator{
		registry:   registry,
		executor:   NewExpertExecutor(registry),
		aggregator: aggregator,
	}
}

func (o *Orchestrator) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, error) {
	if req == nil || req.Decision == nil {
		return nil, fmt.Errorf("route decision is required")
	}
	runCtx := ctx
	if timeoutMS, ok := runtimeTimeout(req.RuntimeContext); ok {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, timeoutMS)
		defer cancel()
	}

	if req.Decision.Strategy == StrategyPrimaryLed {
		if exp, ok := o.registry.GetExpert(req.Decision.PrimaryExpert); ok && exp != nil && exp.Agent != nil {
			stream, err := o.StreamExecute(runCtx, req)
			if err != nil {
				return nil, err
			}
			defer stream.Close()
			var out strings.Builder
			for {
				msg, recvErr := stream.Recv()
				if errors.Is(recvErr, io.EOF) {
					break
				}
				if recvErr != nil {
					return nil, recvErr
				}
				if msg != nil {
					out.WriteString(msg.Content)
				}
			}
			return &ExecuteResult{
				Response: strings.TrimSpace(out.String()),
				Metadata: map[string]any{
					"strategy": req.Decision.Strategy,
					"source":   req.Decision.Source,
				},
			}, nil
		}
		// Fallback for tests/offline mode where experts are registered without live agents.
		if len(req.Decision.OptionalHelpers) > 0 {
			req.Decision.Strategy = StrategySequential
		} else {
			req.Decision.Strategy = StrategySingle
		}
	}

	plan := o.buildPlan(req.Decision)
	results, err := o.executePlan(runCtx, plan, req)
	resp, aggErr := o.aggregateResults(runCtx, results, req)
	if aggErr != nil && err == nil {
		err = aggErr
	}
	traces := make([]ExpertTrace, 0, len(results))
	for _, item := range results {
		status := "success"
		if item.Error != nil {
			status = "failed"
		}
		role := "helper"
		if item.ExpertName == req.Decision.PrimaryExpert {
			role = "primary"
		}
		traces = append(traces, ExpertTrace{
			ExpertName: item.ExpertName,
			Role:       role,
			Output:     item.Output,
			Duration:   item.Duration,
			Status:     status,
		})
	}
	return &ExecuteResult{
		Response: resp,
		Traces:   traces,
		Metadata: map[string]any{
			"strategy": req.Decision.Strategy,
			"source":   req.Decision.Source,
		},
	}, err
}

func (o *Orchestrator) StreamExecute(ctx context.Context, req *ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
	if req == nil || req.Decision == nil {
		return nil, fmt.Errorf("route decision is required")
	}
	if o == nil || o.executor == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	runCtx := ctx
	if timeoutMS, ok := runtimeTimeout(req.RuntimeContext); ok {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, timeoutMS)
		defer cancel()
	}

	if req.Decision.Strategy == StrategySingle || len(req.Decision.OptionalHelpers) == 0 {
		return o.streamSingleExpert(runCtx, req)
	}
	if req.Decision.Strategy == StrategyPrimaryLed {
		return o.streamPrimaryLed(runCtx, req)
	}

	plan := o.buildPlan(req.Decision)
	if plan == nil || len(plan.Steps) == 0 {
		return nil, fmt.Errorf("execution plan is empty")
	}

	sr, sw := schema.Pipe[*schema.Message](64)
	go func() {
		defer sw.Close()
		results := make([]ExpertResult, 0, len(plan.Steps))
		for i := range plan.Steps {
			step := plan.Steps[i]
			stream, err := o.executor.StreamStep(runCtx, &step, req)
			if err != nil {
				sw.Send(schema.AssistantMessage(fmt.Sprintf("专家 %s 执行失败: %v", step.ExpertName, err), nil), nil)
				return
			}
			var content string
			for {
				msg, recvErr := stream.Recv()
				if errors.Is(recvErr, io.EOF) {
					break
				}
				if recvErr != nil {
					sw.Send(schema.AssistantMessage(fmt.Sprintf("专家 %s 流式读取失败: %v", step.ExpertName, recvErr), nil), nil)
					break
				}
				if msg == nil {
					continue
				}
				content += msg.Content
				sw.Send(msg, nil)
			}
			stream.Close()
			results = append(results, ExpertResult{ExpertName: step.ExpertName, Output: content})
		}
		if len(results) > 1 {
			merged, err := o.aggregateResults(runCtx, results, req)
			if err == nil && merged != "" {
				sw.Send(schema.AssistantMessage("\n\n---\n综合分析:\n"+merged, nil), nil)
			}
		}
	}()
	return sr, nil
}

func (o *Orchestrator) streamSingleExpert(ctx context.Context, req *ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
	exp, ok := o.registry.GetExpert(req.Decision.PrimaryExpert)
	if !ok || exp == nil {
		return nil, fmt.Errorf("expert not found: %s", req.Decision.PrimaryExpert)
	}
	if exp.Agent == nil {
		return o.executor.StreamStep(ctx, &ExecutionStep{
			ExpertName: req.Decision.PrimaryExpert,
			Task:       "primary analysis",
		}, req)
	}
	messages := o.buildMessagesWithHistory(req.History, req.Message)
	return exp.Agent.Stream(ctx, messages)
}

func (o *Orchestrator) streamPrimaryLed(ctx context.Context, req *ExecuteRequest) (*schema.StreamReader[*schema.Message], error) {
	sr, sw := schema.Pipe[*schema.Message](64)
	go func() {
		defer sw.Close()
		decision, err := o.primaryDecisionPhase(ctx, req)
		if err != nil {
			sw.Send(schema.AssistantMessage(fmt.Sprintf("决策阶段失败: %v", err), nil), nil)
			return
		}
		if !decision.NeedHelpers {
			o.streamPrimaryAnswer(ctx, req, sw)
			return
		}
		helperResults, _ := o.helperExecutionPhase(ctx, req, decision.HelperRequests)
		o.primarySummaryPhase(ctx, req, helperResults, sw)
	}()
	return sr, nil
}

func (o *Orchestrator) primaryDecisionPhase(ctx context.Context, req *ExecuteRequest) (*PrimaryDecision, error) {
	exp, ok := o.registry.GetExpert(req.Decision.PrimaryExpert)
	if !ok || exp == nil || exp.Agent == nil {
		return nil, fmt.Errorf("primary expert not found: %s", req.Decision.PrimaryExpert)
	}
	decisionPrompt := o.buildDecisionPrompt(req)
	messages := o.buildMessagesWithHistory(req.History, decisionPrompt)
	resp, err := exp.Agent.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}
	content := ""
	if resp != nil {
		content = resp.Content
	}
	return o.parsePrimaryDecision(content, req.Decision.OptionalHelpers), nil
}

func (o *Orchestrator) streamPrimaryAnswer(ctx context.Context, req *ExecuteRequest, sw *schema.StreamWriter[*schema.Message]) {
	exp, ok := o.registry.GetExpert(req.Decision.PrimaryExpert)
	if !ok || exp == nil || exp.Agent == nil {
		sw.Send(schema.AssistantMessage("主专家不可用", nil), nil)
		return
	}
	stream, err := exp.Agent.Stream(ctx, o.buildMessagesWithHistory(req.History, req.Message))
	if err != nil {
		sw.Send(schema.AssistantMessage(fmt.Sprintf("主专家输出失败: %v", err), nil), nil)
		return
	}
	defer stream.Close()
	for {
		msg, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			break
		}
		if msg != nil {
			sw.Send(msg, nil)
		}
	}
}

func (o *Orchestrator) helperExecutionPhase(ctx context.Context, req *ExecuteRequest, helperRequests []HelperRequest) ([]ExpertResult, error) {
	if len(helperRequests) == 0 {
		return nil, nil
	}
	results := make([]ExpertResult, 0, len(helperRequests))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, hr := range helperRequests {
		wg.Add(1)
		go func(helperReq HelperRequest) {
			defer wg.Done()
			o.emitProgress(req.EventEmitter, "expert_progress", ExpertProgressEvent{
				Expert: helperReq.ExpertName,
				Status: "running",
				Task:   helperReq.Task,
			})
			start := time.Now()
			result, err := o.executeHelper(ctx, req, helperReq)
			duration := time.Since(start)
			o.emitProgress(req.EventEmitter, "expert_progress", ExpertProgressEvent{
				Expert:     helperReq.ExpertName,
				Status:     "done",
				DurationMs: duration.Milliseconds(),
			})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results = append(results, ExpertResult{
					ExpertName: helperReq.ExpertName,
					Error:      err,
					Duration:   duration,
				})
				return
			}
			results = append(results, *result)
		}(hr)
	}
	wg.Wait()
	return results, nil
}

func (o *Orchestrator) executeHelper(ctx context.Context, req *ExecuteRequest, helperReq HelperRequest) (*ExpertResult, error) {
	exp, ok := o.registry.GetExpert(helperReq.ExpertName)
	if !ok || exp == nil || exp.Agent == nil {
		return nil, fmt.Errorf("helper expert not found: %s", helperReq.ExpertName)
	}
	start := time.Now()
	taskPrompt := fmt.Sprintf("用户原始请求: %s\n\n你的任务: %s\n\n请执行分析，输出结果供主专家汇总。", req.Message, helperReq.Task)
	messages := o.buildMessagesWithHistory(req.History, taskPrompt)
	resp, err := exp.Agent.Generate(ctx, messages)
	if err != nil {
		return &ExpertResult{ExpertName: helperReq.ExpertName, Error: err, Duration: time.Since(start)}, err
	}
	output := ""
	if resp != nil {
		output = resp.Content
	}
	return &ExpertResult{ExpertName: helperReq.ExpertName, Output: output, Duration: time.Since(start)}, nil
}

func (o *Orchestrator) primarySummaryPhase(ctx context.Context, req *ExecuteRequest, helperResults []ExpertResult, sw *schema.StreamWriter[*schema.Message]) {
	exp, ok := o.registry.GetExpert(req.Decision.PrimaryExpert)
	if !ok || exp == nil || exp.Agent == nil {
		sw.Send(schema.AssistantMessage("主专家不可用", nil), nil)
		return
	}
	summaryPrompt := o.buildSummaryPrompt(req, helperResults)
	messages := o.buildMessagesWithHistory(req.History, summaryPrompt)
	stream, err := exp.Agent.Stream(ctx, messages)
	if err != nil {
		sw.Send(schema.AssistantMessage(fmt.Sprintf("汇总失败: %v", err), nil), nil)
		return
	}
	defer stream.Close()
	for {
		msg, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			break
		}
		if msg != nil {
			sw.Send(msg, nil)
		}
	}
}

func (o *Orchestrator) buildDecisionPrompt(req *ExecuteRequest) string {
	var b strings.Builder
	b.WriteString("用户请求: ")
	b.WriteString(req.Message)
	b.WriteString("\n\n")
	if len(req.Decision.OptionalHelpers) == 0 {
		b.WriteString("如果可以直接回答，请直接回答。")
		return b.String()
	}
	b.WriteString("你可以请求以下助手协助分析（仅在必要时调用）：\n")
	for _, helper := range req.Decision.OptionalHelpers {
		b.WriteString("- ")
		b.WriteString(helper)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString("如果需要助手，请输出：[REQUEST_HELPER: 助手名称: 任务描述]\n")
	b.WriteString("如果不需要助手，请输出: [NO_HELPER]\n")
	return b.String()
}

func (o *Orchestrator) buildSummaryPrompt(req *ExecuteRequest, helperResults []ExpertResult) string {
	var b strings.Builder
	b.WriteString("用户请求: ")
	b.WriteString(req.Message)
	b.WriteString("\n\n")
	if len(helperResults) > 0 {
		b.WriteString("助手分析结果：\n")
		for _, result := range helperResults {
			b.WriteString("\n【")
			b.WriteString(result.ExpertName)
			b.WriteString("】\n")
			if result.Error != nil {
				b.WriteString("执行失败: ")
				b.WriteString(result.Error.Error())
			} else {
				b.WriteString(result.Output)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n请基于以上分析结果，给用户一个完整、连贯的回答。\n")
	}
	return b.String()
}

func (o *Orchestrator) parsePrimaryDecision(content string, availableHelpers []string) *PrimaryDecision {
	decision := &PrimaryDecision{NeedHelpers: false}
	matches := helperRequestPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		expertName := strings.TrimSpace(match[1])
		task := strings.TrimSpace(match[2])
		for _, helper := range availableHelpers {
			if helper == expertName {
				decision.NeedHelpers = true
				decision.HelperRequests = append(decision.HelperRequests, HelperRequest{ExpertName: expertName, Task: task})
				break
			}
		}
	}
	if !decision.NeedHelpers {
		decision.DirectAnswer = strings.TrimSpace(content)
	}
	return decision
}

func (o *Orchestrator) buildMessagesWithHistory(history []*schema.Message, currentMessage string) []*schema.Message {
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
	messages = append(messages, schema.UserMessage(currentMessage))
	return messages
}

func (o *Orchestrator) emitProgress(emitter ProgressEmitter, event string, payload any) {
	if emitter != nil {
		emitter(event, payload)
	}
}

func (o *Orchestrator) buildPlan(decision *RouteDecision) *ExecutionPlan {
	steps := make([]ExecutionStep, 0, 1+len(decision.OptionalHelpers))
	steps = append(steps, ExecutionStep{ExpertName: decision.PrimaryExpert, Task: "primary analysis"})
	switch decision.Strategy {
	case StrategyParallel:
		for _, helper := range decision.OptionalHelpers {
			steps = append(steps, ExecutionStep{ExpertName: helper, Task: "helper analysis"})
		}
	case StrategySequential:
		for i, helper := range decision.OptionalHelpers {
			steps = append(steps, ExecutionStep{ExpertName: helper, Task: "helper analysis", DependsOn: []int{i}, ContextFrom: []int{i}})
		}
	}
	return &ExecutionPlan{Steps: steps}
}

func (o *Orchestrator) executePlan(ctx context.Context, plan *ExecutionPlan, req *ExecuteRequest) ([]ExpertResult, error) {
	if o == nil || o.executor == nil {
		return nil, fmt.Errorf("executor is nil")
	}
	if plan == nil || len(plan.Steps) == 0 {
		return nil, fmt.Errorf("execution plan is empty")
	}
	if req == nil || req.Decision == nil {
		return nil, fmt.Errorf("request decision is empty")
	}

	results := make([]ExpertResult, 0, len(plan.Steps))
	switch req.Decision.Strategy {
	case StrategyParallel:
		type resultWithIndex struct {
			idx int
			val ExpertResult
		}
		ch := make(chan resultWithIndex, len(plan.Steps))
		var wg sync.WaitGroup
		for idx := range plan.Steps {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				step := plan.Steps[i]
				res, err := o.executor.ExecuteStep(ctx, &step, req, nil)
				if res == nil {
					res = &ExpertResult{ExpertName: step.ExpertName, Error: err}
				}
				if err != nil {
					res.Error = err
				}
				ch <- resultWithIndex{idx: i, val: *res}
			}(idx)
		}
		wg.Wait()
		close(ch)
		ordered := make([]ExpertResult, len(plan.Steps))
		for item := range ch {
			ordered[item.idx] = item.val
		}
		results = append(results, ordered...)
		return results, nil
	default:
		var firstErr error
		for i := range plan.Steps {
			step := plan.Steps[i]
			res, err := o.executor.ExecuteStep(ctx, &step, req, results)
			if res == nil {
				res = &ExpertResult{ExpertName: step.ExpertName, Error: err}
			}
			if err != nil && firstErr == nil {
				firstErr = err
				res.Error = err
			}
			results = append(results, *res)
		}
		return results, firstErr
	}
}

func (o *Orchestrator) aggregateResults(ctx context.Context, results []ExpertResult, req *ExecuteRequest) (string, error) {
	if o == nil || o.aggregator == nil {
		return "", fmt.Errorf("aggregator is nil")
	}
	query := ""
	if req != nil {
		query = req.Message
	}
	return o.aggregator.Aggregate(ctx, results, query)
}

func runtimeTimeout(runtime map[string]any) (time.Duration, bool) {
	if len(runtime) == 0 {
		return 0, false
	}
	raw, ok := runtime["timeout_ms"]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case int:
		if v > 0 {
			return time.Duration(v) * time.Millisecond, true
		}
	case int64:
		if v > 0 {
			return time.Duration(v) * time.Millisecond, true
		}
	case float64:
		if v > 0 {
			return time.Duration(v) * time.Millisecond, true
		}
	}
	return 0, false
}
