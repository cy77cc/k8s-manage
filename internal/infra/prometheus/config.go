package prometheus

import (
	"fmt"
	"strings"
	"time"
)

// Config holds Prometheus client settings.
type Config struct {
	Address       string        `yaml:"address" json:"address"`
	Host          string        `yaml:"host" json:"host"`
	Port          string        `yaml:"port" json:"port"`
	Timeout       time.Duration `yaml:"timeout" json:"timeout"`
	MaxConcurrent int           `yaml:"max_concurrent" json:"max_concurrent"`
	RetryCount    int           `yaml:"retry_count" json:"retry_count"`
}

func (c Config) Normalize() Config {
	out := c
	if strings.TrimSpace(out.Address) == "" {
		h := strings.TrimSpace(out.Host)
		p := strings.TrimSpace(out.Port)
		if h != "" {
			if p == "" {
				p = "9090"
			}
			out.Address = fmt.Sprintf("http://%s:%s", h, p)
		}
	}
	if out.Timeout <= 0 {
		out.Timeout = 10 * time.Second
	}
	if out.MaxConcurrent <= 0 {
		out.MaxConcurrent = 10
	}
	if out.RetryCount <= 0 {
		out.RetryCount = 3
	}
	return out
}
