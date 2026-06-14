package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/backup"
	"github.com/via-justa/overpacked-app/backend/internal/db"
	"github.com/via-justa/overpacked-app/backend/internal/migrations"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

// TestHandlersRejectInvalidBody covers the decode-error (400) branch of every
// body-decoding handler. These decode before touching the store, so a nil store
// is fine — no database required.
func TestHandlersRejectInvalidBody(t *testing.T) {
	t.Parallel()

	id := types.UUID(uuid.New())
	check := func(name string, fn func(http.ResponseWriter, *http.Request)) {
		w := httptest.NewRecorder()
		fn(w, httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader([]byte(`{`))))
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400 for invalid body, got %d", name, w.Code)
		}
	}

	m := NewManufacturersHandler(nil)
	check("manufacturers.create", m.CreateManufacturer)
	check("manufacturers.update", func(w http.ResponseWriter, r *http.Request) { m.UpdateManufacturer(w, r, id) })

	it := NewItemTypesHandler(nil)
	check("itemTypes.create", it.CreateItemType)
	check("itemTypes.update", func(w http.ResponseWriter, r *http.Request) { it.UpdateItemType(w, r, "t") })
	check("itemTypes.replaceFields", func(w http.ResponseWriter, r *http.Request) { it.ReplaceItemTypeFields(w, r, "t") })

	pl := NewPackingListsHandler(nil, nil)
	check("packingLists.create", pl.CreatePackingList)
	check("packingLists.update", func(w http.ResponseWriter, r *http.Request) { pl.UpdatePackingList(w, r, id) })
	check("packingLists.addLabel", func(w http.ResponseWriter, r *http.Request) { pl.AddPackingListLabel(w, r, id) })

	tr := NewTripsHandler(nil)
	check("trips.create", tr.CreateTrip)
	check("trips.update", func(w http.ResponseWriter, r *http.Request) { tr.UpdateTrip(w, r, id) })
	check("trips.addPerson", func(w http.ResponseWriter, r *http.Request) { tr.AddTripPerson(w, r, id) })
	check("trips.addPersonItem", func(w http.ResponseWriter, r *http.Request) { tr.AddTripPersonItem(w, r, id, id) })
	check("trips.updatePersonItem", func(w http.ResponseWriter, r *http.Request) { tr.UpdateTripPersonItem(w, r, id, id, id) })
	check("trips.addPersonPack", func(w http.ResponseWriter, r *http.Request) { tr.AddTripPersonPack(w, r, id, id) })
	check("trips.addPackItem", func(w http.ResponseWriter, r *http.Request) { tr.AddTripPersonPackItem(w, r, id, id, id) })
	check("trips.updatePackItem", func(w http.ResponseWriter, r *http.Request) { tr.UpdateTripPersonPackItem(w, r, id, id, id, id) })

	s := NewSetsHandler(nil)
	check("sets.addItem", func(w http.ResponseWriter, r *http.Request) { s.AddSetItem(w, r, id) })
	check("sets.updateItem", func(w http.ResponseWriter, r *http.Request) { s.UpdateSetItem(w, r, id, id) })

	check("settings.update", NewSettingsHandler(nil, "pw").UpdateSettings)
	check("backupConfig.update", NewBackupHandler(nil, nil, nil, "pw").UpdateBackupConfig)
}

var (
	errorPathsMigrationsOnce sync.Once
	errorPathsMigrationsErr  error
)

func newContainerizedStore(t *testing.T) (*store.Store, *sql.DB) {
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
	errorPathsMigrationsOnce.Do(func() {
		errorPathsMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if errorPathsMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", errorPathsMigrationsErr)
	}
	return store.New(dbConn), dbConn
}

// TestHandlersNotFound covers the ErrNotFound (404) branch of get/update/delete
// handlers (and, through them, the store's sql.ErrNoRows → domain.ErrNotFound
// translation) by addressing random ids that don't exist.
func TestHandlersNotFound(t *testing.T) {
	st, dbConn := newContainerizedStore(t)
	defer func() { _ = dbConn.Close() }()

	rid := types.UUID(uuid.New())
	missing := func(name string, want int, fn func(http.ResponseWriter, *http.Request)) {
		w := httptest.NewRecorder()
		fn(w, httptest.NewRequest(http.MethodGet, "/x", http.NoBody))
		if w.Code != want {
			t.Errorf("%s: expected %d for missing id, got %d", name, want, w.Code)
		}
	}

	ih := NewItemsHandler(st)
	missing("items.get", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { ih.GetItem(w, r, rid) })
	missing("items.delete", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { ih.DeleteItem(w, r, rid) })

	mh := NewManufacturersHandler(st)
	missing("manufacturers.get", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { mh.GetManufacturer(w, r, rid) })
	missing("manufacturers.delete", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { mh.DeleteManufacturer(w, r, rid) })

	lh := NewLabelsHandler(st)
	missing("labels.get", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { lh.GetLabel(w, r, rid) })
	missing("labels.delete", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { lh.DeleteLabel(w, r, rid) })

	ph := NewPersonsHandler(st)
	missing("persons.get", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { ph.GetPerson(w, r, rid) })
	missing("persons.delete", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { ph.DeletePerson(w, r, rid) })

	sh := NewSetsHandler(st)
	missing("sets.get", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { sh.GetSet(w, r, rid) })
	missing("sets.delete", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { sh.DeleteSet(w, r, rid) })

	plh := NewPackingListsHandler(st.PackingLists, st.Labels)
	missing("packingLists.get", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { plh.GetPackingListById(w, r, rid) })
	missing("packingLists.delete", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { plh.DeletePackingList(w, r, rid) })

	th := NewTripsHandler(st)
	missing("trips.get", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { th.GetTripById(w, r, rid) })
	missing("trips.delete", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { th.DeleteTrip(w, r, rid) })

	ith := NewItemTypesHandler(st)
	missing("itemTypes.get", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { ith.GetItemType(w, r, "does-not-exist") })
	missing("itemTypes.delete", http.StatusNotFound, func(w http.ResponseWriter, r *http.Request) { ith.DeleteItemType(w, r, "does-not-exist") })
}

// TestHandlersDBUnavailable closes the database connection and then drives every
// handler, covering the store-error (5xx) branch in each handler and the
// error-wrapping branch in each store method.
func TestHandlersDBUnavailable(t *testing.T) {
	st, dbConn := newContainerizedStore(t)
	if err := dbConn.Close(); err != nil { // every subsequent query now fails
		t.Fatalf("close db: %v", err)
	}

	id := types.UUID(uuid.New())
	fail := func(name string, fn func(http.ResponseWriter, *http.Request)) {
		w := httptest.NewRecorder()
		fn(w, httptest.NewRequest(http.MethodGet, "/x", http.NoBody))
		if w.Code < 500 {
			t.Errorf("%s: expected 5xx with DB down, got %d", name, w.Code)
		}
	}
	post := func(name, body string, fn func(http.ResponseWriter, *http.Request)) {
		w := httptest.NewRecorder()
		fn(w, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body)))
		if w.Code < 500 {
			t.Errorf("%s: expected 5xx with DB down, got %d", name, w.Code)
		}
	}

	ih := NewItemsHandler(st)
	fail("items.list", ih.ListItems)
	fail("items.get", func(w http.ResponseWriter, r *http.Request) { ih.GetItem(w, r, id) })
	post("items.create", `{"name":"x","type":"consumable","is_active":true,"manufacturer_id":"`+id.String()+`"}`, ih.CreateItem)

	mh := NewManufacturersHandler(st)
	fail("manufacturers.list", mh.ListManufacturers)
	fail("manufacturers.get", func(w http.ResponseWriter, r *http.Request) { mh.GetManufacturer(w, r, id) })
	post("manufacturers.create", `{"name":"x"}`, mh.CreateManufacturer)

	lh := NewLabelsHandler(st)
	fail("labels.list", lh.ListLabels)
	post("labels.create", `{"name":"x","color":"#fff"}`, lh.CreateLabel)

	ph := NewPersonsHandler(st)
	fail("persons.list", ph.ListPersons)
	post("persons.create", `{"name":"x"}`, ph.CreatePerson)

	sh := NewSetsHandler(st)
	fail("sets.list", sh.ListSets)
	post("sets.create", `{"name":"x","set_category":"consumable"}`, sh.CreateSet)

	plh := NewPackingListsHandler(st.PackingLists, st.Labels)
	fail("packingLists.list", plh.ListPackingLists)
	post("packingLists.create", `{"name":"x"}`, plh.CreatePackingList)

	th := NewTripsHandler(st)
	fail("trips.list", th.ListTrips)
	post("trips.create", `{"name":"x","trip_type":"overnight"}`, th.CreateTrip)

	ith := NewItemTypesHandler(st)
	fail("itemTypes.list", ith.ListItemTypes)
	post("itemTypes.create", `{"id":"x","name":"X"}`, ith.CreateItemType)

	rid := id.String()

	// Update / delete / nested operations — all reach the store and fail.
	post("items.update", `{"name":"x"}`, func(w http.ResponseWriter, r *http.Request) { ih.UpdateItem(w, r, id) })
	fail("items.delete", func(w http.ResponseWriter, r *http.Request) { ih.DeleteItem(w, r, id) })

	post("manufacturers.update", `{"name":"x"}`, func(w http.ResponseWriter, r *http.Request) { mh.UpdateManufacturer(w, r, id) })
	fail("manufacturers.delete", func(w http.ResponseWriter, r *http.Request) { mh.DeleteManufacturer(w, r, id) })

	fail("labels.get", func(w http.ResponseWriter, r *http.Request) { lh.GetLabel(w, r, id) })
	post("labels.update", `{"name":"x","color":"#fff"}`, func(w http.ResponseWriter, r *http.Request) { lh.UpdateLabel(w, r, id) })
	fail("labels.delete", func(w http.ResponseWriter, r *http.Request) { lh.DeleteLabel(w, r, id) })
	fail("labels.listItemLabels", func(w http.ResponseWriter, r *http.Request) { lh.ListItemLabels(w, r, id) })
	post("labels.addItemLabel", `{"label_id":"`+rid+`"}`, func(w http.ResponseWriter, r *http.Request) { lh.AddItemLabel(w, r, id) })
	fail("labels.removeItemLabel", func(w http.ResponseWriter, r *http.Request) { lh.RemoveItemLabel(w, r, id, id) })

	fail("persons.get", func(w http.ResponseWriter, r *http.Request) { ph.GetPerson(w, r, id) })
	post("persons.update", `{"name":"x"}`, func(w http.ResponseWriter, r *http.Request) { ph.UpdatePerson(w, r, id) })
	fail("persons.delete", func(w http.ResponseWriter, r *http.Request) { ph.DeletePerson(w, r, id) })

	fail("sets.get", func(w http.ResponseWriter, r *http.Request) { sh.GetSet(w, r, id) })
	post("sets.update", `{"name":"x","set_category":"consumable"}`, func(w http.ResponseWriter, r *http.Request) { sh.UpdateSet(w, r, id) })
	fail("sets.delete", func(w http.ResponseWriter, r *http.Request) { sh.DeleteSet(w, r, id) })
	fail("sets.listItems", func(w http.ResponseWriter, r *http.Request) { sh.ListSetItems(w, r, id) })
	post("sets.addItem", `{"item_id":"`+rid+`","quantity":1}`, func(w http.ResponseWriter, r *http.Request) { sh.AddSetItem(w, r, id) })
	post("sets.updateItem", `{"quantity":2}`, func(w http.ResponseWriter, r *http.Request) { sh.UpdateSetItem(w, r, id, id) })
	fail("sets.removeItem", func(w http.ResponseWriter, r *http.Request) { sh.RemoveSetItem(w, r, id, id) })

	fail("packingLists.get", func(w http.ResponseWriter, r *http.Request) { plh.GetPackingListById(w, r, id) })
	post("packingLists.update", `{"name":"x"}`, func(w http.ResponseWriter, r *http.Request) { plh.UpdatePackingList(w, r, id) })
	fail("packingLists.delete", func(w http.ResponseWriter, r *http.Request) { plh.DeletePackingList(w, r, id) })
	fail("packingLists.listLabels", func(w http.ResponseWriter, r *http.Request) { plh.ListPackingListLabels(w, r, id) })
	post("packingLists.addLabel", `{"label_id":"`+rid+`"}`, func(w http.ResponseWriter, r *http.Request) { plh.AddPackingListLabel(w, r, id) })
	fail("packingLists.removeLabel", func(w http.ResponseWriter, r *http.Request) { plh.RemovePackingListLabel(w, r, id, id) })

	fail("trips.get", func(w http.ResponseWriter, r *http.Request) { th.GetTripById(w, r, id) })
	post("trips.update", `{"name":"x"}`, func(w http.ResponseWriter, r *http.Request) { th.UpdateTrip(w, r, id) })
	fail("trips.delete", func(w http.ResponseWriter, r *http.Request) { th.DeleteTrip(w, r, id) })
	post("trips.addPerson", `{"person_id":"`+rid+`"}`, func(w http.ResponseWriter, r *http.Request) { th.AddTripPerson(w, r, id) })
	fail("trips.removePerson", func(w http.ResponseWriter, r *http.Request) { th.RemoveTripPerson(w, r, id, id) })
	post("trips.addPersonItem", `{"item_id":"`+rid+`","quantity":1,"carry_status":"worn"}`, func(w http.ResponseWriter, r *http.Request) { th.AddTripPersonItem(w, r, id, id) })
	post("trips.updatePersonItem", `{"quantity":2}`, func(w http.ResponseWriter, r *http.Request) { th.UpdateTripPersonItem(w, r, id, id, id) })
	fail("trips.removePersonItem", func(w http.ResponseWriter, r *http.Request) { th.RemoveTripPersonItem(w, r, id, id, id) })
	post("trips.addPersonPack", `{"name":"p","trip_type":"overnight"}`, func(w http.ResponseWriter, r *http.Request) { th.AddTripPersonPack(w, r, id, id) })
	fail("trips.removePersonPack", func(w http.ResponseWriter, r *http.Request) { th.RemoveTripPersonPack(w, r, id, id, id) })
	post("trips.addPackItem", `{"item_id":"`+rid+`","quantity":1,"carry_status":"packed"}`, func(w http.ResponseWriter, r *http.Request) { th.AddTripPersonPackItem(w, r, id, id, id) })
	post("trips.updatePackItem", `{"quantity":2}`, func(w http.ResponseWriter, r *http.Request) { th.UpdateTripPersonPackItem(w, r, id, id, id, id) })
	fail("trips.removePackItem", func(w http.ResponseWriter, r *http.Request) { th.RemoveTripPersonPackItem(w, r, id, id, id, id) })

	fail("itemTypes.get", func(w http.ResponseWriter, r *http.Request) { ith.GetItemType(w, r, "t") })
	post("itemTypes.update", `{"name":"x"}`, func(w http.ResponseWriter, r *http.Request) { ith.UpdateItemType(w, r, "t") })
	fail("itemTypes.delete", func(w http.ResponseWriter, r *http.Request) { ith.DeleteItemType(w, r, "t") })
	fail("itemTypes.listFields", func(w http.ResponseWriter, r *http.Request) { ith.ListItemTypeFields(w, r, "t") })
	post("itemTypes.replaceFields", `{"fields":[]}`, func(w http.ResponseWriter, r *http.Request) { ith.ReplaceItemTypeFields(w, r, "t") })

	// Backup handlers with the DB down.
	bh := NewBackupHandler(backup.NewService(dbConn, t.TempDir()), st, backup.NewScheduler(backup.NewService(dbConn, t.TempDir()), st.BackupConfig), "pw")
	fail("backup.getConfig", bh.GetBackupConfig)
	post("backup.updateConfig", `{"enabled":false,"schedule":"0 3 * * *","retention_count":7}`, bh.UpdateBackupConfig)
	fail("backup.run", bh.RunBackup)

	fail("settings.get", NewSettingsHandler(st, "pw").GetSettings)

	sw := httptest.NewRecorder()
	NewSearchHandler(st).SearchGlobal(sw,
		httptest.NewRequest(http.MethodGet, "/api/v1/search?q=test", http.NoBody),
		api.SearchGlobalParams{Q: "test"})
	if sw.Code < 500 {
		t.Errorf("search: expected 5xx with DB down, got %d", sw.Code)
	}
}
