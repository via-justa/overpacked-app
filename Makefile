.PHONY: help install install-backend install-frontend up down logs backend frontend build build-backend build-frontend test test-backend test-backend-container test-frontend coverage-frontend e2e check-api-gen check-api-gen-go gen-api gen-api-go clean-frontend seed test-data

COMPOSE ?= docker compose -f dev/docker-compose.yml

help:
	@echo "Overpacked - available targets"
	@echo "  make install           		Install backend and frontend dependencies"
	@echo "  make up                		Start visual testing stack (frontend, backend, db)"
	@echo "  make down              		Stop full dev stack"
	@echo "  make logs              		Tail logs from full dev stack"
	@echo "  make backend           		Start backend container only (with db)"
	@echo "  make frontend          		Start frontend container only"
	@echo "  make build             		Build backend and frontend"
	@echo "  make build-backend     		Build backend binaries"
	@echo "  make build-frontend    		Type-check and build frontend"
	@echo "  make test              		Run backend and frontend checks"
	@echo "  make test-backend      		Run backend Go tests (incl. API-gen drift check)"
	@echo "  make check-api-gen-go  		Fail if backend Go API types are stale vs the spec"
	@echo "  make test-backend-container 	Run backend Go tests (incl. full-stack E2E) against containerized Postgres with coverage"
	@echo "  make test-frontend     		Run frontend type-check, unit tests + lints (incl. API-gen drift check)"
	@echo "  make coverage-frontend 		Run frontend tests with coverage (lcov for SonarQube)"
	@echo "  make e2e               		Run Playwright critical-path E2E (local dev stack)"
	@echo "  make check-api-gen     		Fail if frontend OpenAPI types are stale vs the spec"
	@echo "  make gen-api-go        		Regenerate Go API types from OpenAPI spec"
	@echo "  make gen-api           		Regenerate frontend OpenAPI types"
	@echo "  make seed              		Run database seeds (local)"
	@echo "  make test-data         		Load dev test data (backend auto-seeds on startup, docker-compose)"
	@echo "  make clean-frontend    		Remove frontend dist output"

install: install-backend install-frontend

install-backend:
	cd backend && go mod download

install-frontend:
	cd frontend && npm install

up:
	$(COMPOSE) up --build -d db backend frontend

down:
	$(COMPOSE) down --volumes

restart:
	$(COMPOSE) restart backend frontend

logs:
	$(COMPOSE) logs -f

backend:
	$(COMPOSE) up --build backend

frontend:
	$(COMPOSE) up --build frontend

build: build-backend build-frontend

build-backend:
	cd backend && go build ./...

build-frontend:
	cd frontend && npm run build

test: test-backend test-frontend

test-backend: check-api-gen-go
	cd backend && go test ./...

# Fails if the generated Go API types drift from dev/openapi.yaml.
# Regenerates in place, then fails if the file is modified or not committed.
check-api-gen-go:
	$(MAKE) gen-api-go
	@if [ -n "$$(git status --porcelain -- backend/internal/api/api.gen.go)" ]; then \
		echo "✗ backend Go API types are out of date or uncommitted."; \
		echo "  Run 'make gen-api-go' and commit backend/internal/api/api.gen.go."; \
		git status --porcelain -- backend/internal/api/api.gen.go; \
		exit 1; \
	fi

test-backend-container:
	$(COMPOSE) up -d db
	# -p 1 serializes package test binaries: integration tests across packages share one
	# database, so running them concurrently lets one package's TRUNCATE/migrations clobber another's.
	# -coverpkg=./... credits store/backup/app code exercised by tests in other packages.
	# The profile is written to /workspace/backend/coverage.out, which the compose
	# ../:/workspace bind mount surfaces as host backend/coverage.out (gitignored).
	$(COMPOSE) run --rm -e RUN_CONTAINERIZED_TESTS=true -e JWT_SECRET=test-secret backend \
		go test -p 1 -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...

test-frontend: check-api-gen
	cd frontend && npx vue-tsc -b && npm run test && npm run lint:theme && npm run lint:icons

# Fails if the generated OpenAPI types drift from dev/openapi.yaml.
# Regenerates in place, then fails if the file is modified or not committed.
check-api-gen:
	cd frontend && npm run gen:api
	@if [ -n "$$(git status --porcelain -- frontend/src/lib/api/schema.ts)" ]; then \
		echo "✗ frontend OpenAPI types are out of date or uncommitted."; \
		echo "  Run 'make gen-api' and commit frontend/src/lib/api/schema.ts."; \
		git status --porcelain -- frontend/src/lib/api/schema.ts; \
		exit 1; \
	fi

gen-api-go:
	cd backend && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate models,std-http -package api -o internal/api/api.gen.go ../dev/openapi.yaml

gen-api:
	cd frontend && npm run gen:api

seed:
	cd backend && go run ./cmd/api seed

test-data:
	$(COMPOSE) run --rm test-data

# Run the frontend unit/component suite with coverage. Writes frontend/coverage/lcov.info,
# which the SonarQube workflow ingests (sonar.javascript.lcov.reportPaths).
coverage-frontend:
	cd frontend && npm run coverage

# Playwright critical-path smoke against the local dev stack (vite dev proxies /api on :5173).
# Brings up db+backend+frontend, loads dev test data, installs the browser, then runs the spec.
# CI runs the same spec against the deployment stack on pre-release (.github/workflows/e2e.yml).
e2e:
	$(COMPOSE) up -d db backend frontend
	$(COMPOSE) run --rm test-data
	cd frontend && npx playwright install chromium
	cd frontend && E2E_BASE_URL=$${E2E_BASE_URL:-http://localhost:5173} E2E_USER=$${APP_USERNAME:-admin} E2E_PASS=$${APP_PASSWORD:-pw123} npm run e2e

clean-frontend:
	rm -rf frontend/dist && rm -rf frontend/node_modules