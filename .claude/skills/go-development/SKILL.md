---
name: go-development
description: >-
  Write and review idiomatic Go for this repo (overpacked-app backend) following both Go
  community standards (Effective Go, Go Code Review Comments, Google's Go Style Guide) and
  this codebase's own conventions: a layered domain/store/handlers/app architecture, chi
  routing with oapi-codegen-generated API types, raw database/sql + lib/pq stores, goose
  migrations, and central domain error sentinels. Use this skill whenever the user is
  writing, editing, generating, reviewing, refactoring, or debugging Go source — work
  touching .go files, go.mod, or go.sum under backend/, adding handlers/stores/domain
  types, designing APIs, writing Go tests, handling errors, goroutines/channels, or asking
  whether Go code is idiomatic — even if they don't explicitly say "idiomatic" or "best
  practices." This is about Go source code itself; it does not apply to the Vue frontend, to
  installing the Go toolchain, or to writing Dockerfiles or CI config for the service.
---

# Go Development

Write Go that another Go programmer would nod at: simple, clear, and conventional. The
standards here come from [Effective Go](https://go.dev/doc/effective_go),
[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), and
[Google's Go Style Guide](https://google.github.io/styleguide/go/). They apply both when
you author new Go and when you review someone else's.

This file holds the core principles plus the rules you need on almost every change. Deeper,
situational guidance lives in `references/` — load a reference file only when the task
touches that area, so you keep context lean. The routing table below tells you where to go.

The generic Go standards still apply, but **this codebase has its own concrete conventions**,
and new code should match them so a reviewer can't tell which files are new. Those repo
specifics — the layered architecture, the store/handler patterns, generated API types, error
mapping, and test conventions — live in `references/repo-conventions.md`. Read it for almost
any backend change; it takes precedence over the generic references where they differ.

## This repo at a glance (overpacked-app backend)

A Go 1.26 service under `backend/`, layered so dependencies point inward:

- `internal/domain` — entities, typed enums, and the central `errors.go` (sentinel
  `ErrNotFound`, `ValidationError`). Depends on nothing else internal.
- `internal/store` — data access. One `XStore{db *sql.DB}` per entity (`NewXStore(db)`),
  aggregated by `store.Store` (`store.New(db)`). Raw `database/sql` + `lib/pq`, hand-written
  SQL, null conversions in `sql_helpers.go`.
- `internal/api` — **generated** by oapi-codegen from `dev/openapi.yaml` (`api.gen.go`). Never
  hand-edit; regenerate with `make gen-api-go`.
- `internal/http/handlers` — one `XHandler` per resource holding `*store.Store`; methods match
  the generated server interface, decode/encode JSON via the `decodeJSON`/`writeJSON` helpers,
  and map domain errors to status codes.
- `internal/app` — wiring: `apiServer` implements the generated `ServerInterface` by delegating
  to handlers; chi router in `routes.go`.
- `internal/{auth,config,db,migrations,seeds}` — JWT auth (golang-jwt/v5), env-based config,
  DB connection, goose migrations (`migrations/sql`), yaml seeds.

Key libraries: `go-chi/chi/v5`, `oapi-codegen`, `lib/pq`, `golang-jwt/jwt/v5`, `google/uuid`,
`pressly/goose/v3`. See `references/repo-conventions.md` for how to add to each layer.

## Core philosophy

Favor clarity over cleverness. The reader of Go code matters more than the writer, and Go's
culture rewards code that is obvious over code that is impressive. A few habits carry most of
the weight:

- Keep the happy path left-aligned. Return early on errors and edge cases so the main logic
  isn't buried inside nested `if`/`else`. Prefer `if cond { return ... }` over else blocks.
- Make the zero value useful, so callers can use a type without ceremony.
- Write self-documenting code — clear, descriptive names beat comments that explain what the
  code does. Reserve comments for *why*.
- Lean on the standard library before reaching for a dependency or hand-rolling something
  (`strings.Builder`, `filepath.Join`, `errors.Is`/`As`, the `net/http` mux, etc.).
- Write comments and docs in English by default; translate only if the user asks. Avoid emoji
  in code, comments, and docs.

## Package declarations (read before creating or editing any .go file)

Duplicate or wrong `package` declarations are a compile error and a common automated-editing
mistake, so handle them deliberately:

- Every `.go` file has **exactly one** `package` line. Never add a second.
- Editing an existing file: preserve its current package name. If you rewrite the whole file,
  start with that same package name.
- Creating a file in an existing directory: first check what package the sibling `.go` files
  use, and match it. In a brand-new directory, the package name is normally the directory name.
- Before adding a `package` line via any write/replace tool, verify the target doesn't already
  have one.

## Routing table — where to find detailed guidance

Read the matching reference file when your task touches that area. Don't preload them all.

| If the task involves...                                              | Read |
|----------------------------------------------------------------------|------|
| **Almost any backend change** — layers, stores, handlers, generated API types, error mapping, this repo's test style | `references/repo-conventions.md` |
| Naming packages/vars/funcs/interfaces, formatting, comments, doc comments | `references/naming-style-and-docs.md` |
| Returning, wrapping, checking, or designing errors                   | `references/errors.md` |
| Goroutines, channels, mutexes, `sync` primitives, cancellation       | `references/concurrency.md` |
| Type design, pointer vs value receivers, interfaces, project layout  | `references/types-interfaces-and-structure.md` |
| JSON API shaping and outbound HTTP clients (handler/routing detail is in repo-conventions) | `references/api-design.md` |
| Allocation hot paths, `sync.Pool`, readers/buffers, streaming, pprof | `references/performance-and-io.md` |
| Writing tests, table-driven tests, helpers, input validation, crypto | `references/testing-and-security.md` |

## Validating Go code

For anything mechanically checkable — formatting, unused imports, suspicious constructs,
failing tests — run the real toolchain instead of eyeballing it. `scripts/check.sh` wraps the
standard commands and reports only what's actionable:

```bash
bash scripts/check.sh [path]   # defaults to ./...
```

It runs `gofmt -l`, `go vet`, `golangci-lint run` (if installed), and `go test`. Each step is
optional and skipped gracefully if the tool or files aren't present.

In this repo, the canonical gates are the Makefile targets (run from the repo root) — prefer
them so you match what the team and CI run:

- `make build-backend` — `go build ./...` (compiles; catches type errors)
- `make test-backend` — `go test ./...` (unit tests; integration tests self-skip)
- `make test-backend-container` — runs tests against a containerized Postgres
  (`RUN_CONTAINERIZED_TESTS=true`); needed to exercise store/integration tests
- `make gen-api-go` — regenerate `internal/api/api.gen.go` after changing `dev/openapi.yaml`

Run these after writing or editing Go, and when reviewing, to ground feedback in what actually
fails. Subjective qualities — naming, architecture, interface design — still need your judgment.

## When reviewing Go code

Lead with correctness and clarity, not style nits a formatter would catch. Run `gofmt`/`go vet`
mentally or via the script first so you don't spend review attention on mechanical issues. Then
look for the things tools miss: unchecked errors, goroutine leaks and missing cancellation,
concurrent map access, ignored `defer` cleanup, nil-interface-vs-nil-pointer confusion,
stuttering names, leaky abstractions, and missing context on propagated errors. Explain *why* a
change matters so the author learns the principle, not just the fix.

## Violations — read and record (`dev/backend-violations.md`)

The catalog of backend violation types (unchecked errors, goroutine leaks, missing `defer`
cleanup, stuttering names, `interface{}`, hand-editing `api.gen.go`, SQL leaking into handlers,
raw `sql.ErrNoRows` instead of `domain.ErrNotFound`, bypassing `writeJSON`/`decodeJSON`, wrong
test naming, non-reversible migrations, non-canonical units, etc.) lives in
**`dev/backend-violations.md`** at the repo root, alongside a running log of specific instances
found in the codebase.

- **Before writing or reviewing**, read that file — the catalog says what to avoid, the open log
  entries say what's already known broken (e.g. the `pack_sets` dead code).
- **While working, when you spot a violation**, append it to the log there (`file:line`, the
  catalog type number, status `open`, a short note) rather than fixing-and-forgetting, so the next
  review is cheaper. Mark items `fixed` when resolved.
- Keep that catalog in sync with this skill: if you add a rule here, add the matching violation
  type there.
