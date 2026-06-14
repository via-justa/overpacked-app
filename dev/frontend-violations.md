# Frontend Violations Log

Central catalog of frontend (Vue) code violations **and** a running log of specific instances
found in the codebase. This is the companion to the `vue-development` skill: the skill defines
the *correct* patterns; this file tracks what counts as a violation and where violations have been
spotted, so reviews and cleanups can pick up where the last one left off.

## How agents use this file

- **Before writing or reviewing** frontend code, read the **Catalog** (what to avoid) and the
  **Open** entries below (known outstanding issues).
- **When you find a violation while working**, append a row to the log with `file:line`, the
  catalog type, status `open`, and a short note. Don't silently fix-and-forget — record it so the
  pattern is visible across reviews.
- **When you fix one**, set its status to `fixed` (and note the PR/commit) or remove the row.
- Keep the **Catalog** in sync with the `vue-development` skill; if you add a rule to the skill,
  add the matching violation type here.

## Catalog — frontend violation types

Detail and the correct pattern for each are in the `vue-development` skill (SKILL.md +
`references/`). `(lint)` = also caught by a lint script.

1. **Raw Tailwind palette classes** (`bg-emerald-500`, `text-stone-700`, …) — use semantic theme
   tokens. `(lint: npm run lint:theme)`
2. **Direct PrimeIcons** (`icon="pi pi-…"`, `class="pi pi-…"`) — use `AppIcon` / `iconRegistry`.
   `(lint: npm run lint:icons)`
3. **Hardcoded colors** (hex, `rgb()`, `hsl()`, `bg-[#…]`) or themeable inline `style="…"` —
   breaks theming; use semantic tokens or bind to `var(--color-…)`.
4. **Server data cached in a `ref`/Pinia** instead of a `@tanstack/vue-query` composable.
5. **Pinia used for component-local state** (Pinia is for auth / global UI only).
6. **Untyped or implicit props/emits**, or stray `any`.
7. **Missing loading / empty / error handling** around async data.
8. **Orchestration / domain logic inlined in a component** that should live in a composable.
9. **Hand-written API types** duplicating the generated OpenAPI types.
10. **Broad/deep watchers** where a `computed` would do.
11. **Click-only `<div>`/`<span>`** instead of a real `<button>`/`<a>` (not keyboard-operable).
12. **Custom menu/overlay missing keyboard support** (arrow/Home/End/Enter/Escape) or not
    restoring focus to the trigger on close.
13. **Removed focus outline** with no visible focus replacement.
14. **Direct `localStorage` access** instead of the `lib/storage/localStorage.ts` helpers.
15. **Duplicated viewport/breakpoint logic** instead of a shared composable, or a threshold other
    than 768px (Tailwind `md`).
16. **New/changed logic shipped without tests** — pure utils, composables, stores, api/persistence,
    or behavioral components (conditional rendering, emits, derived display, forms) added or
    modified without a co-located `*.test.ts`. Also covers tests that stub `apiClient` instead of
    mocking the HTTP boundary with MSW, use the app's `queryClient` instead of
    `makeTestQueryClient`, or assert internal state instead of role/label output.
    `(gate: frontend-tests.yml · SonarQube new-code coverage ≥80%)`

## Logged violations

Status key: `open` (needs fixing) · `fixed` · `wontfix` (with reason).

| # | File:line | Type | Status | Note |
|---|-----------|------|--------|------|
| _(none open)_ | | | | |
