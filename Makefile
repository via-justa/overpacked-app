.PHONY: help install install-backend install-frontend up down logs backend frontend build build-backend build-frontend test test-backend test-backend-container test-api-curl-compose test-frontend check-api-gen gen-api gen-api-go clean-frontend seed seed-compose test-data

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
	@echo "  make test-backend      		Run backend Go tests"
	@echo "  make test-backend-container 	Run backend Go tests against containerized Postgres"
	@echo "  make test-api-curl-compose 	Run full endpoint curl test against docker-compose backend"
	@echo "  make test-frontend     		Run frontend type-check (incl. API-gen drift check)"
	@echo "  make check-api-gen     		Fail if frontend OpenAPI types are stale vs the spec"
	@echo "  make gen-api-go        		Regenerate Go API types from OpenAPI spec"
	@echo "  make gen-api           		Regenerate frontend OpenAPI types"
	@echo "  make seed              		Run database seeds (local)"
	@echo "  make seed-compose      		Run database seeds (docker-compose)"
	@echo "  make test-data         		Load dev test data (runs seeds first, docker-compose)"
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

test-backend:
	cd backend && go test ./...

test-backend-container:
	$(COMPOSE) up -d db
	$(COMPOSE) run --rm -e RUN_CONTAINERIZED_TESTS=true -e JWT_SECRET=test-secret backend go test ./...

test-api-curl-compose:
	./dev/scripts/run_full_api_curl_test_with_compose.sh

test-frontend: check-api-gen
	cd frontend && npx vue-tsc -b && npm run lint:theme && npm run lint:icons

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

seed-compose:
	$(COMPOSE) run --rm seed

test-data:
	$(COMPOSE) run --rm test-data

clean-frontend:
	rm -rf frontend/dist && rm -rf frontend/node_modules