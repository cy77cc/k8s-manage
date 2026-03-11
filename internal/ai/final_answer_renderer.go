package ai

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cy77cc/OpsPilot/internal/ai/executor"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/summarizer"
)

type finalAnswerRenderer struct{}

func newFinalAnswerRenderer() *finalAnswerRenderer {
	return &finalAnswerRenderer{}
}

func (r *finalAnswerRenderer) Render(message string, plan *planner.ExecutionPlan, result *executor.Result, summaryOut summarizer.SummaryOutput) []string {
	if result == nil {
		return compactParagraphs([]string{
			firstNonEmpty(summaryOut.Headline, "本轮执行已结束"),
			firstNonEmpty(summaryOut.Conclusion, summaryOut.Summary),
		})
	}
	if looksLikeFleetHostStatus(plan, result) {
		return r.renderFleetHostStatus(summaryOut, result)
	}
	return r.renderGeneric(summaryOut, result)
}

func (r *finalAnswerRenderer) renderGeneric(summaryOut summarizer.SummaryOutput, result *executor.Result) []string {
	paragraphs := []string{firstNonEmpty(summaryOut.Headline, summaryOut.Summary)}
	findings := append([]string(nil), summaryOut.KeyFindings...)
	findings = append(findings, importantObservedFacts(result, 3)...)
	if len(findings) > 0 {
		paragraphs = append(paragraphs, "关键依据：\n- "+strings.Join(dedupeStrings(findings), "\n- "))
	}
	recommendations := summaryOut.Recommendations
	if len(recommendations) == 0 {
		recommendations = summaryOut.NextActions
	}
	if len(recommendations) > 0 {
		paragraphs = append(paragraphs, "建议：\n- "+strings.Join(dedupeStrings(recommendations), "\n- "))
	}
	return compactParagraphs(paragraphs)
}

func (r *finalAnswerRenderer) renderFleetHostStatus(summaryOut summarizer.SummaryOutput, result *executor.Result) []string {
	items := extractHostInventoryEntries(result)
	if len(items) == 0 {
		return r.renderGeneric(summaryOut, result)
	}
	total := len(items)
	abnormal := 0
	for _, item := range items {
		if !isHealthyHost(item) {
			abnormal++
		}
	}

	headline := firstNonEmpty(summaryOut.Headline, "已完成主机状态汇总")
	if abnormal == 0 {
		headline = fmt.Sprintf("共检查 %d 台主机，当前均运行正常。", total)
	} else {
		headline = fmt.Sprintf("共检查 %d 台主机，其中 %d 台需要重点关注。", total, abnormal)
	}

	paragraphs := []string{headline}

	top := summarizeTopHosts(items, abnormal == 0)
	if len(top) > 0 {
		paragraphs = append(paragraphs, "关键依据：\n- "+strings.Join(top, "\n- "))
	}

	recommendations := summaryOut.Recommendations
	if abnormal == 0 {
		recommendations = filterRoutineRestartAdvice(recommendations)
		if len(recommendations) == 0 {
			recommendations = []string{"当前无需额外处理，继续保持常规巡检即可。"}
		}
	}
	if len(recommendations) == 0 {
		recommendations = summaryOut.NextActions
	}
	if len(recommendations) > 0 {
		paragraphs = append(paragraphs, "建议：\n- "+strings.Join(dedupeStrings(recommendations), "\n- "))
	}
	return compactParagraphs(paragraphs)
}

type hostInventoryEntry struct {
	ID       int
	Name     string
	IP       string
	Status   string
	CPUCores int
	MemoryMB int
	DiskGB   int
}

func looksLikeFleetHostStatus(plan *planner.ExecutionPlan, result *executor.Result) bool {
	if plan == nil || result == nil || len(result.Steps) == 0 {
		return false
	}
	scope := plan.Resolved.Scope
	if scope == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(scope.ResourceType), "host") &&
		strings.TrimSpace(scope.Kind) == "all"
}

func extractHostInventoryEntries(result *executor.Result) []hostInventoryEntry {
	if result == nil {
		return nil
	}
	var out []hostInventoryEntry
	for _, step := range result.Steps {
		for _, evidence := range step.Evidence {
			if evidence.Source != "hostops" && evidence.Source != "host_list_inventory" {
				continue
			}
			data := evidence.Data
			if data == nil {
				continue
			}
			list, ok := data["list"].([]map[string]any)
			if !ok {
				if raw, okAny := data["list"].([]any); okAny {
					list = make([]map[string]any, 0, len(raw))
					for _, item := range raw {
						if row, okRow := item.(map[string]any); okRow {
							list = append(list, row)
						}
					}
				}
			}
			for _, item := range list {
				out = append(out, hostInventoryEntry{
					ID:       looseInt(item["id"]),
					Name:     firstNonEmpty(asString(item["name"]), asString(item["hostname"]), asString(item["ip"])),
					IP:       asString(item["ip"]),
					Status:   asString(item["status"]),
					CPUCores: looseInt(item["cpu_cores"]),
					MemoryMB: looseInt(item["memory_mb"]),
					DiskGB:   looseInt(item["disk_gb"]),
				})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func isHealthyHost(item hostInventoryEntry) bool {
	status := strings.ToLower(strings.TrimSpace(item.Status))
	return status == "" || status == "online" || status == "healthy"
}

func summarizeTopHosts(items []hostInventoryEntry, allHealthy bool) []string {
	if len(items) == 0 {
		return nil
	}
	picked := items
	if allHealthy && len(picked) > 3 {
		picked = picked[:3]
	}
	lines := make([]string, 0, len(picked)+1)
	for _, item := range picked {
		line := firstNonEmpty(item.Name, item.IP)
		if line == "" {
			line = fmt.Sprintf("主机 %d", item.ID)
		}
		parts := []string{}
		if item.Status != "" {
			parts = append(parts, "状态 "+item.Status)
		}
		if item.CPUCores > 0 {
			parts = append(parts, fmt.Sprintf("CPU %d 核", item.CPUCores))
		}
		if item.MemoryMB > 0 {
			parts = append(parts, fmt.Sprintf("内存 %.1f GB", float64(item.MemoryMB)/1024.0))
		}
		if item.DiskGB > 0 {
			parts = append(parts, fmt.Sprintf("磁盘 %d GB", item.DiskGB))
		}
		if len(parts) > 0 {
			line += "：" + strings.Join(parts, "，")
		}
		lines = append(lines, line)
	}
	if allHealthy && len(items) > len(picked) {
		lines = append(lines, fmt.Sprintf("其余 %d 台主机状态一致，未见需要单独展开的异常。", len(items)-len(picked)))
	}
	return lines
}

func filterRoutineRestartAdvice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, item := range in {
		text := strings.TrimSpace(item)
		if text == "" {
			continue
		}
		lower := strings.ToLower(text)
		if strings.Contains(text, "重启") && !strings.Contains(text, "异常") && !strings.Contains(text, "补丁") && !strings.Contains(text, "升级") {
			continue
		}
		if strings.Contains(lower, "maintenance window") && !strings.Contains(text, "补丁") {
			continue
		}
		out = append(out, text)
	}
	return out
}

func importantObservedFacts(result *executor.Result, limit int) []string {
	if result == nil || limit <= 0 {
		return nil
	}
	var out []string
	for _, step := range result.Steps {
		for _, evidence := range step.Evidence {
			facts, ok := evidence.Data["observed_facts"].([]string)
			if ok {
				for _, fact := range facts {
					fact = strings.TrimSpace(fact)
					if fact == "" {
						continue
					}
					out = append(out, fact)
					if len(out) >= limit {
						return dedupeStrings(out)
					}
				}
				continue
			}
			if rawFacts, ok := evidence.Data["observed_facts"].([]any); ok {
				for _, fact := range rawFacts {
					text := strings.TrimSpace(asString(fact))
					if text == "" {
						continue
					}
					out = append(out, text)
					if len(out) >= limit {
						return dedupeStrings(out)
					}
				}
			}
		}
	}
	return dedupeStrings(out)
}

func compactParagraphs(in []string) []string {
	out := make([]string, 0, len(in))
	for _, item := range in {
		text := strings.TrimSpace(item)
		if text != "" {
			out = append(out, text)
		}
	}
	return out
}

func dedupeStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func asString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	default:
		return ""
	}
}

func looseInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case float32:
		return int(x)
	default:
		return 0
	}
}
