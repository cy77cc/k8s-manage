package ai

import (
	"strings"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/ai/executor"
	"github.com/cy77cc/OpsPilot/internal/ai/planner"
	"github.com/cy77cc/OpsPilot/internal/ai/summarizer"
)

func TestFinalAnswerRendererFleetHostStatusUsesSummaryAndTopN(t *testing.T) {
	renderer := newFinalAnswerRenderer()
	paragraphs := renderer.Render("查看所有主机状态", &planner.ExecutionPlan{
		Resolved: planner.ResolvedResources{
			Scope: &planner.ResourceScope{Kind: "all", ResourceType: "host"},
		},
	}, &executor.Result{
		Steps: []executor.StepResult{
			{
				StepID:  "step-1",
				Summary: "inventory collected",
				Evidence: []executor.Evidence{{
					Kind:   "tool_result",
					Source: "host_list_inventory",
					Data: map[string]any{
						"list": []any{
							map[string]any{"id": 1, "name": "test", "status": "online", "cpu_cores": 4, "memory_mb": 16384, "disk_gb": 100},
							map[string]any{"id": 2, "name": "火山云服务器", "status": "online", "cpu_cores": 2, "memory_mb": 8192, "disk_gb": 40},
							map[string]any{"id": 3, "name": "香港云服务器", "status": "online", "cpu_cores": 2, "memory_mb": 4096, "disk_gb": 40},
							map[string]any{"id": 4, "name": "备用机", "status": "online", "cpu_cores": 2, "memory_mb": 4096, "disk_gb": 40},
						},
					},
				}},
			},
		},
	}, summarizer.SummaryOutput{
		Headline:        "所有主机运行稳定",
		Recommendations: []string{"建议评估是否安排维护窗口进行计划性重启。"},
	})

	joined := strings.Join(paragraphs, "\n\n")
	if !strings.Contains(joined, "共检查 4 台主机，当前均运行正常。") {
		t.Fatalf("rendered body = %q", joined)
	}
	if !strings.Contains(joined, "其余 1 台主机状态一致") {
		t.Fatalf("rendered body = %q, want topN summary", joined)
	}
	if strings.Contains(joined, "计划性重启") {
		t.Fatalf("rendered body = %q, should suppress routine restart advice", joined)
	}
}

func TestFinalAnswerRendererSuppressesRawCommandDump(t *testing.T) {
	renderer := newFinalAnswerRenderer()
	paragraphs := renderer.Render("查看磁盘使用情况", nil, &executor.Result{}, summarizer.SummaryOutput{
		Headline:    "已在火山云服务器上成功执行 df -h 命令",
		KeyFindings: []string{"完整输出如下：```text\nFilesystem Size Used Avail Use% Mounted on\n/dev/vda2 40G 10G 28G 27% /\n```", "根分区使用率 27%，当前没有磁盘空间压力。"},
	})

	joined := strings.Join(paragraphs, "\n\n")
	if strings.Contains(joined, "Filesystem") || strings.Contains(joined, "完整输出如下") {
		t.Fatalf("rendered body = %q, should not expose raw command dump", joined)
	}
	if !strings.Contains(joined, "根分区使用率 27%") {
		t.Fatalf("rendered body = %q, want summarized finding", joined)
	}
}

func TestFinalAnswerRendererFiltersBoilerplateSummaryScaffold(t *testing.T) {
	renderer := newFinalAnswerRenderer()
	paragraphs := renderer.Render("在火山云服务器上执行 df -h", nil, &executor.Result{
		Steps: []executor.StepResult{{
			StepID:  "step-1",
			Summary: "host command completed",
			Evidence: []executor.Evidence{{
				Kind:   "expert_result",
				Source: "hostops",
				Data: map[string]any{
					"observed_facts": []any{
						"命令 df -h 在 host_id 2 上执行成功，退出码为 0",
						"根文件系统 /dev/vda2 总容量 40G，已使用 10G，可用 28G，使用率 27%",
					},
				},
			}},
		}},
	}, summarizer.SummaryOutput{
		Headline:        "已完成本轮排查",
		KeyFindings:     []string{"已完成步骤 1 个", "收集执行证据 1 条"},
		Recommendations: []string{"查看最终结论正文"},
	})

	joined := strings.Join(paragraphs, "\n\n")
	if strings.Contains(joined, "已完成本轮排查") {
		t.Fatalf("rendered body = %q, should not keep boilerplate headline", joined)
	}
	if strings.Contains(joined, "已完成步骤 1 个") || strings.Contains(joined, "收集执行证据 1 条") {
		t.Fatalf("rendered body = %q, should not keep boilerplate findings", joined)
	}
	if strings.Contains(joined, "查看最终结论正文") {
		t.Fatalf("rendered body = %q, should not keep boilerplate recommendation", joined)
	}
	if !strings.Contains(joined, "根文件系统 /dev/vda2") {
		t.Fatalf("rendered body = %q, want real observed fact", joined)
	}
}
