package logic

import (
	"fmt"
	"strings"
	"time"

	"github.com/cy77cc/OpsPilot/internal/model"
)

// EvaluateOperationalEligibility checks whether a host can participate in
// non-essential operational workflows (deployment candidate selection,
// automation, routine tasks).
func EvaluateOperationalEligibility(host *model.Node) (bool, string) {
	if host == nil {
		return false, "host not found"
	}
	if strings.TrimSpace(host.IP) == "" {
		return false, "host missing ip"
	}
	status := strings.ToLower(strings.TrimSpace(host.Status))
	switch status {
	case "maintenance":
		return false, buildMaintenanceReason(host)
	case "offline", "inactive", "error":
		return false, fmt.Sprintf("host unavailable: %s", status)
	}
	return true, ""
}

func buildMaintenanceReason(host *model.Node) string {
	reason := strings.TrimSpace(host.MaintenanceReason)
	if reason == "" {
		reason = "maintenance mode"
	}
	if host.MaintenanceUntil != nil && !host.MaintenanceUntil.IsZero() {
		return fmt.Sprintf("%s (until %s)", reason, host.MaintenanceUntil.Format(time.RFC3339))
	}
	return reason
}
