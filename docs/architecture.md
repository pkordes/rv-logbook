# Architecture

## Backend — Layer Diagram

```
HTTP Request
     │
     ▼
┌─────────────────────────────────────────────────────┐
│  Middleware chain (main.go)                          │
│  RequestID → RealIP → SlogLogger → Recoverer         │
│  SecurityHeaders → CORS → MaxBodySize                │
└─────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────────────────────────────────────────┐
│  handler/   (HTTP layer)                            │
│                                                     │
│  Implements gen.StrictServerInterface — generated   │
│  by oapi-codegen from openapi.yaml. The compiler    │
│  fails the build if any operation is missing or has │
│  the wrong signature. No business logic here.       │
│                                                     │
│  Responsibilities:                                  │
│  • Parse and validate request types                 │
│  • Map domain errors → HTTP status codes            │
│  • Return typed response objects                    │
└────────────────────┬────────────────────────────────┘
                     │  calls via interface
                     ▼
┌─────────────────────────────────────────────────────┐
│  service/   (business layer)                        │
│                                                     │
│  All business rules live here. Unit-tested with     │
│  mock repos — no database, no HTTP in service tests.│
│                                                     │
│  Responsibilities:                                  │
│  • Field validation (domain.ErrValidation)          │
│  • Orchestration across multiple repos              │
│  • slug normalization for tags                      │
└────────────────────┬────────────────────────────────┘
                     │  calls via interface
                     ▼
┌─────────────────────────────────────────────────────┐
│  repo/      (data layer)                            │
│                                                     │
│  All SQL lives here. Integration-tested against a   │
│  real Postgres instance with per-test transaction   │
│  rollback for isolation.                            │
│                                                     │
│  Responsibilities:                                  │
│  • Parameterized queries (pgx.NamedArgs)            │
│  • Map pgx rows → domain types                      │
│  • Translate pgx errors → domain.ErrNotFound etc.  │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│  domain/    (core types)                            │
│                                                     │
│  Pure Go structs and sentinel errors.               │
│  Zero external dependencies — nothing imports this  │
│  except the layers above.                           │
└─────────────────────────────────────────────────────┘
```

### Interface inversion

Handler and service each define the interface they need in their own package:

```go
// handler/server.go
type TripServicer interface {
    Create(ctx context.Context, trip domain.Trip) (domain.Trip, error)
    ...
}
```

This is the Go convention: "accept interfaces, return concrete types". It means:
- `handler` can be tested with a fake `TripServicer` — no service or DB involved.
- `service` can be tested with a fake `TripRepo` — no DB involved.
- Changing the service signature is a compile error if the interface is not updated.

---

## Frontend — Layer Diagram

```
Route
  │
  ▼
┌───────────────────────────────────────────────────────┐
│  pages/   (route-level components)                    │
│  e.g. TripDetailPage — owns layout, composes features │
└──────────────────────┬────────────────────────────────┘
                       │
                       ▼
┌───────────────────────────────────────────────────────┐
│  features/trips/stops/tags  (product feature slices)  │
│  e.g. StopList, StopForm, TagInput                    │
│  Owns one slice of product functionality.             │
│  Calls TanStack Query hooks for data.                 │
└──────────────────────┬────────────────────────────────┘
                       │
                       ▼
┌───────────────────────────────────────────────────────┐
│  TanStack Query hooks  (server state)                 │
│  useTrips(), useCreateStop(), etc.                    │
│  Handles caching, background refetch, mutation state. │
└──────────────────────┬────────────────────────────────┘
                       │
                       ▼
┌───────────────────────────────────────────────────────┐
│  api/   (typed fetch wrappers)                        │
│  One file per resource: trips.ts, stops.ts, tags.ts  │
│  Only layer that knows the URL shape.                 │
│  All types derived from openapi.d.ts (generated).    │
└──────────────────────┬────────────────────────────────┘
                       │  HTTP /api/* (proxied to :8080 in dev)
                       ▼
                  Go API server
```

### State separation

| State type | Tool | Why |
|---|---|---|
| Server state (trips, stops, tags) | TanStack Query | Cache, invalidation, loading/error states built in |
| Form state (controlled inputs) | React Hook Form | Isolated re-renders; integrates with Zod for validation |
| UI state (dialogs, toggles) | `useState` | Simple enough; no global store needed |

---

## Data Model — ERD

```
┌──────────────────────────────────────────────┐
│  trips                                       │
│  ─────────────────────────────────────────── │
│  id           UUID  PK                       │
│  name         TEXT  NOT NULL                 │
│  start_date   DATE  NOT NULL                 │
│  end_date     DATE  (nullable)               │
│  notes        TEXT  (nullable)               │
│  created_at   TIMESTAMPTZ                    │
│  updated_at   TIMESTAMPTZ                    │
└──────────────────────┬───────────────────────┘
                       │  1
                       │
                       │  ∞  (ON DELETE CASCADE)
┌──────────────────────▼───────────────────────┐
│  stops                                       │
│  ─────────────────────────────────────────── │
│  id           UUID  PK                       │
│  trip_id      UUID  FK → trips.id            │
│  name         TEXT  NOT NULL                 │
│  location     TEXT  (nullable)               │
│  arrived_at   TIMESTAMPTZ NOT NULL           │
│  departed_at  TIMESTAMPTZ (nullable)         │
│  notes        TEXT  (nullable)               │
│  created_at   TIMESTAMPTZ                    │
│  updated_at   TIMESTAMPTZ                    │
└──────────────────────┬───────────────────────┘
                       │  ∞
                       │  (via stop_tags join table)
                       │  ∞
┌──────────────────────▼───────────────────────┐
│  stop_tags  (join table, no surrogate key)   │
│  ─────────────────────────────────────────── │
│  stop_id    UUID  PK, FK → stops.id          │
│  tag_id     UUID  PK, FK → tags.id           │
│             (ON DELETE CASCADE both sides)   │
└──────────────────────┬───────────────────────┘
                       │  ∞
                       │
                       │  1
┌──────────────────────▼───────────────────────┐
│  tags                                        │
│  ─────────────────────────────────────────── │
│  id           UUID  PK                       │
│  name         TEXT  (display text)           │
│  slug         TEXT  UNIQUE (identity key)    │
│  created_at   TIMESTAMPTZ                    │
└──────────────────────────────────────────────┘
```

### Why `slug` is the tag identity (not `name`)

Tags are user-created. "50-amp", "50 AMP", and " 50amp " should all resolve
to the same tag. The service layer normalizes every tag input to a slug
(`strings.ToLower` + replace spaces with `-` + collapse hyphens) before any
DB operation. The unique constraint is on `slug`, so upserts are idempotent.
`name` preserves the text exactly as the first user typed it, but has no
uniqueness constraint.

---

## CI — Three-Tier Pipeline

```
Branch push  ──► backend.yml  ──►  go vet, staticcheck, gosec, build, unit tests
              ──► frontend.yml ──►  ESLint, tsc, vite build, npm audit

PR → main    ──► backend-pr.yml  ──►  unit tests, oasdiff, integration tests,
              │                        govulncheck, CodeQL
              └─► frontend-pr.yml ──►  lint + build (re-run on merge commit)

Push to main ──► (future main.yml) ──►  Playwright E2E, container build/push
```

Tier design principle: a check lives at the *lowest* tier where it provides
meaningful signal. `oasdiff` requires a `main` baseline — it is meaningless on
a branch push. Integration tests require a Postgres container — acceptable cost
at PR time, not on every push. CodeQL takes 3–8 min — acceptable at PR time.
