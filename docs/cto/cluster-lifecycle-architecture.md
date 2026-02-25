# Cluster Lifecycle Architecture (Phase-1)

## Scope

Phase-1 implements lifecycle + namespace multi-tenant isolation + RBAC guard.

## Control Plane Components

- Cluster API handlers: `internal/service/cluster/handler/*`
- Policy layer: permission check + namespace binding + production approval gate
- Resource execution:
  - core resources via `client-go` typed client
  - Argo Rollouts via `dynamic` client (`argoproj.io/v1alpha1`, `rollouts`)

## Tenant Isolation Model

- Team to namespace relation: many-to-many via `cluster_namespace_bindings`
- Non-admin read/write operations must pass namespace binding checks
- `readonly=true` blocks write operations

## Production Safety

- Production namespace action (`deploy/rollback`) requires `k8s:approve` or approved token
- Approval tickets persisted in `cluster_deploy_approvals`

## Audit

- Mutating operations write to `cluster_operation_audits` for traceability
