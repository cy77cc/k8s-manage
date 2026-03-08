# AI Multi-Domain Rollout

## Overview

The AI runtime now supports a rollout toggle for the new multi-domain planning path.

## Configuration

Set `ai.use_multi_domain_arch` in `configs/config.yaml`.

```yaml
ai:
  use_multi_domain_arch: false
```

- `false`: keep the legacy agentic path
- `true`: route agentic requests through the multi-domain planner path

## Current Behavior

- Simple chat mode is unchanged
- Agentic mode checks the rollout toggle
- When enabled, `HybridAgent` uses the multi-domain planner path and summarizes the generated domain plans

## Notes

- This rollout currently wires the planner-oriented multi-domain entrypoint first
- Existing agentic execution remains available as the fallback path
