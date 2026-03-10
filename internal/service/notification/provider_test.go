package notification

import (
	"context"
	"testing"

	"github.com/cy77cc/OpsPilot/internal/model"
)

func TestProviderRegistryRegisterAndGet(t *testing.T) {
	r := NewProviderRegistry()
	p := &LogProvider{}
	r.Register(p)
	got, ok := r.Get("log")
	if !ok || got == nil {
		t.Fatalf("expected provider log")
	}
}

func TestDefaultProviders(t *testing.T) {
	r := NewDefaultProviderRegistry()
	for _, name := range []string{"log", "dingtalk", "wecom", "email", "sms"} {
		if _, ok := r.Get(name); !ok {
			t.Fatalf("missing provider %s", name)
		}
	}
}

func TestLogProviderSend(t *testing.T) {
	p := &LogProvider{}
	err := p.Send(context.Background(), &model.AlertEvent{ID: 1, Title: "A"}, model.AlertNotificationChannel{Name: "test"})
	if err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}
}
