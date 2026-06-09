# Components and Composables

## Components

- Always `<script setup lang="ts">`. Order the block: imports, props, emits, composables/state,
  computeds, then functions.
- **Type props and emits explicitly.** Use the type-based forms so contracts are visible:
  ```ts
  const props = defineProps<{ itemId: string; editable?: boolean }>()
  const emit = defineEmits<{ save: [value: ItemUpdate]; cancel: [] }>()
  ```
  Avoid implicit/string-array event contracts.
- Keep components focused on rendering. When a component starts orchestrating data fetching,
  mutations, or multi-step flows, extract that into a composable and let the component consume
  it. The feature components here (cards, tables, dialogs, form cards) stay presentational and
  lean on `composables/` for behavior.
- Use slots for composition and flexibility; prefer named slots with clear contracts over
  prop-drilling render decisions.
- Group shared components by kind under `src/components/` (`dialogs/`, `forms/`, `display/`,
  etc.); keep feature-specific components inside the feature.

## Page vs. component responsibilities (decoupling)

The repo draws a deliberate line between routed pages and the components they render:

- **Page-level** (the `views/*Page.vue`) owns state management, data fetching, mutations, and
  domain logic — formatting, calculations, filtering, coordinating flows.
- **Component-level** owns UI rendering, user interaction, and local presentation state only.
- **Pass formatted/calculated values down as props.** Components should not do domain logic;
  e.g. `ItemCard` receives a pre-computed image URL and already-formatted values, it doesn't
  fetch or calculate them. The page does that and hands results in.
- **Extract a component** when layout, a dialog, or a complex template grows — even if used in
  only one page — for readability and reuse. **Don't extract** inline formatters, mutation
  handlers, or state-coordination loops; those stay in the page script.
- **Reusability:** shared display components (`ItemCard`, `ItemsListView`, `ItemsTypeTable`) are
  imported across features (items, sets, packs), while feature-specific form/dialog components
  (e.g. `ItemFormCard`) stay within their feature.

## Composables

- Name them `use*` and give each one responsibility. Examples already in the repo:
  `useDeleteConfirmation` (confirmation dialog state), `useMutationWithToast` (mutations with
  toast feedback + invalidation), `useRowActionsMenu`, `useSettings`.
- Return a plain object of refs/computeds/functions; document the composable with a short
  JSDoc block describing what it manages (the repo does this consistently).
- Type internal state precisely — discriminated unions are used for multi-mode state, e.g.:
  ```ts
  export type DeleteConfirmationState =
    | { kind: 'single'; id: string; name: string }
    | { kind: 'bulk'; ids: string[]; count: number }
    | null
  ```
- Composables that fetch data wrap vue-query (see `architecture-and-state.md`); composables for
  pure UI state just manage refs. Don't mix server fetching into a UI-state composable.

## TypeScript

- Build on the generated OpenAPI types and the feature `types.ts`; don't redeclare API shapes.
- Prefer precise types and discriminated unions over `any` or loose `object`. The build runs
  `vue-tsc`, so type errors are build failures, not warnings.
