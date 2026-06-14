# Testing and Migration

## Testing

Tests are a **blocking PR gate** (`.github/workflows/frontend-tests.yml` runs `vue-tsc` + the
Vitest suite), and coverage feeds the SonarQube new-code quality gate (≥80%). New or changed
logic ships with tests in the same change — not as a follow-up.

### Stack & commands

- **Runner:** Vitest (config in `frontend/vitest.config.ts`, merged with `vite.config.ts`).
- **DOM:** happy-dom. **Component testing:** `@testing-library/vue` (primary, query by
  role/label) + `@vue/test-utils` (escape hatch, `flushPromises`). **HTTP mocking:** MSW.
  **Coverage:** `@vitest/coverage-v8`. **E2E:** Playwright.
- Commands: `npm run test` (`vitest run`), `npm run test:watch`, `npm run coverage`,
  `npm run type-check`, `npm run e2e`. Make wrappers: `make test-frontend` (type-check + tests +
  lints), `make coverage-frontend`, `make e2e`.
- The Vitest env pins `TZ=UTC`, so date/age code is deterministic; for code reading
  `new Date()`/`Date.now()`, use `vi.useFakeTimers().setSystemTime(...)`.

### Naming & layout

- Co-locate unit/component tests as `*.test.ts` next to the source (`conversions.ts` +
  `conversions.test.ts`, `LoginView.vue` + `LoginView.test.ts`).
- Shared test infrastructure lives in `src/test/`. Playwright E2E lives in `frontend/e2e/` as
  `*.spec.ts` (kept out of the Vitest `include`; runs on pre-release via `e2e.yml`).
- Generated `src/lib/api/schema.ts` is never tested and is excluded from coverage.

### Shared infrastructure (`src/test/`) — use it, don't hand-roll mocks

- **`renderWithProviders(component, opts)`** — renders with the same plugins as `main.ts`:
  PrimeVue (+ ToastService + tooltip), a testing Pinia, a memory router, and a fresh
  QueryClient. Returns Testing Library's result plus `router`. Use for component tests.
- **`withSetup(() => composable())`** — runs a composable inside a real component instance
  (lifecycle + `provide`/`inject` + `useQuery` work) with the Vue Query plugin installed. Use
  for composable tests. For provide/inject composables (e.g. `useTripPlanner`), call the
  provider function (`provideTripPlanner()`) inside it.
- **`makeTestQueryClient()`** — retry off, `gcTime: 0`, no refetch on focus/reconnect. Always
  use this in tests; never the app's `queryClient`. Clients are auto-drained in `afterEach`.
- **MSW** — `src/test/msw/server.ts` + `handlers/` (one per feature, typed from the OpenAPI
  schema). Mock at the **HTTP boundary**, never stub `apiClient`. Override per test with
  `server.use(...)`. `onUnhandledRequest: 'error'`, so an un-mocked request fails loudly.
- **`fixtures/`** — `personFixture`/`itemFixture`/`settingsFixture`, typed from
  `components['schemas']` with `Partial` overrides.
- **`setup.ts`** — MSW lifecycle, query-client drain, and `ResizeObserver`/`matchMedia` stubs
  (needed by `useIsMobile` and PrimeVue overlays).

**Envelope contract:** `unwrapApiResponse` returns openapi-fetch's `data` (the raw body), and
endpoints return **bare** arrays/objects — `HttpResponse.json([...])` / `HttpResponse.json({...})`,
not `{ data: [...] }`.

### Per-layer patterns

- **Pure util / domain math / zod schema** — import and assert directly; cover documented
  boundaries and every branch. No providers.
- **vue-query composable** (`useSettings`): `withSetup` + a `server.use` handler; assert
  defaults before resolve, then `await vi.waitFor(...)` for the loaded value.
- **provide/inject composable** (`useTripPlanner`): `withSetup(() => provideTripPlanner())`, seed
  the reference queries via MSW, `await vi.waitFor(...)`, then exercise the public mutators and
  assert the pure transforms.
- **Component**: `renderWithProviders`; fill by label, click by role, assert emitted events /
  router navigation. Dialogs **teleport to `<body>`** — query with `screen`. Testing Library
  joins a parent element's text nodes, so for a line split by `<span>`s assert with a regex on
  the combined text.
- **Store + middleware** (`stores/auth.ts`, `lib/api/client.ts`): `setActivePinia(createPinia())`
  / direct calls; drive 401/refresh/retry through MSW handlers.
- **E2E** (`frontend/e2e/*.spec.ts`): one critical-path smoke against the running stack; gated by
  `E2E_USER`/`E2E_PASS` so it self-skips locally. See `playwright.config.ts`.

### Coverage

The local gate (`vitest.config.ts` `coverage.thresholds`) is **90% lines / 85% branches**, scoped
via `coverage.include` to the logic modules under test (`lib/**`, `composables/**`,
`features/**/utils*`, `**/schema.ts`, `stores/auth.ts`, the api/persistence layer). **Widen
`include` as you add tested areas.** Components/views are graded by SonarQube on new code rather
than the local include. When adding a new tested module, add it to `include` and keep it ≥90%.

### Requirements checklist (apply when writing or reviewing)

- [ ] New/changed pure logic, composables, stores, api/persistence have unit tests; components
      with real behavior have a Testing Library test.
- [ ] HTTP mocked via MSW (bare-body shapes), not by stubbing `apiClient`; `makeTestQueryClient`
      used, not the app client.
- [ ] Queries by role/label; assertions on observable output / emits, not internal state.
- [ ] `*.test.ts` co-located; E2E as `e2e/*.spec.ts`.
- [ ] `npm run test` and `npm run type-check` pass; `npm run coverage` meets the threshold.

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
