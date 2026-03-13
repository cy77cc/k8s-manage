# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Development (frontend + backend separate)
make dev-backend    # Start backend at http://127.0.0.1:8080
make dev-frontend   # Start frontend at http://127.0.0.1:5173

# Production build
make build-all      # Build frontend + backend (embeds web/dist)
make run            # Run the compiled binary

# Testing
make test                      # Run all Go tests
make test-coverage             # Generate coverage report
make test-coverage-check       # Fail if coverage < 40%
go test ./internal/service/ai/... -v    # Run specific package tests
make web-test                  # Run frontend tests

# Database migrations
make migrate-up      # Apply migrations
make migrate-status  # Check migration status
make migrate-down    # Rollback one step
```

## Architecture Overview

This is a PaaS platform (OpsPilot) for K8s cluster management with an AI assistant. The stack is:
- **Backend**: Go 1.26 + Gin + GORM + Cobra CLI
- **Frontend**: React 19 + Vite + Ant Design 6 + @ant-design/x-sdk
- **AI**: CloudWeGo Eino framework for LLM orchestration

### Backend Layered Architecture

```
internal/
├── cmd/           # CLI entrypoint (cobra commands)
├── config/        # Viper config loading
├── server/        # HTTP server setup
├── middleware/    # Auth, CORS, rate limiting
├── model/         # GORM models (organized by domain)
├── dao/           # Data access layer (CRUD only)
├── service/       # HTTP handlers + business logic
│   └── <domain>/  # Each domain has routes.go + handler/ + logic/
├── httpx/         # HTTP response helpers
├── xcode/         # Business error codes
└── ai/            # AI orchestration (see below)
```

### AI Module Architecture (`internal/ai/`)

The AI module implements multi-stage orchestration:

```
Rewrite → Plan → Execute → Summarize
```

Key components:
- **orchestrator.go**: Main pipeline orchestrator
- **rewrite/**: Query rewriting with RAG context
- **planner/**: Task decomposition
- **executor/**: Expert-based execution
- **experts/**: Domain experts (hostops, k8s, service, observability, delivery)
- **tools/**: Domain-organized tools (kubernetes/, host/, service/, monitor/)
- **aiv2/**: Single-agent runtime (ChatModelAgent + Runner)

**AIV2** is the new simplified runtime using Eino ADK's `Interrupt/ResumeWithParams` for human-in-the-loop. Toggle via `feature_flags.ai_assistant_v2` in config.

### Frontend Structure

```
web/src/
├── api/modules/<domain>/   # API clients by domain
├── components/<Feature>/   # Feature-grouped components
├── pages/                  # Route pages
├── hooks/<concern>/        # Custom hooks by concern
└── utils/<purpose>/        # Utilities by purpose
```

## Key Conventions

### HTTP Response Format

All business APIs return HTTP 200 with unified body:
```json
{ "code": 1000, "msg": "请求成功", "data": {...} }
```

Use `internal/httpx` functions:
- `httpx.OK(c, data)` - Success (code 1000)
- `httpx.Fail(c, xcode.XXX, "message")` - Business error
- `httpx.BindErr(c, err)` - Parameter binding error

Never use `c.JSON()` with inline `gin.H{}`.

### Code Organization

Directory file count thresholds:
| Category | Max Files |
|----------|-----------|
| Models | 10 |
| Handlers | 8 |
| Components | 12 |
| API Modules | 15 |

When exceeded, split into subdirectories by domain/feature.

### Database

- **No foreign keys** - Relationships handled at application layer
- Use `storage/migration/` for versioned migrations
- Auto-migrate available via `app.auto_migrate: true` in dev

### Comment Style

All Go code MUST have four-level Chinese comments:
1. Package-level (before `package` declaration)
2. Type/struct-level
3. Method/function-level (with parameters, returns, side effects)
4. Inline (explain "why", not "what")

Example:
```go
// Package orchestrator 实现 AI 编排核心逻辑。
//
// 架构概览:
//   Rewrite → Plan → Execute → Summarize
package ai

// Run 启动编排流水线。
//
// 参数:
//   - ctx: 上下文
//   - req: 请求参数
//
// 返回: 成功返回 nil，失败返回错误
func (o *Orchestrator) Run(ctx context.Context, req RunRequest) error
```

### Testing

- Table-driven tests preferred
- Test files alongside source (not separate `test/` directory)
- Coverage threshold: 40% minimum

## Configuration

Config file: `configs/config.yaml` (set via `--config` flag)

Environment variables (`.env` file):
- `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USER`, `MYSQL_PASSWORD`
- `REDIS_ADDR`, `REDIS_PASSWORD`
- `LLM_API_KEY`
- `MILVUS_HOST`, `MILVUS_PORT`, `MILVUS_DATABASE`, `MILVUS_COLLECTION`
- `PROMETHEUS_ADDRESS`

Feature flags in config:
- `feature_flags.ai_assistant_v2: true` - Use AIV2 runtime
- `feature_flags.ai_governed_host_execution: true` - AI approval gates

## API Endpoints

- All API routes prefixed with `/api/v1`
- Frontend served at `/` (embedded in production)
- SSE streaming for AI chat: `/api/v1/ai/chat`, `/api/v1/ai/resume/step/stream`

## Specs and Documentation

Architecture specs are in `openspec/specs/<feature>/spec.md`. Key specs:
- `code-organization-convention` - Directory structure rules
- `code-comment-convention` - Comment format requirements
- `http-response-convention` - API response format
- `ai-module-architecture` - AI pipeline details

User docs in `docs/`:
- `docs/ai/` - AI knowledge base for RAG
- `docs/user/` - Help center and FAQ
