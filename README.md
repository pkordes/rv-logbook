# RV Logbook — Go + React

[![Backend](https://github.com/pkordes/rv-logbook/actions/workflows/backend.yml/badge.svg)](https://github.com/pkordes/rv-logbook/actions/workflows/backend.yml)
[![Backend PR](https://github.com/pkordes/rv-logbook/actions/workflows/backend-pr.yml/badge.svg)](https://github.com/pkordes/rv-logbook/actions/workflows/backend-pr.yml)
[![Frontend](https://github.com/pkordes/rv-logbook/actions/workflows/frontend.yml/badge.svg)](https://github.com/pkordes/rv-logbook/actions/workflows/frontend.yml)
[![Frontend PR](https://github.com/pkordes/rv-logbook/actions/workflows/frontend-pr.yml/badge.svg)](https://github.com/pkordes/rv-logbook/actions/workflows/frontend-pr.yml)

A production-style web application for full-time RV life: log trips and stops,
tag campgrounds for quick recall, view a trip timeline, and export your travel
history to CSV or JSON.

Built as an explicit demonstration of senior-level engineering across a full
Go + React stack — every decision is deliberate and documented.

---

## Features

- **Trip management** — create, edit, and delete trips with date ranges and notes
- **Stop tracking** — log every campground or park with location, dates, cost,
  site number, rating, and free-text notes
- **Tagging** — apply and reuse tags (e.g. `quiet`, `good-cell`, `50-amp`) across
  stops; edit or delete tags globally from the Tags page
- **Timeline view** — visualize stops on a trip as a date-ordered timeline
- **Paginated lists** — all collections support `?page=` and `?limit=` parameters
- **Export** — download full travel history as CSV or JSON from a single endpoint
- **Live API docs** — interactive Scalar UI served at `/docs` with the OpenAPI spec

---

## Architecture Overview

```
openapi.yaml  (single source of truth — all types derived from this)
     │
     └─ oapi-codegen → StrictServerInterface + request/response types
                            │
HTTP Request ──► handler  (implements compiler-enforced interface)
                    │
                 service  (business rules; unit-tested with mocks)
                    │
                   repo   (all SQL; integration-tested against real DB)
                    │
                 domain   (plain structs + sentinel errors; zero deps)
```

Frontend:

```
Page  →  feature  →  TanStack Query hooks  →  api/ (typed fetch wrappers)
                                                │
                                          components/ (dumb UI primitives)
```

See [docs/architecture.md](docs/architecture.md) for the full layer diagram and
data model ERD.

---

## Tech Choices

| Concern | Choice | Why |
|---|---|---|
| Go router | `chi` | Lightweight; accepts stdlib `http.Handler` — no lock-in |
| API contract | `oapi-codegen` (spec-first) | `openapi.yaml` is ground truth; compiler enforces interface conformance |
| Breaking-change CI | `oasdiff` | Fails PRs that break the contract; running on branch has no useful baseline |
| DB driver | `pgx/v5` + raw SQL | Explicit queries; no ORM magic hiding N+1s |
| Migrations | `goose` | SQL-first; embedded in the binary for zero-dep deployment |
| Logging | `log/slog` | stdlib structured logging since Go 1.21; no external dep |
| Config | Hand-written env loader | ~20 lines; deliberate restraint over framework magic |
| React build | Vite | Sub-second HMR; first-class TypeScript support |
| Server state | TanStack Query v5 | Industry standard; clean separation of server vs UI state |
| Forms | React Hook Form + Zod | Zod schema is both the type source and runtime validator |
| Styling | Tailwind + shadcn/ui | Utility-first with accessible primitives; no context switching |
| E2E tests | Playwright | Cross-browser; stable `data-testid` / ARIA selector strategy |
| Security scanning | `gosec` (branch), `govulncheck` + CodeQL (PR) | Layered: fast AST scan on every push, deep taint analysis before merge |

Design decisions are documented as Architecture Decision Records in
[docs/adr/](docs/adr/).

---

## Running Locally

### Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | ≥ 1.25 | https://go.dev/dl |
| Node | ≥ 22 (see `.nvmrc`) | https://nodejs.org |
| Podman / Docker | any | https://podman.io |
| podman-compose | any | `pip install podman-compose` |
| goose | latest | `go install github.com/pressly/goose/v3/cmd/goose@latest` |

### Quick start

```bash
# 1. Clone
git clone https://github.com/pkordes/rv-logbook.git
cd rv-logbook

# 2. Copy environment config
cp .env.example .env
# Edit .env if you need to change DB credentials

# 3. Start Postgres
make db/up

# 4. Run migrations
make db/migrate

# 5. Start the API (terminal 1)
make backend/run

# 6. Start the dev server (terminal 2)
make frontend/dev

# Open http://localhost:5173
```

The API is available at `http://localhost:8080`.
Interactive docs (Scalar UI): `http://localhost:8080/docs`.

### Running tests

```bash
make backend/test/unit   # unit tests, no DB required
make backend/test        # all tests incl. integration (requires DB)
make frontend/test       # Vitest unit tests
make e2e                 # Playwright (requires running API + DB)
```

---

## CI Pipeline

Three-tier structure based on cost and signal:

| Tier | Trigger | Checks |
|---|---|---|
| **Branch** | Push to any non-`main` branch | Build, vet, staticcheck, gosec, unit tests, npm audit, ESLint, type-check |
| **PR** | PR targeting `main` | All branch checks + oasdiff (breaking API), integration tests, govulncheck, CodeQL |
| **Main** *(future)* | Post-merge to `main` | E2E, container build, image scan |

---

## Project Structure

```
rv-logbook/
├── backend/
│   ├── cmd/api/          # main.go — wiring only, no business logic
│   ├── internal/
│   │   ├── domain/       # plain structs, sentinel errors, zero deps
│   │   ├── repo/         # SQL layer; returns domain types
│   │   ├── service/      # business rules, unit-testable
│   │   └── handler/      # HTTP; implements compiler-enforced interface
│   ├── migrations/       # goose SQL migrations (embedded in binary)
│   └── spec/             # openapi.yaml + Go embed
├── frontend/
│   ├── src/
│   │   ├── api/          # typed fetch wrappers (one file per resource)
│   │   ├── features/     # trips/, stops/, tags/ — owns one product slice
│   │   ├── components/   # reusable UI primitives
│   │   └── pages/        # route-level components
│   └── e2e/              # Playwright tests
└── .github/
    ├── workflows/        # CI (backend, frontend, PR, E2E tiers)
    └── dependabot.yml    # automated dependency + security PRs
```

---

## What I Would Do With More Time

- **Authentication** — API key middleware is the natural next step; the handler
  context threading and middleware chain are already set up for it
- **Rate limiting** — per-IP token bucket using `go-chi/httprate`; the middleware
  chain has the right insertion point
- **Map view** — render stops as pins on a Leaflet/Mapbox map using the location
  strings already stored on each stop
- **Offline / PWA** — service worker + IndexedDB for viewing cached trips without
  a connection; natural complement to a travel app used in areas with no cell
- **Geolocation** — auto-populate the stop's location field from the browser's
  Geolocation API on mobile
- **Full-text search** — Postgres `tsvector` columns on `name` and `notes` fields
  to enable real search (vs the current prefix-match on tags)
- **Recurring export schedule** — a cron job that pushes a weekly CSV export to
  an S3 bucket for automatic backup

