# Backend Architecture

## Overview

The backend is a REST API built with:
- **Framework**: Chi (lightweight HTTP router)
- **Database**: PostgreSQL (no ORM, raw SQL)
- **Language**: Go
- **API Contract**: OpenAPI 3.0.3 (source of truth in `dev/openapi.yaml`)
- **Code Generation**: oapi-codegen v2 (spec-first workflow)

## Directory Structure

```
backend/
├── cmd/api/
│   └── main.go                          # Server entry point
├── internal/
│   ├── api/
│   │   └── api.gen.go                   # Generated types, enums, ServerInterface
│   ├── app/
│   │   ├── app.go                       # Dependency container & server bootstrap
│   │   └── routes.go                    # Route registration, wires ServerInterface impl
│   ├── auth/
│   │   ├── service.go                   # JWT token service
│   │   └── service_test.go              # Auth service tests
│   ├── config/
│   │   └── config.go                    # Configuration from environment
│   ├── db/
│   │   └── db.go                        # Database connection
│   ├── domain/
│   │   ├── errors.go                    # Typed domain errors
│   │   ├── settings.go                  # Settings model (units)
│   │   ├── person.go                    # Person model
│   │   ├── manufacturer.go              # Manufacturer model
│   │   ├── item.go                      # Item model (polymorphic)
│   │   ├── set.go                       # Set/SetItem models
│   │   └── pack.go                      # Pack/PackItem/PackSet models
│   ├── http/
│   │   └── handlers/
│   │       ├── auth.go                  # Auth endpoints (login, refresh, logout)
│   │       └── auth_test.go             # Auth handler tests
│   ├── migrations/
│   │   ├── migrations.go                # Goose wrapper
│   │   └── sql/
│   │       └── 00001_initial_schema.sql # Full schema with all tables
│   └── store/
│       ├── store.go                     # Root store container
│       ├── settings.go                  # Settings CRUD
│       ├── person.go                    # Person CRUD
│       ├── manufacturer.go              # Manufacturer CRUD
│       ├── item.go                      # Item CRUD (polymorphic)
│       ├── set.go                       # Set/SetItem CRUD
│       ├── pack.go                      # Pack/PackItem/PackSet CRUD
│       └── sql_helpers.go               # Nullable type converters
├── .oapi-codegen.yaml                   # Code generation config
├── go.mod / go.sum
└── Dockerfile
```

## Key Layers

### 1. HTTP Handlers (implement generated ServerInterface)
- Implement methods from `api.ServerInterface` (generated from OpenAPI spec)
- Receive typed request params/body from oapi-codegen
- Call services/stores for business logic
- Return typed response structs (generated, auto-serialized to JSON)
- Route registration is mounted through `api.HandlerWithOptions(...)` in `backend/internal/app/routes.go`; avoid hand-writing per-endpoint route literals.

### 2. Services (Auth service example)
- `auth.Service`: JWT token issuance (Login, Refresh), validation, logout
- Validates credentials against APP_USERNAME/APP_PASSWORD (env vars)
- Issues access tokens (15m TTL) + refresh tokens (7d TTL) as HS256-signed JWTs

### 3. Store (Data Access)
- Execute raw SQL queries against PostgreSQL
- No transactions yet; can add wrapper for multi-statement operations
- Return domain entities (Person, Item, Pack, etc.)
- Handle relationship operations (set_items, pack_items, pack_sets)

### 4. Domain
- Core data models per entity (settings.go, person.go, item.go, set.go, pack.go)
- Typed errors (`ErrNotFound`) for control flow
- Enums (Gender, CarryStatus, ItemType, TripType, WeightUnit, etc.)

## Configuration

Load from environment variables:
- `DATABASE_URL` (required) - PostgreSQL DSN
- `SERVER_ADDR` (optional, default `0.0.0.0:8000`) - HTTP listen address
- `APP_USERNAME` (required) - Single user's login username
- `APP_PASSWORD` (required) - Single user's login password
- `JWT_SECRET` (required) - Secret key for signing JWT tokens

Example:
```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/overpacked"
export SERVER_ADDR="localhost:8000"
export APP_USERNAME="admin"
export APP_PASSWORD="secret123"
export JWT_SECRET="my-secret-key"
go run ./cmd/api
```

## API Development Workflow

1. **Write/update spec**: Edit `dev/openapi.yaml` with new endpoints, schemas, security
2. **Generate Go code**: Run `make gen-api-go`
   - Generates models (request/response structs)
   - Generates enums with `Valid()` methods
   - Generates `ServerInterface` with method signatures
   - Generates `ServerInterfaceWrapper` to route HTTP requests
3. **Implement handlers**: Create structs that satisfy `ServerInterface`
   - Implement each method with business logic
   - Call stores/services as needed
4. **Wire routes**: Update `routes.go` to bind handler to wrapper
5. **Test**: Unit tests for services, integration tests for handlers

## Middleware Stack

Applied via Chi:
1. **Recovery** - Catches panics, returns 500
2. **Logging** - Logs all requests with method, path, status, duration
3. **RequestID** - Injects unique request ID via `X-Request-ID` header

Route-specific middleware:
- **Auth** - (TODO) Validates bearer token against JWT secret, injects claims into context

## Error Handling

Currently implemented:
- HTTP status codes via standard Go http package
- JSON response bodies using generated models from oapi-codegen
- Example auth errors: `{"error": "message"}` with 400/401 status

(TODO) Typed domain errors:
- Could add `domain.Error` type for structured error classification
- Map domain errors to HTTP status codes and response format

## Request/Response Flow

1. **Request arrives** → Chi router matches route
2. **Middleware chain** → Recovery → Logging → RequestID
3. **Handler wrapper** (generated) → Extracts path params, query params, request body
4. **Handler implementation** → Called with typed params, returns response struct
5. **Handler wrapper** → Serializes response to JSON, writes with correct status code

## Routing Source Of Truth

- Endpoint paths are generated from `dev/openapi.yaml` into `backend/internal/api/api.gen.go`.
- `backend/internal/app/routes.go` should mount generated routing via `api.HandlerWithOptions` and only provide the server implementation and custom error handling.
- Do not duplicate route strings for OpenAPI-covered endpoints in manual Chi route declarations.

## Testing Strategy

- **Auth service tests**: Token generation, validation, TTL enforcement
- **Handler tests**: HTTP request/response marshaling, error cases
- **Route integration tests**: Chi router registration and endpoint availability
- **Store tests**: (TODO) SQL correctness, null handling
- **End-to-end**: (TODO) Docker Compose + API client scripts
