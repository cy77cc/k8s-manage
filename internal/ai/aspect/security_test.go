package aspect

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/eino/compose"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

func TestSecurityAspectPermissionDenied(t *testing.T) {
	t.Parallel()

	aspect := NewSecurityAspect([]core.RegisteredTool{{
		Meta: core.ToolMeta{
			Name:       "service_restart",
			Permission: "service:write",
			Risk:       tools.ToolRiskLow,
		},
	}}, staticPermissionChecker{allowed: false}, nil, &memoryAuditLogger{})

	middleware := aspect.Middleware()
	_, err := middleware.Invokable(func(_ context.Context, _ *compose.ToolInput) (*compose.ToolOutput, error) {
		return &compose.ToolOutput{Result: "ok"}, nil
	})(context.Background(), &compose.ToolInput{Name: "service_restart", Arguments: `{}`})
	if err == nil {
		t.Fatal("middleware error = nil, want permission failure")
	}
	var approvalErr *tools.ApprovalRequiredError
	if !errors.As(err, &approvalErr) {
		t.Fatalf("middleware error = %T, want ApprovalRequiredError", err)
	}
}

func TestSecurityAspectInterruptsRiskyTool(t *testing.T) {
	t.Parallel()

	aspect := NewSecurityAspect([]core.RegisteredTool{{
		Meta: core.ToolMeta{
			Name:       "cluster_delete",
			Permission: "cluster:write",
			Risk:       tools.ToolRiskHigh,
		},
	}}, staticPermissionChecker{allowed: true}, staticInterruptHandler{}, &memoryAuditLogger{})

	middleware := aspect.Middleware()
	_, err := middleware.Invokable(func(_ context.Context, _ *compose.ToolInput) (*compose.ToolOutput, error) {
		return &compose.ToolOutput{Result: "ok"}, nil
	})(context.Background(), &compose.ToolInput{Name: "cluster_delete", Arguments: `{"id":1}`})
	if err == nil {
		t.Fatal("middleware error = nil, want interrupt")
	}
	var approvalErr *tools.ApprovalRequiredError
	if errors.As(err, &approvalErr) {
		t.Fatalf("middleware error = ApprovalRequiredError, want interrupt signal")
	}
}

func TestSecurityAspectAllowsSafeTool(t *testing.T) {
	t.Parallel()

	logger := &memoryAuditLogger{}
	aspect := NewSecurityAspect([]core.RegisteredTool{{
		Meta: core.ToolMeta{
			Name:       "service_status",
			Permission: "service:read",
			Risk:       tools.ToolRiskLow,
		},
	}}, staticPermissionChecker{allowed: true}, nil, logger)

	middleware := aspect.Middleware()
	out, err := middleware.Invokable(func(_ context.Context, _ *compose.ToolInput) (*compose.ToolOutput, error) {
		return &compose.ToolOutput{Result: "ok"}, nil
	})(context.Background(), &compose.ToolInput{Name: "service_status", Arguments: `{}`})
	if err != nil {
		t.Fatalf("middleware error = %v", err)
	}
	if out.Result != "ok" {
		t.Fatalf("middleware result = %q, want ok", out.Result)
	}
	if len(logger.records) != 1 {
		t.Fatalf("audit records = %d, want 1", len(logger.records))
	}
}

type staticPermissionChecker struct {
	allowed bool
}

func (s staticPermissionChecker) HasPermission(context.Context, string) (bool, error) {
	return s.allowed, nil
}

type staticInterruptHandler struct{}

func (staticInterruptHandler) BuildInterrupt(_ context.Context, meta core.ToolMeta, _ map[string]any) (*tools.ApprovalInfo, error) {
	return &tools.ApprovalInfo{
		ToolName: meta.Name,
		Preview:  map[string]any{"mode": "review"},
	}, nil
}

type memoryAuditLogger struct {
	records []AuditRecord
}

func (m *memoryAuditLogger) Log(_ context.Context, record AuditRecord) error {
	m.records = append(m.records, record)
	return nil
}
