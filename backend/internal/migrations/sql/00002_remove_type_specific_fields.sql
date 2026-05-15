-- +goose Up
-- +goose StatementBegin

-- Remove type-specific columns from items table
-- These fields are now managed through item_type_fields and stored in attributes_json

ALTER TABLE items DROP COLUMN IF EXISTS dose_count;
ALTER TABLE items DROP COLUMN IF EXISTS calories;
ALTER TABLE items DROP COLUMN IF EXISTS calories_per_serving;
ALTER TABLE items DROP COLUMN IF EXISTS requires_water;
ALTER TABLE items DROP COLUMN IF EXISTS season;
ALTER TABLE items DROP COLUMN IF EXISTS layer;
ALTER TABLE items DROP COLUMN IF EXISTS waterproof;
ALTER TABLE items DROP COLUMN IF EXISTS size;
ALTER TABLE items DROP COLUMN IF EXISTS color;
ALTER TABLE items DROP COLUMN IF EXISTS capacity_people;
ALTER TABLE items DROP COLUMN IF EXISTS season_rating;
ALTER TABLE items DROP COLUMN IF EXISTS freestanding;
ALTER TABLE items DROP COLUMN IF EXISTS has_footprint;
ALTER TABLE items DROP COLUMN IF EXISTS comfort_temp_c;
ALTER TABLE items DROP COLUMN IF EXISTS fill_type;
ALTER TABLE items DROP COLUMN IF EXISTS r_value;
ALTER TABLE items DROP COLUMN IF EXISTS capacity_mah;
ALTER TABLE items DROP COLUMN IF EXISTS charge_port;
ALTER TABLE items DROP COLUMN IF EXISTS rechargeable;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore type-specific columns (data will be lost)

ALTER TABLE items ADD COLUMN dose_count INT;
ALTER TABLE items ADD COLUMN calories NUMERIC;
ALTER TABLE items ADD COLUMN calories_per_serving NUMERIC;
ALTER TABLE items ADD COLUMN requires_water BOOLEAN;
ALTER TABLE items ADD COLUMN season TEXT CHECK (season IN ('summer', 'winter', 'year_round'));
ALTER TABLE items ADD COLUMN layer TEXT CHECK (layer IN ('base', 'mid', 'shell', 'accessory'));
ALTER TABLE items ADD COLUMN waterproof BOOLEAN;
ALTER TABLE items ADD COLUMN size TEXT;
ALTER TABLE items ADD COLUMN color TEXT;
ALTER TABLE items ADD COLUMN capacity_people NUMERIC;
ALTER TABLE items ADD COLUMN season_rating TEXT CHECK (season_rating IN ('3-season', '4-season'));
ALTER TABLE items ADD COLUMN freestanding BOOLEAN;
ALTER TABLE items ADD COLUMN has_footprint BOOLEAN;
ALTER TABLE items ADD COLUMN comfort_temp_c NUMERIC;
ALTER TABLE items ADD COLUMN fill_type TEXT CHECK (fill_type IN ('down', 'synthetic', 'foam', 'air', 'other'));
ALTER TABLE items ADD COLUMN r_value NUMERIC;
ALTER TABLE items ADD COLUMN capacity_mah INT;
ALTER TABLE items ADD COLUMN charge_port TEXT CHECK (charge_port IN ('usb-c', 'micro-usb', 'lightning', 'dc'));
ALTER TABLE items ADD COLUMN rechargeable BOOLEAN;

-- +goose StatementEnd
