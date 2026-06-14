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

// seedFullDataset inserts one row of every user entity so export/import exercise
// the happy path of every query*/apply* function (not just items/labels).
func seedFullDataset(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	ctx := context.Background()

	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := dbConn.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("seed exec failed: %v\nquery: %s", err, query)
		}
	}

	manuID, labelID, personID := uuid.New(), uuid.New(), uuid.New()
	itemID, customItemID := uuid.New(), uuid.New()
	setID, packID, listID, tripID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	const customType = "seed_custom_type"

	exec(`INSERT INTO manufacturers (id, name, website) VALUES ($1, 'FullCo', 'https://full.co')`, manuID)
	exec(`INSERT INTO labels (id, name, color) VALUES ($1, 'seed-label', '#123456')`, labelID)
	exec(`INSERT INTO item_types (id, name, description, base_profile, is_system)
		VALUES ($1, 'Seed Type', 'desc', 'electronics', FALSE)`, customType)
	exec(`INSERT INTO item_type_fields (id, item_type_id, field_key, field_label, data_type, is_required, sort_order, unit)
		VALUES ($1, $2, 'watts', 'Watts', 'number', FALSE, 1, 'W')`, uuid.New(), customType)
	exec(`INSERT INTO persons (id, name, gender, body_weight_grams, conditioning_level)
		VALUES ($1, 'Seed Person', 'other', 70000, 'average')`, personID)

	img := []byte{0xFF, 0xD8, 0xFF, 0x00, 0x01, 0x02}
	exec(`INSERT INTO items (id, manufacturer_id, type_id, name, is_active, weight_grams,
		default_quantity, default_carry_status, is_default,
		image_blob, image_mime_type, image_size_bytes, image_width_px, image_height_px)
		VALUES ($1, $2, 'consumable', 'Seed Item', TRUE, 50, 1, 'packed', FALSE, $3, 'image/jpeg', $4, 10, 10)`,
		itemID, manuID, img, len(img))
	exec(`INSERT INTO items (id, manufacturer_id, type_id, name, is_active, default_quantity, default_carry_status, is_default, attributes_json)
		VALUES ($1, $2, $3, 'Custom Seed Item', TRUE, 1, 'packed', FALSE, '{"watts":30}')`, customItemID, manuID, customType)
	exec(`INSERT INTO item_labels (item_id, label_id) VALUES ($1, $2)`, itemID, labelID)

	exec(`INSERT INTO item_sets (id, name, description, is_active, set_category)
		VALUES ($1, 'Seed Set', 'desc', TRUE, 'consumable')`, setID)
	exec(`INSERT INTO set_items (set_id, item_id, quantity, notes, sort_order) VALUES ($1, $2, 2, 'note', 1)`, setID, itemID)

	exec(`INSERT INTO packs (id, person_id, name, trip_type, notes)
		VALUES ($1, $2, 'Seed Pack', 'overnight', 'note')`, packID, personID)
	exec(`INSERT INTO pack_items (pack_id, item_id, quantity, carry_status, notes) VALUES ($1, $2, 3, 'packed', 'note')`, packID, itemID)

	exec(`INSERT INTO packing_lists (id, name, description) VALUES ($1, 'Seed List', 'desc')`, listID)
	exec(`INSERT INTO packing_list_labels (packing_list_id, label_id) VALUES ($1, $2)`, listID, labelID)

	exec(`INSERT INTO trips (id, name, trip_type, notes, total_distance_km)
		VALUES ($1, 'Seed Trip', 'overnight', 'note', 12.5)`, tripID)
	tripPersonID := uuid.New()
	exec(`INSERT INTO trip_persons (id, trip_id, person_id) VALUES ($1, $2, $3)`, tripPersonID, tripID, personID)
	exec(`INSERT INTO trip_person_items (trip_person_id, item_id, quantity, carry_status, notes) VALUES ($1, $2, 1, 'worn', 'note')`, tripPersonID, itemID)
	exec(`INSERT INTO trip_person_packs (trip_person_id, pack_id) VALUES ($1, $2)`, tripPersonID, packID)
}

func TestServiceBackupRoundTripFull(t *testing.T) {
	svc, dbConn := newContainerizedService(t)
	defer func() { _ = dbConn.Close() }()
	ctx := context.Background()

	seedFullDataset(t, dbConn)

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
		t.Fatalf("import replace full dataset: %v", err)
	}

	// Every entity kind should round-trip with a non-zero count.
	for _, kind := range []string{"manufacturers", "labels", "item_types", "persons", "items", "item_sets", "packs", "packing_lists", "trips"} {
		if result.Counts[kind] == 0 {
			t.Errorf("expected %s to be restored, got count 0 (counts: %v)", kind, result.Counts)
		}
	}

	// Re-export after import to exercise the query* paths a second time on restored data.
	var reArchive bytes.Buffer
	if err := svc.BuildArchive(ctx, &reArchive); err != nil {
		t.Fatalf("re-export after import: %v", err)
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

	// includeImages=true switches to a ZIP that also writes image entries,
	// exercising writeImageEntry / uniqueImageName / sanitizeFilename.
	var zipOut bytes.Buffer
	if err := svc.ExportItemsCSV(ctx, &zipOut, true); err != nil {
		t.Fatalf("export csv with images: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipOut.Bytes()), int64(zipOut.Len()))
	if err != nil {
		t.Fatalf("open items export zip: %v", err)
	}
	if len(zr.File) == 0 {
		t.Fatal("expected items export zip to contain entries")
	}
}
