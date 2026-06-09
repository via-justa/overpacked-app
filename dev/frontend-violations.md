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

## Logged violations

Status key: `open` (needs fixing) · `fixed` · `wontfix` (with reason).

| # | File:line | Type | Status | Note |
|---|-----------|------|--------|------|
| 1 | `src/components/display/AppItemTableRowContent.vue:58` | 3 | open | Hardcoded contrast colors `#ffffff` / `#111827`; should be theme tokens. |
| 2 | `src/components/display/AppItemTableRowContent.vue:91` | 3 | open | Hardcoded `#6b7280` fallback for `label.color`. Dynamic `label.color` (user data) is fine; the literal fallback should be a token. |
| 3 | `src/features/items/components/ItemLabel.vue:45` | 3 | open | Same `#6b7280` label-color fallback literal. |
| 4 | `src/features/items/views/ItemsPage.vue:494`, `src/components/layout/AppTopNav.vue:46`, `src/components/actions/AppActionsMenu.vue:63` | 15 | open | The `innerWidth < 768` mobile check + resize wiring is duplicated; extract a shared `useIsMobile()` composable. |
| 5 | `src/composables/useRowActionsMenu.ts` | 12 | open | Handles positioning/dismissal but no keyboard navigation; row-action menus may not be keyboard-operable. Fold arrow/Enter/Escape handling in here. |
| 6 | custom Teleport'd menus (`AppActionsMenu.vue`, `ItemsCreateOptionsMenu.vue`) | 12 | open | Verify focus is trapped while open and **restored to the trigger** on close (PrimeVue `Dialog` does this; these custom menus must too). |

### Pending triage

- **Inline `style="…"`** appears at ~15 sites. Many are legitimate dynamic bindings (menu/tooltip
  positioning, user-chosen label colors). Triage each for type 3: a *computed dynamic* value is
  fine, a *themeable literal* is a violation. Not yet individually triaged.
