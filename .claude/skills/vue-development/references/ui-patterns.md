# UI Patterns & Gotchas

Hard-won, repo-specific UI conventions. These reference real shared building blocks — prefer
them over re-implementing, and grep the cited file for current usage before you start.

## Dialogs / form popups

Use the shared popup shell `src/components/dialogs/AppTemplateDialog.vue` for form popups instead
of writing inline PrimeVue `Dialog` markup. Feature dialogs (`ItemFormDialog`, `PersonFormDialog`,
`SetFormDialog`, `PackingListFormDialog`, etc.) wrap it, so a new form dialog should too — you get
consistent sizing, header/footer, and behavior for free.

## Avoiding layout shift

When an element should appear/disappear without the surrounding layout jumping, **reserve its
space** rather than removing it from the DOM. Hide it with `invisible pointer-events-none`
(keeps the box, drops visibility + interaction) instead of `v-if` (which collapses the space).
This is the established pattern for things like inline validation-error slots (see the
`'... ? 'text-danger-500' : 'invisible'` placeholders in `ItemFormCard.vue`).

## Fixed-positioned UI must escape overflow containers

A subtle but recurring bug: an ancestor with `overflow-x/y: auto` (or similar) creates a new
containing block, so a `position: fixed` descendant positions relative to that container, not the
viewport — breaking dropdowns, tooltips, popovers, and floating menus inside scroll areas.

Fix: render fixed UI with `<Teleport to="body">` so it lives outside the overflow container and
positions relative to the viewport while keeping its reactive state and handlers. This is how the
repo does floating menus (see `ItemsTypeTable.vue`, `AppActionsMenu.vue`).

## Row/list action menus

For table or list row-action menus (the "⋯" menu with per-row actions), use the
`src/composables/useRowActionsMenu.ts` composable rather than re-implementing menu open/close
state, viewport-aware positioning, and document-click/cleanup handling in each component.

## localStorage access

Use the SSR-safe helpers in `src/lib/storage/localStorage.ts` (`getStoredValue`,
`setStoredValue`) rather than touching `localStorage` directly — they guard with
`typeof window === 'undefined'` and apply type checks. Match the file's style: prefer
`typeof window === 'undefined'` (not `typeof globalThis.window === 'undefined'`) and regular
exported `function` declarations over arrow-function exports (Sonar preferences the repo follows).

## Units display (canonical storage, converted display)

The backend stores canonical units only (weight in **grams**, volume in **ml**). The frontend
converts for display based on the settings unit, using `src/lib/units/`:

- Weight unit is `g` or `oz` (`WeightUnit`); convert with `gramsToInput` / `inputToGrams`. Body
  weight is stored as `body_weight_grams` and shown in KG (metric) or LB (imperial) accordingly.
- Volume unit is `ml` or `fl_oz` (`VolumeUnit`).

Never store or compare a non-canonical value — convert at the edge (input → grams on the way in,
grams → display on the way out). Person recommended-pack-weight calculations also work from
canonical grams (see `features/persons/utils.ts`).

## Responsiveness

The app is built **mobile-first** and is expected to work from phone width up. Build new UI the
same way rather than treating mobile as an afterthought.

**Layout — prefer CSS via Tailwind responsive utilities.** Start from the unprefixed (mobile)
base and layer larger breakpoints on top, using Tailwind's default screens (`sm` 640, `md` 768,
`lg` 1024, `xl` 1280). This is how the codebase already does layout: grids that collapse on small
screens (`grid-cols-1 … sm:grid-cols-2 … lg:grid-cols-3`), spacing that grows (`p-4 sm:p-5`), and
showing/hiding or reflowing per breakpoint (`md:hidden` / `md:block` / `md:table-row`,
`md:flex-col`). Reach for these before any JavaScript — they cost nothing at runtime and keep
behavior declarative.

**Behavioral switches — JS viewport detection at the `md` (768px) threshold.** When the change
is behavioral, not just styling (e.g. forcing a different view mode, repositioning a floating
menu), the repo detects the viewport in script. `ItemsPage.vue` keeps an `isMobileViewport` ref
(`window.innerWidth < 768`), updates it on a `resize` listener, and derives
`effectiveViewMode` (forcing card view on mobile); `AppTopNav.vue` uses the same `< 768` check for
menu placement. When you do this:

- Keep the threshold at **768px** so it matches Tailwind's `md` breakpoint — don't introduce a
  different magic number.
- Register the `resize` listener in `onMounted` and **remove it in `onUnmounted`** (the existing
  code does; a leaked listener is a real bug).
- Use it only for genuine behavior changes; if a Tailwind utility can express it, prefer that.

**Known inconsistency / improvement:** the `< 768` viewport check and its resize wiring are
duplicated across components. Prefer extracting a small shared composable (e.g.
`useIsMobile()` returning a reactive boolean, owning one listener with cleanup) and reuse it,
rather than re-implementing the detection per component.
