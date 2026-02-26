## 1. Data model and API contract baseline

- [x] 1.1 Add/adjust release and approval persistence fields + indexes in `storage/migrations` for lifecycle, preview token metadata, and timeline correlation IDs
- [x] 1.2 Define unified release lifecycle DTO/response contract in `api/*/v1` and shared domain model used by deployment query/apply/rollback APIs
- [x] 1.3 Add compatibility mapping for historical release states to normalized lifecycle states in backend query layer

## 2. Backend lifecycle and governance implementation

- [x] 2.1 Implement preview token validation (context hash, TTL, parameter match) and block apply when preview is missing/expired/mismatched
- [x] 2.2 Implement production approval gate flow (ticket create, pending state, decision transition, approver identity capture)
- [x] 2.3 Implement runtime-aware execution transition (`approved -> applying -> applied|failed|rollback`) for Kubernetes and Compose adapters

## 3. Audit, diagnostics, and RBAC enforcement

- [x] 3.1 Implement unified timeline event writer for state changes, approval decisions, diagnostics, and rollback actions
- [x] 3.2 Expose release timeline + diagnostics query endpoints with normalized payload and release correlation fields
- [x] 3.3 Enforce Casbin permission checks for release detail, diagnostics, and timeline APIs and add denial-path tests

## 4. Frontend deployment and AI flow alignment

- [x] 4.1 Upgrade `web/src/api/modules/deployment.ts` to consume new lifecycle, preview-required, approval, diagnostics, and timeline contracts
- [x] 4.2 Refactor Deployment page flow to enforce `draft -> preview -> approval/confirm -> apply` with rollback and diagnostics visibility
- [x] 4.3 Align AI command-center deployment actions to reuse unified approval queue and timeline/audit rendering

## 5. Validation and rollout

- [x] 5.1 Add backend regression tests for preview gate, approval gate, lifecycle transitions, rollback path, and timeline payload integrity
- [ ] 5.2 Add frontend state regression coverage for release status rendering, approval pending states, and diagnostics/timeline panels
- [x] 5.3 Run `openspec validate --json` and sync `docs/roadmap.md` + `docs/progress.md` with implementation phase status
