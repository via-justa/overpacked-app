# Styling and Icons

The UI is PrimeVue components + Tailwind v4, themed through `@primeuix/themes`. Two of the
conventions here are enforced by lint scripts and fail CI, so treat them as hard constraints.

## Rule 1 — no raw Tailwind palette classes

Do **not** use Tailwind's raw color-palette utilities in components. That means none of:

```
bg-emerald-500   text-stone-700   border-rose-400   ring-sky-300
from-amber-200   divide-red-500   shadow-orange-400 ...
```

The enforced palette names are: `stone, emerald, rose, red, amber, yellow, orange, sky, white`,
across the prefixes `bg, text, border, ring, shadow, decoration, divide, from, via, to` (and
`file:` / `hover:file:` variants). Instead, use the **semantic theme tokens** the design system
defines, so colors stay themeable and consistent. Raw palette classes are permitted only in
`src/style.css` (where the tokens themselves are defined). The check is
`scripts/check-no-raw-palette-classes.mjs` (`npm run lint:theme`).

When you need a color, reach for the semantic/utility class the app already uses elsewhere
rather than a literal palette shade. If you're unsure which token to use, grep a sibling
component for how it expresses the same intent.

## Rule 2 — no direct PrimeIcons

Do **not** reference PrimeIcons directly. All of these fail the check:

```
icon="pi pi-plus"        class="pi pi-trash"
:class="pi pi-pencil"    icon: 'pi pi-check'   (in a data object)
```

Instead:
- In templates, use the `<AppIcon>` component (`src/components/icons/AppIcon.vue`).
- In data/config objects (menu items, table actions), use the `iconRegistry` template literals
  from `src/lib/icons`.

This keeps icon usage centralized and swappable. The check is
`scripts/check-no-direct-primeicons.mjs` (`npm run lint:icons`); see `src/lib/icons/README.md`
for how the registry is meant to be used.

### Icon registry structure

`src/lib/icons/registry.ts` exports per-category maps plus an aggregated `iconRegistry` object
keyed by **six semantic categories**. Pick the icon by category + semantic name, not by raw
PrimeIcons class:

- **action** — create, delete, edit, cancel, confirm, submit, upload, refresh, search, menu, …
- **navigation** — dashboard, items/gear, sets, packs, lists, trips, persons, settings, login, logout
- **status** — success, error, warning, info, active/inactive
- **content** — image, tag, externalLink, file, folder, building
- **directional** — chevronUp/Down/Left/Right, arrows
- **feedback** — spinner, loading, info

In templates use `<AppIcon category="action" name="delete" />`; in data/config objects use the
registry template-literal form (e.g. `` `pi ${iconRegistry.action.confirm}` ``). If an icon is
missing, add it to the appropriate category in `registry.ts` rather than reaching for a raw
`pi pi-*` class.

## Theme-readiness — keep colors (and other tokens) token-driven

The app must stay **ready to add themes** (e.g. a dark mode or alternate palette) by redefining
tokens, not by editing components. All themeable design values live as tokens in the `@theme`
block of `src/style.css` — colors (`--color-surface-*`, `--color-copy-*`, `--color-line-*`,
`--color-brand-*`, `--color-warning-*`, `--color-danger-*`, `--color-info-*`), plus radius,
shadow, and gradient tokens. Components consume them only through the semantic utility classes
(`bg-surface-elevated`, `text-copy`, `border-line-subtle`, `text-brand-500`, `text-danger-500`, …).

To keep it that way, **don't bake values into components**:

- **No hardcoded colors anywhere in components** — no hex (`#1c1917`), `rgb()`/`hsl()`, named
  colors, or arbitrary-value classes like `bg-[#fff]` / `text-[rgb(...)]`. These ignore the
  active theme and won't change when a theme is swapped.
- **No raw Tailwind palette classes** (`bg-emerald-500`, etc.) — that's the enforced Rule 1 above;
  it's also a theme-readiness rule, since palette shades are fixed, not themeable.
- **Avoid inline `style="..."` for anything themeable.** Inline color/background/border styles
  bypass the token system. Use a token utility class instead; if a value genuinely must be inline
  (e.g. a computed dynamic dimension), bind it to a CSS variable / token (`var(--color-...)`),
  never a literal color.
- **In a `<style>` block** (reserve these for what utilities can't express), reference tokens via
  `var(--color-...)` / `var(--radius-...)` — never literal values.
- **New design values are added as tokens in `style.css`'s `@theme`** (the single source of
  truth), then consumed by class or `var()`. Adding a future theme then means redefining those
  tokens, with no component changes.

Note the `lint:theme` script catches raw palette classes, but it does **not** catch inline styles
or hardcoded hex/`rgb()` literals — those need review discipline, so flag them when you see them.

## stylelint

CSS and `<style>` blocks are linted with stylelint (`stylelint-config-standard`) via
`npm run lint:css`. Keep styles minimal — prefer Tailwind utilities and theme tokens over bespoke
CSS, and reserve `<style>` blocks for what utilities can't express.
