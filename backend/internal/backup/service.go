package backup

import (
	"database/sql"
	"errors"
)

// Package-level sentinels surfaced to handlers for status mapping.
var (
	// ErrBackupDirNotConfigured means BACKUP_BASE_DIR is unset, so server-side
	// backups (scheduled and run-now) are disabled.
	ErrBackupDirNotConfigured = errors.New("backup base directory not configured")
	// ErrInvalidMode means an unknown import mode was requested.
	ErrInvalidMode = errors.New("invalid import mode")
	// ErrUnsupportedVersion means the archive format is newer than this build.
	ErrUnsupportedVersion = errors.New("unsupported backup format version")
	// ErrInvalidArchive means the uploaded file is not a readable backup archive.
	ErrInvalidArchive = errors.New("invalid backup archive")
	// ErrArchiveTooLarge means the archive's decompressed contents exceed the
	// allowed size, guarding against decompression-bomb uploads.
	ErrArchiveTooLarge = errors.New("backup archive contents exceed the maximum allowed size")
)

// Service builds and restores backup archives and writes scheduled backups to disk.
type Service struct {
	db      *sql.DB
	baseDir string // BACKUP_BASE_DIR; empty disables server-side writes
}

// NewService constructs a backup Service. baseDir may be empty (download/restore
// still work; only scheduled and run-now writes require it).
func NewService(db *sql.DB, baseDir string) *Service {
	return &Service{db: db, baseDir: baseDir}
}
