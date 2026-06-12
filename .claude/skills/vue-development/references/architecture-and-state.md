# Architecture, Data Layer, and State

## Feature-based layout

Code is organized by feature under `src/features/<feature>/`. A mature feature looks like:

```
src/features/items/
├── api/itemsApi.ts          # openapi-fetch calls, error normalization, returns typed data
├── composables/useItemQueries.ts  # vue-query wrappers (useQuery/useMutation)
├── components/              # feature-specific components (dialogs, tables, cards, forms)
├── views/ItemsPage.vue      # the routed page
├── types.ts                 # feature types (built on the generated API types)
└── utils/                   # pure helpers
```

Cross-cutting code lives outside features:
- `src/composables/` — reusable composables shared across features (`useDeleteConfirmation`,
  `useMutationWithToast`, `useInlineMutation`, `useRowActionsMenu`, `useSettings`).
- `src/components/` — shared components grouped by kind: `actions/`, `dialogs/`, `display/`,
  `forms/`, `feedback/`, `icons/`, `layout/`.
- `src/lib/` — framework-agnostic building blocks: `api/`, `query/`, `validation/`, `icons/`,
  `format/`, `table/`, `text/`, `units/`, `storage/`.
- `src/stores/` — Pinia stores. `src/router/` — routes. `src/types/` — global types.

When adding to an existing feature, match this structure. When creating a new feature, mirror
it rather than inventing a new shape.

## Data layer: API calls

API functions live in `features/<f>/api/<f>Api.ts` and use the shared `apiClient`
(`openapi-fetch`) imported from `lib/api/client.ts`. They return typed entities from the
feature's `types.ts` (which build on the generated OpenAPI types) — do not hand-write response
types that duplicate the generated ones. Wrap each call with the shared helpers from
`lib/api/request.ts` — `unwrapApiResponse(call, fallback)` for data endpoints and
`ensureApiResponse(call, fallback)` for no-content ones — which check `response.ok`/`data` and
throw a friendly message via `getErrorMessage` (in `lib/api/errors.ts`). Import those helpers;
do **not** re-roll a per-file `getErrorMessage`/`readString` (that duplication was consolidated).
Auth-specific calls live in `lib/api/auth.ts`; the client is configured once with the auth token
+ refresh handler by the auth store.

## Data layer: vue-query composables

Server state is owned by `@tanstack/vue-query`. Each feature exposes query/mutation composables
in `composables/use<F>Queries.ts`:

```ts
export function useItemsQuery() {
  return useQuery({
    queryKey: ['items'],
    queryFn: listItems,
  })
}
```

Conventions:
- **Query keys** are stable arrays, named for the resource: `['items']`, `['item-types']`,
  `['manufacturers']`. Add params as further array elements (`['item', id]`) so caching and
  invalidation are predictable.
- **Mutations** go through the shared composables `useMutationWithToast` / `useInlineMutation`
  rather than raw `useMutation`, so success/error toasts and cache invalidation stay
  consistent.
- The global `queryClient` (`lib/query/client.ts`) sets the defaults: `retry: 1`,
  `refetchOnWindowFocus: false`, `staleTime: 30s`. Don't re-specify these per-query unless a
  case genuinely differs.
- Never copy fetched data into a `ref` or a Pinia store to "hold" it — read it from the query.

## Client state: Pinia

Pinia is for state that is truly app-wide and not server-owned — currently auth and UI
preferences. Use the **setup-store** style:

```ts
export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(...)
  const isAuthenticated = computed(() => ...)
  // ...
  return { accessToken, isAuthenticated, /* actions */ }
})
```

Persisted values (e.g. auth tokens) use namespaced localStorage keys like
`overpacked-app.auth.accessToken`. Do not put component-local UI state or server data in Pinia.
