## 1. Cache Governance Foundation

- [x] 1.1 Define L1 cache key namespace and TTL policy table for phase-1 governed cluster paths.
- [x] 1.2 Implement a shared cache facade (`get`, `set`, `delete`, `get-or-load`) with L1-first behavior and optional Redis hook points.
- [x] 1.3 Add nil-safe Redis handling so Redis-disabled runtime does not break governed request paths.
- [x] 1.4 Add cache telemetry primitives for hit, miss, and source-fallback outcomes.

## 2. Repository Access Boundary (Cluster Phase-1)

- [x] 2.1 Define repository interfaces for phase-1 cluster reads (list/detail/nodes/bootstrap profiles).
- [x] 2.2 Define repository interfaces for phase-1 cluster writes (create/update/delete/sync and related invalidation scopes).
- [x] 2.3 Refactor phase-1 cluster handlers/logic to use repository contracts instead of direct DB calls.
- [x] 2.4 Add explicit transaction boundaries for phase-1 multi-record mutations.

## 3. L1 Cache Integration For Governed Paths

- [x] 3.1 Apply L1 cache-aside to phase-1 cluster read paths via repository/facade integration.
- [x] 3.2 Bind deterministic key invalidation to phase-1 write outcomes.
- [x] 3.3 Tune short TTL class for mutable list endpoints and validate stale-data risk controls.
- [x] 3.4 Verify API response compatibility for migrated endpoints under `/api/v1`.

## 4. Tests And Verification

- [x] 4.1 Add/adjust repository-focused tests for query and mutation semantics on phase-1 cluster scope.
- [x] 4.2 Add cache behavior tests for L1 hit, miss-load, invalidation, and Redis-disabled fallback.
- [x] 4.3 Run regression checks for cluster core endpoints to confirm behavior parity.
- [x] 4.4 Validate OpenSpec artifacts and consistency (`openspec validate --json`).

## 5. Rollout And Follow-up Preparation

- [x] 5.1 Document rollout toggles and rollback procedure for cache path and repository path changes.
- [x] 5.2 Record baseline telemetry before/after migration for cache efficacy and DB fallback.
- [x] 5.3 Identify next-domain candidate list for repository-boundary and L1-governed expansion.
