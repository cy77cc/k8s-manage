package executor

import "testing"

func TestParseExpertResultRecoversHalfStructuredOutput(t *testing.T) {
	raw := `Summary: 已成功检查 payment-api 服务状态
Observed: service_id=42 状态正常
Inference: 当前没有发现异常
Next: 如需更深排查可继续查看最近日志`

	out, err := parseExpertResult(raw)
	if err != nil {
		t.Fatalf("parseExpertResult() error = %v", err)
	}
	if out.Summary == "" {
		t.Fatalf("Summary should not be empty: %#v", out)
	}
	if len(out.ObservedFacts) == 0 {
		t.Fatalf("ObservedFacts should be recovered: %#v", out)
	}
	if len(out.Inferences) == 0 {
		t.Fatalf("Inferences should be recovered: %#v", out)
	}
	if len(out.NextActions) == 0 {
		t.Fatalf("NextActions should be recovered: %#v", out)
	}
}
