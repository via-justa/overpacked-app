# AGENTS.md

Repository guidance for coding agents (Claude Code / Cowork, GitHub Copilot, and others) working
in **overpacked-app**. This file is a concise entry point; the detailed, maintained conventions
live in the skills referenced below.

## Use the skills in `.claude/skills/`

This repo ships its conventions as skills. **Before working, read the relevant `SKILL.md`** —
each has a routing table to deeper reference files:

- **`.claude/skills/repo-general-development/SKILL.md`** — start here. Domain model and data
  tables, sources of truth, the spec-first API and DB-migration workflows, dev setup, the
  fork-based git/PR flow, code-quality and review standards, and the SonarQube MCP workflow.
- **`.claude/skills/go-development/SKILL.md`** — backend Go conventions: the
  domain/store/handlers/app layering, error mapping, generated API types, tests, and the
  `make` gates.
- **`.claude/skills/vue-development/SKILL.md`** — frontend Vue conventions: components and
  composables, vue-query/Pinia, forms, PrimeVue + Tailwind theming, the icon registry, shared UI
  patterns, and accessibility.

> **For agents:** Claude Code / Cowork auto-load these skills. Other agents (e.g. GitHub Copilot)
> do **not** — so open and follow the referenced `SKILL.md` files (and their `references/`)
> explicitly. For any task that spans both backend and frontend, start with
> `repo-general-development`, then use the language skill for the code itself.

## Precedence: code wins

The **code and SQL are the source of truth.** Where any document — including this one — disagrees
with the code, the code wins; update the doc. (This file had drifted and has been reconciled to
the code; the notes below reflect current behavior.)

## Project context

- A **single-user backpacking gear manager** (lighterpack-style): catalog items, group them into
  sets, assemble packs for persons on trips, tracking weight and volume.
- **Monorepo.** `backend/` — Go + PostgreSQL (goose migrations, chi, JWT, oapi-codegen).
  `frontend/` — Vue 3 + TypeScript (Vite, `@tanstack/vue-query`, Pinia for auth/UI only,
  PrimeVue + Tailwind, vee-validate + zod).
- Key docs: `dev/openapi.yaml`, `dev/database-schema.mermaid`, `dev/db-migrations.md`, and the
  migrations under `backend/internal/migrations/sql/`.

## Sources of truth

- **API contract → `dev/openapi.yaml`.** Backend Go types and the server interface are generated
  from it (`make gen-api-go`); never hand-edit `backend/internal/api/api.gen.go`. The frontend has
  `openapi-typescript`/`openapi-fetch` tooling, but **many feature types are hand-maintained** —
  check the feature you touch rather than assuming the client is fully generated.
- **DB schema → `backend/internal/migrations/sql/00001_initial_schema.sql`** (plus later
  migrations). Keep `dev/database-schema.mermaid` in sync; docs follow the schema.

## Core domain rules (current)

- **Auth:** single user via `APP_USERNAME` / `APP_PASSWORD` env vars; all endpoints require a JWT
  after login.
- **Items:** required fields are only `name`, `type`, `is_active`, `manufacturer`; support
  multiple types and images. Weight and volume are stored canonically (**grams**, **ml**).
- **Sets:** builder helpers. Assigning a set to a pack **inflates its items into `pack_items`**;
  pack quantities are then edited independently. The set→pack association is **not persisted** —
  there is no `pack_sets` table.
- **Packs:** managed **only through the trip → person → pack hierarchy** via the API — there are no
  standalone pack routes; the only creation path is
  `POST /api/v1/trips/{tripId}/persons/{personId}/packs`. The schema enforces this too:
  `packs.person_id` is `NOT NULL` with `ON DELETE CASCADE`, so a pack always belongs to a person and
  is removed with them.
- **Trips:** organize multi-person journeys via junctions `trip_persons`, `trip_person_packs`,
  `trip_person_items`; pack items carry `carry_status` (`packed`/`worn`); trip `type` is one of
  `day_hike`, `overnight`, `multi_day`, `thru_hike`. (Trip sets were removed.)
- **Units:** backend is canonical-first (grams, ml) for storage, validation, and calculations; the
  frontend converts for display based on the `settings` table.

Full entity and table detail (the 18-table map and relationships) is in the
`repo-general-development` skill → `references/domain-model.md`.

## Workflow rules

- **Spec-first API:** edit `dev/openapi.yaml` → `make gen-api-go` → implement handlers against the
  generated interface; mount routes via `api.HandlerWithOptions(...)`. Never hand-edit generated
  code or duplicate spec endpoint literals.
- **Schema-first DB:** add a **new** goose migration with both `Up` and `Down`, then sync
  `dev/database-schema.mermaid` (and `dev/openapi.yaml` if API-impacting). Never edit an
  already-applied migration.
- **Keep changes minimal** — no unrelated refactors; small, focused PRs.
- **Prefer named constants** over repeated literal strings (error messages, routes, query keys,
  headers, status labels).
- **Imports** use an "assist, then confirm" flow and must not depend on providers requiring
  registration, affiliate enrollment, paid contracts, or partner approval; prefer open-licensed
  sources.
- **Validate** with `make build` and `make test` before opening a PR. Never commit secrets.

Full workflow detail — dev setup, the fork-based git/PR flow, code-quality and review standards,
and the SonarQube MCP workflow — is in the `repo-general-development` skill.

## Logic notes (running log)

This file also serves as the running log of product and data-model decisions. As new decisions
are made, record them here and reconcile them into the code and the relevant skill so the skills
stay the durable source of conventions.
