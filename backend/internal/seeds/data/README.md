# Seed Data

This directory contains YAML seed files that populate the database with initial data.

## File Naming Convention

Files are processed in alphabetical order, so use numeric prefixes:

- `01_labels.yaml`
- `02_manufacturers.yaml`
- `03_item_types.yaml`

## YAML Format

Each seed file must follow this structure:

```yaml
table: <table_name>
records:
  - field1: value1
    field2: value2
  - field1: value3
    field2: value4
```

## Supported Tables

### labels
- `name` (required, string) - Unique label name
- `color` (optional, string) - Hex color code (e.g., "#FF5733")

### manufacturers
- `name` (required, string) - Unique manufacturer name
- `website` (optional, string) - Manufacturer website URL

More tables will be supported as needed.

## Running Seeds

From the backend directory:

```bash
go run ./cmd/api seed
```

Or using make:

```bash
make seed
```

## Idempotency

Seeds are idempotent and safe to run multiple times. The system checks for existing records by unique key constraints before inserting.

## Validation

All seed files are validated before insertion:
- Required fields must be present
- Field types must match table schema
- Unique constraints are checked
- No duplicate records within the same file
