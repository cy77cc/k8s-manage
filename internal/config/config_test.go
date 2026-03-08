package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestConfigSupportsAIMultiDomainToggle(t *testing.T) {
	v := viper.New()
	v.Set("ai.use_multi_domain_arch", true)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if !cfg.AI.UseMultiDomainArch {
		t.Fatal("expected ai.use_multi_domain_arch to unmarshal into config")
	}
}
