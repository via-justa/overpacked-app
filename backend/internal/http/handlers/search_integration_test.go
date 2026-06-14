package handlers

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"

	"github.com/via-justa/overpacked-app/backend/internal/db"
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/migrations"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

var (
	searchMigrationsOnce sync.Once
	searchMigrationsErr  error
)

const searchSeedToken = "Alpine"

func newContainerizedSearchStore(t *testing.T) (*store.Store, *sql.DB) {
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

	searchMigrationsOnce.Do(func() {
		searchMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if searchMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", searchMigrationsErr)
	}

	truncate := "TRUNCATE TABLE items, item_sets, packing_lists, persons, manufacturers, trips RESTART IDENTITY CASCADE"
	if _, err := dbConn.ExecContext(context.Background(), truncate); err != nil {
		_ = dbConn.Close()
		t.Fatalf("truncate tables: %v", err)
	}

	seedSearchEntities(t, dbConn)

	return store.New(dbConn), dbConn
}

// seedSearchEntities inserts exactly one of each searchable entity, all sharing
// the searchSeedToken in their name so filter and ranking behaviour can be tested.
func seedSearchEntities(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	ctx := context.Background()

	statements := []string{
		`INSERT INTO manufacturers (id, name) VALUES (gen_random_uuid(), 'Alpine Outfitters')`,
		`INSERT INTO items (id, manufacturer_id, type_id, name, is_active)
			SELECT gen_random_uuid(), m.id, 'shelter', 'Alpine Tent', TRUE
			FROM manufacturers m WHERE m.name = 'Alpine Outfitters'`,
		`INSERT INTO item_sets (id, name, set_category) VALUES (gen_random_uuid(), 'Alpine Set', 'shelter')`,
		`INSERT INTO packing_lists (id, name) VALUES (gen_random_uuid(), 'Alpine List')`,
		`INSERT INTO persons (id, name) VALUES (gen_random_uuid(), 'Alpine Hiker')`,
		`INSERT INTO trips (id, name, trip_type) VALUES (gen_random_uuid(), 'Alpine Trek', 'overnight')`,
	}

	for _, stmt := range statements {
		if _, err := dbConn.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seed search entity: %v", err)
		}
	}
}

func TestSearchStoreIntegrationMatchesAllEntities(t *testing.T) {
	st, dbConn := newContainerizedSearchStore(t)
	defer func() { _ = dbConn.Close() }()

	results, err := st.Search.Search(context.Background(), searchSeedToken, nil, 50)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	found := make(map[domain.SearchEntityType]bool)
	for _, r := range results {
		found[r.EntityType] = true
	}

	want := []domain.SearchEntityType{
		domain.SearchEntityItem,
		domain.SearchEntitySet,
		domain.SearchEntityPackingList,
		domain.SearchEntityPerson,
		domain.SearchEntityManufacturer,
		domain.SearchEntityTrip,
	}
	for _, entity := range want {
		if !found[entity] {
			t.Fatalf("expected a result for entity type %q, got results: %+v", entity, results)
		}
	}
}

func TestSearchStoreIntegrationTypeFilter(t *testing.T) {
	st, dbConn := newContainerizedSearchStore(t)
	defer func() { _ = dbConn.Close() }()

	types := []domain.SearchEntityType{domain.SearchEntityPerson}
	results, err := st.Search.Search(context.Background(), searchSeedToken, types, 50)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected exactly 1 person result, got %d: %+v", len(results), results)
	}
	if results[0].EntityType != domain.SearchEntityPerson {
		t.Fatalf("expected person result, got %q", results[0].EntityType)
	}
	if results[0].Title != "Alpine Hiker" {
		t.Fatalf("expected person title 'Alpine Hiker', got %q", results[0].Title)
	}
}

func TestSearchStoreIntegrationTypoTolerance(t *testing.T) {
	st, dbConn := newContainerizedSearchStore(t)
	defer func() { _ = dbConn.Close() }()

	// "Alpne Outfitters" is a one-character typo of the manufacturer name and is
	// not a substring, so the match relies on trigram similarity.
	results, err := st.Search.Search(context.Background(), "Alpne Outfitters", nil, 50)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	var matched bool
	for _, r := range results {
		if r.EntityType == domain.SearchEntityManufacturer && r.Title == "Alpine Outfitters" {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("expected typo query to match manufacturer via trigram, got: %+v", results)
	}
}
