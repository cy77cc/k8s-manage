## Migration Notes

### Database
- Added migration: `storage/migrations/20260303_000023_unify_release_trigger_context.sql`
- New columns in `deployment_releases`:
  - `trigger_source` (`manual` by default)
  - `trigger_context_json`
  - `ci_run_id` (`0` by default)
- New indexes:
  - `idx_deploy_release_trigger_source`
  - `idx_deploy_release_ci_run`

### Backward Compatibility
- `/api/v1/services/:id/deploy` is preserved and now maps to unified deployment release orchestration.
- `/api/v1/cicd/releases*` is preserved and now acts as a compatibility facade over deployment releases.
- Legacy release tables remain readable. Query compatibility is implemented by preferring unified releases and falling back/merging legacy records where appropriate.

### Rollback Strategy
- If rollout needs reversal:
  1. Revert application code to previous commit.
  2. Run migration down for `20260303_000023_unify_release_trigger_context.sql`.
  3. Keep legacy endpoints active (already preserved) while validating rollback behavior.

### Operational Checks
- Validate OpenSpec change: `openspec validate --changes --json`
- Backend domain tests:
  - `go test ./internal/service/deployment ./internal/service/service ./internal/service/cicd`
- Frontend checks:
  - `npm --prefix web run test:run -- src/pages/CICD/CICDPage.test.tsx`
  - `npm --prefix web run build`
