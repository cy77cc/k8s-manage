## ADDED Requirements

### Requirement: Critical Read Paths SHALL Use L1 Cache-Aside
The system SHALL serve designated critical read paths through an L1 in-process cache-aside strategy with deterministic load-from-source behavior on cache miss.

#### Scenario: L1 hit returns cached payload
- **WHEN** a request targets a capability marked as L1-governed and a valid L1 entry exists
- **THEN** the system MUST return the cached payload without querying the primary database

#### Scenario: L1 miss loads from source and populates cache
- **WHEN** a request targets a capability marked as L1-governed and no valid L1 entry exists
- **THEN** the system MUST load data from the source of truth
- **AND** the system MUST store the loaded payload in L1 using the configured key and TTL policy

### Requirement: Cache Key Namespace And TTL Policy SHALL Be Standardized
The system SHALL define and enforce a cache governance table for key namespace, TTL class, and invalidation scope for each governed endpoint or repository read method.

#### Scenario: Key naming follows governance convention
- **WHEN** maintainers review cache configuration for a governed read path
- **THEN** the key MUST use the domain-governed namespace format
- **AND** the path MUST have an explicit TTL class defined in the governance table

#### Scenario: Mutable list endpoints use short TTL class
- **WHEN** a governed endpoint is classified as mutable list data
- **THEN** the endpoint MUST use the short TTL class defined by governance

### Requirement: Write Paths SHALL Invalidate Deterministic Key Sets
The system SHALL invalidate deterministic key sets on governed write operations so subsequent reads do not rely on stale cache entries.

#### Scenario: Entity update invalidates detail and list keys
- **WHEN** a governed entity is created, updated, deleted, or synchronized
- **THEN** the system MUST invalidate that entity's detail key
- **AND** the system MUST invalidate affected list keys defined by governance scope

### Requirement: L1 Governance SHALL Degrade Gracefully Without Redis
The system SHALL preserve governed read/write behavior when Redis is disabled or unavailable, with L1 continuing as the primary cache mechanism.

#### Scenario: Redis disabled still satisfies governed reads
- **WHEN** Redis is disabled in runtime configuration
- **THEN** governed reads MUST continue using L1 cache-aside behavior
- **AND** the request MUST NOT fail solely because Redis is unavailable

#### Scenario: Redis runtime error falls back safely
- **WHEN** an optional Redis access attempt fails for a governed path
- **THEN** the system MUST continue request processing via L1 and/or source-of-truth fallback
- **AND** the system MUST record fallback telemetry

### Requirement: Cache Telemetry SHALL Be Observable
The system SHALL expose cache observability signals for governed paths, including hit, miss, and source-fallback outcomes.

#### Scenario: Cache metrics are emitted on governed reads
- **WHEN** a governed read request is processed
- **THEN** the system MUST emit telemetry indicating whether the result came from L1 hit, miss-load, or fallback path
