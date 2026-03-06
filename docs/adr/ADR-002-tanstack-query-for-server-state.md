# ADR-002: TanStack Query for Server State (vs. Redux)

**Date:** 2024-11 (Phase 9)
**Status:** Accepted

---

## Context

A React frontend that talks to a REST API needs a strategy for managing server state:
the data that lives in the database and is fetched over HTTP. This is the overwhelming
majority of "state" in this application — trips, stops, tags.

The traditional choice in the React ecosystem is Redux (or Redux Toolkit). It is the
most-mentioned tool in job postings and tutorials. The newer alternative is TanStack
Query (formerly React Query).

---

## Decision

Use **TanStack Query v5** for all server state (data that comes from the API).
Use plain `useState` for UI-only state (whether a dialog is open, which view mode is active).

**No global state store** (Redux, Zustand, Jotai, etc.) is used anywhere in this application.

---

## Rationale

### The fundamental mismatch in Redux for server state

Redux models state as a single, synchronous, in-memory object. Server state is
fundamentally different: it is remote, asynchronously fetched, can be stale, can be
loading, can have errors, and can change independently of user actions (another tab
updates it; the server changes it).

Using Redux for server state means manually managing:
- `isLoading`, `isError`, `data` flags per resource
- Cache invalidation (when to refetch after a mutation)
- Background refresh
- Stale-while-revalidate behavior
- Deduplication of concurrent requests for the same data

TanStack Query handles all of this automatically.

### What TanStack Query provides out-of-the-box

```typescript
const { data: trips, isLoading, isError } = useQuery({
  queryKey: ['trips'],
  queryFn: listTrips,
})
```

This single hook gives you:
- Automatic caching keyed by `queryKey`
- Loading and error states without any extra code
- Background refetch when the window regains focus
- Stale-time configuration to avoid over-fetching
- Optimistic update helpers for mutations
- `invalidateQueries` to refetch after a mutation — cache coherence without manual coordination

### Redux would be appropriate if...

- The application had complex *client-side* business logic that doesn't map to server state
  (e.g., a multi-step workflow with derived state across many entities)
- The team needed time-travel debugging via Redux DevTools for hard-to-reproduce bugs
- The app had significant real-time state (websocket events mutating shared state)

None of those apply here. This app is: fetch trips, show trips, create trip, refetch.

---

## State model in this app

| State | Tool | Reason |
|---|---|---|
| Trips, stops, tags from API | TanStack Query | Server state — async, cacheable, stale |
| Form field values | React Hook Form | Isolated re-renders; Zod validation co-located with schema |
| Dialog open/close | `useState` | Two-value UI state; no sharing needed |
| Active view (list/timeline) | `useState` | Local component preference |
| Toast notifications | Sonner | Fire-and-forget; no state needed |

---

## Consequences

**Positive:**
- ~60% less boilerplate than equivalent Redux Toolkit code for the same fetch/cache/invalidate pattern
- Error and loading states handled uniformly across all features
- Cache invalidation is explicit: `queryClient.invalidateQueries({ queryKey: ['trips'] })` after
  every mutation keeps the UI consistent without manual array splicing
- Easy to test: mock the query client in unit tests, or use `msw` to intercept HTTP in integration tests

**Negative / Trade-offs:**
- TanStack Query is less known to entry-level developers than Redux (though this gap is closing fast)
- Does not solve *client-side* state; if complex non-server state is added later, a separate tool
  (Zustand is the lightest option) would need to be added alongside it
- Query cache is in-memory — refreshing the page re-fetches everything. Acceptable for this use case.

---

## Alternatives Considered

| Option | Why rejected |
|---|---|
| Redux Toolkit + RTK Query | RTK Query is good, but adding Redux itself brings ~15 KB and conceptual overhead (actions, reducers, selectors) that this app does not need |
| SWR (Vercel) | Solid library; TanStack Query has richer invalidation and mutation APIs |
| Apollo Client | GraphQL-only; this app uses REST |
| Raw `fetch` + `useState` | Requires re-implementing caching, deduplication, and stale detection manually |
