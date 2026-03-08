# Frontend/Backend Separate Dev Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make local development run frontend and backend separately while preserving the existing embedded frontend flow for production builds.

**Architecture:** Add an explicit backend development mode and gate static asset registration on that mode. Keep Vite as the frontend dev server, rely on its existing `/api` and `/ws` proxy, and leave production embed behavior unchanged.

**Tech Stack:** Go, Gin, Vite, React, Make

---

### Task 1: Locate and centralize backend runtime mode

**Files:**
- Modify: `internal/config/config.go`
- Modify: `main.go`
- Modify: `internal/service/service.go`

**Step 1: Inspect current config loading path**

Run: `sed -n '1,220p' internal/config/config.go && sed -n '1,220p' main.go`
Expected: find the existing config bootstrap path and the safest place to derive a dev/prod mode flag.

**Step 2: Write the failing test**

Add a focused backend test that asserts static frontend routes are not mounted in development mode.

Suggested target:
- `internal/service/service_static_routes_test.go`

Test cases:
- dev mode request to `/` returns 404
- prod mode request to `/` uses existing frontend fallback behavior or at least does not behave like dev mode

**Step 3: Run the test to verify it fails**

Run: `go test ./internal/service -run TestStaticRouteRegistration -count=1`
Expected: FAIL because the current code always mounts frontend routes.

**Step 4: Implement minimal runtime mode support**

- Introduce a small config/helper for backend runtime mode
- Avoid spreading raw env checks throughout the code
- Make `service.Init` or `registerWebStaticRoutes` depend on a single dev-mode decision

**Step 5: Run the test to verify it passes**

Run: `go test ./internal/service -run TestStaticRouteRegistration -count=1`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/config/config.go main.go internal/service/service.go internal/service/service_static_routes_test.go
git commit -m "feat: add backend dev mode for separate frontend workflow"
```

### Task 2: Gate embedded frontend routes behind non-dev mode

**Files:**
- Modify: `internal/service/service.go`
- Test: `internal/service/service_static_routes_test.go`

**Step 1: Extend the failing test coverage**

Add explicit assertions for:
- dev mode keeps `/api/health` working
- dev mode does not serve SPA fallback
- non-dev mode preserves existing static route behavior

**Step 2: Run the test to verify it fails where coverage is new**

Run: `go test ./internal/service -run TestStaticRouteRegistration -count=1`
Expected: FAIL on one or more newly added assertions.

**Step 3: Implement minimal routing change**

- Only call `registerWebStaticRoutes` when not in dev mode
- Keep WebSocket and API route registration unchanged
- Do not alter production embed loading semantics

**Step 4: Run the test to verify it passes**

Run: `go test ./internal/service -run TestStaticRouteRegistration -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/service.go internal/service/service_static_routes_test.go
git commit -m "feat: skip embedded frontend routes in backend dev mode"
```

### Task 3: Add developer entrypoints for separate startup

**Files:**
- Modify: `Makefile`
- Optionally Modify: `README.md`

**Step 1: Add failing workflow check**

Document the desired commands before implementing:
- `make dev-backend`
- `make dev-frontend`
- optional `make dev`

There may not be an automated test here; verification is command-based.

**Step 2: Implement minimal command surface**

Add:
- `dev-backend`: run backend with explicit dev-mode env
- `dev-frontend`: run `npm run dev` inside `web/`
- optional `dev`: start both if there is a lightweight, non-fragile way in this repo

Prefer keeping `dev-backend` and `dev-frontend` as the primary supported commands even if `dev` is omitted.

**Step 3: Verify commands are correct**

Run:
- `make -n dev-backend`
- `make -n dev-frontend`

Expected:
- backend command includes explicit dev-mode env
- frontend command runs from `web/`

**Step 4: Commit**

```bash
git add Makefile README.md
git commit -m "chore: add separate frontend and backend dev commands"
```

### Task 4: Update documentation for dev vs production workflows

**Files:**
- Modify: `README.md`
- Optionally Modify: `openspec/project.md`

**Step 1: Add development workflow docs**

Document:
- backend dev mode only serves API/WS
- frontend dev server runs independently
- frontend URL and backend URL
- Vite proxy behavior

**Step 2: Preserve production workflow docs**

Keep:
- `make web-build`
- `make build`
- embed-based serving in production

**Step 3: Verify docs are accurate**

Run:
- `sed -n '90,160p' README.md`
- `sed -n '1,140p' openspec/project.md`

Expected: docs clearly separate development and production workflows.

**Step 4: Commit**

```bash
git add README.md openspec/project.md
git commit -m "docs: clarify separate dev workflow and embedded production flow"
```

### Task 5: End-to-end verification

**Files:**
- No new source files expected

**Step 1: Verify backend tests**

Run: `go test ./internal/service -run TestStaticRouteRegistration -count=1`
Expected: PASS

**Step 2: Verify backend still builds**

Run: `go build ./...`
Expected: PASS

**Step 3: Verify frontend dev config still matches backend**

Run: `sed -n '1,220p' web/vite.config.ts`
Expected: `/api` and `/ws` proxy still target backend.

**Step 4: Manual dev verification**

Run in terminal 1:

```bash
make dev-backend
```

Run in terminal 2:

```bash
make dev-frontend
```

Expected:
- backend listens on `127.0.0.1:8080`
- frontend listens on Vite default port
- visiting frontend works without `web/dist` rebuild
- API requests succeed through proxy

**Step 5: Manual production verification**

Run:

```bash
make web-build
make build
```

Expected:
- backend binary still embeds current frontend output

**Step 6: Final commit**

```bash
git add .
git commit -m "feat: support separate frontend and backend dev workflow"
```
