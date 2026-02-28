## 1. OpenSpec And Contract Alignment

- [ ] 1.1 Review and finalize spec wording for `user-access-governance`, `role-aware-navigation-visibility`, and `platform-capability-baseline` deltas with product/security owners.
- [x] 1.2 Validate change artifacts locally with `openspec validate --json` and resolve any schema/scenario format issues.

## 2. Backend RBAC And Audit Enforcement

- [x] 2.1 Inventory governance-related `/api/v1` endpoints for users/roles/permissions and map each endpoint to required Casbin policy actions.
- [x] 2.2 Implement or update backend middleware/logic to enforce role-aware access checks for all governance endpoints.
- [x] 2.3 Standardize unauthorized responses to HTTP 403 with consistent error body for frontend route/API handling.
- [x] 2.4 Add audit logging fields (`actor`, `resource`, `action`, `timestamp`) for denied governance API access attempts.
- [x] 2.5 Add/adjust backend tests for authorized and unauthorized route/API access scenarios.

## 3. Frontend Navigation And Route Guard Refactor

- [x] 3.1 Refactor right-side navigation configuration to introduce an independent governance group with Users, Roles, and Permissions entries.
- [x] 3.2 Remove duplicated governance entries from system settings and add transitional redirect for legacy paths.
- [x] 3.3 Implement role-aware menu rendering from permission snapshot and keep route-guard fallback for manual URL access.
- [x] 3.4 Build a unified access-denied page/state and wire it to route guard and 403 API responses.
- [x] 3.5 Add frontend route and menu visibility tests covering authorized/unauthorized users.

## 4. UI/UX Interaction Standardization

- [x] 4.1 Apply consistent list-detail-edit interaction pattern across Users, Roles, and Permissions pages.
- [x] 4.2 Implement risk-action confirmations (role unbind / permission revoke) with affected-object count preview.
- [x] 4.3 Ensure accessibility baselines (keyboard navigation, visible focus, contrast >= 4.5:1, touch target >= 44px).
- [x] 4.4 Add explicit loading/empty/error states and success feedback for save/authorize operations.

## 5. Release, Migration, And Verification

- [x] 5.1 Introduce feature flag for new governance menu rollout and define fallback to legacy settings entry.
- [ ] 5.2 Execute migration plan: backend-first deployment, frontend rollout, and 30-day legacy redirect observation.
- [ ] 5.3 Verify post-release metrics (403 rate, old-path traffic, governance task completion time) and confirm cutover readiness.
- [ ] 5.4 Remove legacy redirects and stale navigation config after migration completion and stakeholder sign-off.
