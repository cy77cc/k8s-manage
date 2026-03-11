package availability

type Layer string

const (
	LayerRewrite    Layer = "rewrite"
	LayerPlanner    Layer = "planner"
	LayerExpert     Layer = "expert"
	LayerSummarizer Layer = "summarizer"
)

func UnavailableMessage(layer Layer) string {
	switch layer {
	case LayerRewrite:
		return "AI 理解模块当前不可用，请稍后重试或手动在页面中执行操作。"
	case LayerPlanner:
		return "AI 规划模块当前不可用，请稍后重试或手动在页面中执行操作。"
	case LayerExpert:
		return "AI 执行专家当前不可用，请稍后重试或手动在页面中执行操作。"
	case LayerSummarizer:
		return "AI 总结模块当前不可用，你可以直接查看原始执行结果。"
	default:
		return "AI 模块当前不可用，请稍后重试。"
	}
}

func InvalidOutputMessage(layer Layer) string {
	switch layer {
	case LayerRewrite:
		return "AI 理解模块返回了无效结果，请稍后重试或手动在页面中执行操作。"
	case LayerPlanner:
		return "AI 规划模块返回了无效结果，请稍后重试或手动在页面中执行操作。"
	case LayerExpert:
		return "AI 执行专家返回了无效结果，请稍后重试或手动在页面中执行操作。"
	case LayerSummarizer:
		return "AI 总结模块返回了无效结果，你可以直接查看原始执行结果。"
	default:
		return "AI 模块返回了无效结果，请稍后重试。"
	}
}
