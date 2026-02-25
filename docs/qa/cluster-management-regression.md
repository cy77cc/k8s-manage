# Cluster Management Regression Plan (Phase-1)

## Backend Cases

1. Namespace binding isolation
- non-admin without binding cannot read/write namespace resources.
- readonly binding cannot perform write operations.

2. Rollout lifecycle
- preview/apply succeeds when CRD exists.
- CRD missing returns `rollout_crd_missing`.
- promote/abort/rollback returns action result or explicit CLI missing message.

3. HPA lifecycle
- create/update/delete HPA with CPU/MEM metrics.
- invalid metrics payload rejected.

4. Quota/LimitRange lifecycle
- create/update/delete quotas.
- create/update limit ranges with quantity validation.

5. Production gate
- deploy/rollback on prod namespace requires `k8s:approve` or approved token.

## Frontend Cases

1. K8s drawer tabs show Namespaces/Rollouts/HPA/Quotas panels.
2. Namespace create and binding update flows refresh list.
3. Rollout preview and apply flow updates rollout list.
4. HPA and Quota editor create/update/delete flows work.

## Build Gates

- `go test ./...`
- `cd web && npm run build`
