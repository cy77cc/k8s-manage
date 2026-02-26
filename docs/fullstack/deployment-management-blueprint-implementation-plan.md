# Deployment Management Blueprint Implementation Plan

## 1. Backend Sequence

### 1.1 Deployment Domain (`internal/service/deployment`)

1. Standardize lifecycle states and compatibility mapping:
   - canonical: `preview/pending_approval/approved/applying/applied/failed/rollback`
   - compatibility read mapping for historical rows: `succeeded -> applied`, `rolled_back -> rollback`
2. Enforce preview-first apply:
   - apply must include `preview_ref`
   - reject with `preview_required|preview_expired|preview_mismatch`
3. Extend release response contract:
   - include `approval_required`, `approval_ticket`, `lifecycle_state`, `preview_ref`
4. Extend timeline/audit persistence:
   - event taxonomies and required payload fields

Migration checkpoints:
- DB migration for preview reference and timeline enrichments (if needed)
- Backfill/mapping rules for historical states

### 1.2 Approval Inbox Domain (new global approval module)

1. Introduce inbox aggregate entities:
   - `approval_inbox_items`
   - `approval_inbox_actions`
2. Unify approval source for UI and AI:
   - ticket created once
   - single review endpoints
3. Support filtering and assignment:
   - `mine/pending/all`
   - project/team/risk/priority/status dimensions

Migration checkpoints:
- add inbox tables
- migrate legacy ticket references from scattered domain tables into normalized links

### 1.3 Environment Provisioning Domain (new)

1. Introduce environment model:
   - `environments`, `environment_nodes`
2. Introduce install orchestration model:
   - `environment_install_tasks`, `environment_install_logs`
3. Introduce artifacts and script bundle validation:
   - `environment_artifacts` + manifest checksum validation
4. Introduce cluster connection model:
   - `cluster_connections` for `platform_managed|external_managed`
5. Support optional `lvs_vip` post-create task

Migration checkpoints:
- create environment and install task tables
- create cluster connection table
- compatibility read for existing `clusters.kubeconfig/ca_cert/token`

## 2. Frontend Sequence

### 2.1 Deployment Page (`web/src/pages/Deployment/DeploymentPage.tsx`)

1. Align lifecycle labels and colors to canonical states
2. Enforce preview-confirm-apply interaction contract
3. Expose approval-required states and jump-to-inbox actions
4. Improve timeline + diagnostics detail panel structure

### 2.2 API Module (`web/src/api/modules/deployment.ts`)

1. Align type contracts with backend response fields
2. Add preview reference required for apply
3. Add global approval inbox API bindings

### 2.3 AI Command Center Entry (`web/src/pages/AI/*`)

1. Surface release and approval progress from shared ticket/release IDs
2. Reuse inbox and release timelines instead of custom local state

## 3. Script/Runtime Delivery Plan

1. Define and validate `script/manifests/index.yaml`
2. Implement runtime bundle checks for k8s and compose
3. Execute hooks through SSH with structured step status/log capture
4. Support optional `lvs_vip` hooks in post-create phase

## 4. Cross-Domain Validation Gates

1. API contract gate:
   - release/apply/approval/inbox/environment endpoints must pass schema checks
2. Security gate:
   - RBAC for release read/apply/approve and inbox review
3. Observability gate:
   - timeline completeness and diagnostics payload integrity
4. Compatibility gate:
   - old release status rows render with canonical mapping

