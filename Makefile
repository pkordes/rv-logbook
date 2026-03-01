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
        backend/run backend/build backend/test backend/lint backend/generate \
        frontend/dev frontend/build frontend/test frontend/lint \
        db/up db/down db/migrate db/rollback db/reset

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
	$(info     make backend/test       Run all Go tests)
	$(info     make backend/lint       Run go vet + staticcheck)
	$(info     make backend/generate   Regenerate Go code from openapi.yaml)
	$(info )
	$(info   Frontend)
	$(info     make frontend/dev       Start Vite dev server (http://localhost:5173))
	$(info     make frontend/build     Build production bundle to frontend/dist/)
	$(info     make frontend/test      Run Vitest unit tests)
	$(info     make frontend/lint      Run ESLint)
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

## Run all Go tests via gotestsum.
## -race is omitted locally: it requires CGO (a C compiler).
## The race detector runs in CI on Linux where gcc is available.
backend/test:
	cd $(BACKEND_DIR) && gotestsum --format pkgname -- -count=1 ./...

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

## Run ESLint across all TypeScript/TSX source files.
frontend/lint:
	npm --prefix $(FRONTEND_DIR) run lint

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
