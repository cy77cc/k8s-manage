## Why

The current codebase mixes business logic with direct GORM access across multiple domains, making transaction boundaries, cache invalidation, and performance behavior difficult to reason about. We need a cache-first and repository-first baseline now to stabilize core-path latency and reduce operational risk without requiring external middleware.

## What Changes

- Introduce a unified L1 in-process cache capability for critical read-heavy business paths, using a consistent cache-aside policy and standardized key/TTL conventions.
- Define a repository-boundary requirement for selected domains so handlers/logic no longer perform scattered direct database operations.
- Standardize cache invalidation rules for write paths (create/update/delete/sync), including deterministic key invalidation instead of ad hoc logic.
- Require graceful degradation when Redis is disabled or unavailable, with L1 remaining the default runtime path.
- Add observability requirements for cache hit/miss and DB fallback behavior to support safe rollout and tuning.

## Capabilities

### New Capabilities
- `l1-cache-governance`: Establishes L1 cache behavior, keying, TTL strategy, invalidation rules, and degradation semantics for critical business reads.
- `repository-access-boundary`: Defines mandatory data-access layering and repository contracts for domain services that currently mix business flow and DB operations.

### Modified Capabilities
- `platform-capability-baseline`: Clarifies platform baseline expectations to include cache/DB resilience behavior when optional middleware (e.g., Redis) is unavailable.

## Impact

- Affected backend code under `internal/service/*`, especially high-traffic domains such as `cluster`, plus data-access abstractions under `internal/dao` or domain-local repositories.
- No immediate external API contract change is required, but internal request handling behavior and performance characteristics will be standardized.
- Service context and runtime wiring may be adjusted to make L1 primary and Redis optional fallback/enhancement.
- Additional metrics and logging fields will be introduced for cache efficacy and DB fallback monitoring.
