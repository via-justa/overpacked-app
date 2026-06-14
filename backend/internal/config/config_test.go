package config

import "testing"

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected an error when DATABASE_URL is missing")
	}
}

func TestLoadDefaultsAndValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("SERVER_ADDR", "")
	t.Setenv("APP_USERNAME", "admin")
	t.Setenv("APP_PASSWORD", "pw")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("ENABLE_SEED_DATA", "true")
	t.Setenv("BACKUP_BASE_DIR", "/tmp/backups")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ServerAddr != defaultServerAddr {
		t.Fatalf("expected default server addr %q, got %q", defaultServerAddr, cfg.ServerAddr)
	}
	if cfg.AppUsername != "admin" || cfg.AppPassword != "pw" || cfg.JWTSecret != "secret" {
		t.Fatalf("unexpected auth config: %+v", cfg)
	}
	if !cfg.EnableSeedData || cfg.BackupBaseDir != "/tmp/backups" {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	// Explicit SERVER_ADDR is honored; a non-truthy seed flag stays false.
	t.Setenv("SERVER_ADDR", "127.0.0.1:9999")
	t.Setenv("ENABLE_SEED_DATA", "no")
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg2.ServerAddr != "127.0.0.1:9999" {
		t.Fatalf("expected explicit server addr, got %q", cfg2.ServerAddr)
	}
	if cfg2.EnableSeedData {
		t.Fatal("expected EnableSeedData false for non-truthy value")
	}
}
