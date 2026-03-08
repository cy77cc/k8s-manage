package planner

import (
	"fmt"
	"strings"

	"github.com/cy77cc/k8s-manage/internal/ai/orchestrator"
	"github.com/cy77cc/k8s-manage/internal/ai/types"
)

func BuildPrompt(domain types.Domain, req orchestrator.DomainRequest) string {
	return strings.TrimSpace(fmt.Sprintf(`你是 %s 领域规划器。请只输出 DomainPlan JSON。
必须只规划，不执行工具；可用 Discovery 工具补全参数；必须声明 depends_on / produces / requires。

用户意图：%s
上下文：%v`, domain, strings.TrimSpace(req.UserIntent), req.Context))
}
