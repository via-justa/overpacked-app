# Sources of Truth & Spec-First Architecture

Two artifacts are authoritative. Most cross-cutting mistakes come from drifting from them.

## The two sources of truth

- **API contract → `dev/openapi.yaml`.** Every endpoint's request/response shape originates here.
  Go models and the server interface are generated from it (authoritative on the backend side).
  On the frontend, generation tooling (`make gen-api`) exists but many feature types are
  hand-maintained, so check the feature you touch before assuming the client is generated (see the
  Known drifts in `dev/repo-violations.md`).
- **Database schema → `backend/internal/migrations/sql/00001_initial_schema.sql`** (plus later
  migrations). The real schema is whatever the migrations produce.

Supporting docs (not authoritative, kept in sync):
- `dev/database-schema.mermaid` — entity-relationship diagram.
- `dev/db-migrations.md` — migration how-to.

**Divergence rule:** if docs disagree with the SQL or the spec, fix the docs to match — unless
the user explicitly asks to change the schema or contract.

## Spec-first API development

Adding or changing an endpoint is always spec-first:

1. **Edit `dev/openapi.yaml`** — add/modify the path, request body, and response schemas.
2. **Regenerate:** `make gen-api-go`. This runs oapi-codegen (v2, spec-first; config at
   `backend/.oapi-codegen.yaml`) and rewrites `backend/internal/api/api.gen.go` with models,
   enums, and the `ServerInterface`.
3. **Implement handlers** against the regenerated interface (see the `go-development` skill for
   handler conventions).
4. **Wire routes** by mounting the generated routes via `api.HandlerWithOptions(...)` in
   `backend/internal/app/routes.go`. Don't duplicate OpenAPI endpoint literals in manual route
   registration.
5. **Update the frontend types** if the frontend will consume the change. Run `make gen-api` for
   the generated OpenAPI types, but also check whether the feature uses a hand-maintained type
   (many do) — if so, update it to match the new contract.

Never edit `api.gen.go` by hand — it's overwritten on the next generate.

## Schema-change workflow

1. Add a **new** goose migration under `backend/internal/migrations/sql/` with both `Up` and
   `Down` (see `references/database-and-migrations.md`).
2. Sync `dev/database-schema.mermaid` to reflect the new schema.
3. If the change affects the API, update `dev/openapi.yaml` and regenerate (above).
4. Never edit a migration that may already be applied in a shared environment.

## Monorepo layout

```
backend/    Go service (cmd/, internal/{domain,store,http,app,api,auth,config,db,migrations,seeds})
frontend/   Vue 3 + TS app (src/{features,components,composables,lib,stores,router})
dev/        openapi.yaml, database-schema.mermaid, db-migrations.md, docker-compose, scripts
deployment/ helm charts + compose for deploy
Makefile    common tasks (install, build, test, gen-api, up/down, seed)
```

Generated/derived code (`backend/internal/api/api.gen.go`, the frontend OpenAPI types) is
produced from `dev/openapi.yaml` — treat it as build output, not source.
