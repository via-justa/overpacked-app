package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/via-justa/overpacked-app/backend/internal/domain"
)

const singletonBackupConfigID = 1

type BackupConfigStore struct {
	db *sql.DB
}

func NewBackupConfigStore(db *sql.DB) *BackupConfigStore {
	return &BackupConfigStore{db: db}
}

func (s *BackupConfigStore) Get(ctx context.Context) (*domain.BackupConfig, error) {
	query := `
		SELECT id, enabled, schedule, retention_count,
			last_run_at, last_status, last_error, updated_at
		FROM backup_config
		WHERE id = $1`

	var cfg domain.BackupConfig
	var lastRunAt sql.NullTime
	var lastStatus sql.NullString
	var lastError sql.NullString

	err := s.db.QueryRowContext(ctx, query, singletonBackupConfigID).Scan(
		&cfg.ID,
		&cfg.Enabled,
		&cfg.Schedule,
		&cfg.RetentionCount,
		&lastRunAt,
		&lastStatus,
		&lastError,
		&cfg.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get backup config: %w", err)
	}

	cfg.LastRunAt = timePtr(lastRunAt)
	if lastStatus.Valid {
		status := domain.BackupRunStatus(lastStatus.String)
		cfg.LastStatus = &status
	}
	cfg.LastError = strPtr(lastError)

	return &cfg, nil
}

// Update persists the user-editable fields (enabled, schedule, destination, retention).
func (s *BackupConfigStore) Update(ctx context.Context, cfg *domain.BackupConfig) error {
	query := `
		UPDATE backup_config
		SET enabled = $2,
			schedule = $3,
			retention_count = $4,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		singletonBackupConfigID,
		cfg.Enabled,
		cfg.Schedule,
		cfg.RetentionCount,
	).Scan(&cfg.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update backup config: %w", err)
	}

	cfg.ID = singletonBackupConfigID
	return nil
}

// UpdateRunStatus records the outcome of the latest backup run. A nil runErr marks success.
func (s *BackupConfigStore) UpdateRunStatus(ctx context.Context, ranAt time.Time, runErr error) error {
	status := domain.BackupRunStatusSuccess
	var errMsg sql.NullString
	if runErr != nil {
		status = domain.BackupRunStatusError
		errMsg = sql.NullString{String: runErr.Error(), Valid: true}
	}

	query := `
		UPDATE backup_config
		SET last_run_at = $2,
			last_status = $3,
			last_error = $4
		WHERE id = $1`

	res, err := s.db.ExecContext(ctx, query, singletonBackupConfigID, ranAt, string(status), errMsg)
	if err != nil {
		return fmt.Errorf("update backup run status: %w", err)
	}

	return rowsAffectedOrNotFound(res, "rows affected on update backup run status")
}
