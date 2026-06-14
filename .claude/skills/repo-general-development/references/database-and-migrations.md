# Database & Migrations

Migrations are managed with **goose** and embedded into the binary at compile time. All
migration files live in `backend/internal/migrations/sql/`. They run automatically on the next
binary start and can be driven explicitly via `cmd/api`.

## Running migrations

Set `DATABASE_URL`, then from `backend/`:

```sh
export DATABASE_URL="postgres://user:password@localhost:5432/overpacked?sslmode=disable"

go run ./cmd/api            # apply all pending (default)
go run ./cmd/api up         # same, explicit
go run ./cmd/api down       # roll back the most recent
go run ./cmd/api reset      # roll back all
go run ./cmd/api status     # show current status
go run ./cmd/api up-to 2    # migrate to a specific version
go run ./cmd/api down-to 1
```

## Creating a migration

File naming: `<version>_<description>.sql`, e.g. `00002_add_person_notes.sql`,
`00003_add_pack_sharing.sql`. Place it in `backend/internal/migrations/sql/`; it's embedded and
picked up automatically.

Every file needs goose annotations, with a `StatementBegin`/`StatementEnd` wrapper around each
statement:

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE persons ADD COLUMN notes TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE persons DROP COLUMN notes;
-- +goose StatementEnd
```

## Conventions (enforced expectations)

- **Always write a `Down` block** — every migration must be reversible.
- Wrap every statement in `-- +goose StatementBegin` / `-- +goose StatementEnd` so
  multi-statement migrations run correctly.
- **Prefer additive changes** (add columns/tables) in `Up`; put destructive operations
  (drop/rename) in the `Down` of the migration that introduced the object.
- **Never edit an already-applied migration** — create a new one instead.
- Sync `dev/database-schema.mermaid` after a schema change (and `dev/openapi.yaml` if the change
  is API-impacting).

## Canonical unit columns

Storage is always canonical; conversion happens at the application layer. Follow the column
suffix convention:

| Field type             | DB column suffix | Unit |
|------------------------|------------------|------|
| Weight                 | `_grams`         | g    |
| Volume                 | `_ml`            | ml   |
| Temperature            | `_c`             | °C   |
| Capacity (electronics) | `_mah`           | mAh  |

Weights are stored as `NUMERIC` grams, volumes as millilitres. Never introduce a column that
stores a non-canonical unit.
