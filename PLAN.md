# RV Logbook — Implementation Plan

> **How to use this file**
> - Work through steps in order. Each checkbox is one logical unit of work.
> - Check off items as they are completed (`[x]`).
> - You can stop at any checkbox boundary and resume safely.
> - Each phase ends with a working, committable state.
> - No step is started until the previous step is understood and confirmed.

---

## How Phases Are Structured

Each step follows this sequence:
1. **Explain** — what we're building, why this way, how it connects
2. **Red** — write the failing test(s) first, commit with `test:` prefix
3. **Green** — write the minimum implementation to pass, commit with `feat:`/`fix:` prefix
4. **Refactor** (if needed) — clean up, commit with `refactor:` prefix
5. **Verify** — compile, run tests, lint passes
6. **Q&A** — questions answered before moving on

---

## Phase 0 — Repo & Tooling Scaffold

> Goal: A clean, working skeleton that every future phase builds on.
> After this phase: `make help` works, Docker Compose starts Postgres,
> and the repo structure matches the architecture doc.

- [x] **0.1** Initialize Go module (`backend/go.mod`) and install core dependencies
      (`chi`, `pgx/v5`, `goose`, `testify`, `oapi-codegen` runtime)
      Config will be a hand-written loader — no external library
- [x] **0.2** Initialize Vite + React + TypeScript project under `frontend/`
- [x] **0.3** Create `docker-compose.yml` with a Postgres service and a
      named volume (no app containers yet — dev runs locally)
- [x] **0.4** Write the root `Makefile` with targets:
      `help`, `backend/run`, `backend/test`, `backend/lint`,
      `frontend/dev`, `frontend/test`, `frontend/lint`, `db/migrate`
- [x] **0.5** Add `.env.example` documenting all required environment variables
- [x] **0.6** Install `oapi-codegen` and `oasdiff` CLI tools
- [x] **0.7** Add `.gitignore` covering Go, Node, Docker, and editor artifacts
- [x] **0.8** Create empty directory scaffold matching the architecture
      (`backend/cmd/api/`, `backend/internal/{domain,repo,service,handler}/`,
       `backend/migrations/`, `backend/testutil/`,
       `frontend/src/{api,components,features,pages}/`)
- [x] **0.9** Set up GitHub Actions CI workflows (backend + frontend)
      so branch protection is in place before Phase 1 branching begins
- [x] **0.10** Commit everything to `main`: `chore: project scaffold and CI`
- [x] **0.11** Enable branch protection on `main` in GitHub
      (require PR, require CI to pass, no direct pushes)

---

## Phase 1 — Backend Foundation

> Goal: An HTTP server that starts, serves a health endpoint, connects to Postgres,
> and logs structured JSON. This is the backbone every endpoint attaches to.

- [x] **1.1** Write `backend/cmd/api/main.go` — wire config, DB pool, router, server
- [x] **1.2** Write config struct using `env` tags (port, DB DSN, log level)
- [x] **1.3** Write `openapi.yaml` with the `/healthz` endpoint and run `oapi-codegen`
      to generate the initial server interface
- [x] **1.4** Write `internal/handler/health.go` — implements the generated interface,
      returns `{"status":"ok"}` with the DB ping result
- [x] **1.5** Add `slog` structured logger; attach request-ID middleware
- [x] **1.6** Add chi middleware stack: `RequestID`, `RealIP`, `Logger`, `Recoverer`
- [x] **1.7** Write smoke test for the health handler (httptest, no DB needed)
- [x] **1.8** Verify: `make backend/run` starts; `curl localhost:8080/healthz` returns 200
- [x] **1.9** Commit: `feat(backend): server bootstrap with health endpoint`

---

## Phase 2 — Database Foundation

> Goal: A repeatable migration workflow and an integration test harness.
> After this phase we can create tables, run them, and roll back cleanly.

- [x] **2.1** Write `migrations/001_create_trips.sql` (goose up/down)
- [x] **2.2** Write `migrations/002_create_stops.sql`
- [x] **2.3** Write `migrations/003_create_tags.sql` and
      `migrations/004_create_stop_tags.sql` (join table)
- [ ] **2.4** Document the schema in a `backend/migrations/README.md` with an ERD
      (text-based is fine)
- [x] **2.5** Add `make db/migrate` and `make db/rollback` targets using goose
- [x] **2.6** Write `backend/testutil/db.go` — spins up a real DB connection
      for integration tests (reads `TEST_DATABASE_URL` env var)
- [x] **2.7** Write one integration test that runs all migrations against a real DB
      and then rolls them all back
- [x] **2.8** Commit: `feat(db): migrations and integration test harness`

---

## Phase 3 — Trips Domain (Backend)

> Goal: Full CRUD for Trips, wired end-to-end through all four layers.
> This is the "reference implementation" — every subsequent domain follows
> exactly the same pattern.

- [x] **3.1** Define `domain.Trip` struct (id, name, start_date, end_date, notes,
      created_at, updated_at)
- [x] **3.2** Define `domain.ErrNotFound`, `domain.ErrValidation` sentinel errors
- [x] **3.3** Write `repo.TripRepo` interface + `pgTripRepo` implementation
      (Create, GetByID, List, Update, Delete)
- [x] **3.4** Write unit tests for repo using a real DB (`testutil.DB`)
- [x] **3.5** Write `service.TripService` with validation logic
      (name required, end_date must be after start_date)
- [x] **3.6** Write unit tests for service (mock the repo interface)
- [x] **3.7** Write HTTP handlers (`POST /trips`, `GET /trips`,
      `GET /trips/{id}`, `PUT /trips/{id}`, `DELETE /trips/{id}`)
- [x] **3.8** Write handler tests using `httptest`
- [x] **3.9** Wire handlers into `main.go` router
- [x] **3.10** Manual verification: create a trip with `curl`, retrieve it, update it,
       delete it
- [x] **3.10a** Add Scalar API browser UI (`/docs`) and serve embedded spec at
       `/openapi.yaml` — enables browser-based manual testing for all future phases
- [x] **3.11** Commit: `feat(trips): full CRUD — domain, repo, service, handler`

---

## Phase 4 — Stops Domain (Backend)

> Goal: Stops belong to Trips. This introduces a parent-child relationship and
> demonstrates how the layered architecture handles relational data.

- [x] **4.1** Define `domain.Stop` struct (id, trip_id, name, location,
      arrived_at, departed_at, notes, created_at, updated_at)
- [x] **4.2** Write `repo.StopRepo` interface + `pgStopRepo` implementation
- [x] **4.3** Write repo integration tests
- [x] **4.4** Write `service.StopService` (validate trip exists before creating stop;
      validate arrived_at < departed_at)
- [x] **4.5** Write service unit tests (mock both TripRepo and StopRepo)
- [x] **4.6** Write handlers (`POST /trips/{tripID}/stops`,
      `GET /trips/{tripID}/stops`, `GET /trips/{tripID}/stops/{stopID}`,
      `PUT /trips/{tripID}/stops/{stopID}`,
      `DELETE /trips/{tripID}/stops/{stopID}`)
- [x] **4.7** Write handler tests
- [x] **4.8** Wire into router
- [x] **4.9** Commit: `feat(stops): CRUD with trip ownership`

---

## Phase 5 — Tags Domain (Backend)

> Goal: Tags are a many-to-many relationship between Tags and Stops.
> Demonstrates join-table handling and bulk operations.

- [x] **5.1** Define `domain.Tag` struct (id, name, slug, created_at)
- [x] **5.2** Write `repo.TagRepo` with upsert-by-slug logic and
      `List(ctx, prefix)` for autocomplete (`prefix=""` returns all tags;
      `prefix="wal"` returns tags whose slug starts with "wal")
- [x] **5.2a** Add `GET /tags?q=` endpoint to spec + handler
      (global tag list with optional prefix filter for typeahead UI)
- [x] **5.3** Write repo integration tests
- [x] **5.4** Write `service.TagService` (normalize slug: lowercase, hyphenated)
- [x] **5.5** Write service unit tests
- [x] **5.6** Extend StopService: `AddTag(stopID, tagName)`, `RemoveTagFromStop(stopID, slug)`,
      `ListTagsByStop(stopID)` (returns tags linked to a stop)
- [x] **5.7** Write handlers:
      `POST /trips/{tripId}/stops/{stopId}/tags`,
      `DELETE /trips/{tripId}/stops/{stopId}/tags/{slug}`,
      `GET /tags?q=` (global autocomplete)
- [x] **5.8** Write handler tests
- [x] **5.9** Wire into router (`main.go` wired tagService + NewServer now takes 3 args)
- [x] **5.10** Commit: `feat(tags): many-to-many stop tagging`

---

## Phase 6 — Export Endpoint (Backend)

> Goal: Demonstrate a non-CRUD use case — streaming a structured response.
> Introduces content negotiation (CSV vs JSON) and response streaming.
> Exports all trips as a flat table: one row per stop, trip fields repeated per stop.
> Trips with no stops yield one row with empty stop fields.
> Row fields: trip_name, trip_start_date, stop_name, stop_location, arrived_at, departed_at, tags (comma-joined slugs)

- [x] **6.1** Define `domain.ExportRow` struct — flat row type
- [x] **6.2** Write `service.ExportService.Export(ctx) ([]ExportRow, error)` —
      queries all trips, then stops+tags per trip, assembles rows
- [x] **6.3** Write service unit tests
- [x] **6.4** Write `GET /export` handler
      — returns JSON array by default, CSV when `Accept: text/csv` or `?format=csv`
- [x] **6.5** Write handler tests for both formats
- [x] **6.6** Wire into router
- [x] **6.7** Commit: `feat(export): CSV and JSON full-data export`

---

## Phase 7 — API Polish (Backend)

> Goal: Production-level API hygiene. These are the things that separate a demo
> from something you'd actually ship.

- [x] **7.1** Consistent error response shape:
      `{"error": {"code": "...", "message": "...", "fields": {...}}}`
- [x] **7.2** Centralized error-to-HTTP-status mapping in `handler/errors.go`
- [x] **7.3** Input size limits and timeout middleware
- [x] **7.4** Pagination on all list endpoints (`?page=1&limit=20`)
- [x] **7.5** Add `CORS` middleware (needed for the React dev server)
- [x] **7.6** Add `oasdiff` breaking-change check to the CI workflow — PRs targeting
      `main` fail if `openapi.yaml` introduces a breaking change without a version bump
- [x] **7.7** Commit: `feat(api): error shapes, pagination, CORS, oasdiff gate`

---

## Phase 8 — CI Enhancements

> Goal: Expand the CI pipeline established in Phase 0 with deeper checks
> now that real code and an OpenAPI spec exist.

- [x] **8.1** Add `staticcheck` to the backend workflow (requires Go source to be meaningful)
- [x] **8.2** Add integration tests to the backend workflow (Postgres service container)
- [x] **8.3** ~~Add `oasdiff` breaking-change gate to the backend workflow~~ — done in step 7.6 (`backend-pr.yml`)
- [x] **8.4** Add build-status badges to `README.md`
- [x] **8.5** Gate integration test files behind `//go:build integration` tag so they are
      excluded by the compiler (not just skipped at runtime) when the tag is absent
- [x] **8.6** Add API-level integration tests (`backend/internal/apitest/`)
      Wires the real stack end-to-end: `httptest.NewServer` → handler → service → repo → Postgres.
      Written as raw Go using `net/http` + `testify` + `httptest` — the idiomatic Go approach.
      No framework: the stdlib is expressive enough and this is what production Go codebases use.
      One test file per resource; covers at minimum the happy-path CRUD cycle for Trips.
      Tagged `//go:build integration` — runs in the same PR CI job as the repo tests.
- [x] **8.7** Verify: open PR from `feature/phase-8-ci`, watch all CI checks go green
- [x] **8.8** Commit: `test(api): full-stack API integration tests`

---

## Phase 9 — Frontend Foundation

> Goal: A running React app with routing, a typed API client, and a working
> TanStack Query setup. This is the frontend equivalent of Phase 1.
>
> **Learning focus:** What is the DOM? What is a component? What does Vite do?
> Why TypeScript? What does "server state" mean vs "UI state"?

- [x] **9.1** Configure Vite proxy (`/api` → `localhost:8080`) for local dev
- [x] **9.2** Set up React Router v6 with a root layout (nav + outlet)
- [x] **9.3** Create `src/api/client.ts` — a thin `fetch` wrapper that:
      - prefixes `/api`
      - throws a typed `ApiError` on non-2xx
      - attaches `Content-Type: application/json`
- [x] **9.4** Configure TanStack Query `QueryClient` with sane defaults
      (staleTime, retry policy)
- [x] **9.5** Write a smoke test for `client.ts` (mock `fetch`, verify error throwing)
- [x] **9.6** Create `src/components/ErrorBoundary.tsx` and
      `src/components/LoadingSpinner.tsx`
- [x] **9.7** Add Tailwind CSS and verify a styled "Hello World" page
- [x] **9.8** Commit: `feat(frontend): app shell, router, API client, TanStack Query`

---

## Phase 10 — Trips Feature (Frontend)

> Goal: List, create, and delete trips from the UI. First real TanStack Query
> + React Hook Form + Zod usage.
>
> **Learning focus:** useQuery vs useMutation. What is a controlled component?
> What does form validation on the client buy you vs server validation?

- [x] **10.1** Write `src/api/trips.ts` — typed functions: `listTrips`, `createTrip`,
       `deleteTrip`, `updateTrip` (mirrors the backend endpoints)
- [x] **10.2** Write `src/features/trips/useTripQueries.ts` — TanStack Query hooks
       wrapping the API functions
- [x] **10.3** Write `src/features/trips/TripList.tsx` — renders trips, loading state,
       empty state, error state
- [x] **10.4** Write RTL test for `TripList` (mock the query hook)
- [x] **10.5** Write `src/features/trips/TripForm.tsx` — React Hook Form + Zod schema
       for create/edit
- [x] **10.6** Write RTL test for `TripForm` (submit valid/invalid data)
- [x] **10.7** Compose into `src/pages/TripsPage.tsx`
- [x] **10.8** Wire `TripsPage` into the router
- [x] **10.9** Manual end-to-end: create a trip in the UI, see it in the list,
       delete it — all reflected in the DB
- [x] **10.10** Commit: `feat(frontend/trips): list, create, delete`

---

## Phase 11 — Stops Feature (Frontend)

> Goal: View and manage stops within a trip. Introduces nested routing and
> URL-driven state.
>
> **Learning focus:** React Router `useParams`. Why URL state beats useState for
> "which trip am I looking at".

- [x] **11.1** Write `src/api/stops.ts`
- [x] **11.2** Write `src/features/stops/useStopQueries.ts`
- [x] **11.3** Write `src/features/stops/StopList.tsx`
- [x] **11.4** Write `src/features/stops/StopForm.tsx` with tag input
       (comma-separated, validated with Zod)
- [x] **11.5** RTL tests for both components
- [x] **11.6** Compose into `src/pages/TripDetailPage.tsx` (shows trip info + stop list)
- [x] **11.7** Wire `TripDetailPage` as a nested route under `/trips/:id`
- [x] **11.8** Commit: `feat(frontend/stops): stop management inside trip detail`

---

## Phase 12 — Tag Input Polish (Frontend)

> Goal: Replace the raw comma-separated tag text field with a proper pill-based
> tag input with autocomplete. Display tags as pills in the stop list.
> Requires two backend additions done spec-first before the frontend work begins:
> (a) embed `tags` inline on every `Stop` response to avoid N+1 fetches, and
> (b) add `PATCH /tags/{slug}` for the tag rename feature (Phase 13).
>
> **Learning focus:** Controlled vs uncontrolled inputs. Managing a list as local
> component state before committing it to the server. How to debounce a query
> (avoid hammering the API on every keystroke). SQL `json_agg` aggregation as an
> alternative to N+1 queries.

- [x] **12.0a** Update `openapi.yaml`:
       - Add `tags: Tag[]` (optional, defaults to `[]`) to the `Stop` schema
       - Add `PATCH /tags/{slug}` endpoint with `PatchTagRequest` body (`name` only)
       - Add `PatchTagRequest` schema
- [x] **12.0b** Regenerate backend handler interface (`make backend/generate`)
- [x] **12.0c** Write failing repo tests (RED):
       - `repo/stop_test.go`: assert `Tags` field is populated after a tag is added to a stop
       - `repo/tag_test.go`: add `TestTagRepo_UpdateName`
- [x] **12.0d** Implement repo (GREEN):
       - Add `Tags []domain.Tag` to `domain.Stop`
       - Add `UpdateName(ctx, slug, name string) (domain.Tag, error)` to `TagRepo` interface + `pgTagRepo`
       - Update `GetByID`, `ListByTripID`, `ListByTripIDPaged` in `repo/stop.go` to use a
         `LEFT JOIN stop_tags / tags` with `json_agg(...) FILTER (WHERE t.id IS NOT NULL)`
         aggregation — one query per call, no N+1
- [x] **12.0e** Write failing service + handler tests (RED):
       - `service/tag_test.go`: `TestTagService_UpdateName` (empty name → ErrValidation, valid → updated tag)
       - `handler/tag_test.go`: `TestPatchTag` (200 with renamed tag, 404 for unknown slug, 422 for empty name)
       - `handler/stop_test.go`: assert stop responses include a `tags` array
- [x] **12.0f** Implement service + handler (GREEN):
       - `service/tag.go`: add `UpdateName(ctx, slug, name string) (domain.Tag, error)`
       - `handler/tag.go`: add `PatchTag` handler
       - `handler/stop.go`: update `stopToResponse` to map `s.Tags → []gen.Tag`
       - Wire `PatchTag` — `oapi-codegen` enforces this at compile time
- [x] **12.0g** Regenerate frontend types (`make frontend/generate`),
       update `schemas.ts`: add `TagSchema`, add `tags` field to `StopSchema`
- [x] **12.0h** Commit backend + schema: `feat(tags): embed tags on Stop responses; PATCH /tags/{slug}`

- [x] **12.1** Write `src/api/tags.ts` — typed wrapper around `GET /tags?q=`
       (used for autocomplete) and `GET /tags` (used for the tags management page)
- [x] **12.2** Write `src/components/TagPill.tsx` — shared display primitive.
       Renders a coloured badge; accepts an optional `onRemove` callback that shows
       an `×` button when provided. No domain knowledge — pure display.
- [x] **12.3** Write RTL test for `TagPill` (renders label, calls onRemove when × clicked)
- [x] **12.4** Update `StopList.tsx` to render `stop.tags` as `<TagPill>` rows
       below each stop's name/location line. No interaction needed here.
- [x] **12.5** Update `StopList` RTL test to assert tags are rendered as pills
- [x] **12.6** Write `src/components/TagInput.tsx` — the reusable controlled input:
       - Renders pending tags as `<TagPill onRemove=…>` — removing before save
         discards the tag without creating it
       - As the user types, calls `GET /tags?q=<value>` and shows a dropdown of
         matching existing tags
       - **Enter** with no suggestion (or suggestion ignored) creates a new pending pill
       - **Backspace** on empty input removes the last pill
       - Clicking a dropdown suggestion adds the canonical tag name
       - Exposes `value: string[]` + `onChange: (tags: string[]) => void` — fits
         directly into React Hook Form via `Controller`
- [x] **12.7** Write RTL tests for `TagInput`:
       - typing and pressing Enter adds a pending pill
       - clicking a suggestion adds it and clears the input
       - × on a pill removes it
       - Backspace on empty removes last pill
       - duplicate tag names are silently ignored
- [x] **12.8** Replace `tagsRaw` text field in `StopForm.tsx` with `<TagInput>`
       wired via RHF `Controller`. Update the Zod schema: `tags: z.array(z.string())`
       replaces `tagsRaw: z.string()`.
- [x] **12.9** Update `StopForm` RTL tests to use the new tag input interactions
- [x] **12.10** Commit: `feat(frontend/tags): pill-based tag input with autocomplete`

---

## Phase 13 — Tags Management Page (Frontend + Backend)

> Goal: A `/tags` page where the user can view all tags, edit a tag's display name,
> and delete a tag. Editing the display name (not the slug) requires one new
> backend endpoint.
>
> **Learning focus:** When a UI feature requires a new API — and how to add it
> spec-first (edit `openapi.yaml` → regenerate → implement) rather than
> bolting it on ad-hoc.
>
> **Backend dependency:** `PATCH /tags/{slug}` — updates `name` only; the slug is
> the stable identifier and must never change (it would break existing stop-tag links).

- [x] **13.1** Add `PATCH /tags/{slug}` to `openapi.yaml`, regenerate handler interface  *(done in Phase 12)*
- [x] **13.2** Write service method `TagService.UpdateName(ctx, slug, newName) error`
       (validates non-empty, re-normalises display name) and unit test  *(done in Phase 12)*
- [x] **13.3** Write handler + handler test for `PATCH /tags/{slug}`  *(done in Phase 12)*
- [x] **13.4** Wire into router; commit backend: `feat(tags): PATCH /tags/{slug} rename endpoint`  *(done in Phase 12)*
- [ ] **13.0** Backend prereq — `DELETE /tags/{slug}` (not in original plan; needed for delete button):
       - Add endpoint to `openapi.yaml`, regenerate
       - `repo/tag_test.go`: `TestTagRepo_Delete_OK`, `TestTagRepo_Delete_NotFound`
       - `service/tag_test.go`: `TestTagService_Delete`
       - `handler/tag_test.go`: `TestDeleteTag_204`, `TestDeleteTag_404`
       - Implement `TagRepo.Delete`, `TagService.Delete`, `handler.DeleteTag`
       - Commit: `feat(tags): DELETE /tags/{slug}`
- [x] **13.5** Add `deleteTag` to `src/api/tags.ts` + test
- [x] **13.6** Write `src/features/tags/useTagQueries.ts` — TanStack Query hooks
       for list, update, delete
- [x] **13.7** Write `src/pages/TagsPage.tsx` — table of all tags; inline edit of
       display name (click pencil → text input → save); delete button per row
- [x] **13.8** RTL tests for `TagsPage` (renders tags, inline edit, delete)
- [x] **13.9** Wire `/tags` route into router and add link to nav
- [x] **13.10** Commit: `feat(frontend/tags): tags management page`

---

## Phase 14 — Timeline View (Frontend)

> Goal: Display stops in chronological order on a visual "timeline" within a trip.
> Demonstrates pure derivation — no new API, just a different rendering of
> existing data.
>
> **Learning focus:** Derived data in React. When NOT to add a new API endpoint.

- [x] **14.1** Write `src/features/trips/TripTimeline.tsx` — sorts stops by
       `arrived_at`, renders as a vertical timeline
- [x] **14.2** RTL test: given unsorted stops, output is ordered correctly
- [x] **14.3** Add a "Timeline" tab toggle to `TripDetailPage`
       (list view ↔ timeline view, no routing needed — pure UI state)
- [x] **14.4** Commit: `feat(frontend/timeline): chronological stop timeline`

---

## Phase 15 — Export (Frontend)

> Goal: Trigger a file download from the browser using a signed link or a
> direct fetch-to-blob approach.
>
> **Learning focus:** How browser downloads work. `fetch` + `Blob` + `URL.createObjectURL`.

- [x] **15.1** Write `src/api/export.ts` — fetches the export endpoint and
       returns a `Blob`
- [x] **15.2** Write `src/features/trips/ExportButton.tsx` — triggers download on click
- [x] **15.3** RTL test: clicking the button calls the API and creates a download link
- [x] **15.4** Add button to `TripDetailPage`
- [x] **15.5** Commit: `feat(frontend/export): CSV and JSON download`

---

## Phase 16 — General Search (Frontend) ⏭ SKIPPED

> Goal: A search bar at the top of the trips list that filters across all fields —
> trip name/notes and stop name/location/notes/tags — within the already-fetched
> data. No new backend endpoint needed for an MVP; a full-text `GET /search?q=`
> endpoint can be added later if performance demands it.
>
> **Learning focus:** `useSearchParams` hook. Client-side derived filtering as a
> stepping stone before deciding whether to push work to the server.
>
> **Skipped** — deferred to allow focus on deployment and polish phases first.

- [ ] **16.1** Write `src/features/trips/TripSearch.tsx` — controlled text input
       that updates `?q=` in the URL query string via `useSearchParams`
- [ ] **16.2** Filter the trips list client-side: match against trip name, notes,
       and any stop name/location/notes/tag name within the trip
- [ ] **16.3** RTL test: typing in the search box updates the URL and filters the list
- [ ] **16.4** Wire `TripSearch` into `TripsPage` above the trip list
- [ ] **16.5** Commit: `feat(frontend/search): cross-field trip and stop search`

---

## Phase 17 — Frontend Polish

> Goal: The UX details that make a portfolio project stand out.
> Approach taken: adopted **shadcn/ui** as a full design system instead of
> hand-rolling individual polish items. Covers steps A–E below.

- [x] **17.A** Init shadcn/ui v3.8.5 (New York style, stone OKLCH palette);
       add `@/*` path alias to tsconfig + vite; add `src/components/ui/` with
       Button, Card, Input, Label, Badge, Skeleton, Sonner components;
       add `src/lib/utils.ts` (`cn()` helper)
- [x] **17.B** Styled nav header with active-link highlighting; dark stone
       `--primary` bar, `NavLink` + `cn()` for active state
- [x] **17.C** Badge + Skeleton: `TagPill` → `Badge variant="default"` (dark,
       high contrast); `LoadingSpinner` → animated `Skeleton` bars;
       pre-boot HTML loader in `index.html` for JS bundle gap;
       camper-van emoji favicon replacing Vite default
- [x] **17.D** All raw form elements replaced: `Input`/`Label`/`Button` across
       TripForm, StopForm, TripList, StopList, TagsPage, ExportButton,
       TripDetailPage; delete actions use `ghost + text-destructive`
- [x] **17.E** Card layout on TripsPage, TripDetailPage, TagsPage; headings
       remain real `<h1>`/`<h2>` elements (not `CardTitle` which is a `div`)
- [x] **17.F** Sonner toast notifications: inline `role="alert"` error
       paragraphs removed; `toast.error()` via mutation `onError` callbacks;
       `<Toaster />` mounted once in `main.tsx`; `addError`/`editError`
       useState dropped from TripDetailPage
- [x] **17.G** Commit and PR: `feat(ui): Phase 17 — shadcn/ui design system` (PR #20)

---

## Phase 18 — Deployment (Fly.io)

> Goal: Ship the full stack to a real public URL.
> Hosting platform: **Fly.io** (Docker-native, managed Postgres, single-region, ~$5-7/month).
> This phase must complete before Phase 19 (E2E) because Playwright runs against
> the live deployed environment.
>
> **Cost model (demo-only use):** The app machine can be scaled to zero between demos
> (`fly scale count 0`), so compute cost is ~$0 when idle. The managed Postgres dev
> cluster runs continuously at ~$1.94/month. Total expected cost: ~$2/month.
> Before a demo, run `fly scale count 1` — Go cold starts on Fly take ~3-5 seconds.
>
> **Prerequisite:** Create a Fly.io account at https://fly.io and install the
> `flyctl` CLI (`winget install flyctl` or see https://fly.io/docs/hands-on/install-flyctl/).

- [ ] **18.1** Write a production `Dockerfile` for the Go backend
       (multi-stage: build stage → minimal runtime image)
- [ ] **18.2** Write `fly.toml` for the backend app
       (internal port 8080, health check `/healthz`, single machine to keep costs low)
- [ ] **18.3** Provision a Fly Postgres cluster (`fly postgres create`)
       and attach it to the app — this sets `DATABASE_URL` automatically
- [ ] **18.4** Add `make fly/deploy` target that runs `fly deploy`
- [ ] **18.5** Run database migrations at deploy time
       (add a `[deploy] release_command` in `fly.toml` that runs `goose up`)
- [ ] **18.6** Configure Fly secrets for any env vars not set by Postgres attachment
       (e.g. `LOG_LEVEL`, `PORT`)
- [ ] **18.7** Write `fly.frontend.toml` for the Vite frontend
       (static build served via a lightweight nginx or Fly static site)
- [ ] **18.8** Add GitHub Actions deploy job to the **Main tier** CI
       (triggers on push to `main` after all tests pass; uses `superfly/flyctl-actions`)
- [ ] **18.9** Verify: push to `main` → CI deploys → `https://<appname>.fly.dev/healthz` returns 200
- [ ] **18.10** Commit: `feat(infra): Fly.io deployment config`

---

## Phase 19 — E2E Tests (Playwright)

> Goal: One full user journey, automated and runnable in CI.
> Demonstrates ownership of the whole stack, not just unit tests.

- [x] **19.1** Install and configure Playwright (`frontend/e2e/`)
- [x] **19.2** Write journey: create trip → add stop with tags → verify timeline →
       export CSV → delete trip
- [x] **19.3** Add `make e2e` target (starts API + frontend, runs Playwright, tears down)
- [x] **19.4** Add E2E job to GitHub Actions (only runs on `main`, runs against the Fly.io URL)
- [x] **19.5** Commit: `test(e2e): full trip lifecycle journey`

---

## Phase 20 — README & Portfolio Wrap-Up

> Goal: Someone landing on the repo should immediately understand what it is,
> how to run it, and what makes it interesting. This is your pitch.

- [x] **20.1** Rewrite `README.md` with: project summary, feature list, tech choices
       (with one-line rationale for each), and a "What I'd do with more time" section
- [x] **20.2** Add `docs/architecture.md` with layer diagram and data model ERD
- [x] **20.3** Add `docs/adr/` — at least two Architecture Decision Records:
       one for the layered backend, one for TanStack Query vs Redux
- [x] **20.4** Final `make lint` and `make test` — everything green
- [ ] **20.5** Tag `v1.0.0`
- [x] **20.6** Commit: `docs: portfolio-ready README and architecture docs`

---

## Resuming After a Break

1. Find the first unchecked item (`[ ]`) in this file.
2. Read the phase header above it for context.
3. Tell Copilot: *"Resume from step X.Y"* and it will explain what that step
   does before writing any code.

---

## Completed Phases

- **Phase 0** — Repo & Tooling Scaffold ✅
- **Phase 1** — Backend Foundation ✅
- **Phase 2** — Database Foundation ✅
- **Phase 3** — Trips Domain (Backend) ✅
- **Phase 4** — Stops Domain (Backend) ✅
- **Phase 5** — Tags Domain (Backend) ✅
- **Phase 6** — Export Endpoint (Backend) ✅
- **Phase 7** — API Polish (Backend) ✅
- **Phase 8** — CI Enhancements ✅
- **Phase 9** — Frontend Foundation ✅
- **Phase 10** — Trips Feature (Frontend) ✅
- **Phase 11** — Stops Feature (Frontend) ✅
- **Phase 12** — Tag Input Polish (Frontend) ✅
- **Phase 13** — Tags Management Page (Frontend + Backend) ✅
- **Phase 14** — Timeline View (Frontend) ✅
- **Phase 15** — Export (Frontend) ✅
- **Phase 16** — General Search (Frontend) ⏭ skipped
- **Phase 17** — Frontend Polish (shadcn/ui design system) ✅
