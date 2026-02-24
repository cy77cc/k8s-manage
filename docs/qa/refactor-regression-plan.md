# Refactor Regression Plan (QA)

## Migration

- `make migrate-up` on empty DB: pass
- `make migrate-status`: shows applied/pending correctly
- incremental upgrade on existing DB: pass

## Host Onboarding

- probe success -> create success
- probe fail -> create blocked (non-admin)
- expired/consumed token -> `probe_expired`/`probe_not_found`
- credentials update fail keeps old host credentials

## Node Compatibility

- `/api/v1/node/add` still works
- response has `Deprecation` and `Sunset` headers
- created row consistent with `/api/v1/hosts` schema

## Frontend

- `HostOnboardingPage` can complete 3-step flow
- probe failure supports retry path
- admin sees force option; non-admin hidden

## Global Regression

- `go test ./...`
- `cd web && npm run build`
- spot-check login, hosts list/detail, ai chat, rbac pages
