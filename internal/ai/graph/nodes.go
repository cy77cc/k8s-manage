package graph

import (
	"context"
	"fmt"
	"strings"
	"sync"

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
