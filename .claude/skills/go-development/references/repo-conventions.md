# overpacked-app Backend Conventions

This is the concrete shape of the `backend/` service. When it differs from the generic Go
references, follow this — new code should be indistinguishable from what's already here.

## Layered architecture (dependencies point inward)

```
cmd/api            → builds config, db, store, handlers, router; serves
internal/domain    → entities, enums, errors        (no internal deps)
internal/store     → data access over domain types  (depends on domain)
internal/http/handlers → HTTP transport             (depends on store, domain, api)
internal/app       → wiring + chi routes            (depends on handlers, store, api)
internal/api       → GENERATED from dev/openapi.yaml (do not hand-edit)
internal/{auth,config,db,migrations,seeds}
```

Rule of thumb: SQL lives in `store`, entity types and business rules in `domain`, HTTP concerns
in `handlers`, and wiring in `app`. Don't let SQL or transport details leak across those lines.

## Domain layer

Entities are plain structs using `uuid.UUID`, `time.Time`, and pointers (`*string`, `*float64`)
for nullable columns. Enums are typed strings with a `const` block:

```go
type CarryStatus string

const (
	CarryStatusPacked CarryStatus = "packed"
	CarryStatusWorn   CarryStatus = "worn"
)
```

Errors are centralized in `internal/domain/errors.go`:

```go
var ErrNotFound = errors.New("not found")

type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return fmt.Sprintf("validation error: %s", e.Message) }
```

Reuse these. A "not found" condition is `domain.ErrNotFound` (checked with `errors.Is`); a bad
input is a `domain.ValidationError`. Don't invent parallel error types per package.

## Store layer

One store struct per entity, holding only `*sql.DB`, with a `New…Store` constructor; all are
aggregated by `store.Store` and built in `store.New(db)`:

```go
type ItemStore struct{ db *sql.DB }

func NewItemStore(db *sql.DB) *ItemStore { return &ItemStore{db: db} }
```

Conventions:
- Every method takes `ctx context.Context` first and uses `QueryRowContext` / `QueryContext` /
  `ExecContext`. SQL is hand-written with `$1`-style placeholders and `RETURNING` for generated
  columns (`id, created_at, updated_at`).
- Convert nullable domain pointers to/from SQL null types with the helpers in `sql_helpers.go`
  (`toNullString`, `toNullFloat64`, `toNullInt`, `strPtr`, `floatPtr`, …). Add to that file
  rather than re-implementing conversions inline.
- Wrap errors with context using `%w` and lowercase messages: `fmt.Errorf("create item: %w", err)`.
- Map "no rows" to the domain sentinel — a lookup that finds nothing returns
  `domain.ErrNotFound` (translate `sql.ErrNoRows`), so handlers can `errors.Is` on it. For
  deletes/updates, map an Exec that affected no rows with the shared
  `rowsAffectedOrNotFound(res, op)` helper in `sql_helpers.go` instead of re-writing the
  `RowsAffected()` check inline.
- `attributes`/JSON columns are marshaled to/from `map[string]any` with `encoding/json`.

## API types are generated

`internal/api/api.gen.go` is produced by `oapi-codegen` (models + std-http server) from
`dev/openapi.yaml`. **Never edit it by hand.** To change request/response shapes or endpoints,
edit `dev/openapi.yaml` and run `make gen-api-go` (and `make gen-api` for the frontend types).
Handlers convert between `api.*` DTOs and `domain.*` entities explicitly.

## Handler layer

One handler per resource, holding the dependencies it needs (usually `*store.Store`):

```go
type ItemsHandler struct{ store *store.Store }

func NewItemsHandler(st *store.Store) *ItemsHandler { return &ItemsHandler{store: st} }
```

Method signatures match the generated server interface — `(w http.ResponseWriter,
r *http.Request)` plus typed path params where the spec has them (e.g.
`itemId types.UUID`). Inside a handler:

1. Decode the body into the generated `api.*Create`/`api.*Update` type with `decodeJSON(r, &req)`
   (or `json.NewDecoder`), and `defer r.Body.Close()`.
2. Build the `domain` entity (generate IDs with `uuid.New()`, set defaults).
3. Call the store with `r.Context()`.
4. Map errors to status and respond with `writeJSON`:

```go
item, err := h.store.Items.GetByID(r.Context(), uuid.UUID(itemId))
if errors.Is(err, domain.ErrNotFound) {
	writeError(w, http.StatusNotFound, itemsErrNotFound)
	return
}
if err != nil {
	writeError(w, http.StatusInternalServerError, "failed to get item")
	return
}
writeJSON(w, http.StatusOK, itemToAPI(item))
```

- Write error responses with `writeError(w, status, msg)` (it wraps the standard
  `map[string]string{"error": msg}` body); reuse per-handler message consts (e.g.
  `const itemsErrNotFound = "item not found"`). Success bodies use `writeJSON`.
- Status mapping: `domain.ErrNotFound` → 404, `domain.ValidationError` / bad body → 400,
  anything else → 500.
- Convert domain→api with `xToAPI` functions and the pointer helpers (`intPtr`, `boolPtr`,
  `float32PtrFromFloat64`, …). `writeJSON`, `writeError`, and `decodeJSON` live in `handlers/auth.go`.

## Wiring (app layer)

`apiServer` in `internal/app` holds all handlers and satisfies the generated `ServerInterface`
by delegating each method one-to-one. Routes use `go-chi/chi/v5`. When you add an endpoint:
update the OpenAPI spec, regenerate, add the handler method, and add the delegating method on
`apiServer`.

## Configuration, auth, db, migrations

- `config.Load()` reads env vars into a `Config` struct and returns an error if a required one
  (e.g. `DATABASE_URL`) is missing. Add new settings there.
- `internal/auth` is a `Service` issuing/validating JWTs (`golang-jwt/jwt/v5`) with a sentinel
  `ErrInvalidCredentials`.
- DB connection via `db.Connect(url)`; schema changes are goose migrations under
  `internal/migrations/sql`, run with `migrations.Run(ctx, db, "up", nil)`.

## Test conventions

- Tests are white-box (same `package handlers`, etc.), in `_test.go` files next to the code.
- **Name tests `TestXxxHandlerScenario` in CamelCase** — e.g. `TestItemsHandlerCreateInvalidBody`,
  not the `Test_func_scenario` underscore style. Match the existing names.
- Use `t.Parallel()` in tests that allow it, and `t.Helper()` in helpers.
- Handler unit tests use `net/http/httptest` (`httptest.NewRequest` + `NewRecorder`) and assert
  on `w.Code` / decoded body.
- Integration tests that need a real Postgres gate themselves: skip unless
  `RUN_CONTAINERIZED_TESTS == "true"` and `DATABASE_URL` is set, run migrations once via
  `sync.Once`, and are exercised with `make test-backend-container`. Don't write store tests
  that assume a DB is always present — follow the skip pattern.
- **Keep integration tests isolated.** They share one database across the whole run, so each
  test must stand on its own — passing alone and in the full suite, and on a re-run against a
  persisted volume:
  - If a setup helper truncates user tables but deliberately leaves a shared table (e.g.
    `item_types`) intact, insert into that table **idempotently** (`ON CONFLICT (id) DO NOTHING`)
    or use a per-test unique id. A fixed id inserted by every test collides on the second run.
  - Don't assert against (or mutate) a **pre-seeded fixture id** when the handler under test
    *creates a new row and returns its id* — operate on the id from the API response. (A create
    endpoint like `AddTripPersonPack` mints a fresh pack; the pre-inserted `packID` is unrelated.)
