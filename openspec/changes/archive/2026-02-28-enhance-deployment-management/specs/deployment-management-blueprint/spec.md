## MODIFIED Requirements

### Requirement: Deployment blueprint SHALL define runtime abstraction contract
The blueprint SHALL define a runtime abstraction contract so Kubernetes and Compose share one release workflow while supporting runtime-specific execution adapters, and SHALL include environment bootstrap semantics (SSH installation, package validation, and runtime health verification) before release operations.

#### Scenario: Runtime-neutral release workflow
- **WHEN** an operator triggers preview/apply/rollback for different runtimes
- **THEN** the system SHALL keep consistent lifecycle semantics and SHALL delegate execution to runtime-specific adapters

#### Scenario: Runtime-neutral environment bootstrap workflow
- **WHEN** an operator creates a new environment for runtime `k8s` or `compose`
- **THEN** the system SHALL execute a shared bootstrap state model and invoke runtime-specific installation adapters with unified audit output

## ADDED Requirements

### Requirement: Deployment blueprint SHALL define cluster access source model
The blueprint SHALL define a cluster access source model that distinguishes platform-managed and external-managed credentials and enforces a consistent security and governance chain.

#### Scenario: Unified cluster source governance
- **WHEN** a deployment target is bound to cluster credentials from either source model
- **THEN** the system SHALL apply the same RBAC, approval, and timeline policies regardless of credential origin
