package replanner

import "github.com/cy77cc/k8s-manage/internal/ai/executor"

func HasFailures(result executor.ExecutionResult) bool {
	for _, step := range result.Results {
		if step.Error != "" {
			return true
		}
	}
	return false
}
