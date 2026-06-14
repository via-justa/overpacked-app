package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	backupFilePrefix = "overpacked-backup-"
	backupFileExt    = ".zip"
	backupTimeLayout = "20060102-150405"
)

// Run builds an archive and writes it atomically into BACKUP_BASE_DIR, then prunes
// older backups beyond retention. It returns the absolute path of the written file.
// ts stamps the filename.
func (s *Service) Run(ctx context.Context, retention int, ts time.Time) (string, error) {
	if s.baseDir == "" {
		return "", ErrBackupDirNotConfigured
	}

	dir, err := filepath.Abs(s.baseDir)
	if err != nil {
		return "", fmt.Errorf("resolve backup base dir: %w", err)
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".backup-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temp backup: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once renamed

	if err := s.BuildArchive(ctx, tmp); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close temp backup: %w", err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return "", fmt.Errorf("set backup permissions: %w", err)
	}

	filename := backupFilePrefix + ts.UTC().Format(backupTimeLayout) + backupFileExt
	finalPath := filepath.Join(dir, filename)
	if err := os.Rename(tmpName, finalPath); err != nil {
		return "", fmt.Errorf("finalize backup: %w", err)
	}

	if err := pruneBackups(dir, retention); err != nil {
		return finalPath, fmt.Errorf("prune old backups: %w", err)
	}

	return finalPath, nil
}

// pruneBackups keeps only the newest `retention` backup files in dir, deleting older
// ones. It only ever touches files matching the backup naming pattern.
func pruneBackups(dir string, retention int) error {
	if retention <= 0 {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read backup dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, backupFilePrefix) && strings.HasSuffix(name, backupFileExt) {
			files = append(files, name)
		}
	}

	if len(files) <= retention {
		return nil
	}

	// Timestamped names sort chronologically; drop everything but the newest `retention`.
	sort.Strings(files)
	for _, name := range files[:len(files)-retention] {
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return fmt.Errorf("remove old backup %s: %w", name, err)
		}
	}

	return nil
}
