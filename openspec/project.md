# k8s-manage Project Context

## Project Overview

`k8s-manage` is a full-stack Kubernetes/PaaS operations platform.
It combines:

- cluster and host lifecycle management
- service and deployment management
- RBAC and user/auth management
- AI-assisted operations (chat + tool calling + approval flow)

The backend serves API and embedded frontend assets in one process.

## Tech Stack

### Backend

- Language: Go `1.25.x`
- HTTP framework: Gin
- CLI: Cobra
- Config: Viper (+ `.env`/environment variable overrides)
- ORM/DB access: GORM
- Databases: MySQL (primary), optional Postgres/SQLite support
- Cache/session infra: Redis
- AuthN/AuthZ: JWT + Casbin (RBAC)
- Logging: Zap
- API docs: Swag (`docs/swagger.*`)
- K8s integration: `client-go`
- AI orchestration: CloudWeGo Eino + MCP integration (`mcp-go`)

### Frontend

- React `19` + TypeScript
- Vite build tooling
- UI: Ant Design (`antd`)
- Styling: Tailwind CSS (plus custom CSS)
- Data/HTTP: Axios
- Routing: React Router

## Runtime Architecture

- Single Go service process.
- API prefix: `/api/v1`.
- Health endpoint: `/api/health`.
- Frontend built to `web/dist` and embedded via Go `embed`.
- Non-`/api/*` routes fallback to SPA `index.html`.
- Service context wires DB, Redis, K8s clientset, Casbin enforcer, and AI platform agent.

## Repository Structure

- `api/`: request/response contracts and versioned API structs
- `internal/`: core business logic and runtime modules
- `internal/service/<domain>/`: domain modules (typically `routes.go`, `handler/`, `logic/`)
- `storage/`: DB setup and SQL migration runner/files
- `resource/`: SQL and Casbin model resources
- `web/`: frontend application and build assets
- `docs/`: architecture, product, QA, and contract documentation
- `openspec/`: spec-driven change artifacts

## Development Workflow

- Build frontend: `make web-build`
- Build backend: `make build`
- Build both: `make build-all`
- Run locally: `make run`
- Run migrations:
  - `make migrate-up`
  - `make migrate-status`
  - `make migrate-down`

Notes:

- Backend embeds current frontend build output at compile time.
- Bootstrap migrations run at startup; `app.auto_migrate` is for explicit dev usage only.

## Conventions

### Coding Conventions

- Keep backend domain code in `internal/service/<domain>` with clear route/handler/logic separation.
- Keep API contracts versioned under `api/<domain>/v1`.
- Keep middleware cross-cutting concerns in `internal/middleware`.
- Prefer small, focused handlers that delegate business logic to logic/service layers.
- Follow idiomatic Go formatting (`gofmt`) and table-driven tests where practical.

### API Conventions

- Use `/api/v1` namespace for new endpoints.
- Protect business endpoints with JWT auth middleware unless explicitly public.
- Apply RBAC checks consistently (Casbin + permission codes).
- Keep compatibility routes explicit and documented when deprecating paths.

### Data & Migration Conventions

- Use versioned SQL files in `storage/migrations/`.
- Every migration should include both `-- +migrate Up` and `-- +migrate Down`.
- Do not rely on implicit schema changes in production code paths.

### Security Conventions

- Never commit secrets (`.env` is ignored).
- Encrypt sensitive host/cloud key material before persistence (AES-GCM pattern in current codebase).
- Keep JWT, encryption keys, and DB credentials environment-driven.

### Frontend Conventions

- Use TypeScript for page/component modules.
- Keep API calls centralized under `web/src/api/modules`.
- Keep RBAC-aware UI logic in dedicated RBAC components/context.
- Align route/page naming with domain modules (hosts, services, deployment, ai, settings).

### Git & Documentation Conventions

- Commit style follows Conventional Commits (`feat:`, `feat(scope):`, etc.).
- Product and architecture docs are maintained in `docs/` and should be updated with major behavior changes.
- Keep OpenSpec changes (`openspec/changes/*`) aligned with implemented behavior.

## Domain Notes

- Primary domain is platform engineering / DevOps control plane for Kubernetes and host resources.
- AI assistant is not only Q&A; it can trigger tool execution with approval gating for mutating actions.
- Multi-tenant governance is implemented through project/team ownership and RBAC permission models.
