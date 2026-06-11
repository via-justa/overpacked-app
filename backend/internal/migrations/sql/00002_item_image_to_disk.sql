-- +goose Up
-- +goose StatementBegin
ALTER TABLE items DROP CONSTRAINT chk_items_image_metadata;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE items DROP COLUMN image_blob;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE items ADD COLUMN image_path TEXT;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE items ADD CONSTRAINT chk_items_image_metadata CHECK (
    image_path IS NULL OR (
        image_mime_type IS NOT NULL AND
        image_size_bytes IS NOT NULL AND
        image_width_px IS NOT NULL AND
        image_height_px IS NOT NULL
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE items DROP CONSTRAINT chk_items_image_metadata;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE items DROP COLUMN image_path;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE items ADD COLUMN image_blob BYTEA;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE items ADD CONSTRAINT chk_items_image_metadata CHECK (
    image_blob IS NULL OR (
        image_mime_type IS NOT NULL AND
        image_size_bytes IS NOT NULL AND
        image_width_px IS NOT NULL AND
        image_height_px IS NOT NULL
    )
);
-- +goose StatementEnd
