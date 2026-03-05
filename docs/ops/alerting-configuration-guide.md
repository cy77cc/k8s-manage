# Alerting Configuration Guide

## 1. Start Components

- Prometheus compose: `deploy/compose/prometheus/`
- Alertmanager compose: `deploy/compose/alertmanager/`

Prometheus references:
- `prometheus.yml`
- `alerting_rules.yml`

Alertmanager references:
- `alertmanager.yml`
- receiver webhook: `http://k8s-manage:8080/api/v1/alerts/receiver`

## 2. Rule Lifecycle

1. Manage rules through `/api/v1/alert-rules`.
2. Rule CRUD triggers automatic sync.
3. You can force sync with `POST /api/v1/alerts/rules/sync`.
4. Sync writes Prometheus rule file and calls `POST /-/reload`.

## 3. Channel Configuration

Use `alert_notification_channels` with fields:
- `provider`: `log`, `dingtalk`, `wecom`, `email`, `sms`
- `config_json`: provider-specific JSON configuration

Examples:
- DingTalk: `{"webhook":"https://oapi.dingtalk.com/robot/send?..."}`
- WeCom: `{"webhook":"https://qyapi.weixin.qq.com/cgi-bin/webhook/send?..."}`
- Email: `{"smtp_host":"smtp.example.com","smtp_port":465,...}`

## 4. Delivery Behavior

- Async per-channel dispatch.
- Retry: `1s`, `2s`, `4s` (max 3 attempts).
- Failures are recorded in `alert_notification_deliveries`.

## 5. Troubleshooting

- Prometheus reload failed: verify `prometheus` reachable from backend.
- No deliveries: verify channels `enabled=1` and provider config.
- Webhook not received: verify Alertmanager route receiver URL and network.
