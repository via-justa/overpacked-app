---
name: repo-general-development
description: >-
  Repo-wide guidance for the overpacked-app monorepo (backpacking gear manager: Go backend +
  Vue frontend + PostgreSQL) covering everything the language-specific skills don't: the domain
  model and its rules, the sources of truth (dev/openapi.yaml, the initial schema SQL, the
  mermaid diagram), the spec-first API workflow, database/goose migration rules, canonical units,
  dev setup and the fork-based git/PR workflow, code-quality and review standards, and the
  SonarQube MCP workflow. Use this skill for cross-cutting work — adding or changing an API
  endpoint end to end, schema/migration changes, understanding how trips/packs/persons/items/sets
  relate, onboarding/build/test commands, repo conventions, PR and review practices, or any task
  that spans both backend and frontend. For Go-specific code style use the go-development skill;
  for Vue-specific code style use the vue-development skill; this skill covers the rest and ties
  them together.
---

# overpacked-app — Repo-General Development

This skill captures how the **overpacked-app** repository works as a whole: its domain, its
sources of truth, its cross-cutting workflows, and its quality bar. It's the orientation layer
that sits above the two language skills.

**Pick the right skill:**

- **Backend Go code style** (handlers, stores, domain types, errors, Go tests) → `go-development`.
- **Frontend Vue code style** (components, composables, vue-query, forms, theming) → `vue-development`.
- **Everything else and anything spanning both** → this skill: the domain model, the API/DB
  contracts, migrations, the spec-first workflow, dev setup, git/PR flow, review standards, Sonar.

When a task crosses layers (e.g. "add an endpoint"), start here for the workflow, then use the
language skills for the code itself.

## The project in one paragraph

overpacked-app is a single-user **backpacking gear management** app (think lighterpack): you
catalog gear items, group them into sets, and assemble packs for people on trips, tracking
weight and volume. It's a monorepo: a **Go + PostgreSQL** backend under `backend/` (goose
migrations, JWT auth, chi + oapi-codegen) and a **Vue 3 + TypeScript** frontend under
`frontend/` (Vite, vue-query, Pinia, PrimeVue + Tailwind). The backend stores canonical units
and is the source of truth for data; the frontend handles display conversion.

## Sources of truth (do not let these drift)

Two artifacts are authoritative, and docs follow them — never the reverse:

- **API contract:** `dev/openapi.yaml`. All endpoint shapes come from here; Go types and the
  server interface are generated from it.
- **Database schema:** `backend/internal/migrations/sql/00001_initial_schema.sql`. The live
  schema is whatever the migrations produce.
- Supporting docs: `dev/database-schema.mermaid` (ER diagram) and `dev/db-migrations.md`.

If docs and SQL/spec diverge, **update the docs to match**, unless the user explicitly asks for a
schema or contract change. See `references/sources-of-truth-and-architecture.md`.

## Golden workflow rules

These apply to almost every change and are the things most easily gotten wrong:

- **Spec-first API.** To add or change an endpoint: edit `dev/openapi.yaml` first, run
  `make gen-api-go`, then implement handlers against the generated interface. Never hand-edit
  generated code, and never hand-register routes that duplicate spec literals.
- **Schema-first DB.** To change the schema: add a new goose migration (with `Up` *and* `Down`),
  then sync `dev/database-schema.mermaid` (and `dev/openapi.yaml` if the change is API-impacting).
  Never edit an already-applied migration — add a new one.
- **Canonical units.** Weight is stored in **grams**, volume in **millilitres**, always. Backend
  is canonical-first for storage, validation, and calculations; unit display/conversion happens
  on the frontend per the settings table.
- **Keep changes minimal.** No unrelated refactors; keep PRs small and focused.
- **Prefer named constants** over repeated literal strings (error messages, routes, query keys,
  header names, status labels).
- **Imports are assistive, not authoritative.** Any external-data import must use an
  "assist, then confirm" flow and must not depend on providers requiring registration, affiliate
  enrollment, paid contracts, or partner approval. Prefer open-licensed sources.
- **Validate after changes** with the repo's build/test gates (below).
- **Never commit secrets** (tokens, credentials, keys).

## Routing table — where to find detail

Read the matching reference when your task touches that area; don't preload them all.

| If the task involves...                                                        | Read |
|--------------------------------------------------------------------------------|------|
| Understanding entities & rules: persons, packs, items, item types, sets, trips, units | `references/domain-model.md` |
| The OpenAPI/DB sources of truth, spec-first API workflow, code generation       | `references/sources-of-truth-and-architecture.md` |
| Writing/running goose migrations, naming, Up/Down, canonical unit columns       | `references/database-and-migrations.md` |
| Local setup, make targets, the fork-based git flow, branches, PRs, commits      | `references/dev-workflow-and-git.md` |
| Reviewing code, quality bar, complexity, the repo review checklist              | `references/code-quality-and-review.md` |
| Using the SonarQube MCP server during a task                                    | `references/sonarqube-mcp.md` |

## Repo-wide validation

Run the Makefile gates from the repo root so you match CI and the team. `scripts/check.sh`
wraps the repo-wide ones and degrades gracefully if a toolchain is missing:

```bash
bash scripts/check.sh        # runs: make build, make test
```

The full set:

- `make install` — install backend + frontend deps
- `make build` / `make build-backend` / `make build-frontend`
- `make test` — backend (`go test ./...`) + frontend (`vue-tsc` + theme/icon lint)
- `make gen-api-go` — regenerate Go API types after editing `dev/openapi.yaml`
- `make up` / `make backend` / `make frontend` — run the dev stack via docker compose

For language-specific checks (golangci-lint, stylelint), use the `go-development` /
`vue-development` skills' own scripts.

## Violations & drifts — read and record (`dev/repo-violations.md`)

The catalog of cross-cutting violation types (skipping spec-first, editing generated code,
non-reversible migrations, unsynced mermaid, non-canonical units, unsafe imports, committed
secrets, sprawling PRs) **and the list of known code/doc drifts** live in
**`dev/repo-violations.md`** at the repo root, alongside a running log.

- **Before cross-cutting work**, read that file — especially the **Known drifts** section, which
  records where the docs or an intended rule differ from what the code actually does (e.g. packs
  being looser at the DB level than the API rule, the `pack_sets` dead code, the partly
  hand-maintained frontend types).
- **When you find a violation or a new drift while working**, append it there (`file:line`, type,
  status `open`, a short note) rather than fixing-and-forgetting.
- Keep that catalog in sync with this skill.

Stack-specific violations live in `dev/frontend-violations.md` and `dev/backend-violations.md`.
