# GitHub Copilot Instructions — RV Logbook

## Who I Am & What This Project Is For

This project serves three equally important purposes:

1. **Portfolio / Demo** — Demonstrate senior-level proficiency in Go and React.
   Every decision should be justifiable. Code should look like it was written by
   someone who has shipped production systems, not just "made it work".

2. **Best Practices Reference** — Every layer of the stack (Go API, React UI,
   Docker, CI) should follow current community best practices. This is the project
   someone points to when asked "show me how you build things".

3. **Active Learning** — The developer is comfortable in Java and Python and has
   prior Go experience, but is *new to frontend development*. React, TypeScript,
   browser concepts, and frontend tooling must be explained as we go.

---

## Core Behavior Rules (Non-Negotiable)

### 1. Incremental Changes Only
- Never make more than one logical change at a time.
- A "logical change" = one endpoint, one component, one migration, one test file.
- If a request is large, break it down and ask which piece to start with.

### 2. Explain Before You Code
Before writing any code, briefly explain:
- **What** we are building (a sentence or two)
- **Why** it is done this way (the design principle or best practice being applied)
- **How** it connects to what came before

This is not optional. The goal is learning, not just output.

### 3. Q&A Gate — Never Auto-Proceed
After completing each step, always:
1. Summarize what was just done.
2. Highlight anything that may be surprising or worth questioning.
3. Ask: *"Do you have any questions before we move on?"*
4. **Wait for explicit confirmation** ("ready", "next", "looks good", etc.)
   before beginning the next step.

### 4. Frontend Concepts Need Extra Care
The developer has never built a UI before. When touching React/TypeScript:
- Explain browser/DOM concepts that a backend developer might not know.
- Explain the *mental model* (component tree, state vs props, lifecycle) before
  jumping into syntax.
- Relate concepts to Go equivalents where possible (e.g., "a React component is
  like a function that returns HTML — similar to a Go template function but it
  re-renders automatically when its inputs change").

### 5. Code Must Be Production-Quality
- Go: follow `go vet`, `staticcheck`, and the Uber Go style guide conventions.
- React: use TypeScript strictly (`strict: true`), no `any`, named exports.
- No TODO comments left in committed code unless they are tracked issues.
- Every public function/type has a doc comment.

### 6. TDD Is the Workflow — No Exceptions
Every coding step follows Red → Green → Commit, in that order:

1. **Red** — write the test(s) first. They must fail for the right reason
   (i.e. the function doesn't exist yet, or returns wrong data — not a
   compile error caused by a typo). Commit with prefix `test:`.
2. **Green** — write the minimum implementation to make the tests pass.
   No extra code, no premature generalisation. Commit with prefix `feat:` or `fix:`.
3. **Refactor** (if needed) — clean up without changing behaviour. Tests
   must still pass. Commit with prefix `refactor:`.

This applies to every layer:
- Go service functions → unit test with mocked repo
- Go repo functions → integration test against real DB
- Go HTTP handlers → httptest-based handler test
- React components → RTL smoke test
- React API client functions → fetch-mock unit test

Never write implementation code before a failing test exists for it.
Never commit a test that is already passing (that is not a TDD test — it is
a retrofit, which defeats the purpose).

The two-commit pattern per step makes the Red→Green transition visible in
the PR diff and demonstrates disciplined TDD to anyone reviewing the repo.

---

### 7. Frontend E2E Testability — Non-Negotiable

Every interactive UI element added to this project **must have a stable,
non-positional, non-fragile identifier** before it is committed. This is a
hard rule. No exceptions.

#### Why this matters

E2E tests (Playwright) find elements on the page. There are several ways to do
that; most of them break the moment a designer renames text or moves an element.
We enforce the one approach that survives both.

#### The selector priority hierarchy (Playwright's own recommendation)

Use the highest applicable level. Do not skip levels. Do not add redundant
identifiers at multiple levels.

| Priority | Selector | When to use |
|---|---|---|
| 1 | `getByRole()` | Element has semantic meaning AND a computable name (link, heading, dialog box) |
| 2 | `getByLabel()` | Form field associated with a `<label>` via `htmlFor` / `id` |
| 3 | `getByTestId()` | Everything else — submit buttons whose label changes during loading, view toggles, structural containers |
| 4 | `getByText()` | **Last resort only.** Never use for interactive elements or anything that might change. |

CSS class names (`className`) **must never** be used as test selectors. Classes
belong to styling. Changing a class name to fix a visual bug must not break
tests. These concerns are fully separate.

#### Rules for each element type

| Element type | Required identifier |
|---|---|
| Per-row action button (Edit, Delete) | `aria-label="{Action} {item.name}"` e.g. `aria-label="Delete camping"` |
| Submit button with changing text ("Add" → "Saving…") | `data-testid="{resource}-form-submit"` e.g. `data-testid="trip-form-submit"` |
| Cancel / secondary form button | `aria-label="Cancel {context}"` or `data-testid="{resource}-form-cancel"` |
| View / mode toggle button | `data-testid="view-toggle-{mode}"` e.g. `data-testid="view-toggle-timeline"` |
| Mutation error message | `role="alert"` — this doubles as an accessibility requirement |
| Named navigational region | `aria-label` on the `<nav>` landmark element |
| Form container | `data-testid="{resource}-form"` |
| Read-only data cell that tests will assert | `data-testid="{resource}-{field}"` e.g. `data-testid="stop-name"` |

#### What `aria-label` is (background for a backend developer)

ARIA (Accessible Rich Internet Applications) is a W3C web standard. It exists to
make UI elements understandable to screen readers — software used by visually
impaired people. `aria-label="Delete camping"` tells the browser: "announce this
button to a screen reader as 'Delete camping'". This serves two goals at once:
accessibility compliance AND a stable test target. When an element needs an
aria-label for accessibility reasons, that label IS the test handle —
no separate `data-testid` is needed. If the element has no accessibility need for
a label, use `data-testid`.

#### Naming convention for `data-testid`

Format: `{resource}-{element}` or `{resource}-{element}-{qualifier}`

Examples:
- `trip-form-submit`
- `stop-form-cancel`
- `view-toggle-list`
- `view-toggle-timeline`
- `stop-name` (data display, not interactive)

When a testid appears inside a repeating list item, include the item's ID or
name as a suffix to make it unique: `stop-list-item-{id}`.

#### Copilot enforcement rules

1. **Never generate a button, input, link, or interactive element without first
   checking the table above.** Apply the required identifier before declaring
   the component done.
2. **Never use** `className` as a test selector in Playwright test files.
3. **Never use** `getByText()` for interactive elements. Only acceptable for
   asserting that static content is visible.
4. **When writing E2E tests**, add a comment on the first `getByTestId()` call
   in the file that reads:  
   `// See CONTRIBUTING.md § "E2E Testability" for the selector strategy.`
5. **When writing new components**, if the component contains an interactive
   element that maps to none of the priority-1 or priority-2 cases above,
   add the `data-testid` in the same commit as the component — never as a
   follow-up.

---

## Project Architecture at a Glance

```
rv-logbook/
├── backend/               # Go API
│   ├── cmd/api/           # main.go — wire everything together
│   ├── internal/
│   │   ├── domain/        # pure types, no dependencies
│   │   ├── repo/          # DB layer (sqlc-generated or hand-written)
│   │   ├── service/       # business logic, unit-testable
│   │   └── handler/       # HTTP handlers (chi router)
│   ├── migrations/        # SQL migrations (goose)
│   └── testutil/          # shared test helpers
├── frontend/              # React + TypeScript (Vite)
│   ├── src/
│   │   ├── api/           # typed fetch wrappers (one file per resource)
│   │   ├── components/    # reusable UI primitives
│   │   ├── features/      # feature folders (trips/, stops/, tags/)
│   │   └── pages/         # route-level components
│   └── e2e/               # Playwright tests
├── docker-compose.yml
├── Makefile
└── .github/
    ├── copilot-instructions.md
    └── workflows/         # CI
```

### Backend Layers (Go)
```
openapi.yaml  (source of truth — edited by hand)
  → oapi-codegen  (generates StrictServerInterface + request/response types)

HTTP Request
  → handler   (implements StrictServerInterface — compiler enforces spec conformance)
  → service   (business rules, orchestration — NO SQL here)
  → repo      (all SQL lives here, returns domain types)
  → domain    (plain structs + constants — zero dependencies)
```

### Frontend Layers (React)
```
Page component   (route owner, composes features)
  → feature      (owns one slice of product functionality)
  → TanStack Query hooks (server state — data fetching/mutation)
  → api/         (typed fetch calls — the only place that knows the URL)
  → components/  (dumb, reusable UI — no business logic)
```

---

## Technology Choices (and Why)

| Concern            | Choice                  | Why |
|--------------------|-------------------------|-----|
| Go HTTP router     | `chi`                   | Lightweight, idiomatic, uses stdlib `net/http` types natively |
| API contract       | `oapi-codegen` (spec-first) | `openapi.yaml` is the source of truth; generates Go interfaces the compiler enforces |
| Breaking change CI | `oasdiff` (`oasdiff/oasdiff`) | CI gate on every PR — fails the build if a breaking API change is detected |
| DB access (Go)     | `pgx/v5` + raw SQL      | Keeps SQL explicit and readable; no ORM magic |
| Migrations         | `goose`                 | Simple, SQL-first, supports embed |
| Config             | Hand-written loader     | ~20 lines, zero dependency, shows deliberate restraint |
| Logging            | `log/slog`              | stdlib, structured, no dependency |
| React build tool   | Vite                    | Fast, modern, first-class TS support |
| Server state       | TanStack Query v5       | Industry standard; separates server vs UI state cleanly |
| Forms              | React Hook Form + Zod   | Zod schema doubles as type source-of-truth |
| Styling            | Tailwind CSS            | Utility-first; fast to iterate without context switching |
| E2E tests          | Playwright              | Cross-browser, good DX |

---

## CI Pipeline Tiers

The CI pipeline has three tiers. Each tier is slower and more expensive than the
one below it, so checks are placed at the *lowest* tier where they can provide
meaningful signal. This is the "shift left, gate right" principle.

| Tier | Trigger | Goal | Time budget |
|------|---------|------|-------------|
| **Branch** | Push to any non-`main` branch | Fast feedback — catch obvious breaks immediately | < 2 min |
| **PR** | `pull_request` targeting `main` | Gate quality — validate the full working unit before merging | < 10 min |
| **Main** | Push to `main` (post-merge) | Confirm the integrated state is releasable | uncapped |

### What belongs at each tier

**Branch** (`.github/workflows/backend.yml`, `frontend.yml`):
- `go vet` + `go build` — catches compile errors instantly
- Unit tests (`-short`, no DB) — validates logic in under a minute
- TypeScript type-check + Vite build
- ESLint

**PR** (`.github/workflows/backend-pr.yml`, `frontend-pr.yml`):
- Everything from branch (re-run on the actual merge commit)
- `oasdiff breaking` — only meaningful when compared against `main`; comparing
  within a branch has no baseline to diff against
- Integration tests (Phase 8) — need a Postgres service container; DB spin-up
  cost is acceptable at PR time, not on every branch push
- `govulncheck` (Phase 8) — security scan; slow enough to belong here, not branch

**Main** (future `main.yml`):
- Playwright E2E tests — require full stack; acceptable post-merge, not on every PR
- Container build + push
- Any SBOM / image scanning

### Rules for Copilot when touching CI

1. **Before adding any check**, ask which tier it belongs to by applying the
   table above. Do not default to "the existing workflow file".
2. **oasdiff always lives in the PR tier** — it requires a `main` baseline.
3. **Integration tests always live in the PR tier** — they need a service container.
4. **E2E tests live in the Main tier** — they need a full running stack.
5. When a new tier file is created, note it in `CONTRIBUTING.md`.

---

## Definition of "Done" for Any Step

A step is complete when:
- [ ] A failing test was written and committed first (`test:` commit)
- [ ] The implementation makes it pass (`feat:`/`fix:` commit)
- [ ] The code compiles / lints clean
- [ ] The Makefile target for that layer works (`make test`, `make lint`, etc.)
- [ ] A brief explanation has been given and questions have been answered

---

## Keeping CONTRIBUTING.md Up to Date

`CONTRIBUTING.md` is the living record of how to work on this project. After any
step that introduces a new tool, make target, environment variable, convention, or
workflow decision, update `CONTRIBUTING.md` to reflect it. Do not defer this to the
end of a phase — update it in the same step the change is made.

Specifically:
- New prerequisite tool added → add it to the Prerequisites table with version and install command
- New `make` target added → add it to the Common Commands table
- New environment variable added → add it to the Environment Variables section
- New branching or PR convention decided → update the Workflow section

---

## What to Do If Instructions Are Ambiguous

1. State the ambiguity explicitly.
2. Propose the most sensible default.
3. Ask for confirmation before proceeding.

Do not silently pick an approach and bury the decision in the code.

---

## Tone & Communication

- Treat the developer as an experienced engineer who is *new to this specific
  stack* — not a beginner, not an expert in these tools.
- Use analogies to Java/Python/backend concepts when explaining frontend ideas.
- Be direct. Skip filler phrases. Short sentences preferred over long paragraphs.
- If something is a best practice, say so and name the principle.
