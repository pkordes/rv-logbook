# Contributing to RV Logbook

This file covers everything you need to develop on this project locally.
For project purpose and architecture, see [README.md](README.md) and
[docs/architecture.md](docs/architecture.md) (added in Phase 17).

---

## Prerequisites

Install all of the following before attempting to run or build anything.

| Tool | Minimum Version | Purpose | Install |
|------|----------------|---------|---------|
| Go | 1.24 | Backend language | `winget install GoLang.Go` (Windows) or [go.dev/dl](https://go.dev/dl) |
| Node.js | 22 LTS | Frontend build tooling | [nodejs.org](https://nodejs.org) |
| npm | 10+ | Frontend package manager | Bundled with Node |
| Podman | 5+ | Runs local Postgres container | [podman.io](https://podman.io) |
| podman-compose | 1.5+ | Orchestrates containers via compose file | `pip install podman-compose` |
| make | 3.81+ | Task runner (Makefile) | `winget install GnuWin32.Make` (Windows) |
| Git | any recent | Version control | [git-scm.com](https://git-scm.com) |

### Go CLI Tools

After installing Go, add `$GOPATH/bin` (default: `~/go/bin`) to your `PATH`,
then install:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
go install github.com/oasdiff/oasdiff@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install gotest.tools/gotestsum@latest
```

`gotestsum` replaces bare `go test` for `make backend/test`. It produces a readable per-package summary and a pass/fail count. All `go test` flags (e.g. `-race`, `-count=1`) are passed through after `--`.

### VS Code Extensions (recommended)

| Extension ID | Purpose |
|-------------|---------|
| `golang.go` | Go IntelliSense, vet on save, test runner |
| `dbaeumer.vscode-eslint` | TypeScript/JS linting |
| `esbenp.prettier-vscode` | Auto-format on save |
| `bradlc.vscode-tailwindcss` | Tailwind class autocomplete |

---

## Environment Variables

Copy `.env.example` to `.env` at the repo root and fill in any required values:

```bash
cp .env.example .env
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | no | `8080` | Port the Go API listens on |
| `DATABASE_URL` | yes | — | Postgres connection string for dev DB |
| `TEST_DATABASE_URL` | yes | — | Postgres connection string for test DB (integration tests) |
| `LOG_LEVEL` | no | `info` | `debug`, `info`, `warn`, `error` |
| `CORS_ORIGINS` | no | `http://localhost:5173` | Comma-separated list of allowed CORS origins |
| `MAX_BODY_BYTES` | no | `1048576` (1 MiB) | Maximum request body size; larger bodies get HTTP 413 |

> `.env` is gitignored. Never commit real credentials.
> The defaults in `.env.example` match the `docker-compose.yml` credentials and work out of the box.

---

## Local Setup

### Start the database

```bash
# Start Postgres in a container (runs in background)
podman-compose up -d

# Run all pending migrations
make db/migrate
```

### Run the backend

```bash
make backend/run
# API available at http://localhost:8080
# Health check: curl http://localhost:8080/healthz
```

### Run the frontend (dev server)

```bash
make frontend/dev
# UI available at http://localhost:5173
# Proxies /api requests to the backend automatically
```

---

## Common Commands

| Command | What it does |
|---------|-------------|
| `make help` | List all available targets |
| `make backend/run` | Start the Go API server |
| `make backend/build` | Compile Go binary to `backend/bin/api` |
| `make backend/check` | Compile all packages without producing a binary (fast refactor check) |
| `make backend/test` | Run all Go tests — all packages, DB required |
| `make backend/test/service` | Run service-layer unit tests only (no DB — fast TDD loop) |
| `make backend/test/handler` | Run handler-layer unit tests only (no DB — fast TDD loop) |
| `make backend/lint` | Run `go vet` + `staticcheck` |
| `make backend/generate` | Regenerate Go stubs from `openapi.yaml` |
| `make frontend/dev` | Start Vite dev server (port 5173) |
| `make frontend/build` | Build production bundle to `frontend/dist/` |
| `make frontend/test` | Run Vitest unit tests (single run) |
| `make frontend/lint` | Run ESLint |
| `make db/up` | Start the Postgres container (background) |
| `make db/down` | Stop containers — data volume persists |
| `make db/migrate` | Apply pending migrations (`goose up`) |
| `make db/rollback` | Roll back last migration (`goose down`) |
| `make db/reset` | Wipe dev DB and re-apply all migrations |

---

## Branching & PR Workflow

- **`main`** — stable, always deployable. Direct commits permitted during initial
  project scaffolding (Phase 0) only.
- **Feature branches** — required from Phase 1 onward. Branch from `main`,
  name as `feat/short-description` or `fix/short-description`.
- **CI is tiered** — `backend.yml` / `frontend.yml` run fast checks on every branch
  push (vet, build, unit tests, lint). `backend-pr.yml` / `frontend-pr.yml` run
  those same checks *plus* PR-only gates (oasdiff, and integration tests in Phase 8)
  when a PR targets `main`.
- **PRs require CI to pass** — lint, unit tests, and the `oasdiff` breaking-change
  check must all be green before merging.
- **Breaking API changes** — if you intentionally change the OpenAPI contract in a
  breaking way, the PR description must explicitly acknowledge it. The `oasdiff`
  CI gate will fail and must be overridden deliberately.

---

## Code Style

### Go
- Follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- `go vet` and `staticcheck` must pass with zero warnings
- Every exported function and type has a doc comment
- No `TODO` comments in committed code — open a GitHub issue instead

### TypeScript / React
- `strict: true` in `tsconfig.json` — no `any`
- Named exports only (no default exports)
- Component files are named in PascalCase, all others in camelCase
- Prettier handles all formatting — do not manually adjust whitespace

---

## API Conventions

### Pagination

All list endpoints support cursor-free page/offset pagination via query parameters:

| Parameter | Default | Max  | Description              |
|-----------|---------|------|--------------------------|
| `page`    | `1`     | —    | 1-based page number      |
| `limit`   | `20`    | `100`| Items per page           |

Paginated responses use a consistent envelope shape:

```json
{
  "data": [ ... ],
  "pagination": { "page": 1, "limit": 20, "total": 42 }
}
```

`total` is the count of all matching records (not just the current page), enabling
clients to calculate the number of pages: `ceil(total / limit)`.

The defaults and the `limit` cap are enforced in `domain.NewPaginationParams`.

---

## Architecture Overview

See [docs/architecture.md](docs/architecture.md) for the full layer diagram and
data model. A quick summary:

```
openapi.yaml  →  oapi-codegen  →  handler  →  service  →  repo  →  Postgres
                (generates interface)  (implements it)
```

The OpenAPI spec is the source of truth. Never change handler signatures manually —
change the spec and regenerate.
