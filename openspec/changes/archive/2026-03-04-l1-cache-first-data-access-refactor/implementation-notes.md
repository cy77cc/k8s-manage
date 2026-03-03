# Implementation Notes: L1 Cache First + Repository Boundary

## 1. Rollout Toggles And Rollback (Task 5.1)

### Rollout toggles used in this change
- `redis.enable`:
  - `true`: L2 Redis hook is active (best-effort), with L1 as primary.
  - `false`: L2 hook is disabled; governed paths continue through L1 + source-of-truth fallback.
- Cluster phase-1 governed paths (implemented):
  - cluster list/detail/nodes
  - bootstrap profile list
- Write-driven deterministic invalidation paths (implemented):
  - cluster import/update/delete/sync
  - bootstrap profile create/update/delete

### Rollback procedure
1. Disable Redis dependency impact immediately by setting `redis.enable=false` (L1-only path remains available).
2. If behavioral regression is observed on governed read paths, rollback by reverting this change set for `internal/service/cluster/*` and `internal/cache/*`.
3. Keep repository delete transaction behavior as atomic baseline (`DeleteClusterWithRelations`) to avoid partial write side effects during rollback.

## 2. Telemetry Baseline Record (Task 5.2)

### Before migration baseline
- Cache telemetry primitives on governed paths: **not available** (no standardized L1 hit/miss/fallback counters before this change).
- DB fallback observability on governed paths: **not available** in unified form before this change.

### After migration baseline (available signals)
- `Facade.Stats().L1Hits`
- `Facade.Stats().L2Hits`
- `Facade.Stats().Misses`
- `Facade.Stats().FallbackLoads`

### Local verification evidence
- Cache facade behavior validated with unit tests in `internal/cache/facade_test.go` (L1 hit/miss, L2 fallback, delete invalidation, Redis-error fallback).
- Cluster repository and migrated core path behavior validated via `go test ./internal/service/cluster`.

## 3. Next-Domain Candidates (Task 5.3)

Recommended next domains for repository-boundary + L1-governed expansion:
1. `service` domain
- High query frequency and configuration reads.
- Existing deploy/config flows benefit from standardized invalidation.

2. `deployment` domain
- Release/target list/detail and approval flows can use short-TTL list caching with deterministic invalidation.

3. `host` domain
- Host detail/list and credentials metadata reads have clear cache-aside candidates.

4. `rbac` domain
- Permission graph and role-permission read paths are good L1 candidates, with explicit write-side invalidation.
