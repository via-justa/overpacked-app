---
name: vue-development
description: >-
  Write and review Vue 3 + TypeScript frontend code the way this repo (overpacked-app)
  does it: script-setup Composition API, @tanstack/vue-query for server state, Pinia for
  auth/UI only, vee-validate + zod forms, PrimeVue + Tailwind theming, and the feature-based
  src/features layout. Use this skill whenever the user is writing, editing, generating,
  reviewing, refactoring, or debugging frontend code — work touching .vue files or files
  under frontend/src (components, composables, stores, views, the api or query libs),
  building forms, wiring data fetching, styling with Tailwind/PrimeVue, or asking whether a
  component is idiomatic — even if they don't say "Vue" or "best practices." This is about
  Vue and TypeScript source in this app; it does not apply to the Go backend, to build/CI/Docker
  config, or to installing tooling.
---

# Vue Development (overpacked-app frontend)

Write Vue that fits this codebase, not generic Vue. The frontend is a Vue 3 + TypeScript +
Vite app, and the team has settled on specific tools and conventions. New code should look
like the code already here; a reviewer should not be able to tell which files are new. The
goal of this skill is to encode those conventions so that what you produce drops straight in.

This file holds the core principles and the rules that apply to almost every change. Deeper,
situational detail lives in `references/` — load a reference file only when the task touches
that area, so context stays lean. The routing table below points the way.

## The stack (what to reach for)

- **Framework:** Vue 3, always `<script setup lang="ts">` and the Composition API.
- **Server state:** `@tanstack/vue-query`. Remote data is never hand-cached in a ref or a
  store — it lives in query/mutation composables.
- **Client state:** Pinia, and only for state that is genuinely app-wide (auth, UI prefs).
  Component-local state stays in the component; server data stays in vue-query.
- **Forms & validation:** `vee-validate` with `zod` schemas, wired through
  `buildTypedSchema` (`lib/validation/schema.ts`).
- **API access:** the shared `openapi-fetch` client in `lib/api/client.ts`; types come from
  the generated OpenAPI types, not hand-written duplicates.
- **UI & styling:** PrimeVue components + Tailwind v4, themed via `@primeuix/themes`.
- **Icons:** the `AppIcon` component and the registry in `lib/icons` — never raw PrimeIcons.

## Two rules that are mechanically enforced (don't break them)

The repo has lint scripts that fail CI on these, so getting them wrong is not a style opinion
— it's a broken build:

1. **No raw Tailwind palette classes.** Never write color utilities like `bg-emerald-500`,
   `text-stone-700`, `border-rose-400`, etc. Use the semantic theme tokens instead. Raw
   palette classes are allowed only in `src/style.css`. (Enforced by
   `scripts/check-no-raw-palette-classes.mjs`.)
2. **No direct PrimeIcons.** Never write `icon="pi pi-..."`, `class="pi pi-..."`, or
   `:class="pi pi-..."`. Use the `<AppIcon>` component or the `iconRegistry` template
   literals from `lib/icons`. (Enforced by `scripts/check-no-direct-primeicons.mjs`.)

When you touch markup that uses an icon or a color, check it against these before finishing.

## Core principles

- **Composition-centric.** Pull reusable logic into composables (`use*`) with a single clear
  responsibility. Don't duplicate orchestration across components.
- **Type-safe by default.** Explicitly type props and emits; lean on the generated API types;
  avoid `any`. The build runs `vue-tsc`, so type errors break it.
- **Separate UI from orchestration.** As a component grows, push data fetching and mutations
  into a composable and keep the component focused on rendering.
- **Handle every async state.** Loading, empty, success, and error are all real UI states —
  render each explicitly rather than assuming the happy path.
- **Accessible by default.** Prefer semantic HTML and keyboard-operable controls; don't reach
  for raw DOM manipulation unless it's required and isolated.
- **Performance-aware.** Use `computed`/`watch` intentionally; avoid broad or deep watchers
  without a reason; lazy-load feature routes.

## Routing table — where to find detailed guidance

Read the matching reference file when your task touches that area. Don't preload them all.

| If the task involves...                                                      | Read |
|------------------------------------------------------------------------------|------|
| Where files go, the feature layout, the api + vue-query data layer, Pinia scope, query keys | `references/architecture-and-state.md` |
| Building components/composables, `<script setup>` patterns, props/emits, slots | `references/components-and-composables.md` |
| Forms, vee-validate, zod schemas, `buildTypedSchema`, shared validators       | `references/forms-and-validation.md` |
| Tailwind theme tokens, PrimeVue, the palette rule, AppIcon/iconRegistry, stylelint | `references/styling-and-icons.md` |
| Rendering performance, `computed`/`watch` discipline, code splitting, a11y    | `references/performance-and-accessibility.md` |
| Shared UI patterns/gotchas: dialogs, Teleport for fixed UI, row-action menus, localStorage, unit display, responsive layout | `references/ui-patterns.md` |
| Adding tests, or migrating Options API / Vue 2 code to Composition API        | `references/testing-and-migration.md` |

## Validating frontend code

For the checks the team automates, run them rather than eyeballing — they're the same gates
CI applies. `scripts/check.sh` wraps them and reports only what's actionable:

```bash
bash scripts/check.sh          # auto-detects the frontend/ directory
```

It runs `vue-tsc` (type check), `stylelint`, and the two custom lint scripts (`lint:theme`,
`lint:icons`). Each step is skipped gracefully if the tool or directory isn't present. Run it
after writing or editing frontend code, and when reviewing, so feedback is grounded in what
actually fails. Note: this repo has **no unit/e2e test runner installed** (no Vitest, no
Playwright) — see `references/testing-and-migration.md` before assuming you can run tests.

## When reviewing Vue code

Lead with the things tools can't catch: is server state in vue-query rather than a ref or
Pinia? Are props/emits explicitly typed? Are loading/empty/error states all handled? Is logic
that belongs in a composable inlined in the component? Then verify the two enforced rules
(palette classes, direct PrimeIcons) and run the script for type/lint issues. Explain *why* a
change matters so the author internalizes the convention.

Also check **keyboard accessibility** on any interactive change: real elements (not click-only
divs), keyboard-operable custom menus/overlays (arrow/Home/End/Enter/Escape), and focus that
moves into an overlay on open and returns to the trigger on close. See
`references/performance-and-accessibility.md` for the full keyboard rules — this is a valid basis
for suggesting fixes in a review, including a later scan of existing components.

## Violations — read and record (`dev/frontend-violations.md`)

The catalog of frontend violation types (raw palette classes, direct PrimeIcons, hardcoded
colors/inline styles, server data in a ref, untyped props, missing async states, inaccessible
controls, etc.) lives in **`dev/frontend-violations.md`** at the repo root, alongside a running
log of specific instances found in the codebase.

- **Before writing or reviewing**, read that file — the catalog tells you what to avoid, and the
  open log entries tell you what's already known to be broken.
- **While working, when you spot a violation**, append it to the log there (`file:line`, the
  catalog type number, status `open`, a short note) instead of fixing-and-forgetting. This makes
  the next review cheaper. Mark items `fixed` when you resolve them.
- Keep that catalog in sync with this skill: if you add a rule here, add the matching violation
  type there.
