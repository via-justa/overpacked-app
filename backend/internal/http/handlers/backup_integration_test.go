package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/backup"
	"github.com/via-justa/overpacked-app/backend/internal/db"
	"github.com/via-justa/overpacked-app/backend/internal/migrations"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

const backupTestPassword = "pw123"

var (
	backupHandlerMigrationsOnce sync.Once
	backupHandlerMigrationsErr  error
)

func newContainerizedBackupHandler(t *testing.T) (*BackupHandler, *sql.DB, string) {
	t.Helper()

	if os.Getenv("RUN_CONTAINERIZED_TESTS") != "true" {
		t.Skip("containerized integration tests are disabled")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is required for integration tests")
	}

	dbConn, err := db.Connect(databaseURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}

	backupHandlerMigrationsOnce.Do(func() {
		backupHandlerMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if backupHandlerMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", backupHandlerMigrationsErr)
	}

	if _, err := dbConn.ExecContext(context.Background(), `
		TRUNCATE TABLE trip_person_items, trip_person_packs, trip_persons, trips,
			pack_items, packs, set_items, item_sets, item_labels, packing_list_labels,
			items, packing_lists, labels, manufacturers, persons RESTART IDENTITY CASCADE`); err != nil {
		_ = dbConn.Close()
		t.Fatalf("truncate tables: %v", err)
	}

	baseDir := t.TempDir()
	st := store.New(dbConn)
	svc := backup.NewService(dbConn, baseDir)
	scheduler := backup.NewScheduler(svc, st.BackupConfig)
	return NewBackupHandler(svc, st, scheduler, backupTestPassword), dbConn, baseDir
}

// seedBackupItem inserts a manufacturer and one item so exports have content.
func seedBackupItem(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	ctx := context.Background()

	manuID := uuid.New()
	if _, err := dbConn.ExecContext(ctx,
		`INSERT INTO manufacturers (id, name) VALUES ($1, 'HandlerCo')`, manuID); err != nil {
		t.Fatalf("insert manufacturer: %v", err)
	}
	if _, err := dbConn.ExecContext(ctx, `
		INSERT INTO items (id, manufacturer_id, type_id, name, is_active, weight_grams)
		VALUES ($1, $2, 'consumable', 'Handler Bar', TRUE, 42)`, uuid.New(), manuID); err != nil {
		t.Fatalf("insert item: %v", err)
	}
}

func newImportRequest(t *testing.T, archive []byte, mode, password string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", "backup.zip")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(archive); err != nil {
		t.Fatalf("write archive part: %v", err)
	}
	if err := mw.WriteField("mode", mode); err != nil {
		t.Fatalf("write mode field: %v", err)
	}
	if err := mw.WriteField("password", password); err != nil {
		t.Fatalf("write password field: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/backup/import", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestBackupHandlerExportImportRoundTrip(t *testing.T) {
	h, dbConn, _ := newContainerizedBackupHandler(t)
	defer func() { _ = dbConn.Close() }()

	seedBackupItem(t, dbConn)

	exportReq := httptest.NewRequest(http.MethodGet, "/api/v1/backup/export", http.NoBody)
	exportW := httptest.NewRecorder()
	h.ExportBackup(exportW, exportReq)

	if exportW.Code != http.StatusOK {
		t.Fatalf("expected 200 from export, got %d", exportW.Code)
	}
	if ct := exportW.Header().Get("Content-Type"); ct != "application/zip" {
		t.Fatalf("expected application/zip, got %q", ct)
	}
	archive := exportW.Body.Bytes()
	if _, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive))); err != nil {
		t.Fatalf("export is not a valid zip: %v", err)
	}

	importW := httptest.NewRecorder()
	h.ImportBackup(importW, newImportRequest(t, archive, "merge", backupTestPassword))

	if importW.Code != http.StatusOK {
		t.Fatalf("expected 200 from import, got %d (body: %s)", importW.Code, importW.Body.String())
	}
	var result api.BackupImportResult
	if err := json.NewDecoder(importW.Body).Decode(&result); err != nil {
		t.Fatalf("decode import result: %v", err)
	}
	if result.Counts["items"] != 1 {
		t.Fatalf("expected 1 item in import counts, got %d", result.Counts["items"])
	}
}

func TestBackupHandlerImportWrongPassword(t *testing.T) {
	h, dbConn, _ := newContainerizedBackupHandler(t)
	defer func() { _ = dbConn.Close() }()

	seedBackupItem(t, dbConn)

	exportW := httptest.NewRecorder()
	h.ExportBackup(exportW, httptest.NewRequest(http.MethodGet, "/api/v1/backup/export", http.NoBody))

	importW := httptest.NewRecorder()
	h.ImportBackup(importW, newImportRequest(t, exportW.Body.Bytes(), "merge", "wrong-password"))

	if importW.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong password, got %d", importW.Code)
	}
}

func TestBackupHandlerConfigGetAndUpdate(t *testing.T) {
	h, dbConn, _ := newContainerizedBackupHandler(t)
	defer func() { _ = dbConn.Close() }()

	getW := httptest.NewRecorder()
	h.GetBackupConfig(getW, httptest.NewRequest(http.MethodGet, "/api/v1/backup/config", http.NoBody))
	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 from get config, got %d", getW.Code)
	}

	body := []byte(`{"enabled":true,"schedule":"*/5 * * * *","retention_count":3}`)
	updW := httptest.NewRecorder()
	h.UpdateBackupConfig(updW, httptest.NewRequest(http.MethodPut, "/api/v1/backup/config", bytes.NewReader(body)))
	if updW.Code != http.StatusOK {
		t.Fatalf("expected 200 from update config, got %d (body: %s)", updW.Code, updW.Body.String())
	}

	var updated api.BackupConfig
	if err := json.NewDecoder(updW.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated config: %v", err)
	}
	if !updated.Enabled || updated.Schedule != "*/5 * * * *" || updated.RetentionCount != 3 {
		t.Fatalf("config not persisted as sent: %+v", updated)
	}
}

func TestBackupHandlerUpdateConfigRejectsInvalidCron(t *testing.T) {
	h, dbConn, _ := newContainerizedBackupHandler(t)
	defer func() { _ = dbConn.Close() }()

	body := []byte(`{"enabled":true,"schedule":"not a cron","retention_count":3}`)
	w := httptest.NewRecorder()
	h.UpdateBackupConfig(w, httptest.NewRequest(http.MethodPut, "/api/v1/backup/config", bytes.NewReader(body)))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid cron, got %d", w.Code)
	}
}

func TestBackupHandlerRunBackupWritesFile(t *testing.T) {
	h, dbConn, baseDir := newContainerizedBackupHandler(t)
	defer func() { _ = dbConn.Close() }()

	seedBackupItem(t, dbConn)

	runW := httptest.NewRecorder()
	h.RunBackup(runW, httptest.NewRequest(http.MethodPost, "/api/v1/backup/run", http.NoBody))
	if runW.Code != http.StatusOK {
		t.Fatalf("expected 200 from run backup, got %d (body: %s)", runW.Code, runW.Body.String())
	}

	var result api.BackupRunResult
	if err := json.NewDecoder(runW.Body).Decode(&result); err != nil {
		t.Fatalf("decode run result: %v", err)
	}
	if result.Path == "" {
		t.Fatal("expected a non-empty backup path")
	}
	if _, err := os.Stat(result.Path); err != nil {
		t.Fatalf("expected backup file to exist at %s: %v", result.Path, err)
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatalf("read backup dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 backup file written, got %d", len(entries))
	}
}

func TestBackupHandlerExportItemsCSV(t *testing.T) {
	h, dbConn, _ := newContainerizedBackupHandler(t)
	defer func() { _ = dbConn.Close() }()

	seedBackupItem(t, dbConn)

	includeImages := false
	w := httptest.NewRecorder()
	h.ExportItems(w, httptest.NewRequest(http.MethodGet, "/api/v1/export/items", http.NoBody),
		api.ExportItemsParams{IncludeImages: &includeImages})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from export items, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv" {
		t.Fatalf("expected text/csv, got %q", ct)
	}
	csv := w.Body.String()
	if !bytes.Contains([]byte(csv), []byte("name,manufacturer,item_type")) {
		t.Fatalf("csv missing header: %q", csv)
	}
	if !bytes.Contains([]byte(csv), []byte("Handler Bar")) || !bytes.Contains([]byte(csv), []byte("HandlerCo")) {
		t.Fatalf("csv missing seeded data: %q", csv)
	}
}
