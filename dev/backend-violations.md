# Backend Violations Log

Central catalog of backend (Go) code violations **and** a running log of specific instances found
in the codebase. Companion to the `go-development` skill: the skill defines the *correct*
patterns; this file tracks what counts as a violation and where violations have been spotted, so
reviews and cleanups carry over between sessions.

## How agents use this file

- **Before writing or reviewing** backend code, read the **Catalog** (what to avoid) and the
  **Open** entries below.
- **When you find a violation while working**, append a row with `file:line`, the catalog type,
  status `open`, and a short note.
- **When you fix one**, set status to `fixed` (note the PR/commit) or remove the row.
- Keep the **Catalog** in sync with the `go-development` skill.

## Catalog — backend violation types

Detail and the correct pattern for each are in the `go-development` skill (SKILL.md +
`references/`, especially `references/repo-conventions.md`).

1. **Unchecked errors**, or `_`-ignored errors without a documented reason.
2. **Goroutine leaks** — a goroutine with no defined exit path.
3. **Concurrent map writes / unguarded shared mutable state.**
4. **Missing `defer`** for closing files, response bodies, connections.
5. **nil interface vs nil pointer** confusion.
6. **Stuttering names** (`http.HTTPServer` instead of `http.Server`).
7. **`interface{}`** instead of `any`, or better a concrete type / constrained generic.
8. **Ignoring the zero value** of a type.
9. **Duplicate `package` declarations.**
10. **Hand-editing `internal/api/api.gen.go`** instead of editing `dev/openapi.yaml` + `make gen-api-go`.
11. **SQL or business logic leaking into handlers** — queries belong in `store`, types/rules in `domain`.
12. **Returning a raw DB error / `sql.ErrNoRows`** instead of mapping to `domain.ErrNotFound`.
13. **Not mapping domain errors to HTTP status** in handlers (`ErrNotFound`→404, `ValidationError`→400).
14. **Bypassing the `writeJSON` / `decodeJSON` helpers** for ad-hoc encode/decode.
15. **Test names not in `TestXxxHandlerScenario` CamelCase** (don't use `Test_func_scenario`).
16. **Migration without a `Down` block**, or editing an already-applied migration.
17. **Non-canonical unit columns** (weights not `_grams`, volumes not `_ml`, etc.).
18. **Integration tests not isolated/idempotent** — fixed IDs inserted into tables the truncation helper skips (collide across runs of a shared DB), or assertions/calls made against state the handler doesn't actually use (e.g. a pre-inserted row the handler ignores while creating its own).

## Logged violations

Status key: `open` (needs fixing) · `fixed` · `wontfix` (with reason).

| # | File:line | Type | Status | Note |
|---|-----------|------|--------|------|
| 1 | `internal/store/pack.go:355` (`AssignSet`), `:369` (`ListSets`), `:398` (`RemoveSet`); `internal/domain/pack.go:40` (`PackSet`) | dead code | open | These query a `pack_sets` table (`INSERT/SELECT/DELETE FROM pack_sets`) that **no migration creates** — they fail at runtime. Either add the migration or remove the dead code + `PackSet` type. Not wired to any route. |
| 2 | `internal/http/handlers/trips_test.go:178` (`insertTripTestData`) | 18 | fixed | Inserted a fixed `item_type` id `"test-item-type"`, but the truncation helper deliberately skips `item_types`, so the row survived and the insert hit `item_types_pkey` on the 2nd+ test/run against a shared DB. Now generates a unique id per call (`"test-item-type-"+uuid`). |
| 3 | `internal/http/handlers/trips_test.go` (`TestTripsHandlerIntegrationNestedGet`, `TestTripsHandlerIntegrationTripPersonPacks`) | 18 | fixed | Asserted against / removed the pre-inserted `packID` from `insertTripTestData`, but `AddTripPersonPack` creates a brand-new pack with a generated id and ignores the pre-inserted one. Nested GET showed the linked (new) pack with 0 items; remove targeted an unlinked pack → 404. Now capture and use the handler-returned pack id. |
| 4 | `internal/http/handlers/trips.go:636` (`AddTripPersonPackItem`) | 11 | open | Validates only that the pack *exists* (`Packs.GetByID`), not that it is linked to the trip-person via `trip_person_packs`. Lets items be added to an unlinked pack; this is what let test #3's bug masquerade as working. Should verify the pack belongs to the trip-person. |
