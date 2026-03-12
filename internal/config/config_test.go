package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestConfigSupportsAIMultiDomainToggle(t *testing.T) {
	v := viper.New()
	v.Set("ai.use_multi_domain_arch", true)
	v.Set("ai.use_turn_block_streaming", true)
	v.Set("feature_flags.ai_model_first_runtime", true)
	v.Set("feature_flags.ai_legacy_semantic_fallback", false)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if !cfg.AI.UseMultiDomainArch {
		t.Fatal("expected ai.use_multi_domain_arch to unmarshal into config")
	}
	if !cfg.AI.UseTurnBlockStreaming {
		t.Fatal("expected ai.use_turn_block_streaming to unmarshal into config")
	}
	if cfg.FeatureFlags.AIModelFirstRuntime == nil || !*cfg.FeatureFlags.AIModelFirstRuntime {
		t.Fatal("expected feature_flags.ai_model_first_runtime to unmarshal into config")
	}
	if cfg.FeatureFlags.AILegacySemanticFallback == nil || *cfg.FeatureFlags.AILegacySemanticFallback {
		t.Fatal("expected feature_flags.ai_legacy_semantic_fallback to unmarshal into config")
	}
}
