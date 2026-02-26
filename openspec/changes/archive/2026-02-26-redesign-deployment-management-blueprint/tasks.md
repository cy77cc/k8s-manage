## 1. Blueprint Baseline

- [x] 1.1 Publish deployment management capability map (Target/Release/Governance/Observability/AI Bridge)
- [x] 1.2 Define cross-runtime abstraction contract for Kubernetes and Compose release execution
- [x] 1.3 Define phased milestones and acceptance checklist for blueprint rollout

## 2. Lifecycle and API Contract Alignment

- [x] 2.1 Standardize release lifecycle states and response fields across deployment APIs
- [x] 2.2 Align approval-gated flow contract for production releases (ticket metadata, decision lifecycle)
- [x] 2.3 Align rollback contract and compatibility mapping for historical states

## 3. Approval and Audit Governance

- [x] 3.1 Define unified approval policy matrix (environment, action, approver role, expiry)
- [x] 3.2 Define release audit/timeline event taxonomy and required payload fields
- [x] 3.3 Define RBAC requirements for release query, diagnostics, and timeline access

## 4. UI and Command-Center Experience

- [x] 4.1 Redesign deployment page interaction blueprint for lifecycle, approval, and rollback actions
- [x] 4.2 Define timeline/diagnostics presentation contract for release detail views
- [x] 4.3 Define AI command-center operation flow that reuses deployment approval and audit chain

## 5. Delivery and Verification Plan

- [x] 5.1 Plan backend implementation sequence by domain (`deployment`, `cicd`, `ai`) with migration checkpoints
- [x] 5.2 Plan frontend implementation sequence (`DeploymentPage`, API module, command-center entry)
- [x] 5.3 Define validation checklist (`openspec validate`, API regression, UI state regression, approval-path tests)
- [x] 5.4 Sync roadmap/progress documentation with blueprint phase status
