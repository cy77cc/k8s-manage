# Refactor Task Breakdown (Fullstack)

## Delivered

1. Migration framework
- added `storage/migration/runner.go`
- added SQL files in `storage/migrations/`
- added CLI: `k8s-manage migrate up|down|status`
- startup bootstrap: run migrations before server start

2. Config/Build
- `app.auto_migrate` added (default false)
- Makefile: `migrate-up`, `migrate-status`, `migrate-down`

3. Service structure
- host refactor:
  - `internal/service/host/routes.go`
  - `internal/service/host/handler/*.go`
  - `internal/service/host/logic/*.go`
- cluster/rbac handlers moved to sub-packages:
  - `internal/service/cluster/handler/resource.go`
  - `internal/service/rbac/handler/permission.go`

4. Host onboarding APIs
- `POST /api/v1/hosts/probe`
- `POST /api/v1/hosts`
- `PUT /api/v1/hosts/:id/credentials`
- model added: `internal/model/host_probe.go`

5. Node compatibility
- `/api/v1/node/add` delegates to host logic
- response headers:
  - `Deprecation: true`
  - `Sunset: Mon, 30 Jun 2026 00:00:00 GMT`

6. Frontend onboarding
- rewrote `web/src/pages/Hosts/HostOnboardingPage.tsx`
- 3-step flow wired to `probe -> create`
- `web/src/api/modules/hosts.ts` added `probeHost` / `updateCredentials`

## Pending (Phase 2)

- split rbac handler into finer-grained files (`permission.go`, `user_role.go`)
- split cluster resource handler into smaller files by domain
- deprecate/remove legacy host create modal in HostList page
