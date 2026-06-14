# Repo Violations & Drift Log

Cross-cutting (repo-wide) violations and known **code/doc drifts** for overpacked-app. Companion
to the `repo-general-development` skill. Stack-specific issues live in `dev/frontend-violations.md`
and `dev/backend-violations.md`; this file covers cross-cutting workflow violations and the places
where documentation or an intended rule diverges from what the code actually does.

## How agents use this file

- **Before cross-cutting work** (API endpoints, schema/migrations, anything spanning both stacks),
  read the **Catalog**, the **Known drifts**, and the **Open** log entries below.
- **When you find a violation or a new drift while working**, append it (with `file:line` where
  applicable, type, status `open`, a short note) instead of fixing-and-forgetting.
- **When you fix or reconcile one**, mark it `fixed` (note the PR/commit) or remove the row.
- Keep the **Catalog** in sync with the `repo-general-development` skill.

## Catalog — cross-cutting violation types

Detail and the correct workflow for each are in the `repo-general-development` skill.

1. **Implementing an endpoint without spec-first** — not updating `dev/openapi.yaml` + running
   `make gen-api-go` before writing handlers.
2. **Editing generated code** (`internal/api/api.gen.go`, generated frontend types) by hand.
3. **Migration missing a `Down` block**, or editing an already-applied migration.
4. **Schema changed but `dev/database-schema.mermaid` not synced.**
5. **Storing or comparing weights/volumes in non-canonical units** (must be grams / ml).
6. **Import path that trusts external data without user confirmation**, or depends on a provider
   requiring registration / affiliate / paid contract / partner approval.
7. **Secrets committed** to the repo or baked into a Docker image.
8. **Sprawling PRs** that mix unrelated changes.

## Known drifts / code quirks

The **code and SQL are the source of truth.** These are places where documentation (including
`AGENTS.md`, historically) or an intended rule differs from what the code actually does. Keep them
in mind and reconcile when you touch the area.

- **Packs — API rule vs DB schema:** the API enforces the trip → person → pack hierarchy (no
  standalone pack routes; only `AddTripPersonPack`), but the *database* is looser:
  `packs.person_id` is nullable (`ON DELETE SET NULL`) and there's an `is_template` flag. Don't
  rely on "a pack always has a person" as a DB invariant — the hierarchy is enforced at the
  API/handler layer. (`backend/internal/store/pack.go`, schema)

## Logged violations (running)

Status key: `open` · `fixed` · `wontfix` (with reason). (No cross-cutting instances logged yet
beyond the known drifts above — append as found.)

| # | File:line | Type | Status | Note |
|---|-----------|------|--------|------|
