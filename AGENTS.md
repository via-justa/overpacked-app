# Logic Notes

This file tracks product and data-model logic decisions provided by the user.
It should be updated continuously as new decisions are made.

## Scope
- App domain: backpacking gear management (similar to lighterpack/lightweight packing tools)
- Diagram/modeling baseline is maintained in database-schema.mermaid

# AGENTS.md

This file defines repository-level guidance for coding agents working in this project.

## Project Context
- Domain: backpacking gear management (similar to lighterpack-style apps).
- Backend stack: Go + PostgreSQL + goose migrations.
- Frontend stack:
  - Framework: Vue 3 (`<script setup>`) + TypeScript
  - Build: Vite
  - Routing: Vue Router v4
  - Server state: `@tanstack/vue-query`
  - Global app state: Pinia (auth, UI preferences only)
  - Forms: VeeValidate + Zod
  - UI: PrimeVue 4 + Tailwind CSS
  - Typed API client: `openapi-typescript` + `openapi-fetch`

- Key architecture docs:
  - `dev/database-schema.mermaid`
  - `dev/openapi.yaml`
  - `dev/db-migrations.md`

## Source Of Truth
- Database schema truth: `backend/internal/migrations/sql/00001_initial_schema.sql`.
- API contract truth: `dev/openapi.yaml`.
- If docs and SQL diverge, update docs to match SQL unless the user explicitly requests a schema change.

## Core Domain Rules

### Login and Auth
- Single user application with username and password from ENV vars APP_USERNAME and APP_PASSWORD
- All backend endpoints after login use JWT tokens for auth

### Persons
- Peaple who carry packs.
- Pack recommanded weight is calculated based on the age, gender and body weight

### Packs
- A pack is used by exactly one person.
- Packs contain items through a many-to-many relationship (PACK_ITEMS).
- Packs can track which sets contributed items through PACK_SETS.
- Sets added to packs are inflated
- items added directly to packs take priority over inflated sets

### Items
- Items support multiple types; defined in ITEM_TYPES table.
- Items have type-specific detail fields.
- Item image support is required.
- Item weight and volume uses canonical storage in only grams and ml respectively.

### Sets
- Items can belong to multiple sets through SET_ITEMS.
- A set can be assigned to many packs, and a pack can include many sets through PACK_SETS.
- Assigning a set to a pack inflates set items into PACK_ITEMS; pack quantities remain independently editable.

### Item Required Fields
Only these are mandatory for items:
- name
- type
- is_active
- manufacturer

### Units
- Store canonical units in DB only:
  - Weight: grams
  - Volume: ml
- Backend remains canonical-first for storage, validation, and calculations.
- Displayed units are converted on the frontend side based on the values in the backend settings table
- For exports/reports, the converted values are used based on the current settings values

## Import Policy (Legal/Operational)
- Do not depend on providers that require account registration, affiliate enrollment, paid contracts, or partner approval.
- Prefer legal, open-licensed sources when available.
- Imported data should be assistive and user-confirmed, not blindly trusted.

## DB Migration Rules
- Additive changes preferred; destructive changes require clear rollback.
- Every migration must include both `Up` and `Down` blocks.
- Never edit historical migrations after they are applied in shared environments; add a new migration instead.

## Agent Workflow Rules
- **Spec-first API development**: Update `dev/openapi.yaml` first with new endpoint schemas, then run `make gen-api-go` to generate Go types and handler interface.
- When changing schema: update SQL migration first, then sync `dev/database-schema.mermaid`
- Validate with build checks after backend changes.
- Keep changes minimal and avoid unrelated refactors.
- Prefer named constants over literal strings in code (especially for repeated values such as error messages, routes, query keys, header names, and status labels).
- Use an "assist, then confirm" flow for any external import attempt.

## Frontend UI Learnings (Required)
- Reusable popup shell: use `frontend/src/components/AppTemplateDialog.vue` for form popups instead of inline `Dialog` markup.
- Reusable create action: use `frontend/src/components/AppCreateButton.vue` for top-nav create actions that may be reused in other sections.
- Prevent top-nav layout jump: do not conditionally remove the create button with `v-if`; keep it rendered and hide it with `invisible pointer-events-none` so width is preserved.
- Persons form UX: in popups, render form fields in a single vertical column (no responsive two-column split).
- Date display consistency: avoid locale-based date rendering for persons; keep fixed day-month-year formatting for list display and form input.
- Body-weight unit behavior: backend remains canonical grams only; settings unit is display-only (`g` => KG input/display, `oz` => LB input/display).
- Theme-token guardrail: keep raw Tailwind palette utilities out of app templates and feature CSS. Use semantic tokens from `frontend/src/style.css` instead, and run `npm run lint:theme` in `frontend/` before committing UI changes.

## OpenAPI & Code Generation
- **Generator**: oapi-codegen v2.7.0 (spec-first, not code-first)
- **Config**: `backend/.oapi-codegen.yaml`
- **Generated code**: `backend/internal/api/api.gen.go` (models, enums, ServerInterface)
- **Workflow**: Edit `dev/openapi.yaml` → run `make gen-api-go` → implement handlers against generated interface
- **No custom generators**: Removed old `cmd/openapi` code-first generator; oapi-codegen is authoritative
- **Route wiring**: Mount generated routes via `api.HandlerWithOptions(...)` in `backend/internal/app/routes.go`; avoid duplicating OpenAPI endpoint literals in manual route registration.

## Frontend Component Architecture

### Component Decoupling Pattern
- **Page-level responsibility**: State management, data fetching, mutations, domain logic (formatting, calculations)
- **Component-level responsibility**: UI rendering, user interaction, local state for presentation
- **Extraction rule**: UI components that manage layout, dialogs, or complex templates should be extracted to dedicated `.vue` files for readability and reuse, even if used in only one page
- **When NOT to extract**: Inline logic like formatters, mutation handlers, and state coordination loops stay in the page script

### Examples
- `ItemsPage.vue` (page): Manages items state, queries, mutations, filtering, and domain formatting logic
- `ItemDetailsDialog.vue` (component): Renders item details in a dialog; receives formatted data and emits events
- `ItemFormDialog.vue` (component): Wraps form in a dialog; delegates form logic to `ItemFormCard`
- `ItemsCreateOptionsMenu.vue` (component): Renders a floating menu; receives options array and emits selection events
- `ItemCard.vue` (component): Renders a single item card; accepts pre-calculated image URL and formatted values

### Reusability Pattern
- Shared display components (`ItemCard`, `ItemsListView`, `ItemsTypeTable`) are imported by multiple features (items, sets, packs)
- Feature-specific form/dialog components stay within the feature folder and are not widely reused (e.g., `ItemFormCard` is items-specific)
- Pass all formatted/calculated values as props to components; avoid components doing domain logic
