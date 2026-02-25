# Cluster API Contract (Phase-1)

## New Endpoints

- Namespaces
  - `GET /api/v1/clusters/:id/namespaces`
  - `POST /api/v1/clusters/:id/namespaces`
  - `DELETE /api/v1/clusters/:id/namespaces/:name`
  - `GET /api/v1/clusters/:id/namespaces/bindings`
  - `PUT /api/v1/clusters/:id/namespaces/bindings/:teamId`
- Rollouts
  - `GET /api/v1/clusters/:id/rollouts`
  - `POST /api/v1/clusters/:id/rollouts/preview`
  - `POST /api/v1/clusters/:id/rollouts/apply`
  - `POST /api/v1/clusters/:id/rollouts/:name/promote|abort|rollback`
- HPA
  - `GET /api/v1/clusters/:id/hpa`
  - `POST /api/v1/clusters/:id/hpa`
  - `PUT /api/v1/clusters/:id/hpa/:name`
  - `DELETE /api/v1/clusters/:id/hpa/:name?namespace=...`
- Quota/LimitRange
  - `GET /api/v1/clusters/:id/quotas`
  - `POST|PUT /api/v1/clusters/:id/quotas/:name?`
  - `DELETE /api/v1/clusters/:id/quotas/:name?namespace=...`
  - `GET /api/v1/clusters/:id/limit-ranges`
  - `POST /api/v1/clusters/:id/limit-ranges`
- Approval
  - `POST /api/v1/clusters/:id/approvals`
  - `POST /api/v1/clusters/:id/approvals/:ticket/confirm`

## Response Envelope

- Success: `{ code: 1000, msg: "ok", data: ... }`
- List: `{ code: 1000, msg: "ok", data: { list: [], total: n } }`
- Forbidden: `{ code: 2004, msg: "forbidden" | detailed reason }`

## Frontend Mapping

- `web/src/api/modules/kubernetes.ts` contains typed wrappers and list normalization.
- K8s page split panels under `web/src/components/K8s/`.
