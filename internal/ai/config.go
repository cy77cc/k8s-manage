package ai

import "github.com/cy77cc/OpsPilot/internal/config"

type RolloutConfig struct {
	UseMultiDomainArch          bool `json:"use_multi_domain_arch"`
	UseAssistantV2              bool `json:"use_assistant_v2"`
	UseModelFirstRuntime        bool `json:"use_model_first_runtime"`
	AllowLegacySemanticFallback bool `json:"allow_legacy_semantic_fallback"`
}

type RolloutThresholds struct {
	MaxPlannerErrorRate    float64 `json:"max_planner_error_rate"`
	MaxResumeFailureRate   float64 `json:"max_resume_failure_rate"`
	MinRewriteSuccessRate  float64 `json:"min_rewrite_success_rate"`
	AllowCompatibilityPath bool    `json:"allow_compatibility_path"`
}

type RolloutDecision struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}

func CurrentRolloutConfig() RolloutConfig {
	cfg := RolloutConfig{
		UseMultiDomainArch:   config.CFG.AI.UseMultiDomainArch,
		UseModelFirstRuntime: true,
	}
	if config.CFG.FeatureFlags.AIAssistantV2 != nil {
		cfg.UseAssistantV2 = *config.CFG.FeatureFlags.AIAssistantV2
	}
	if config.CFG.FeatureFlags.AIModelFirstRuntime != nil {
		cfg.UseModelFirstRuntime = *config.CFG.FeatureFlags.AIModelFirstRuntime
	}
	if config.CFG.FeatureFlags.AILegacySemanticFallback != nil {
		cfg.AllowLegacySemanticFallback = *config.CFG.FeatureFlags.AILegacySemanticFallback
	}
	return cfg
}

func (c RolloutConfig) AgenticEnabled() bool {
	return c.UseMultiDomainArch
}

func (c RolloutConfig) ModelFirstEnabled() bool {
	return c.UseModelFirstRuntime
}

func (c RolloutConfig) CompatibilityEnabled() bool {
	return !c.UseMultiDomainArch || c.UseAssistantV2 || c.AllowLegacySemanticFallback
}

func (c RolloutConfig) RuntimeMode() string {
	switch {
	case c.ModelFirstEnabled():
		return "model_first"
	case c.AllowLegacySemanticFallback:
		return "compatibility"
	default:
		return "disabled"
	}
}

func DefaultRolloutThresholds() RolloutThresholds {
	return RolloutThresholds{
		MaxPlannerErrorRate:    0.10,
		MaxResumeFailureRate:   0.02,
		MinRewriteSuccessRate:  0.90,
		AllowCompatibilityPath: true,
	}
}

func (c RolloutConfig) Decide(th RolloutThresholds, plannerErrorRate, resumeFailureRate, rewriteSuccessRate float64) RolloutDecision {
	if !c.AgenticEnabled() {
		return RolloutDecision{Enabled: false, Reason: "agentic rollout disabled by config"}
	}
	if rewriteSuccessRate < th.MinRewriteSuccessRate {
		return RolloutDecision{Enabled: false, Reason: "rewrite success rate below rollout threshold"}
	}
	if plannerErrorRate > th.MaxPlannerErrorRate {
		return RolloutDecision{Enabled: false, Reason: "planner error rate above rollout threshold"}
	}
	if resumeFailureRate > th.MaxResumeFailureRate {
		return RolloutDecision{Enabled: false, Reason: "resume failure rate above rollout threshold"}
	}
	return RolloutDecision{Enabled: true, Reason: "rollout thresholds satisfied"}
}
