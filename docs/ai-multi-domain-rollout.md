# AI Model-First Rollout

## Overview

The stage-based AI runtime now supports an explicit rollout mode for the model-first contract.

This rollout is not intended to silently re-enable deterministic semantic fallbacks. Its purpose is to make rollout state explicit and give operators a controlled rollback posture.

## Configuration

```yaml
ai:
  use_multi_domain_arch: true

feature_flags:
  ai_model_first_runtime: true
  ai_legacy_semantic_fallback: false
```

Flags:

- `ai.use_multi_domain_arch`
  - enables the stage-based orchestrator runtime
- `feature_flags.ai_model_first_runtime`
  - enables the model-first semantic contract
- `feature_flags.ai_legacy_semantic_fallback`
  - marks compatibility rollback mode for operators

## Runtime Modes

- `model_first`
  - stage-based runtime is enabled
  - model-first contract is enabled
  - no hidden code semantic fallback should impersonate AI stages
- `compatibility`
  - model-first contract is disabled for rollback
  - operators should direct users to compatibility/manual paths as needed
- `disabled`
  - model-first contract is disabled and no compatibility mode is declared

The active mode is exposed via:

- response header `X-AI-Runtime-Mode`
- SSE `meta.runtime_mode`

## Rollback Guidance

If production rollback is required:

1. Keep `ai.use_multi_domain_arch=true` unless the entire stage runtime must be disabled.
2. Set `feature_flags.ai_model_first_runtime=false`.
3. If operations require an explicit rollback marker, set `feature_flags.ai_legacy_semantic_fallback=true`.
4. In rollback mode, instruct operators and users to use manual operations or compatibility endpoints where needed.

Do not:

- silently restore deterministic semantic fallback code
- make the system appear fully AI-functional when a model layer is unavailable

## Validation Checklist

- startup logs include `rewrite`, `planner`, `expert`, and `summarizer`
- SSE `meta` includes rollout flags
- response headers include rollout mode
- stage failures stay explicit during rollback or degraded mode
