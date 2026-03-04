# ADR-001: Layered Backend Architecture with Compiler-Enforced Interface Boundaries

**Date:** 2024-10 (Phase 1)
**Status:** Accepted

---

## Context

We need a backend architecture that is:
1. Testable at each layer independently (unit tests for business logic, integration tests for SQL)
2. Explicit about data flow and dependencies
3. Easy to navigate for someone reading the code for the first time

The common alternatives are:
- **Flat package** — all code in one or two files. Fast to start, becomes unreadable at scale.
- **MVC** — a common pattern from frameworks like Rails or Django. Mixes concerns (controllers tend to contain business logic).
- **Hexagonal / ports-and-adapters** — a strict decoupling pattern; more indirection than this project needs.
- **Layered (chose this)** — four layers with one direction of dependency: handler → service → repo → domain.

---

## Decision

Adopt a strict four-layer architecture:

```
handler → service → repo → domain
```

Each layer depends only on the layer below it, and only through an interface defined
in the *consuming* package (not the providing package). This follows the Go idiom:
"accept interfaces, return concrete types."

The `handler` package defines `TripServicer`, `StopServicer`, etc. — not the `service` package.
This means `handler` tests can inject a fake service with zero real dependencies.
Same relationship between `service` and `repo`.

The API spec (`openapi.yaml`) is the single source of truth. `oapi-codegen` generates
`StrictServerInterface` — a Go interface the `handler` package must implement. If an
endpoint is missing, or its parameter types change, the build fails. This makes the
spec and implementation impossible to drift apart silently.

---

## Consequences

**Positive:**
- Handler tests are pure unit tests — they call real handler methods with mock service dependencies.
  No HTTP server, no DB, no network.
- Service tests are pure unit tests — they call real service methods with mock repo dependencies.
  All business logic is tested at the speed of in-process function calls.
- Repo tests are integration tests that run real SQL against a real DB. They are isolated by wrapping
  each test in a transaction that rolls back — fast, no teardown scripts needed.
- The compiler enforces the API contract. An unimplemented endpoint or wrong return type is a build error.
- New engineers can orient themselves: "where does field validation happen?" → service. "where is the SQL?" → repo.

**Negative / Trade-offs:**
- More files than a flat approach. For a project this size, that's a minor cost.
- Interface definitions in the consumer package are unconventional in languages other than Go.
  Engineers from Java or Python backgrounds may expect interfaces next to their implementations.

---

## Alternatives Considered

| Option | Why rejected |
|---|---|
| Single-package flat layout | Untestable without a real DB; SQL and business logic intermingled |
| GORM ORM | Hides SQL, makes query debugging harder, N+1 queries non-obvious |
| GraphQL | Adds complexity without benefit for a simple CRUD app; REST is sufficient |
| sqlc code generation | Good option; raw SQL with pgx.NamedArgs was chosen to keep the dependency list small and queries readable |
