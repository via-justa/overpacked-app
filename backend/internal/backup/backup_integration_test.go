package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/via-justa/overpacked-app/backend/internal/db"
	"github.com/via-justa/overpacked-app/backend/internal/migrations"
)

var (
	backupMigrationsOnce sync.Once
	backupMigrationsErr  error
)

func newContainerizedService(t *testing.T) (*Service, *sql.DB) {
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

	backupMigrationsOnce.Do(func() {
		backupMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if backupMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", backupMigrationsErr)
	}

	if _, err := dbConn.ExecContext(context.Background(), `
		TRUNCATE TABLE trip_person_items, trip_person_packs, trip_persons, trips,
			pack_items, packs, set_items, item_sets, item_labels, packing_list_labels,
			items, packing_lists, labels, manufacturers, persons RESTART IDENTITY CASCADE`); err != nil {
		_ = dbConn.Close()
		t.Fatalf("truncate tables: %v", err)
	}

	return NewService(dbConn, t.TempDir()), dbConn
}

// seedRoundTripData inserts a manufacturer, a label, and an item (with image + label)
// and returns the item id. Item type "consumable" is a system type from the migration.
func seedRoundTripData(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	ctx := context.Background()

	manuID := uuid.New()
	if _, err := dbConn.ExecContext(ctx,
		`INSERT INTO manufacturers (id, name, website) VALUES ($1, 'TestCo', 'https://test.co')`, manuID); err != nil {
		t.Fatalf("insert manufacturer: %v", err)
	}

	labelID := uuid.New()
	if _, err := dbConn.ExecContext(ctx,
		`INSERT INTO labels (id, name, color) VALUES ($1, 'tent', '#FF0000')`, labelID); err != nil {
		t.Fatalf("insert label: %v", err)
	}

	itemID := uuid.New()
	image := []byte{0xFF, 0xD8, 0xFF, 0x00, 0x01, 0x02}
	if _, err := dbConn.ExecContext(ctx, `
		INSERT INTO items (id, manufacturer_id, type_id, name, is_active, weight_grams,
			image_blob, image_mime_type, image_size_bytes, image_width_px, image_height_px)
		VALUES ($1, $2, 'consumable', 'Test Bar', TRUE, 50, $3, 'image/jpeg', $4, 10, 10)`,
		itemID, manuID, image, len(image)); err != nil {
		t.Fatalf("insert item: %v", err)
	}

	if _, err := dbConn.ExecContext(ctx,
		`INSERT INTO item_labels (item_id, label_id) VALUES ($1, $2)`, itemID, labelID); err != nil {
		t.Fatalf("insert item label: %v", err)
	}
}

func TestServiceBackupRoundTripReplace(t *testing.T) {
	svc, dbConn := newContainerizedService(t)
	defer func() { _ = dbConn.Close() }()
	ctx := context.Background()

	seedRoundTripData(t, dbConn)

	var archive bytes.Buffer
	if err := svc.BuildArchive(ctx, &archive); err != nil {
		t.Fatalf("build archive: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}

	result, err := svc.Import(ctx, zr, ModeReplace)
	if err != nil {
		t.Fatalf("import replace: %v", err)
	}
	if result.Counts["items"] != 1 {
		t.Fatalf("expected 1 item imported, got %d", result.Counts["items"])
	}

	var (
		itemCount    int
		manufacturer string
		imageLen     int
		labelCount   int
	)
	if err := dbConn.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(m.name), MAX(octet_length(i.image_blob))
		FROM items i JOIN manufacturers m ON m.id = i.manufacturer_id`).
		Scan(&itemCount, &manufacturer, &imageLen); err != nil {
		t.Fatalf("query items: %v", err)
	}
	if itemCount != 1 {
		t.Fatalf("expected 1 item after replace, got %d", itemCount)
	}
	if manufacturer != "TestCo" {
		t.Fatalf("expected manufacturer TestCo, got %q", manufacturer)
	}
	if imageLen != 6 {
		t.Fatalf("expected image blob of 6 bytes restored, got %d", imageLen)
	}

	if err := dbConn.QueryRowContext(ctx, `SELECT COUNT(*) FROM item_labels`).Scan(&labelCount); err != nil {
		t.Fatalf("query item_labels: %v", err)
	}
	if labelCount != 1 {
		t.Fatalf("expected 1 item label after restore, got %d", labelCount)
	}
}

func TestServiceExportItemsCSV(t *testing.T) {
	svc, dbConn := newContainerizedService(t)
	defer func() { _ = dbConn.Close() }()
	ctx := context.Background()

	seedRoundTripData(t, dbConn)

	var out bytes.Buffer
	if err := svc.ExportItemsCSV(ctx, &out, false); err != nil {
		t.Fatalf("export csv: %v", err)
	}

	csv := out.String()
	if !strings.Contains(csv, "name,manufacturer,item_type") {
		t.Fatalf("csv missing header, got: %q", csv)
	}
	if !strings.Contains(csv, "Test Bar") || !strings.Contains(csv, "TestCo") || !strings.Contains(csv, "tent") {
		t.Fatalf("csv missing denormalized data, got: %q", csv)
	}
}
