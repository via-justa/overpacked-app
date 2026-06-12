-- +goose Up
-- +goose StatementBegin
CREATE TABLE backup_config (
    id INT PRIMARY KEY CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    schedule TEXT NOT NULL DEFAULT '0 2 * * *',
    retention_count INT NOT NULL DEFAULT 7 CHECK (retention_count > 0),
    last_run_at TIMESTAMPTZ,
    last_status TEXT CHECK (last_status IN ('success', 'error')),
    last_error TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO backup_config (id) VALUES (1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS backup_config;
-- +goose StatementEnd
