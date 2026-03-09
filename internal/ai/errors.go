package ai

import "errors"

// AI 模块错误定义。
var (
	// ErrToolNotFound 工具未找到错误。
	ErrToolNotFound = errors.New("tool not found")
	// ErrPermissionDenied 权限拒绝错误。
	ErrPermissionDenied = errors.New("permission denied")
	// ErrApprovalExpired 审批已过期错误。
	ErrApprovalExpired = errors.New("approval expired")
)

// EventPlanCreated 是计划创建事件名称。
const EventPlanCreated = "plan_created"
