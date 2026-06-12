package domain

import "time"

// BackupRunStatus records the outcome of the most recent scheduled/run-now backup.
type BackupRunStatus string

const (
	BackupRunStatusSuccess BackupRunStatus = "success"
	BackupRunStatusError   BackupRunStatus = "error"
)

// BackupConfig is the singleton configuration for scheduled server-side backups.
type BackupConfig struct {
	ID             int
	Enabled        bool
	Schedule       string // cron expression, e.g. "0 2 * * *"
	RetentionCount int    // number of backup files to keep
	LastRunAt      *time.Time
	LastStatus     *BackupRunStatus
	LastError      *string
	UpdatedAt      time.Time
}
