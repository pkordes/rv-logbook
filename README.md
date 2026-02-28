# RV Logbook — Go + React

A small, production-style web app for full-time RV life:
- log trips and stops (campgrounds/parks)
- add notes + tags for quick filtering (e.g. `quiet`, `good-cell`, `50amp`)
- view a trip timeline
- export data (CSV/JSON)

This repo is intentionally designed as a demo project for best practices in Go and React.

---

## Tech Stack

**Backend**
- Go (HTTP API)
- Postgres (primary DB)
- SQL migrations
- REST + JSON (OpenAPI optional)
- Tests: unit + integration

**Frontend**
- React + TypeScript (Vite)
- TanStack Query for server state
- React Hook Form + Zod for forms/validation
- Playwright smoke tests (optional)

**DevEx**
- Docker Compose for local dev
- Makefile/scripts for common tasks
- GitHub Actions CI

---

## Goals (Product + Engineering)

### Product Goals (MVP)
1. Create a trip
2. Add stops to a trip
3. Tag and search stops/notes
4. Trip timeline view (stops ordered by date)
5. Export trips/stops to CSV/JSON

### Engineering Goals
- Clear domain model and clean separation: transport → service → repo
- Input validation + consistent error responses
- Structured logging + request IDs
- DB migrations and a repeatable local setup
- Tests for services and at least one DB integration test
- CI that runs lint + tests + build

---

## Data Model (v1)

### Trip
- id (uuid)
- name (string)
- start_date (date)
- end_date (date, nullable)
- created_at, updated_at

### Stop
- id (uuid)
- trip_id (uuid)
- name (string)            // campground/park name
- location (string)        // "City, ST" or freeform
- arrival_date (date)
- departure_date (date, nullable)
- cost_cents (int, nullable)
- site (string, nullable)  // e.g., "B12"
- notes (text, nullable)
- rating (int 1-5, nullable)
- tags (many-to-many)
- created_at, updated_at

### Tag
- id (uuid)
- name (string, unique, lowercase)

---

## API (v1)

Base URL: `/api`

### Health
- `GET /healthz` → 200 OK

### Trips
- `GET /trips?query=&limit=&offset=` (search by name)
- `POST /trips`
- `GET /trips/{tripId}`
- `PATCH /trips/{tripId}`
- `DELETE /trips/{tripId}` (optional)

### Stops
- `GET /trips/{tripId}/stops?tag=&query=&limit=&offset=`
- `POST /trips/{tripId}/stops`
- `GET /stops/{stopId}`
- `PATCH /stops/{stopId}`
- `DELETE /stops/{stopId}` (optional)

### Export
- `GET /export/trips.json`
- `GET /export/stops.csv`

**Conventions**
- JSON request/response
- IDs are UUID strings
- All error responses:
  ```json
  { "error": { "code": "VALIDATION_ERROR", "message": "…" , "details": { } } }