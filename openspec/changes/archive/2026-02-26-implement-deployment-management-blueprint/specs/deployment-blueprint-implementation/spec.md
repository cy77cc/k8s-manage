## ADDED Requirements

### Requirement: Unified release lifecycle contract
The system MUST expose a unified lifecycle contract for release workflows under `/api/v1`, including normalized states and response fields for preview, approval, apply, and rollback across Kubernetes and Compose runtimes.

#### Scenario: Query release with normalized lifecycle
- **WHEN** an authorized user requests a release detail
- **THEN** the response MUST include normalized lifecycle state, runtime type, approval status, and current execution stage fields

### Requirement: Preview is mandatory before apply
The system MUST reject apply requests that do not reference a valid preview artifact generated from the same draft context and within configured TTL.

#### Scenario: Reject apply without valid preview
- **WHEN** a user calls apply without preview token, with expired token, or with mismatched draft parameters
- **THEN** the API MUST reject the request and return a preview-required error with machine-readable reason code

### Requirement: Production release approval gate
The system MUST place production-risk release requests into a global approval queue and MUST prevent execution before an authorized approver decision is recorded.

#### Scenario: Production release enters pending approval
- **WHEN** a release targets an environment with approval policy enabled
- **THEN** the system MUST create an approval ticket, set release state to pending approval, and withhold runtime execution

### Requirement: Shared audit and timeline events
The system MUST persist release timeline events for state transitions, approval decisions, diagnostics, and rollback actions, and SHALL make those events queryable by release identifier.

#### Scenario: Render end-to-end release timeline
- **WHEN** a user opens release timeline in Deployment UI or AI command context
- **THEN** the system MUST return ordered events containing event type, timestamp, actor, and correlated release metadata

### Requirement: Runtime-aware diagnostics visibility
The system MUST return structured diagnostics for failed or interrupted releases and MUST enforce RBAC on diagnostics/timeline query endpoints.

#### Scenario: Deny unauthorized diagnostics access
- **WHEN** a user without required deployment read permission requests diagnostics data
- **THEN** the system MUST return authorization failure and MUST NOT disclose diagnostics payload details

### Requirement: Consistent UI and AI operation flow
The system SHALL provide equivalent operational steps (draft, preview, submit approval, confirm apply, observe status) in Deployment page and AI command center entry, both bound to the same approval and audit chain.

#### Scenario: AI-triggered release follows same governance chain
- **WHEN** an operator initiates release from AI command center
- **THEN** the request MUST produce the same approval metadata and timeline traceability as the Deployment page flow
