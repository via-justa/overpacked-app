package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/via-justa/overpacked-app/backend/internal/api"
	"github.com/via-justa/overpacked-app/backend/internal/db"
	"github.com/via-justa/overpacked-app/backend/internal/migrations"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

// Unit tests for invalid request bodies

func TestTripsHandlerCreateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewTripsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.CreateTrip(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestTripsHandlerUpdateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewTripsHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/trips/"+uuid.NewString(), bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.UpdateTrip(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestTripsHandlerAddPersonPackInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewTripsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+uuid.NewString()+"/persons/"+uuid.NewString()+"/packs", bytes.NewReader([]byte(`{"pack_id":`)))
	w := httptest.NewRecorder()

	h.AddTripPersonPack(w, req, types.UUID(uuid.New()), types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestTripsHandlerAddPersonItemInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewTripsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+uuid.NewString()+"/persons/"+uuid.NewString()+"/items", bytes.NewReader([]byte(`{"item_id":`)))
	w := httptest.NewRecorder()

	h.AddTripPersonItem(w, req, types.UUID(uuid.New()), types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestTripsHandlerUpdatePersonItemInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewTripsHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/trips/"+uuid.NewString()+"/persons/"+uuid.NewString()+"/items/"+uuid.NewString(), bytes.NewReader([]byte(`{"quantity":`)))
	w := httptest.NewRecorder()

	h.UpdateTripPersonItem(w, req, types.UUID(uuid.New()), types.UUID(uuid.New()), types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestTripsHandlerAddPersonPackItemInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewTripsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+uuid.NewString()+"/persons/"+uuid.NewString()+"/packs/"+uuid.NewString()+"/items", bytes.NewReader([]byte(`{"item_id":`)))
	w := httptest.NewRecorder()

	h.AddTripPersonPackItem(w, req, types.UUID(uuid.New()), types.UUID(uuid.New()), types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

// Integration tests

var (
	tripsMigrationsOnce sync.Once
	tripsMigrationsErr  error
)

func newContainerizedTripsHandler(t *testing.T) (*TripsHandler, *sql.DB) {
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

	tripsMigrationsOnce.Do(func() {
		tripsMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if tripsMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", tripsMigrationsErr)
	}

	// Truncate user data tables (avoid system tables like item_types)
	truncateTables := []string{
		"packing_list_labels", "packing_lists",
		"trip_person_items", "trip_person_packs", "trip_persons", "trips",
		"pack_items", "packs",
		"set_items", "item_sets",
		"item_labels", "labels",
		"items",
		"persons",
		"manufacturers",
	}
	for _, table := range truncateTables {
		if _, err := dbConn.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)); err != nil {
			_ = dbConn.Close()
			t.Fatalf("truncate table %s: %v", table, err)
		}
	}

	return NewTripsHandler(store.New(dbConn)), dbConn
}

func insertTripTestData(t *testing.T, dbConn *sql.DB) (manufacturerID, personID, itemID, setID, packID types.UUID) {
	t.Helper()

	// Insert manufacturer
	manufacturerID = types.UUID(uuid.New())
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO manufacturers (id, name) VALUES ($1, $2)", manufacturerID, "Test Manufacturer"); err != nil {
		t.Fatalf("insert manufacturer: %v", err)
	}

	// Insert person
	personID = types.UUID(uuid.New())
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO persons (id, name) VALUES ($1, $2)", personID, "Test Person"); err != nil {
		t.Fatalf("insert person: %v", err)
	}

	// Insert item type. item_types is intentionally not truncated between tests
	// (it can hold shared/system rows), so this fixed id is inserted idempotently.
	itemTypeID := "test-item-type"
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO item_types (id, name, base_profile) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING", itemTypeID, "Test Type", "shelter"); err != nil {
		t.Fatalf("insert item type: %v", err)
	}

	// Insert item
	itemID = types.UUID(uuid.New())
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO items (id, manufacturer_id, type_id, name, is_active) VALUES ($1, $2, $3, $4, $5)", itemID, manufacturerID, itemTypeID, "Test Item", true); err != nil {
		t.Fatalf("insert item: %v", err)
	}

	// Insert set
	setID = types.UUID(uuid.New())
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO item_sets (id, name, set_category, is_active) VALUES ($1, $2, $3, $4)", setID, "Test Set", itemTypeID, true); err != nil {
		t.Fatalf("insert set: %v", err)
	}

	// Insert pack
	packID = types.UUID(uuid.New())
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO packs (id, person_id, name, trip_type) VALUES ($1, $2, $3, $4)", packID, personID, "Test Pack", "overnight"); err != nil {
		t.Fatalf("insert pack: %v", err)
	}

	return
}

func mustCreateTrip(t *testing.T, h *TripsHandler) api.Trip {
	t.Helper()

	reqBody := api.TripCreate{
		Name:     "Integration Test Trip",
		TripType: api.TripCreateTripTypeOvernight,
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", bytes.NewReader(reqBytes))
	w := httptest.NewRecorder()

	h.CreateTrip(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for trip create, got %d: %s", w.Code, w.Body.String())
	}

	var trip api.Trip
	if err := json.NewDecoder(w.Body).Decode(&trip); err != nil {
		t.Fatalf("decode trip response: %v", err)
	}

	return trip
}

func TestTripsHandlerIntegrationCRUD(t *testing.T) {
	h, dbConn := newContainerizedTripsHandler(t)
	defer func() { _ = dbConn.Close() }()

	// Create trip
	trip := mustCreateTrip(t, h)

	if trip.Name != "Integration Test Trip" {
		t.Errorf("expected name 'Integration Test Trip', got %s", trip.Name)
	}
	if trip.TripType != api.TripTripTypeOvernight {
		t.Errorf("expected trip type overnight, got %s", trip.TripType)
	}

	// Get trip
	req := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String(), nil)
	w := httptest.NewRecorder()
	h.GetTripById(w, req, trip.Id)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for trip get, got %d", w.Code)
	}

	var gotTrip api.Trip
	if err := json.NewDecoder(w.Body).Decode(&gotTrip); err != nil {
		t.Fatalf("decode trip response: %v", err)
	}

	if gotTrip.Id != trip.Id {
		t.Errorf("expected trip ID %s, got %s", trip.Id, gotTrip.Id)
	}

	// Update trip
	distanceKm := float32(15.5)
	updateReqBody := api.TripUpdate{
		Name:            strPtr("Updated Trip Name"),
		TotalDistanceKm: &distanceKm,
	}
	updateReqBytes, _ := json.Marshal(updateReqBody)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/trips/"+trip.Id.String(), bytes.NewReader(updateReqBytes))
	updateW := httptest.NewRecorder()
	h.UpdateTrip(updateW, updateReq, trip.Id)

	if updateW.Code != http.StatusOK {
		t.Fatalf("expected 200 for trip update, got %d: %s", updateW.Code, updateW.Body.String())
	}

	var updatedTrip api.Trip
	if err := json.NewDecoder(updateW.Body).Decode(&updatedTrip); err != nil {
		t.Fatalf("decode updated trip response: %v", err)
	}

	if updatedTrip.Name != "Updated Trip Name" {
		t.Errorf("expected name 'Updated Trip Name', got %s", updatedTrip.Name)
	}
	if updatedTrip.TotalDistanceKm == nil || *updatedTrip.TotalDistanceKm != distanceKm {
		t.Errorf("expected distance %f, got %v", distanceKm, updatedTrip.TotalDistanceKm)
	}

	// List trips
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/trips", nil)
	listW := httptest.NewRecorder()
	h.ListTrips(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 for trips list, got %d", listW.Code)
	}

	var trips []api.Trip
	if err := json.NewDecoder(listW.Body).Decode(&trips); err != nil {
		t.Fatalf("decode trips list response: %v", err)
	}

	if len(trips) != 1 {
		t.Errorf("expected 1 trip, got %d", len(trips))
	}

	// Delete trip
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/trips/"+trip.Id.String(), nil)
	deleteW := httptest.NewRecorder()
	h.DeleteTrip(deleteW, deleteReq, trip.Id)

	if deleteW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for trip delete, got %d", deleteW.Code)
	}

	// Verify deletion
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String(), nil)
	getW := httptest.NewRecorder()
	h.GetTripById(getW, getReq, trip.Id)

	if getW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for deleted trip, got %d", getW.Code)
	}
}

func TestTripsHandlerIntegrationNestedGet(t *testing.T) {
	h, dbConn := newContainerizedTripsHandler(t)
	defer func() { _ = dbConn.Close() }()

	_, personID, itemID, _, _ := insertTripTestData(t, dbConn)
	trip := mustCreateTrip(t, h)

	// Add person to trip
	addPersonReqBody := api.TripPersonCreate{PersonId: personID}
	addPersonReqBytes, _ := json.Marshal(addPersonReqBody)
	addPersonReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons", bytes.NewReader(addPersonReqBytes))
	addPersonW := httptest.NewRecorder()
	h.AddTripPerson(addPersonW, addPersonReq, trip.Id)

	if addPersonW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add person, got %d: %s", addPersonW.Code, addPersonW.Body.String())
	}

	// Create pack for person. The API mints a new pack; use the returned id.
	addPackReqBody := api.TripPersonPackCreate{
		Name:     "Test Pack",
		TripType: "overnight",
	}
	addPackReqBytes, _ := json.Marshal(addPackReqBody)
	addPackReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/packs", bytes.NewReader(addPackReqBytes))
	addPackW := httptest.NewRecorder()
	h.AddTripPersonPack(addPackW, addPackReq, trip.Id, personID)

	if addPackW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add pack, got %d: %s", addPackW.Code, addPackW.Body.String())
	}

	var createdPack api.TripPersonPackWithDetails
	if err := json.NewDecoder(addPackW.Body).Decode(&createdPack); err != nil {
		t.Fatalf("decode created pack: %v", err)
	}
	packID := createdPack.PackId

	// Add item to pack
	addPackItemReqBody := api.PackItemCreate{
		ItemId:      itemID,
		Quantity:    1.0,
		CarryStatus: api.PackItemCreateCarryStatusPacked,
	}
	addPackItemReqBytes, _ := json.Marshal(addPackItemReqBody)
	addPackItemReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/packs/"+packID.String()+"/items", bytes.NewReader(addPackItemReqBytes))
	addPackItemW := httptest.NewRecorder()
	h.AddTripPersonPackItem(addPackItemW, addPackItemReq, trip.Id, personID, packID)

	if addPackItemW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add pack item, got %d: %s", addPackItemW.Code, addPackItemW.Body.String())
	}

	// Add item directly to person
	addPersonItemReqBody := api.TripPersonItemCreate{
		ItemId:      itemID,
		Quantity:    1,
		CarryStatus: api.TripPersonItemCreateCarryStatusWorn,
	}
	addPersonItemReqBytes, _ := json.Marshal(addPersonItemReqBody)
	addPersonItemReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/items", bytes.NewReader(addPersonItemReqBytes))
	addPersonItemW := httptest.NewRecorder()
	h.AddTripPersonItem(addPersonItemW, addPersonItemReq, trip.Id, personID)

	if addPersonItemW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add person item, got %d: %s", addPersonItemW.Code, addPersonItemW.Body.String())
	}

	// Get full trip with nested data
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String(), nil)
	getW := httptest.NewRecorder()
	h.GetTripById(getW, getReq, trip.Id)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 for get trip, got %d", getW.Code)
	}

	var tripDetails api.TripWithDetails
	if err := json.NewDecoder(getW.Body).Decode(&tripDetails); err != nil {
		t.Fatalf("decode trip details: %v", err)
	}

	// Verify nested structure
	if len(tripDetails.Persons) != 1 {
		t.Fatalf("expected 1 person, got %d", len(tripDetails.Persons))
	}

	person := tripDetails.Persons[0]
	if person.PersonId != personID {
		t.Errorf("expected person ID %s, got %s", personID, person.PersonId)
	}

	if len(person.Packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(person.Packs))
	}

	pack := person.Packs[0]
	if pack.PackId != packID {
		t.Errorf("expected pack ID %s, got %s", packID, pack.PackId)
	}

	if len(pack.Items) != 1 {
		t.Fatalf("expected 1 item in pack, got %d", len(pack.Items))
	}

	packItem := pack.Items[0]
	if packItem.ItemId != itemID {
		t.Errorf("expected pack item ID %s, got %s", itemID, packItem.ItemId)
	}

	if len(person.Items) != 1 {
		t.Fatalf("expected 1 item on person, got %d", len(person.Items))
	}

	personItem := person.Items[0]
	if personItem.ItemId != itemID {
		t.Errorf("expected person item ID %s, got %s", itemID, personItem.ItemId)
	}
	if personItem.CarryStatus != api.Worn {
		t.Errorf("expected carry status worn, got %s", personItem.CarryStatus)
	}
}

func TestTripsHandlerIntegrationTripPersonPacks(t *testing.T) {
	h, dbConn := newContainerizedTripsHandler(t)
	defer func() { _ = dbConn.Close() }()

	_, personID, _, _, _ := insertTripTestData(t, dbConn)
	trip := mustCreateTrip(t, h)

	// Add person to trip
	addPersonReqBody := api.TripPersonCreate{PersonId: personID}
	addPersonReqBytes, _ := json.Marshal(addPersonReqBody)
	addPersonReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons", bytes.NewReader(addPersonReqBytes))
	addPersonW := httptest.NewRecorder()
	h.AddTripPerson(addPersonW, addPersonReq, trip.Id)

	if addPersonW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add person, got %d: %s", addPersonW.Code, addPersonW.Body.String())
	}

	// Create pack for person. The API mints a new pack; use the returned id.
	addPackReqBody := api.TripPersonPackCreate{
		Name:     "Test Pack",
		TripType: "overnight",
	}
	addPackReqBytes, _ := json.Marshal(addPackReqBody)
	addPackReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/packs", bytes.NewReader(addPackReqBytes))
	addPackW := httptest.NewRecorder()
	h.AddTripPersonPack(addPackW, addPackReq, trip.Id, personID)

	if addPackW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add pack, got %d: %s", addPackW.Code, addPackW.Body.String())
	}

	var packDetails api.TripPersonPackWithDetails
	if err := json.NewDecoder(addPackW.Body).Decode(&packDetails); err != nil {
		t.Fatalf("decode pack details: %v", err)
	}

	packID := packDetails.PackId
	if packDetails.Pack.Name != "Test Pack" {
		t.Errorf("expected created pack name 'Test Pack', got %q", packDetails.Pack.Name)
	}

	// Verify pack in nested GET
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String(), nil)
	getW := httptest.NewRecorder()
	h.GetTripById(getW, getReq, trip.Id)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 for get trip, got %d", getW.Code)
	}

	var tripDetails api.TripWithDetails
	if err := json.NewDecoder(getW.Body).Decode(&tripDetails); err != nil {
		t.Fatalf("decode trip details: %v", err)
	}

	if len(tripDetails.Persons) != 1 || len(tripDetails.Persons[0].Packs) != 1 {
		t.Errorf("expected 1 person with 1 pack")
	}

	// Remove pack
	removeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/packs/"+packID.String(), nil)
	removeW := httptest.NewRecorder()
	h.RemoveTripPersonPack(removeW, removeReq, trip.Id, personID, packID)

	if removeW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for remove pack, got %d", removeW.Code)
	}

	// Verify removal
	getReq2 := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String(), nil)
	getW2 := httptest.NewRecorder()
	h.GetTripById(getW2, getReq2, trip.Id)

	if getW2.Code != http.StatusOK {
		t.Fatalf("expected 200 for get trip, got %d", getW2.Code)
	}

	var tripDetails2 api.TripWithDetails
	if err := json.NewDecoder(getW2.Body).Decode(&tripDetails2); err != nil {
		t.Fatalf("decode trip details: %v", err)
	}

	if len(tripDetails2.Persons) != 1 || len(tripDetails2.Persons[0].Packs) != 0 {
		t.Errorf("expected 1 person with 0 packs after removal")
	}
}

func TestTripsHandlerIntegrationTripPersonItems(t *testing.T) {
	h, dbConn := newContainerizedTripsHandler(t)
	defer func() { _ = dbConn.Close() }()

	_, personID, itemID, _, _ := insertTripTestData(t, dbConn)
	trip := mustCreateTrip(t, h)

	// Add person to trip
	addPersonReqBody := api.TripPersonCreate{PersonId: personID}
	addPersonReqBytes, _ := json.Marshal(addPersonReqBody)
	addPersonReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons", bytes.NewReader(addPersonReqBytes))
	addPersonW := httptest.NewRecorder()
	h.AddTripPerson(addPersonW, addPersonReq, trip.Id)

	if addPersonW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add person, got %d: %s", addPersonW.Code, addPersonW.Body.String())
	}

	// Add item to person
	notes := "Test item notes"
	addItemReqBody := api.TripPersonItemCreate{
		ItemId:      itemID,
		Quantity:    2,
		CarryStatus: api.TripPersonItemCreateCarryStatusWorn,
		Notes:       &notes,
	}
	addItemReqBytes, _ := json.Marshal(addItemReqBody)
	addItemReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/items", bytes.NewReader(addItemReqBytes))
	addItemW := httptest.NewRecorder()
	h.AddTripPersonItem(addItemW, addItemReq, trip.Id, personID)

	if addItemW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add item, got %d: %s", addItemW.Code, addItemW.Body.String())
	}

	var itemDetails api.TripPersonItemWithDetails
	if err := json.NewDecoder(addItemW.Body).Decode(&itemDetails); err != nil {
		t.Fatalf("decode item details: %v", err)
	}

	if itemDetails.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", itemDetails.Quantity)
	}
	if itemDetails.CarryStatus != api.Worn {
		t.Errorf("expected carry status worn, got %s", itemDetails.CarryStatus)
	}

	// Update item
	newQuantity := 3
	updateItemReqBody := api.TripPersonItemUpdate{Quantity: &newQuantity}
	updateItemReqBytes, _ := json.Marshal(updateItemReqBody)
	updateItemReq := httptest.NewRequest(http.MethodPatch, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/items/"+itemDetails.Id.String(), bytes.NewReader(updateItemReqBytes))
	updateItemW := httptest.NewRecorder()
	h.UpdateTripPersonItem(updateItemW, updateItemReq, trip.Id, personID, itemDetails.Id)

	if updateItemW.Code != http.StatusOK {
		t.Fatalf("expected 200 for update item, got %d: %s", updateItemW.Code, updateItemW.Body.String())
	}

	var updatedItem api.TripPersonItemWithDetails
	if err := json.NewDecoder(updateItemW.Body).Decode(&updatedItem); err != nil {
		t.Fatalf("decode updated item: %v", err)
	}

	if updatedItem.Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", updatedItem.Quantity)
	}

	// Remove item
	removeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/items/"+itemDetails.Id.String(), nil)
	removeW := httptest.NewRecorder()
	h.RemoveTripPersonItem(removeW, removeReq, trip.Id, personID, itemDetails.Id)

	if removeW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for remove item, got %d", removeW.Code)
	}
}

func TestTripsHandlerIntegrationTripPersonPackItems(t *testing.T) {
	h, dbConn := newContainerizedTripsHandler(t)
	defer func() { _ = dbConn.Close() }()

	_, personID, itemID, _, _ := insertTripTestData(t, dbConn)
	trip := mustCreateTrip(t, h)

	// Add person and pack
	addPersonReqBody := api.TripPersonCreate{PersonId: personID}
	addPersonReqBytes, _ := json.Marshal(addPersonReqBody)
	addPersonReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons", bytes.NewReader(addPersonReqBytes))
	addPersonW := httptest.NewRecorder()
	h.AddTripPerson(addPersonW, addPersonReq, trip.Id)

	if addPersonW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add person, got %d", addPersonW.Code)
	}

	addPackReqBody := api.TripPersonPackCreate{
		Name:     "Test Pack",
		TripType: "overnight",
	}
	addPackReqBytes, _ := json.Marshal(addPackReqBody)
	addPackReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/packs", bytes.NewReader(addPackReqBytes))
	addPackW := httptest.NewRecorder()
	h.AddTripPersonPack(addPackW, addPackReq, trip.Id, personID)

	if addPackW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add pack, got %d", addPackW.Code)
	}

	var packDetails api.TripPersonPackWithDetails
	if err := json.NewDecoder(addPackW.Body).Decode(&packDetails); err != nil {
		t.Fatalf("decode pack details: %v", err)
	}
	packID := packDetails.PackId

	// Add item to pack
	addItemReqBody := api.PackItemCreate{
		ItemId:      itemID,
		Quantity:    2.0,
		CarryStatus: api.PackItemCreateCarryStatusPacked,
	}
	addItemReqBytes, _ := json.Marshal(addItemReqBody)
	addItemReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/packs/"+packID.String()+"/items", bytes.NewReader(addItemReqBytes))
	addItemW := httptest.NewRecorder()
	h.AddTripPersonPackItem(addItemW, addItemReq, trip.Id, personID, packID)

	if addItemW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add pack item, got %d: %s", addItemW.Code, addItemW.Body.String())
	}

	var itemDetails api.PackItemWithDetails
	if err := json.NewDecoder(addItemW.Body).Decode(&itemDetails); err != nil {
		t.Fatalf("decode item details: %v", err)
	}

	if itemDetails.Quantity != 2.0 {
		t.Errorf("expected quantity 2, got %f", itemDetails.Quantity)
	}

	// Update pack item
	newQuantity := float32(3.0)
	updateItemReqBody := api.PackItemUpdate{Quantity: &newQuantity}
	updateItemReqBytes, _ := json.Marshal(updateItemReqBody)
	updateItemReq := httptest.NewRequest(http.MethodPatch, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/packs/"+packID.String()+"/items/"+itemID.String(), bytes.NewReader(updateItemReqBytes))
	updateItemW := httptest.NewRecorder()
	h.UpdateTripPersonPackItem(updateItemW, updateItemReq, trip.Id, personID, packID, itemID)

	if updateItemW.Code != http.StatusOK {
		t.Fatalf("expected 200 for update pack item, got %d: %s", updateItemW.Code, updateItemW.Body.String())
	}

	var updatedItem api.PackItemWithDetails
	if err := json.NewDecoder(updateItemW.Body).Decode(&updatedItem); err != nil {
		t.Fatalf("decode updated item: %v", err)
	}

	if updatedItem.Quantity != 3.0 {
		t.Errorf("expected quantity 3, got %f", updatedItem.Quantity)
	}

	// Remove pack item
	removeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String()+"/packs/"+packID.String()+"/items/"+itemID.String(), nil)
	removeW := httptest.NewRecorder()
	h.RemoveTripPersonPackItem(removeW, removeReq, trip.Id, personID, packID, itemID)

	if removeW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for remove pack item, got %d", removeW.Code)
	}
}

func TestTripsHandlerIntegrationTripPersons(t *testing.T) {
	h, dbConn := newContainerizedTripsHandler(t)
	defer func() { _ = dbConn.Close() }()

	_, personID, _, _, _ := insertTripTestData(t, dbConn)
	trip := mustCreateTrip(t, h)

	// Add person to trip
	addPersonReqBody := api.TripPersonCreate{PersonId: personID}
	addPersonReqBytes, _ := json.Marshal(addPersonReqBody)
	addPersonReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/persons", bytes.NewReader(addPersonReqBytes))
	addPersonW := httptest.NewRecorder()
	h.AddTripPerson(addPersonW, addPersonReq, trip.Id)

	if addPersonW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add person, got %d: %s", addPersonW.Code, addPersonW.Body.String())
	}

	// Verify person was added by getting full trip details
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String(), nil)
	getW := httptest.NewRecorder()
	h.GetTripById(getW, getReq, trip.Id)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 for get trip, got %d", getW.Code)
	}

	var tripDetails api.TripWithDetails
	if err := json.NewDecoder(getW.Body).Decode(&tripDetails); err != nil {
		t.Fatalf("decode trip details: %v", err)
	}

	if len(tripDetails.Persons) != 1 {
		t.Errorf("expected 1 person in trip, got %d", len(tripDetails.Persons))
	}

	// Remove person from trip
	removeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/trips/"+trip.Id.String()+"/persons/"+personID.String(), nil)
	removeW := httptest.NewRecorder()
	h.RemoveTripPerson(removeW, removeReq, trip.Id, personID)

	if removeW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for remove person, got %d", removeW.Code)
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
