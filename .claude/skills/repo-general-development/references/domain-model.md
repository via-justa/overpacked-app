# Domain Model

overpacked-app manages backpacking gear: a catalog of **items** (optionally grouped into
**sets**) that get assembled into **packs** for **persons** on **trips**, with weight and volume
tracked throughout. Understanding these entities and their relationships is essential before
changing schema, endpoints, or UI.

> **Source of truth is the code/SQL, not prose.** This model is derived from the migrations,
> `backend/internal/domain/`, and the handlers. Where any doc disagrees with the code, trust the
> schema and code. Known doc/code drifts and quirks are tracked in `dev/repo-violations.md`
> (Known drifts) — consult it when a behavior here looks looser or different in practice.

## Auth (single-user)

- The app is single-user. Credentials come from env vars `APP_USERNAME` and `APP_PASSWORD`.
- After login, all backend endpoints require a JWT.

## Items

- Items support multiple **types**, defined in the `ITEM_TYPES` table; each type can have
  type-specific detail fields.
- **Required fields (only these are mandatory):** `name`, `type`, `is_active`, `manufacturer`.
  Everything else is optional.
- Item image support is required (image blob + mime/size/dimensions metadata).
- **Weight and volume are stored canonically** — grams and millilitres only. Never store or
  compare other units server-side.

## Manufacturers, Labels, Item Types

- Items reference a **manufacturer** and a **type** (`ITEM_TYPES`).
- **Labels** are user tags that can be attached to items.

## Sets

- Items can belong to multiple **sets** via `SET_ITEMS` (many-to-many).
- A set can be assigned to many packs, and a pack can include many sets via `PACK_SETS`.
- **Assigning a set to a pack inflates the set's items into `PACK_ITEMS`.** After inflation, pack
  quantities are edited independently of the set.
- Sets (and packing lists) are **builder helpers** for assembling packs — not stored with trips
  at runtime.
- The set→pack association is not persisted; there is no `pack_sets` table (some store code
  references one — a known quirk tracked in `dev/repo-violations.md` / `dev/backend-violations.md`).

## Persons

- Persons are the people who carry packs.
- A pack's **recommended weight** is calculated from the person's age, gender, and body weight.

## Packs

- **Through the API, packs are managed only via the trip → person → pack hierarchy** — there are
  no standalone pack routes. The only creation path is
  `POST /api/v1/trips/{tripId}/persons/{personId}/packs` (handled by `AddTripPersonPack`); a
  comment in `packs_helpers.go` states packs are no longer managed as standalone resources. (The
  DB schema is looser than this rule — a known quirk tracked in `dev/repo-violations.md`; enforce
  the hierarchy at the API layer, not as a DB invariant.)
- Packs contain items through `PACK_ITEMS` (many-to-many). Each pack item has a `carry_status`
  of `packed` or `worn`.

## Trips

Trips organize multi-person journeys with person-specific gear:

- A trip contains persons via the `TRIP_PERSONS` junction.
- Each person in a trip can have multiple packs via `TRIP_PERSON_PACKS`.
- Each person in a trip can also have items directly via `TRIP_PERSON_ITEMS` (worn or packed).
- **Trip sets were removed** — sets/packing lists are builder helpers only, not persisted with
  trips at runtime.
- Trip metadata: `name`, `type` (`day_hike`, `overnight`, `multi_day`, `thru_hike`), duration,
  distance, route URLs.

## Settings & units

- A backend `settings` table holds display preferences.
- Storage is always canonical (grams, ml). The frontend converts for display based on settings
  (e.g. body-weight unit: `g` → KG input/display, `oz` → LB input/display).
- Exports/reports use the converted values based on current settings.

## Import policy (legal/operational)

- Imported gear data is **assistive and user-confirmed**, never blindly trusted — use an
  "assist, then confirm" flow.
- Do not build on providers that require account registration, affiliate enrollment, paid
  contracts, or partner approval. Prefer legal, open-licensed sources.

## Table map

The schema is **18 tables** (one migration so far: `00001_initial_schema.sql`). This map is for
orientation — it changes slowly. **For exact columns, types, defaults, and `ON DELETE` behavior,
read `backend/internal/migrations/sql/` (the source of truth) and `dev/database-schema.mermaid`
(ER diagram); don't assume column-level detail from here.** Arrows show the foreign key
(`child → parent`); tables with two FKs are junction/many-to-many tables.

Catalog:
- `manufacturers` — gear makers.
- `item_types` — item categories (PK is a `TEXT` id, not a UUID).
- `item_type_fields → item_types` — type-specific field definitions.
- `items → manufacturers, item_types` — the gear catalog.
- `labels` — user tags. `item_labels → items, labels` — M:N item↔label.

Sets (builder helper):
- `item_sets → item_types` (`set_category`) — named sets.
- `set_items → item_sets, items` — M:N set↔item.

People, trips, and the pack hierarchy:
- `persons` — people who carry packs.
- `trips` — journeys.
- `trip_persons → trips, persons` — M:N who's on a trip.
- `trip_person_packs → trip_persons, packs` — a participant's packs.
- `trip_person_items → trip_persons, items` — items carried directly (worn/packed).
- `packs → persons` (**nullable** — see the Packs drift note) — a pack.
- `pack_items → packs, items` — M:N pack↔item, carries `carry_status` (`packed`/`worn`).

Packing lists (builder helper):
- `packing_lists` — list definitions. `packing_list_labels → packing_lists, labels`.

Settings:
- `settings` — single-row display preferences (units, etc.).

Note there is no `pack_sets` table despite store code referencing one (a known quirk — see
`dev/repo-violations.md`).

## Where this lives in code

- Schema truth: `backend/internal/migrations/sql/00001_initial_schema.sql`.
- ER diagram: `dev/database-schema.mermaid`.
- Backend entities: `backend/internal/domain/` (one file per entity).
- API shapes: `dev/openapi.yaml` (generated into `backend/internal/api`).
