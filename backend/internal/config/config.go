package config

import (
	"fmt"
	"os"
)

const (
	defaultServerAddr = "0.0.0.0:8000"
)

type Config struct {
	DatabaseURL string
	ServerAddr  string
	AppUsername string
	AppPassword string
	JWTSecret   string
	// EnableSeedData runs the seed catalog on startup when true ("true"/"1").
	EnableSeedData bool
	// BackupBaseDir confines server-side backup writes. Empty disables scheduled/run-now backups.
	BackupBaseDir string
}

func Load() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = defaultServerAddr
	}

	return &Config{
		DatabaseURL:    databaseURL,
		ServerAddr:     serverAddr,
		AppUsername:    os.Getenv("APP_USERNAME"),
		AppPassword:    os.Getenv("APP_PASSWORD"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		EnableSeedData: isTruthy(os.Getenv("ENABLE_SEED_DATA")),
		BackupBaseDir:  os.Getenv("BACKUP_BASE_DIR"),
	}, nil
}

func isTruthy(v string) bool {
	switch v {
	case "true", "TRUE", "True", "1", "yes":
		return true
	default:
		return false
	}
}
