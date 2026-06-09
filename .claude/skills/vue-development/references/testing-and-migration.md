# Testing and Migration

## Testing — current state

This repo does **not** currently have a unit or e2e test runner installed. There is no Vitest,
no Vue Test Utils, and no Playwright/Cypress in `frontend/package.json`. So:

- Don't fabricate a test command or claim `npm test` works — it isn't configured.
- Don't add a test file in a style that assumes a runner that isn't present.
- The available automated gates are type-checking (`vue-tsc`) and the lint scripts; treat
  those as the safety net for now.

If the user explicitly wants to add testing, recommend the conventional Vue 3 stack and set it
up deliberately rather than scattering ad-hoc tests:
- **Unit/component:** Vitest + `@vue/test-utils` (jsdom environment), tests co-located as
  `*.spec.ts` next to the unit or under a `__tests__/` folder in the feature.
- **e2e:** Playwright.
Structure components and composables to be testable regardless: keep logic in composables
(easy to test in isolation), keep components presentational, and avoid hidden module-level
side effects.

## Migrating Options API / Vue 2 code

New code is Composition API with `<script setup lang="ts">`. If you encounter Options API or
Vue 2 patterns:

- Prefer **incremental migration** over a rewrite. Convert one component (or one concern) at a
  time and keep behavior parity first, then modernize internals.
- Map the pieces directly: `data` → `ref`/`reactive`; `computed` (options) → `computed()`;
  `methods` → plain functions; `watch` (options) → `watch()`; lifecycle hooks → their
  `onMounted`/`onUnmounted` equivalents; mixins → composables.
- Move shared logic that was in mixins or repeated across components into a `use*` composable
  as you go.
- Call out any behavioral risk in the migration (timing of watchers, `this` semantics, reactive
  caveats) so the change can be reviewed safely.
- Don't migrate purely for its own sake mid-task; if a file is out of scope, note it as a
  follow-up rather than expanding the change.
