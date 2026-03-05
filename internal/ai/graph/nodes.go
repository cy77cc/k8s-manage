package graph

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/experts"
)

// PrimaryRunner executes the primary expert step.
type PrimaryRunner interface {
	RunPrimary(ctx context.Context, req *experts.ExecuteRequest) (string, error)
}

// HelperRunner executes helper expert steps.
type HelperRunner interface {
	RunHelper(ctx context.Context, req *experts.ExecuteRequest, helper experts.HelperRequest) (experts.ExpertResult, error)
}

// StreamPrimaryRunner executes a primary step and returns a stream.
type StreamPrimaryRunner interface {
	RunPrimaryStream(ctx context.Context, req *experts.ExecuteRequest) (*schema.StreamReader[*schema.Message], error)
}

// StreamHelperRunner executes a helper step and returns a stream.
type StreamHelperRunner interface {
	RunHelperStream(ctx context.Context, req *experts.ExecuteRequest, helper experts.HelperRequest) (*schema.StreamReader[*schema.Message], error)
}

func runPrimary(ctx context.Context, runner PrimaryRunner, in *GraphInput) (*GraphInput, error) {
	if in == nil {
		return nil, fmt.Errorf("graph input is nil")
	}
	if strings.TrimSpace(in.PrimaryOutput) != "" {
		return in, nil
	}
	if runner != nil && in.Request != nil {
		out, err := runner.RunPrimary(ctx, in.Request)
		if err != nil {
			return nil, err
		}
		in.PrimaryOutput = strings.TrimSpace(out)
	}
	if strings.TrimSpace(in.PrimaryOutput) == "" {
		in.PrimaryOutput = strings.TrimSpace(in.Message)
	}
	if strings.TrimSpace(in.PrimaryOutput) == "" {
		in.PrimaryOutput = "已接收请求，等待专家执行。"
	}
	return in, nil
}

func runHelpersParallel(ctx context.Context, runner HelperRunner, in *GraphInput) (*GraphInput, error) {
	if in == nil {
		return nil, fmt.Errorf("graph input is nil")
	}
	if runner == nil || in.Request == nil || len(in.HelperRequests) == 0 {
		return in, nil
	}
	results := make([]experts.ExpertResult, len(in.HelperRequests))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i := range in.HelperRequests {
		i := i
		h := in.HelperRequests[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := runner.RunHelper(ctx, in.Request, h)
			if err != nil {
				res = experts.ExpertResult{ExpertName: h.ExpertName, Error: err}
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
			results[i] = res
		}()
	}
	wg.Wait()
	in.HelperResults = append(in.HelperResults, results...)
	return in, firstErr
}

func runHelpersSequential(ctx context.Context, runner HelperRunner, in *GraphInput) (*GraphInput, error) {
	if in == nil {
		return nil, fmt.Errorf("graph input is nil")
	}
	if runner == nil || in.Request == nil || len(in.HelperRequests) == 0 {
		return in, nil
	}
	for _, h := range in.HelperRequests {
		res, err := runner.RunHelper(ctx, in.Request, h)
		if err != nil {
			in.HelperResults = append(in.HelperResults, experts.ExpertResult{ExpertName: h.ExpertName, Error: err})
			return in, err
		}
		in.HelperResults = append(in.HelperResults, res)
	}
	return in, nil
}

func aggregateResults(in *GraphInput) (*GraphOutput, error) {
	if in == nil {
		return nil, fmt.Errorf("graph input is nil")
	}
	out := &GraphOutput{
		Response: strings.TrimSpace(in.PrimaryOutput),
		Results:  append([]experts.ExpertResult(nil), in.HelperResults...),
		Metadata: map[string]any{"strategy": in.Strategy},
	}
	if out.Response == "" {
		out.Response = "未获得主专家输出，请补充上下文后重试。"
	}
	return out, nil
}

func runPrimaryStream(ctx context.Context, streamRunner StreamPrimaryRunner, runner PrimaryRunner, in *GraphInput) (*GraphInput, error) {
	if in == nil {
		return nil, fmt.Errorf("graph input is nil")
	}
	if strings.TrimSpace(in.PrimaryStream) != "" {
		return in, nil
	}
	var content string
	if streamRunner != nil && in.Request != nil {
		stream, err := streamRunner.RunPrimaryStream(ctx, in.Request)
		if err != nil {
			return nil, err
		}
		content, err = collectStreamContent(stream)
		if err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(content) == "" && runner != nil && in.Request != nil {
		out, err := runner.RunPrimary(ctx, in.Request)
		if err != nil {
			return nil, err
		}
		content = out
	}
	content = strings.TrimSpace(content)
	if content == "" {
		content = strings.TrimSpace(in.Message)
	}
	if content == "" {
		content = "已接收请求，等待专家执行。"
	}
	in.PrimaryStream = content
	return in, nil
}

func runHelpersParallelStream(ctx context.Context, streamRunner StreamHelperRunner, runner HelperRunner, in *GraphInput) (*GraphInput, error) {
	if in == nil {
		return nil, fmt.Errorf("graph input is nil")
	}
	if len(in.HelperRequests) == 0 {
		return in, nil
	}
	results := make([]string, len(in.HelperRequests))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	for i := range in.HelperRequests {
		i := i
		h := in.HelperRequests[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			text, err := runSingleHelperStream(ctx, streamRunner, runner, in.Request, h)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				results[i] = fmt.Sprintf("%s: %v", h.ExpertName, err)
				return
			}
			results[i] = strings.TrimSpace(text)
		}()
	}
	wg.Wait()
	in.HelperStreams = append(in.HelperStreams, results...)
	return in, firstErr
}

func runHelpersSequentialStream(ctx context.Context, streamRunner StreamHelperRunner, runner HelperRunner, in *GraphInput) (*GraphInput, error) {
	if in == nil {
		return nil, fmt.Errorf("graph input is nil")
	}
	if len(in.HelperRequests) == 0 {
		return in, nil
	}
	for _, h := range in.HelperRequests {
		text, err := runSingleHelperStream(ctx, streamRunner, runner, in.Request, h)
		if err != nil {
			in.HelperStreams = append(in.HelperStreams, fmt.Sprintf("%s: %v", h.ExpertName, err))
			return in, err
		}
		in.HelperStreams = append(in.HelperStreams, strings.TrimSpace(text))
	}
	return in, nil
}

func aggregateStreamResults(in *GraphInput) (*schema.StreamReader[*schema.Message], error) {
	if in == nil {
		return nil, fmt.Errorf("graph input is nil")
	}
	parts := make([]string, 0, 1+len(in.HelperStreams))
	if strings.TrimSpace(in.PrimaryStream) != "" {
		parts = append(parts, strings.TrimSpace(in.PrimaryStream))
	}
	for _, item := range in.HelperStreams {
		msg := strings.TrimSpace(item)
		if msg == "" {
			continue
		}
		parts = append(parts, msg)
	}
	if len(parts) == 0 {
		parts = append(parts, "未获得主专家输出，请补充上下文后重试。")
	}
	sr, sw := schema.Pipe[*schema.Message](len(parts))
	for i := range parts {
		sw.Send(schema.AssistantMessage(parts[i], nil), nil)
	}
	sw.Close()
	return sr, nil
}

func runSingleHelperStream(ctx context.Context, streamRunner StreamHelperRunner, runner HelperRunner, req *experts.ExecuteRequest, helper experts.HelperRequest) (string, error) {
	if streamRunner != nil && req != nil {
		stream, err := streamRunner.RunHelperStream(ctx, req, helper)
		if err == nil && stream != nil {
			return collectStreamContent(stream)
		}
		if err != nil {
			return "", err
		}
	}
	if runner != nil && req != nil {
		res, err := runner.RunHelper(ctx, req, helper)
		if err != nil {
			return "", err
		}
		return res.Output, nil
	}
	return "", nil
}

func collectStreamContent(stream *schema.StreamReader[*schema.Message]) (string, error) {
	if stream == nil {
		return "", nil
	}
	defer stream.Close()
	var out strings.Builder
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if msg != nil {
			out.WriteString(msg.Content)
		}
	}
	return out.String(), nil
}
