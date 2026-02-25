## Why

监控与告警是 roadmap Phase 2 核心目标，但目前缺少独立 OpenSpec change 管理指标模型、告警规则与通知链路。

## What Changes

- 建立 Monitoring/Alerting 独立 change，定义指标、规则、告警事件与通知行为。
- 规范监控数据查询与告警策略管理的后续实施范围。

## Capabilities

### New Capabilities
- `monitoring-alerting-phase`: 定义监控采集、告警规则、通知与审计的阶段需求。

### Modified Capabilities
- None.

## Impact

- Affected modules: `internal/service/monitoring/*`, `web/src/api/modules/monitoring.ts`.
- Affected docs: `docs/roadmap.md`, `docs/qa/*monitoring*`.
