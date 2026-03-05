# Monitoring & Alerting API

Base path: `/api/v1`

## Metrics

- `GET /metrics`
  - query: `metric` (required), `start_time`, `end_time`, `granularity_sec`, `source`
  - returns Prometheus-backed series payload:
    - `window.start`
    - `window.end`
    - `window.granularity_sec`
    - `dimensions`
    - `series[]` with `timestamp`, `value`, optional `labels`

## Alert Rules

- `GET /alert-rules`
- `POST /alert-rules`
- `PUT /alert-rules/:id`
- `POST /alert-rules/:id/enable`
- `POST /alert-rules/:id/disable`
- `POST /alerts/rules/sync` (manual sync to Prometheus `alerting_rules.yml` and reload)

## Alerts

- `GET /alerts`
- `GET /alert-rules/:id/evaluations`
- `POST /alerts/receiver` (Alertmanager webhook receiver)

## Notification Channels

- `GET /alert-channels`
- `POST /alert-channels`
- `PUT /alert-channels/:id`
- `GET /alert-deliveries`
