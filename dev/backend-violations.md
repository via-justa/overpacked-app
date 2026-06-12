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

## Logged violations

Status key: `open` (needs fixing) · `fixed` · `wontfix` (with reason).

| # | File:line | Type | Status | Note |
|---|-----------|------|--------|------|
| 1 | `internal/store/pack.go` (`AssignSet`, `ListSets`, `RemoveSet`); `internal/domain/pack.go` (`PackSet`) | dead code | fixed | Removed — queried a `pack_sets` table no migration creates, wired to no route. Also removed unused `domain.PackingListLabel`. |
