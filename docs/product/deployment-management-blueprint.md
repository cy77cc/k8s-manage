# Deployment Management Blueprint

## 1. Capability Map

### 1.1 Domains

| Domain | Scope | Key APIs |
| --- | --- | --- |
| Target | Deployment endpoint modeling for `k8s|compose`, environment binding, host/cluster association | `/api/v1/deploy/targets*`, `/api/v1/environments*` |
| Release | Preview, apply, approval, rollback, lifecycle state machine | `/api/v1/deploy/releases/*` |
| Governance | Approval policy, risk classification, RBAC gate, expiry strategy | `/api/v1/approvals/inbox*` |
| Observability | Release timeline, diagnostics payload, install logs, runtime status | `/api/v1/deploy/releases/:id/timeline`, `/api/v1/environments/:id/install/tasks/*` |
| AI Bridge | Command-center intent flow reusing release and approval contracts | `/api/v1/ai/*`, `/api/v1/deploy/*`, `/api/v1/approvals/*` |

### 1.2 Runtime Abstraction Contract

The platform defines a runtime-neutral release contract:

1. Input: service + target + env + strategy + variables
2. Mandatory preview: materialize release draft and validate risk/checks
3. Apply confirmation: accepted only with valid preview reference
4. Approval gate (production risk): pending decision before runtime execution
5. Runtime adapter execution:
   - K8s adapter: kubeconfig/client-go or kubeadm-provisioned cluster
   - Compose adapter: SSH + docker compose executor
6. Unified timeline and diagnostics output

Runtime-specific differences are isolated in adapters and install scripts; lifecycle and policy semantics are shared.

## 2. Lifecycle and Response Standard

### 2.1 Canonical Lifecycle

`preview -> pending_approval/approved -> applying -> applied|failed -> rollback`

Compatibility:
- Existing `succeeded` is mapped to `applied`
- Existing `rolled_back` is mapped to `rollback`

### 2.2 Standard Response Fields

Release APIs SHALL expose:
- `release_id`
- `status`
- `lifecycle_state`
- `runtime_type`
- `approval_required`
- `approval_ticket`
- `preview_ref`
- `diagnostics_json`

## 3. Approval and Audit Governance

### 3.1 Unified Approval Policy Matrix

| Environment | Action | Required Role | Ticket Expiry | SLA |
| --- | --- | --- | --- | --- |
| development | apply | `deploy:release:apply` | optional | best effort |
| staging | apply/rollback | `deploy:release:apply` | optional | 2h |
| production | apply/rollback/install | `deploy:release:approve` | 30m | P1/P0 |

### 3.2 Event Taxonomy

Core events:
- `release.previewed`
- `release.pending_approval`
- `release.approved`
- `release.rejected`
- `release.applying`
- `release.applied`
- `release.failed`
- `release.rollback_requested`
- `release.rollback_completed`
- `release.rollback_failed`
- `environment.install_step_started|succeeded|failed`

Required event payload fields:
- `event_id`, `action`, `resource_type`, `resource_id`
- `project_id`, `team_id`, `environment_id`
- `actor_id`, `source`
- `status`, `occurred_at`
- `detail` object with runtime-specific context

### 3.3 RBAC Baseline

- Query permissions:
  - `deploy:release:read`
  - `deploy:k8s:read`
  - `deploy:compose:read`
- Mutating permissions:
  - `deploy:release:apply`
  - `deploy:release:rollback`
  - `deploy:release:approve`
  - `deploy:k8s:apply|rollback`
  - `deploy:compose:apply|rollback`
- Inbox permissions:
  - `deploy:approval:read`
  - `deploy:approval:review`

## 4. Entry Experience Blueprint

### 4.1 Default Entry

Default for project-group users is task-oriented flow:
1. choose target/environment
2. preview
3. confirm
4. track status/timeline

Advanced platform controls remain secondary paths.

### 4.2 AI + UI Shared Flow

Both entry points SHALL reuse one release contract and one approval inbox:

1. AI or UI creates preview
2. apply request references preview
3. production requests generate ticket in global inbox
4. approver decides in inbox
5. execution and timeline update both entry points

### 4.3 Timeline and Diagnostics Presentation Contract

Release detail views SHALL display:
- lifecycle timeline (ordered events)
- diagnostics summary + raw detail payload
- approval ticket state and comments
- rollback origin and result

## 5. Environment Deployment Blueprint (Compose + K8s)

### 5.1 Goal

Support environment provisioning and runtime deployment over SSH with offline/binary artifacts from `script/`.

### 5.2 Script Bundle Structure

`script/` SHALL include:
- `manifests/index.yaml`
- `bin/linux-amd64/*`
- `images/*` (for offline k8s/cni bundles)
- `templates/*`
- `hooks/*`
- `checksums/SHA256SUMS`

### 5.3 Cluster Access Model

- `platform_managed`: cluster created by platform, kubeconfig/cert refs persisted for direct access
- `external_managed`: imported by kubeconfig or cert+endpoint/token

### 5.4 Optional LVS/VIP Capability

Cluster creation supports optional `lvs_vip` mode:
- default disabled
- enabled flow is post-cluster async task
- recommended mode: single VIP + keepalived/ipvsadm + health check `:6443` + retry(3)

