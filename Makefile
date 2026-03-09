# RV Logbook — Makefile
# Run `make help` to see all available targets.
# All targets are run from the repo root.
#
# Prerequisites: Go, Node/npm, podman-compose, goose, staticcheck, oapi-codegen
# See CONTRIBUTING.md for install instructions.

# Load .env file if present — sets DATABASE_URL etc. for local dev.
# The leading dash means "don't fail if the file doesn't exist".
-include .env
export

# ---------------------------------------------------------------------------
# Configuration — override any of these in .env or on the command line
# ---------------------------------------------------------------------------

DATABASE_URL      ?= postgres://rvlogbook:rvlogbook@localhost:5432/rvlogbook?sslmode=disable
TEST_DATABASE_URL ?= postgres://rvlogbook:rvlogbook@localhost:5432/rvlogbook_test?sslmode=disable

BACKEND_DIR  := backend
FRONTEND_DIR := frontend
COMPOSE      := podman-compose
GOOSE        := goose

# ---------------------------------------------------------------------------
# Phony targets (no output file is produced — always re-run)
# ---------------------------------------------------------------------------

.PHONY: help \
        backend/run backend/build backend/check backend/test backend/test/unit \
		backend/test/service backend/test/handler \
		backend/spec backend/spec/integration \
		backend/lint backend/generate \
        frontend/dev frontend/build frontend/test frontend/lint frontend/generate \
        db/up db/down db/migrate db/rollback db/reset \
        e2e \
        lint/encoding hooks/install

# ---------------------------------------------------------------------------
# help — self-documenting target list
# ---------------------------------------------------------------------------

help: ## Show this help message
	$(info )
	$(info RV Logbook -- available make targets)
	$(info )
	$(info   Backend)
	$(info     make backend/run        Start the Go API server)
	$(info     make backend/build      Compile Go binary to backend/bin/api)
	$(info     make backend/check      Compile all packages without producing a binary)
	$(info     make backend/test       Run all Go tests (all packages, DB required))
	$(info     make backend/test/unit  Run all tests, skip integration tests (no DB required))
	$(info     make backend/test/service  Run service-layer unit tests only (no DB))
	$(info     make backend/test/handler  Run handler-layer unit tests only (no DB))
	$(info     make backend/spec             Print unit tests as a human-readable spec (no DB))
	$(info     make backend/spec/integration Print integration test names as a spec (no DB required))
	$(info     make backend/lint       Run go vet + staticcheck)
	$(info     make backend/generate   Regenerate Go code from openapi.yaml)
	$(info )
	$(info   Frontend)
	$(info     make frontend/dev       Start Vite dev server (http://localhost:5173))
	$(info     make frontend/build     Build production bundle to frontend/dist/)
	$(info     make frontend/test      Run Vitest unit tests)
	$(info     make frontend/lint      Run ESLint + TypeScript type-check)
	$(info     make frontend/generate  Regenerate TypeScript types from openapi.yaml)
	$(info )
	$(info   Repo health)
	$(info     make lint/encoding      Check all .md/.txt files for UTF-8 BOM and mojibake)
	$(info     make hooks/install      Configure Git to use .githooks/ pre-push hook)
	$(info )
	$(info   E2E Tests)
	$(info     make e2e               Run Playwright E2E tests (requires: make db/up && make backend/run in another terminal))
	$(info )
	$(info   Database)
	$(info     make db/up              Start the Postgres container (background))
	$(info     make db/down            Stop containers (data volume persists))
	$(info     make db/migrate         Apply all pending migrations (goose up))
	$(info     make db/rollback        Roll back the last migration (goose down))
	$(info     make db/reset           Wipe DB and re-apply all migrations (dev only))
	$(info )
	@:

# ---------------------------------------------------------------------------
# Backend targets
# ---------------------------------------------------------------------------

## Start the Go API server.
## Reads config from environment / .env file.
backend/run:
	cd $(BACKEND_DIR) && go run ./cmd/api

## Compile the Go binary to backend/bin/api.
backend/build:
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api

## Compile all packages without producing a binary.
## Faster than backend/build — use this to verify a refactor compiles cleanly.
backend/check:
	cd $(BACKEND_DIR) && go build ./...

## Run all Go tests including integration tests.
## -tags integration includes files gated by //go:build integration.
## -p 1 runs one test package at a time — required because integration tests
## share a single Postgres database and would conflict if run in parallel.
## Requires TEST_DATABASE_URL to be set (see .env.example).
backend/test:
	cd $(BACKEND_DIR) && gotestsum --format pkgname -- -tags integration -count=1 -p 1 ./...

## Run service-layer unit tests only. No database required.
## Use during TDD inner loop for fast feedback on service logic.
backend/test/service:
	cd $(BACKEND_DIR) && gotestsum --format pkgname -- -count=1 ./internal/service/...

## Run handler-layer unit tests only. No database required.
## Uses httptest — no live server or DB needed.
backend/test/handler:
	cd $(BACKEND_DIR) && gotestsum --format pkgname -- -count=1 ./internal/handler/...

## Run all tests excluding integration tests (no database required).
## Integration test files are gated by //go:build integration and are not
## compiled at all without the tag — no env var trick needed.
## Used by branch CI and by developers without a running DB.
backend/test/unit:
	cd $(BACKEND_DIR) && gotestsum --format pkgname -- -count=1 ./...

## Print all unit tests as a human-readable specification using gotestdox.
## Uses ./... — automatically picks up any new packages. No DB required.
## Integration test files (//go:build integration) are excluded by default.
backend/spec:
	cd $(BACKEND_DIR) && go test -json ./... | gotestdox

## Print all integration tests as a human-readable specification.
## Scans source files directly — no database or build tags required.
backend/spec/integration:
	python scripts/spec-format.py $(BACKEND_DIR)/internal/repo $(BACKEND_DIR)/internal/apitest $(BACKEND_DIR)/testutil
## Run go vet and staticcheck.
## Both must pass with zero warnings — this mirrors the CI check.
backend/lint:
	cd $(BACKEND_DIR) && go vet ./...
	cd $(BACKEND_DIR) && staticcheck ./...

## Regenerate Go server stubs and types from backend/openapi.yaml.
## Run this any time openapi.yaml changes.
backend/generate:
	cd $(BACKEND_DIR) && go generate ./...

# ---------------------------------------------------------------------------
# Frontend targets
# ---------------------------------------------------------------------------

## Start the Vite dev server with hot module replacement.
## Proxies /api requests to the Go backend (configure in vite.config.ts).
frontend/dev:
	npm --prefix $(FRONTEND_DIR) run dev

## Build the production frontend bundle to frontend/dist/.
frontend/build:
	npm --prefix $(FRONTEND_DIR) run build

## Run Vitest unit tests (single run, no watch mode).
frontend/test:
	npm --prefix $(FRONTEND_DIR) run test -- --run

## Run ESLint and TypeScript type-check across all source files.
frontend/lint:
	npm --prefix $(FRONTEND_DIR) run lint
	npm --prefix $(FRONTEND_DIR) run typecheck

## Regenerate TypeScript types from backend/spec/openapi.yaml.
## Run this whenever the OpenAPI spec changes.
frontend/generate:
	npm --prefix $(FRONTEND_DIR) run generate

# ---------------------------------------------------------------------------
# Database targets (all use goose against the local compose Postgres)
# ---------------------------------------------------------------------------

## Start the Postgres container in the background.
db/up:
	$(COMPOSE) up -d
	@echo Waiting for Postgres to be ready...
	@$(COMPOSE) exec postgres pg_isready -U rvlogbook
	@echo Postgres is ready.

## Stop containers. Data volume is preserved — use db/reset to wipe data.
db/down:
	$(COMPOSE) down

## Apply all pending migrations using goose.
db/migrate:
	$(GOOSE) -dir $(BACKEND_DIR)/migrations postgres "$(DATABASE_URL)" up

## Roll back the most recently applied migration.
db/rollback:
	$(GOOSE) -dir $(BACKEND_DIR)/migrations postgres "$(DATABASE_URL)" down

## Wipe the dev database and re-apply all migrations from scratch.
## WARNING: destroys all local data. Never run against production.
db/reset:
	$(GOOSE) -dir $(BACKEND_DIR)/migrations postgres "$(DATABASE_URL)" reset
	$(GOOSE) -dir $(BACKEND_DIR)/migrations postgres "$(DATABASE_URL)" up

# ---------------------------------------------------------------------------
# E2E targets
# ---------------------------------------------------------------------------

## Run Playwright end-to-end tests.
## Prerequisites (must be running before invoking this target):
##   make db/up && make db/migrate   (once, to start Postgres and apply schema)
##   make backend/run                (in a separate terminal)
## Playwright starts the Vite dev server automatically via playwright.config.ts.
e2e:
	npm --prefix $(FRONTEND_DIR) run e2e

# ---------------------------------------------------------------------------
# Repo health targets
# ---------------------------------------------------------------------------

## Check all .md and .txt files for UTF-8 BOM and cp1252 mojibake sequences.
## Mojibake happens when a UTF-8 file is decoded as cp1252 and re-saved,
## turning em-dashes (—) into the garbled sequence â€" and box-drawing
## characters into â"‚.
## This same check is enforced by the pre-push hook (make hooks/install)
## and by the branch CI workflow.
lint/encoding:
	python scripts/check-encoding.py .

## Configure Git to use the .githooks/ directory for local hooks.
## Run this once after cloning the repo.
## The pre-push hook will then block any push that contains encoding issues.
hooks/install:
	git config core.hooksPath .githooks
	@echo "Hooks installed. .githooks/pre-push will run before each git push."
