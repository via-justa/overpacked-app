-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE TABLE settings (
    id INT PRIMARY KEY CHECK (id = 1),
    weight_unit TEXT NOT NULL CHECK (weight_unit IN ('g', 'oz')),
    distance_unit TEXT NOT NULL CHECK (distance_unit IN ('km', 'mi')),
    temperature_unit TEXT NOT NULL CHECK (temperature_unit IN ('c', 'f')),
    volume_unit TEXT NOT NULL CHECK (volume_unit IN ('ml', 'fl_oz')),
    currency TEXT NOT NULL DEFAULT 'eur' CHECK (currency IN ('usd', 'eur'))
);

INSERT INTO settings (id, weight_unit, distance_unit, temperature_unit, volume_unit)
VALUES (1, 'g', 'km', 'c', 'ml');

CREATE TABLE persons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    gender TEXT CHECK (gender IN ('male', 'female', 'other')),
    birthdate DATE,
    body_weight_grams NUMERIC CHECK (body_weight_grams > 0),
    conditioning_level TEXT CHECK (conditioning_level IN ('sedentary', 'average', 'athletic', 'military')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE item_types (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    base_profile TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT item_types_base_profile_check CHECK (base_profile IS NULL OR base_profile IN ('consumable', 'wearable', 'shelter', 'sleep', 'electronics'))
);

INSERT INTO item_types (id, name, base_profile, is_system) VALUES
    ('consumable', 'Consumable', 'consumable', TRUE),
    ('wearable',   'Wearable',   'wearable',   TRUE),
    ('shelter',    'Shelter',    'shelter',    TRUE),
    ('sleep',      'Sleep',      'sleep',      TRUE),
    ('electronics','Electronics','electronics', TRUE);

CREATE TABLE item_type_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_type_id TEXT NOT NULL REFERENCES item_types(id) ON DELETE CASCADE,
    field_key TEXT NOT NULL,
    field_label TEXT NOT NULL,
    data_type TEXT NOT NULL CHECK (data_type IN ('string', 'integer', 'number', 'boolean', 'enum')),
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    enum_options_json JSONB,
    min_value NUMERIC,
    max_value NUMERIC,
    unit TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (item_type_id, field_key)
);

CREATE INDEX idx_item_type_fields_item_type_id ON item_type_fields(item_type_id);

INSERT INTO item_type_fields (item_type_id, field_key, field_label, data_type, is_required, enum_options_json, sort_order)
VALUES
    ('consumable', 'dose_count',          'Dose count',          'integer', FALSE, NULL,                                              10),
    ('consumable', 'calories',            'Calories',            'number',  FALSE, NULL,                                              20),
    ('consumable', 'calories_per_serving','Calories per serving', 'number', FALSE, NULL,                                              30),
    ('consumable', 'requires_water',      'Requires water',      'boolean', FALSE, NULL,                                              40),

    ('wearable',   'season',              'Season',              'enum',    FALSE, '["summer", "winter", "year_round"]'::jsonb,        10),
    ('wearable',   'layer',               'Layer',               'enum',    FALSE, '["base", "mid", "shell", "accessory"]'::jsonb,     20),
    ('wearable',   'waterproof',          'Waterproof',          'boolean', FALSE, NULL,                                              30),
    ('wearable',   'size',                'Size',                'string',  FALSE, NULL,                                              40),
    ('wearable',   'color',               'Color',               'string',  FALSE, NULL,                                              50),

    ('shelter',    'capacity_people',     'Capacity people',     'number',  FALSE, NULL,                                              10),
    ('shelter',    'season_rating',       'Season rating',       'enum',    FALSE, '["3-season", "4-season"]'::jsonb,                 20),
    ('shelter',    'freestanding',        'Freestanding',        'boolean', FALSE, NULL,                                              30),
    ('shelter',    'has_footprint',       'Has footprint',       'boolean', FALSE, NULL,                                              40),

    ('sleep',      'comfort_temp_c',      'Comfort temp C',      'number',  FALSE, NULL,                                              10),
    ('sleep',      'fill_type',           'Fill type',           'enum',    FALSE, '["down", "synthetic", "foam", "air", "other"]'::jsonb, 20),
    ('sleep',      'r_value',             'R value',             'number',  FALSE, NULL,                                              30),

    ('electronics','capacity_mah',        'Capacity mAh',        'integer', FALSE, NULL,                                              10),
    ('electronics','charge_port',         'Charge port',         'enum',    FALSE, '["usb-c", "micro-usb", "lightning", "dc"]'::jsonb, 20),
    ('electronics','rechargeable',        'Rechargeable',        'boolean', FALSE, NULL,                                              30);

CREATE TABLE manufacturers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    website TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manufacturer_id UUID NOT NULL REFERENCES manufacturers(id) ON DELETE RESTRICT,
    type_id TEXT NOT NULL REFERENCES item_types(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL,
    description TEXT,
    source_url TEXT,
    price NUMERIC,
    weight_grams NUMERIC CHECK (weight_grams > 0),
    volume_ml NUMERIC CHECK (volume_ml > 0),
    default_quantity INT NOT NULL DEFAULT 1 CHECK (default_quantity > 0),
    default_carry_status TEXT NOT NULL DEFAULT 'packed' CHECK (default_carry_status IN ('packed', 'worn')),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    image_blob BYTEA,
    image_mime_type TEXT,
    image_size_bytes INT,
    image_width_px INT,
    image_height_px INT,
    attributes_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_items_image_metadata CHECK (
        image_blob IS NULL OR (
            image_mime_type IS NOT NULL AND
            image_size_bytes IS NOT NULL AND
            image_width_px IS NOT NULL AND
            image_height_px IS NOT NULL
        )
    )
);

CREATE TABLE item_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    set_category TEXT NOT NULL REFERENCES item_types(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE set_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    set_id UUID NOT NULL REFERENCES item_sets(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    notes TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    UNIQUE (set_id, item_id)
);

CREATE TABLE packs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id UUID REFERENCES persons(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    trip_type TEXT CHECK (trip_type IN ('day_hike', 'overnight', 'multi_day', 'thru_hike')),
    notes TEXT,
    is_template BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE pack_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pack_id UUID NOT NULL REFERENCES packs(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    carry_status TEXT NOT NULL DEFAULT 'packed' CHECK (carry_status IN ('packed', 'worn')),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (pack_id, item_id)
);

-- Create trips table
CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    trip_type TEXT CHECK (trip_type IN ('day_hike', 'overnight', 'multi_day', 'thru_hike')) NOT NULL,
    duration INTERVAL,
    notes TEXT,
    trip_komoot_url TEXT,
    trip_strava_url TEXT,
    trip_wanderer_url TEXT,
    total_distance_km NUMERIC,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create trip_persons junction table (trips can have multiple persons, persons can go on multiple trips)
CREATE TABLE trip_persons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(trip_id, person_id)
);

-- Create trip_person_packs junction table (trip persons can have multiple packs)
CREATE TABLE trip_person_packs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_person_id UUID NOT NULL REFERENCES trip_persons(id) ON DELETE CASCADE,
    pack_id UUID NOT NULL REFERENCES packs(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(trip_person_id, pack_id)
);

-- Create trip_person_items junction table (trip persons can have items - worn or in packs)
CREATE TABLE trip_person_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_person_id UUID NOT NULL REFERENCES trip_persons(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    carry_status TEXT CHECK (carry_status IN ('packed', 'worn')) DEFAULT 'packed',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(trip_person_id, item_id)
);

-- Create labels table for item categorization
CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    color TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create item_labels junction table for many-to-many relationship
CREATE TABLE item_labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(item_id, label_id)
);

-- Create global packing_lists table
CREATE TABLE packing_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create packing_list_labels junction table
CREATE TABLE packing_list_labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    packing_list_id UUID NOT NULL REFERENCES packing_lists(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(packing_list_id, label_id)
);

-- Create indexes for better query performance
CREATE INDEX idx_items_type_active ON items (type_id, is_active);
CREATE INDEX idx_set_items_set_id ON set_items (set_id);
CREATE INDEX idx_pack_items_pack_id ON pack_items (pack_id);
CREATE INDEX idx_item_sets_set_category ON item_sets(set_category);
CREATE INDEX idx_trips_trip_type ON trips(trip_type);
CREATE INDEX idx_trip_persons_trip_id ON trip_persons(trip_id);
CREATE INDEX idx_trip_persons_person_id ON trip_persons(person_id);
CREATE INDEX idx_trip_person_packs_trip_person_id ON trip_person_packs(trip_person_id);
CREATE INDEX idx_trip_person_packs_pack_id ON trip_person_packs(pack_id);
CREATE INDEX idx_trip_person_items_trip_person_id ON trip_person_items(trip_person_id);
CREATE INDEX idx_trip_person_items_item_id ON trip_person_items(item_id);
CREATE INDEX idx_item_labels_item_id ON item_labels(item_id);
CREATE INDEX idx_item_labels_label_id ON item_labels(label_id);
CREATE INDEX idx_labels_name ON labels(name);
CREATE INDEX idx_packing_list_labels_packing_list_id ON packing_list_labels(packing_list_id);
CREATE INDEX idx_packing_list_labels_label_id ON packing_list_labels(label_id);
CREATE INDEX idx_packing_lists_name ON packing_lists(name);

-- Trigram indexes for fuzzy global search
CREATE INDEX idx_items_name_trgm ON items USING gin (name gin_trgm_ops);
CREATE INDEX idx_items_description_trgm ON items USING gin (description gin_trgm_ops);
CREATE INDEX idx_item_sets_name_trgm ON item_sets USING gin (name gin_trgm_ops);
CREATE INDEX idx_item_sets_description_trgm ON item_sets USING gin (description gin_trgm_ops);
CREATE INDEX idx_packing_lists_name_trgm ON packing_lists USING gin (name gin_trgm_ops);
CREATE INDEX idx_packing_lists_description_trgm ON packing_lists USING gin (description gin_trgm_ops);
CREATE INDEX idx_persons_name_trgm ON persons USING gin (name gin_trgm_ops);
CREATE INDEX idx_manufacturers_name_trgm ON manufacturers USING gin (name gin_trgm_ops);
CREATE INDEX idx_trips_name_trgm ON trips USING gin (name gin_trgm_ops);
CREATE INDEX idx_trips_notes_trgm ON trips USING gin (notes gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_trips_notes_trgm;
DROP INDEX IF EXISTS idx_trips_name_trgm;
DROP INDEX IF EXISTS idx_manufacturers_name_trgm;
DROP INDEX IF EXISTS idx_persons_name_trgm;
DROP INDEX IF EXISTS idx_packing_lists_description_trgm;
DROP INDEX IF EXISTS idx_packing_lists_name_trgm;
DROP INDEX IF EXISTS idx_item_sets_description_trgm;
DROP INDEX IF EXISTS idx_item_sets_name_trgm;
DROP INDEX IF EXISTS idx_items_description_trgm;
DROP INDEX IF EXISTS idx_items_name_trgm;

DROP INDEX IF EXISTS idx_packing_lists_name;
DROP INDEX IF EXISTS idx_packing_list_labels_label_id;
DROP INDEX IF EXISTS idx_packing_list_labels_packing_list_id;
DROP INDEX IF EXISTS idx_labels_name;
DROP INDEX IF EXISTS idx_item_labels_label_id;
DROP INDEX IF EXISTS idx_item_labels_item_id;
DROP INDEX IF EXISTS idx_trip_person_items_item_id;
DROP INDEX IF EXISTS idx_trip_person_items_trip_person_id;
DROP INDEX IF EXISTS idx_trip_person_packs_pack_id;
DROP INDEX IF EXISTS idx_trip_person_packs_trip_person_id;
DROP INDEX IF EXISTS idx_trip_persons_person_id;
DROP INDEX IF EXISTS idx_trip_persons_trip_id;

DROP INDEX IF EXISTS idx_trips_trip_type;
DROP INDEX IF EXISTS idx_item_sets_set_category;
DROP INDEX IF EXISTS idx_pack_items_pack_id;
DROP INDEX IF EXISTS idx_set_items_set_id;
DROP INDEX IF EXISTS idx_items_type_active;
DROP INDEX IF EXISTS idx_item_type_fields_item_type_id;

DROP TABLE IF EXISTS packing_list_labels;
DROP TABLE IF EXISTS packing_lists;
DROP TABLE IF EXISTS item_labels;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS trip_person_items;
DROP TABLE IF EXISTS trip_person_packs;
DROP TABLE IF EXISTS trip_persons;
DROP TABLE IF EXISTS trips;
DROP TABLE IF EXISTS pack_items;
DROP TABLE IF EXISTS packs;
DROP TABLE IF EXISTS set_items;
DROP TABLE IF EXISTS item_sets;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS item_type_fields;
DROP TABLE IF EXISTS manufacturers;
DROP TABLE IF EXISTS item_types;
DROP TABLE IF EXISTS persons;
DROP TABLE IF EXISTS settings;

DROP EXTENSION IF EXISTS "pg_trgm";
-- +goose StatementEnd
