# Deployment Management Blueprint Regression Checklist

## 1. OpenSpec Validation

- [ ] `openspec validate --changes --json` passes
- [ ] `openspec validate --specs --json` passes
- [ ] change artifacts status is `done` for proposal/design/specs/tasks

## 2. API Contract Regression

### 2.1 Release lifecycle

- [ ] preview API returns canonical fields and `preview_ref`
- [ ] apply without preview returns `preview_required`
- [ ] apply with expired preview returns `preview_expired`
- [ ] apply with mismatch params returns `preview_mismatch`
- [ ] production apply enters `pending_approval`
- [ ] approved flow transitions to `applying -> applied`
- [ ] reject flow transitions to `rejected` and does not execute runtime apply

### 2.2 Approval inbox

- [ ] ticket generated from UI and AI routes to same inbox list
- [ ] approve/reject updates release status and writes audit actions
- [ ] inbox list filtering works for `mine/pending/all`

### 2.3 Timeline and diagnostics

- [ ] release timeline includes required event chain
- [ ] diagnostics are available in detail responses for failed runtime execution

## 3. UI State Regression

- [ ] deployment page lifecycle tags match canonical state names
- [ ] pending approval rows show review actions
- [ ] detail dialog shows timeline + diagnostics
- [ ] command center can track approval and execution status by ticket/release

## 4. Approval-path Regression

- [ ] user without approval permission cannot review ticket
- [ ] approver can process ticket in global inbox
- [ ] approval decision comments are persisted and visible in timeline

## 5. Environment Deployment Regression

- [ ] environment install preview validates script manifest and required artifacts
- [ ] environment install apply creates task and step logs
- [ ] k8s platform-managed cluster persists connection material for direct access
- [ ] external cluster import by kubeconfig works
- [ ] external cluster import by cert/token works
- [ ] optional `lvs_vip` mode creates post-create task and reports status

## 6. Failure-path Regression

- [ ] `artifact_missing`
- [ ] `ssh_unreachable`
- [ ] `kubeconfig_invalid`
- [ ] `approval_ticket_expired`
- [ ] `release_rollback_failed`

