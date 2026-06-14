# Performance and Accessibility

## Rendering performance

- Use `computed` for derived values so results are cached and only recompute when their
  dependencies change. Don't recompute the same derivation inline in the template.
- Use `watch`/`watchEffect` intentionally and narrowly. Avoid `deep: true` and broad watchers
  unless you genuinely need them — they're a common source of reactive overwork. Prefer
  watching a specific getter over watching a whole object.
- Give `v-for` stable, meaningful `:key`s (entity ids, not array indices) so the diff is
  cheap and component state isn't reused incorrectly.
- For list-heavy and dashboard views, be deliberate about what re-renders: keep heavy
  subtrees in their own components so reactivity is scoped, and consider virtualization for
  very large lists.
- Don't make data reactive that doesn't need to be; vue-query already holds server data
  reactively, so don't wrap it again.

## Code splitting

- Lazy-load routed feature views in the router with dynamic imports so each feature is its own
  chunk:
  ```ts
  { path: '/items', component: () => import('@/features/items/views/ItemsPage.vue') }
  ```
- Keep feature code importable as a unit; avoid cross-feature imports that pull everything into
  one chunk.

## Accessibility

- Prefer semantic HTML elements; reach for ARIA only to fill gaps, not to replace semantics.
- Ensure interactive controls are keyboard operable (focusable, activatable with Enter/Space)
  and have visible focus states.
- Don't convey state by color alone — pair it with text, icon, or ARIA state (this dovetails
  with the no-raw-palette rule: state should be semantic, not a hardcoded shade).
- When building on PrimeVue components, preserve their built-in a11y (labels, roles); when
  composing custom controls, label them and manage focus explicitly.
- Render loading/empty/error states as real, announced content rather than silent gaps.

### Concrete a11y conventions used here

These are the established patterns in the app — apply them when you add the corresponding UI:

- **Active nav links** get `aria-current="page"` so screen readers convey location
  (see `AppTopNav.vue`).
- **Dynamic announcements** use `aria-live="polite"` — Toast notifications and loading/status
  regions (see `App.vue`) so updates are announced.
- **Form fields** link their validation message via `aria-describedby` and set `aria-invalid`
  when in error.
- **Data tables** include a `<caption class="sr-only">` for screen-reader context without visual
  display (see `ItemsTypeTable.vue`).
- **Loading indicators** carry `role="status"` so they're announced.
- The page `<title>` in `index.html` should be descriptive (it's "Overpacked", not a placeholder
  like "frontend").

### Keyboard accessibility (build it in)

Every feature must be fully operable with the keyboard alone — this is a default expectation for
new UI, not a later polish step. The rules that keep it consistent:

- **Use real interactive elements.** A thing you click must be a `<button>`, `<a>`, or form
  control — never a click-only `<div>`/`<span>`. Native elements are focusable and
  Enter/Space-activatable for free, which is why the app already leans on `<button>` and PrimeVue
  components. (Full-screen backdrop overlays used only to dismiss a popup are the one acceptable
  exception — they aren't keyboard targets because Escape handles dismissal.)
- **If you must make a non-native element interactive,** give it `role`, `tabindex="0"`, and
  keyboard handlers mirroring the native behavior (Enter/Space to activate). Prefer not to.
- **Preserve PrimeVue's built-in keyboard support** — don't override its focus, Enter/Space,
  Escape, or arrow-key behavior. Reach for PrimeVue components before hand-rolling interactive UI.
- **Custom menus / listboxes / popovers** follow the established pattern in
  `AppActionsMenu.vue`: ArrowUp/ArrowDown to move, Home/End to jump, Enter/Space to activate,
  Escape to close, with `role="menu"`/`role="menuitem"` and programmatic focus on the items.
- **Manage focus on overlays.** When an overlay opens, move focus into it (first item or a
  sensible target); while open, keep focus within it; on close, **restore focus to the element
  that opened it.** PrimeVue `Dialog` does this for you (so `AppTemplateDialog` inherits it) — for
  custom Teleport'd menus you must do it yourself.
- **Row/list action menus** built via `useRowActionsMenu` should carry the same keyboard
  navigation as the menu above; if you extend that composable, add the key handling there so every
  consumer gets it rather than re-implementing per component.
- **Visible focus.** Never remove focus outlines without replacing them with an equally clear
  visible focus state — keyboard users navigate by it.
- **Consider a skip-to-content link** in the app shell so keyboard users can bypass the nav.

When reviewing, tab through the change mentally (or in the browser): can you reach every control,
activate it, escape every overlay, and does focus land somewhere sensible afterward? If not, it's
not done.
