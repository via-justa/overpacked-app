package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strconv"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var migrationFiles embed.FS

func Run(ctx context.Context, db *sql.DB, command string, args []string) error {
	subFS, err := fs.Sub(migrationFiles, "sql")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	goose.SetBaseFS(subFS)
	goose.SetDialect("postgres")

	switch command {
	case "up":
		return goose.UpContext(ctx, db, ".", goose.WithAllowMissing())
	case "down":
		return goose.DownContext(ctx, db, ".")
	case "reset":
		return goose.ResetContext(ctx, db, ".")
	case "status":
		return goose.StatusContext(ctx, db, ".")
	case "up-to":
		version, parseErr := parseVersionArg(args)
		if parseErr != nil {
			return parseErr
		}
		return goose.UpToContext(ctx, db, ".", version)
	case "down-to":
		version, parseErr := parseVersionArg(args)
		if parseErr != nil {
			return parseErr
		}
		return goose.DownToContext(ctx, db, ".", version)
	default:
		return fmt.Errorf("unsupported migration command: %s", command)
	}
}

func parseVersionArg(args []string) (int64, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("migration version argument is required")
	}

	version, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid migration version %q: %w", args[0], err)
	}

	return version, nil
}
