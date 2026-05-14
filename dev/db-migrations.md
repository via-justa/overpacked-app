# Database Migrations

Migrations are managed with [goose](https://github.com/pressly/goose) and embedded directly into the binary at compile time. All migration files live in `backend/internal/migrations/sql/`.

## Prerequisites

- Go 1.24+
- A running PostgreSQL instance
- `DATABASE_URL` environment variable set

```sh
export DATABASE_URL="postgres://user:password@localhost:5432/packing_light?sslmode=disable"
```

## Running Migrations

From the `backend/` directory:

```sh
# Apply all pending migrations (default)
go run ./cmd/api

# Equivalent explicit command
go run ./cmd/api up

# Roll back the most recent migration
go run ./cmd/api down

# Roll back all migrations
go run ./cmd/api reset

# Show current migration status
go run ./cmd/api status

# Migrate to a specific version
go run ./cmd/api up-to 2
go run ./cmd/api down-to 1
```

## Creating a New Migration

Migration files must follow the naming convention:

```
<version>_<description>.sql
```

Examples:
- `00002_add_person_notes.sql`
- `00003_add_pack_sharing.sql`

Each file must include goose annotations:

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

Place the file in `backend/internal/migrations/sql/` and it will be automatically embedded and picked up on the next run.

## Migration Conventions

- Always write a `Down` block — every migration must be reversible.
- Use `-- +goose StatementBegin` / `-- +goose StatementEnd` for every statement to handle multi-statement migrations correctly.
- Prefer additive changes (add columns/tables) over destructive changes (drop/rename) in `Up`.
- Destructive operations (drop column, rename) belong in the `Down` block of the migration that introduced the object.
- Never edit an already-applied migration — create a new one instead.
- All weight values must be stored as grams (`NUMERIC`), all volumes as millilitres. Unit conversion is handled at the application layer.

## Canonical Units Reference

| Field type    | DB column suffix | Unit  |
|---------------|-----------------|-------|
| Weight        | `_grams`        | g     |
| Volume        | `_ml`           | ml    |
| Temperature   | `_c`            | °C    |
| Capacity (electronics) | `_mah` | mAh |
