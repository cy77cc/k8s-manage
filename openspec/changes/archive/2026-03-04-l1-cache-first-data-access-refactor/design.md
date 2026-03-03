## Context

The current backend has mixed layering: some domains use repository abstraction (for example, CICD), while others place direct GORM access in handlers/logic (notably cluster-related paths). Caching is also inconsistent: L1 cache exists in service context, Redis is used in selective DAO/logic paths, and key/TTL/invalidation patterns vary by module.

This change is cross-cutting because it affects domain service structure, cache behavior, and runtime resilience expectations. The design must preserve current API contracts while introducing a consistent internal data-access boundary and L1-first caching policy for critical business reads.

## Goals / Non-Goals

**Goals:**
- Establish an L1-first cache governance model for critical read paths, with deterministic cache-aside and write invalidation semantics.
- Define and enforce repository access boundaries so domain handlers/logic do not perform scattered direct DB access.
- Ensure graceful runtime behavior when Redis is disabled or unavailable; L1 remains primary for the targeted paths.
- Provide measurable observability for hit/miss/fallback so rollout can be tuned with evidence.
- Deliver migration in incremental slices that do not break existing `/api/v1` contracts.

**Non-Goals:**
- Rewriting all domains in one iteration.
- Introducing a distributed cache invalidation bus in this change.
- Changing external request/response contracts unless required by explicit future specs.
- Replacing all existing DAO packages immediately.

## Decisions

1. L1-first cache architecture with optional L2 Redis
- Decision: Standardize reads on `GetOrLoad` semantics where L1 is checked first; Redis (if enabled) is optional enhancement, not a hard dependency.
- Rationale: Core business paths should remain performant and available without middleware dependency.
- Alternative considered: Redis-first strategy. Rejected because it increases operational coupling and adds outage blast radius for core reads.

2. Cache policy: cache-aside + deterministic invalidation
- Decision: Use cache-aside for targeted read endpoints and deterministic key invalidation on write paths.
- Rationale: Predictable behavior and easier correctness review compared with ad hoc cache writes and delayed double-delete patterns.
- Alternative considered: write-through. Rejected for now due to higher coupling between write flow and cache state management across mixed legacy code paths.

3. Repository boundary for domain data access
- Decision: Introduce/expand domain repository interfaces (starting with cluster and related hot paths); handlers/logic delegate persistence to repos.
- Rationale: Centralizes query behavior, transaction boundaries, and cache invalidation call sites.
- Alternative considered: keep direct GORM in handlers with coding conventions only. Rejected because conventions alone are weakly enforceable in a large codebase.

4. Incremental domain rollout (cluster-first)
- Decision: First migrate high-traffic cluster read/write paths, then extend pattern to additional domains.
- Rationale: Largest immediate latency and maintainability benefit, with controlled blast radius.
- Alternative considered: horizontal framework-first refactor across all domains. Rejected due to delivery risk and verification complexity.

5. Observability as rollout gate
- Decision: Add metrics for cache hit/miss, load fallback, and query latency on migrated paths before broader adoption.
- Rationale: Prevents regressions hidden by local improvements and supports evidence-based TTL/key tuning.
- Alternative considered: defer metrics to later. Rejected because migration safety depends on runtime feedback.

## Risks / Trade-offs

- [Cross-instance inconsistency with L1-only] -> Mitigation: Scope initial guarantees to single-instance consistency; document behavior and keep optional L2 hooks for later multi-instance hardening.
- [Stale data from overly long TTL] -> Mitigation: Use short TTL defaults on mutable lists, deterministic invalidation on writes, and endpoint-level tuning from metrics.
- [Refactor churn in cluster module] -> Mitigation: Migrate in slices (read endpoints first, then write paths), preserve existing API tests, and avoid broad unrelated edits.
- [Hidden dependency on Redis in legacy code] -> Mitigation: Guard all Redis access with nil-safe adapter semantics and fallback to DB/L1 behavior.
- [N+1 query patterns still expensive on cache miss] -> Mitigation: During repository migration, normalize critical list queries and aggregate counting logic.

## Migration Plan

1. Introduce shared cache facade and key conventions
- Add a small internal cache facade for L1 operations (`get/set/delete/get-or-load`) and optional L2 hook points.
- Define cache key namespaces and TTL policy table for targeted cluster endpoints.

2. Cluster read-path migration (non-breaking)
- Move selected cluster reads behind repository methods.
- Add L1 cache-aside on those methods.
- Validate behavior parity with existing response contracts.

3. Cluster write-path invalidation alignment
- Route selected write paths through repositories.
- Attach deterministic invalidation for impacted key sets.

4. Resilience and observability hardening
- Enforce nil-safe Redis behavior and fallback policy.
- Add metrics/log fields for cache behavior and DB fallback.

5. Expand pattern to next domains (follow-up)
- Apply the same repository/cache governance model to other high-value domains after cluster stabilization.

Rollback strategy
- Keep feature-flag style toggles at cache facade or endpoint-level wiring where practical.
- If regressions occur, disable cache path for affected endpoints while retaining repository boundary changes.
- Revert by module slice, not whole-system rollback.

## Open Questions

- Which specific endpoints should be in phase-1 mandatory L1 scope beyond cluster list/detail/nodes/profile reads?
- Should L1 entries be strictly per-process only, or should we add optional versioned keys anticipating multi-instance rollout?
- What default TTL baseline should be adopted for mutable list endpoints (e.g., 15s vs 30s), and who owns tuning decisions?
- Do we need a hard lint/check rule to prevent new direct DB calls in handlers/logic for migrated domains?
