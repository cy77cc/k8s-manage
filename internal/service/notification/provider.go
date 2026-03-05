package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cy77cc/k8s-manage/internal/model"
)

// Provider is the pluggable notification delivery interface.
type Provider interface {
	Name() string
	Send(ctx context.Context, alert *model.AlertEvent, channel model.AlertNotificationChannel) error
	ValidateConfig(config map[string]any) error
}

// ProviderRegistry manages notification providers.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{providers: make(map[string]Provider)}
}

func NewDefaultProviderRegistry() *ProviderRegistry {
	r := NewProviderRegistry()
	r.Register(&LogProvider{})
	r.Register(&DingTalkProvider{client: &http.Client{Timeout: 5 * time.Second}})
	r.Register(&WeComProvider{client: &http.Client{Timeout: 5 * time.Second}})
	r.Register(&EmailProvider{})
	r.Register(&SMSProvider{})
	return r
}

func (r *ProviderRegistry) Register(p Provider) {
	if r == nil || p == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[strings.ToLower(strings.TrimSpace(p.Name()))] = p
}

func (r *ProviderRegistry) Get(name string) (Provider, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[strings.ToLower(strings.TrimSpace(name))]
	return p, ok
}

func ParseChannelConfig(channel model.AlertNotificationChannel) map[string]any {
	out := map[string]any{}
	if strings.TrimSpace(channel.ConfigJSON) == "" {
		return out
	}
	_ = json.Unmarshal([]byte(channel.ConfigJSON), &out)
	return out
}

type LogProvider struct{}

func (p *LogProvider) Name() string { return "log" }

func (p *LogProvider) ValidateConfig(_ map[string]any) error { return nil }

func (p *LogProvider) Send(_ context.Context, alert *model.AlertEvent, channel model.AlertNotificationChannel) error {
	_ = alert
	_ = channel
	return nil
}

type DingTalkProvider struct{ client *http.Client }

func (p *DingTalkProvider) Name() string { return "dingtalk" }

func (p *DingTalkProvider) ValidateConfig(config map[string]any) error {
	webhook := strings.TrimSpace(fmt.Sprintf("%v", config["webhook"]))
	if webhook == "" {
		return fmt.Errorf("dingtalk webhook is required")
	}
	return nil
}

func (p *DingTalkProvider) Send(ctx context.Context, alert *model.AlertEvent, channel model.AlertNotificationChannel) error {
	cfg := ParseChannelConfig(channel)
	if err := p.ValidateConfig(cfg); err != nil {
		return err
	}
	webhook := strings.TrimSpace(fmt.Sprintf("%v", cfg["webhook"]))
	if webhook == "" {
		webhook = strings.TrimSpace(channel.Target)
	}
	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]any{
			"title": alert.Title,
			"text":  fmt.Sprintf("### %s\n\n- 状态: %s\n- 级别: %s\n- 指标: %s\n- 当前值: %.2f\n- 阈值: %.2f", alert.Title, alert.Status, alert.Severity, alert.Metric, alert.Value, alert.Threshold),
		},
	}
	return postJSON(ctx, p.client, webhook, payload)
}

type WeComProvider struct{ client *http.Client }

func (p *WeComProvider) Name() string { return "wecom" }

func (p *WeComProvider) ValidateConfig(config map[string]any) error {
	webhook := strings.TrimSpace(fmt.Sprintf("%v", config["webhook"]))
	if webhook == "" {
		return fmt.Errorf("wecom webhook is required")
	}
	return nil
}

func (p *WeComProvider) Send(ctx context.Context, alert *model.AlertEvent, channel model.AlertNotificationChannel) error {
	cfg := ParseChannelConfig(channel)
	if err := p.ValidateConfig(cfg); err != nil {
		return err
	}
	webhook := strings.TrimSpace(fmt.Sprintf("%v", cfg["webhook"]))
	if webhook == "" {
		webhook = strings.TrimSpace(channel.Target)
	}
	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]any{
			"content": fmt.Sprintf("**%s**\n> 状态: %s\n> 级别: %s\n> 指标: %s\n> 当前值: %.2f\n> 阈值: %.2f", alert.Title, alert.Status, alert.Severity, alert.Metric, alert.Value, alert.Threshold),
		},
	}
	return postJSON(ctx, p.client, webhook, payload)
}

type EmailProvider struct{}

func (p *EmailProvider) Name() string { return "email" }

func (p *EmailProvider) ValidateConfig(config map[string]any) error {
	if strings.TrimSpace(fmt.Sprintf("%v", config["smtp_host"])) == "" {
		return fmt.Errorf("email smtp_host is required")
	}
	return nil
}

func (p *EmailProvider) Send(_ context.Context, alert *model.AlertEvent, _ model.AlertNotificationChannel) error {
	_ = alert
	return nil
}

type SMSProvider struct{}

func (p *SMSProvider) Name() string { return "sms" }

func (p *SMSProvider) ValidateConfig(_ map[string]any) error { return nil }

func (p *SMSProvider) Send(_ context.Context, alert *model.AlertEvent, _ model.AlertNotificationChannel) error {
	_ = alert
	return nil
}

func postJSON(ctx context.Context, c *http.Client, endpoint string, payload map[string]any) error {
	if strings.TrimSpace(endpoint) == "" {
		return fmt.Errorf("notification endpoint is empty")
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("provider response status: %d", resp.StatusCode)
	}
	return nil
}
