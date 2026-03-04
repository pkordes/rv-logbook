# RV Logbook â€” Go + React

[![Backend](https://github.com/pkordes/rv-logbook/actions/workflows/backend.yml/badge.svg)](https://github.com/pkordes/rv-logbook/actions/workflows/backend.yml)
[![Backend PR](https://github.com/pkordes/rv-logbook/actions/workflows/backend-pr.yml/badge.svg)](https://github.com/pkordes/rv-logbook/actions/workflows/backend-pr.yml)
[![Frontend](https://github.com/pkordes/rv-logbook/actions/workflows/frontend.yml/badge.svg)](https://github.com/pkordes/rv-logbook/actions/workflows/frontend.yml)
[![Frontend PR](https://github.com/pkordes/rv-logbook/actions/workflows/frontend-pr.yml/badge.svg)](https://github.com/pkordes/rv-logbook/actions/workflows/frontend-pr.yml)

A production-style web application for full-time RV life: log trips and stops,
tag campgrounds for quick recall, view a trip timeline, and export your travel
history to CSV or JSON.

Built as an explicit demonstration of senior-level engineering across a full
Go + React stack â€” every decision is deliberate and documented.

---

## Features

- **Trip management** â€” create, edit, and delete trips with date ranges and notes
- **Stop tracking** â€” log every campground or park with location, dates, cost,
  site number, rating, and free-text notes
- **Tagging** â€” apply and reuse tags (e.g. `quiet`, `good-cell`, `50-amp`) across
  stops; edit or delete tags globally from the Tags page
- **Timeline view** â€” visualize stops on a trip as a date-ordered timeline
- **Paginated lists** â€” all collections support `?page=` and `?limit=` parameters
- **Export** â€” download full travel history as CSV or JSON from a single endpoint
- **Live API docs** â€” interactive Scalar UI served at `/docs` with the OpenAPI spec

---

## Architecture Overview

```
openapi.yaml  (single source of truth â€” all types derived from this)
     â”‚
     â””â”€ oapi-codegen â†’ StrictServerInterface + request/response types
                            â”‚
HTTP Request â”€â”€â–º handler  (implements compiler-enforced interface)
                    â”‚
                 service  (business rules; unit-tested with mocks)
                    â”‚
                   repo   (all SQL; integration-tested against real DB)
                    â”‚
                 domain   (plain structs + sentinel errors; zero deps)
```

Frontend:

```
Page  â†’  feature  â†’  TanStack Query hooks  â†’  api/ (typed fetch wrappers)
                                                â”‚
                                          components/ (dumb UI primitives)
```

See [docs/architecture.md](docs/architecture.md) for the full layer diagram and
data model ERD.

---

## Tech Choices

| Concern | Choice | Why |
|---|---|---|
| Go router | `chi` | Lightweight; accepts stdlib `http.Handler` â€” no lock-in |
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
| Go | â‰¥ 1.25 | https://go.dev/dl |
| Node | â‰¥ 22 (see `.nvmrc`) | https://nodejs.org |
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
â”œâ”€â”€ backend/
â”‚   â”œâ”€â”€ cmd/api/          # main.go â€” wiring only, no business logic
â”‚   â”œâ”€â”€ internal/
â”‚   â”‚   â”œâ”€â”€ domain/       # plain structs, sentinel errors, zero deps
â”‚   â”‚   â”œâ”€â”€ repo/         # SQL layer; returns domain types
â”‚   â”‚   â”œâ”€â”€ service/      # business rules, unit-testable
â”‚   â”‚   â””â”€â”€ handler/      # HTTP; implements compiler-enforced interface
â”‚   â”œâ”€â”€ migrations/       # goose SQL migrations (embedded in binary)
â”‚   â””â”€â”€ spec/             # openapi.yaml + Go embed
â”œâ”€â”€ frontend/
â”‚   â”œâ”€â”€ src/
â”‚   â”‚   â”œâ”€â”€ api/          # typed fetch wrappers (one file per resource)
â”‚   â”‚   â”œâ”€â”€ features/     # trips/, stops/, tags/ â€” owns one product slice
â”‚   â”‚   â”œâ”€â”€ components/   # reusable UI primitives
â”‚   â”‚   â””â”€â”€ pages/        # route-level components
â”‚   â””â”€â”€ e2e/              # Playwright tests
â””â”€â”€ .github/
    â”œâ”€â”€ workflows/        # CI (backend, frontend, PR, E2E tiers)
    â””â”€â”€ dependabot.yml    # automated dependency + security PRs
```

---

## What I Would Do With More Time

- **Authentication** â€” API key middleware is the natural next step; the handler
  context threading and middleware chain are already set up for it
- **Rate limiting** â€” per-IP token bucket using `go-chi/httprate`; the middleware
  chain has the right insertion point
- **Map view** â€” render stops as pins on a Leaflet/Mapbox map using the location
  strings already stored on each stop
- **Offline / PWA** â€” service worker + IndexedDB for viewing cached trips without
  a connection; natural complement to a travel app used in areas with no cell
- **Geolocation** â€” auto-populate the stop's location field from the browser's
  Geolocation API on mobile
- **Full-text search** â€” Postgres `tsvector` columns on `name` and `notes` fields
  to enable real search (vs the current prefix-match on tags)
- **Recurring export schedule** â€” a cron job that pushes a weekly CSV export to
  an S3 bucket for automatic backup

