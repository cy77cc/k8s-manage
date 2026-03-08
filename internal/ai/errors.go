package ai

import "errors"

var (
	ErrToolNotFound     = errors.New("tool not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrApprovalExpired  = errors.New("approval expired")
)

const EventPlanCreated = "plan_created"
