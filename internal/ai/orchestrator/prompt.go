package orchestrator

import (
	"fmt"
	"strings"
)

func BuildDomainSelectionPrompt(message string, descriptors []DomainDescriptor) string {
	parts := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		parts = append(parts, fmt.Sprintf("- %s: %s", descriptor.Domain, descriptor.Description))
	}
	return strings.TrimSpace(fmt.Sprintf(`你是多领域任务编排器。请识别用户请求涉及的领域，只返回 JSON 数组。
每个元素格式为 {"domain":"...","user_intent":"...","context":{}}。
若无法判断，返回 [{"domain":"general","context":{}}]。

可用领域：
%s

用户请求：%s`, strings.Join(parts, "\n"), strings.TrimSpace(message)))
}
