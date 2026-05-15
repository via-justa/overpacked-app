package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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

func TestTripsHandlerAddTripItemInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewTripsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+uuid.NewString()+"/items", bytes.NewReader([]byte(`{"item_id":`)))
	w := httptest.NewRecorder()

	h.AddTripItem(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestTripsHandlerUpdateTripItemInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewTripsHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/trips/"+uuid.NewString()+"/items/"+uuid.NewString(), bytes.NewReader([]byte(`{"quantity":`)))
	w := httptest.NewRecorder()

	h.UpdateTripItem(w, req, types.UUID(uuid.New()), types.UUID(uuid.New()))

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

	if _, err := dbConn.ExecContext(context.Background(), "TRUNCATE TABLE packing_list_labels, packing_lists, trip_persons, trip_sets, trip_items, trip_packs, trips, pack_items, packs, set_items, item_sets, item_labels, labels, items, persons, manufacturers, item_types RESTART IDENTITY CASCADE"); err != nil {
		_ = dbConn.Close()
		t.Fatalf("truncate trip-related tables: %v", err)
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

	// Insert item
	itemID = types.UUID(uuid.New())
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO items (id, manufacturer_id, type_id, name, is_active) VALUES ($1, $2, $3, $4, $5)", itemID, manufacturerID, "shelter", "Test Item", true); err != nil {
		t.Fatalf("insert item: %v", err)
	}

	// Insert set
	setID = types.UUID(uuid.New())
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO item_sets (id, name, set_category, is_active) VALUES ($1, $2, $3, $4)", setID, "Test Set", "consumable", true); err != nil {
		t.Fatalf("insert set: %v", err)
	}

	// Insert pack
	packID = types.UUID(uuid.New())
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO packs (id, name, trip_type, is_template) VALUES ($1, $2, $3, $4)", packID, "Test Pack", "overnight", false); err != nil {
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

func TestTripsHandlerIntegrationTripPacks(t *testing.T) {
	h, dbConn := newContainerizedTripsHandler(t)
	defer func() { _ = dbConn.Close() }()

	_, _, _, _, packID := insertTripTestData(t, dbConn)
	trip := mustCreateTrip(t, h)

	// Add pack to trip
	addPackReqBody := api.TripPackCreate{PackId: packID}
	addPackReqBytes, _ := json.Marshal(addPackReqBody)
	addPackReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/packs", bytes.NewReader(addPackReqBytes))
	addPackW := httptest.NewRecorder()
	h.AddTripPack(addPackW, addPackReq, trip.Id)

	if addPackW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add pack, got %d: %s", addPackW.Code, addPackW.Body.String())
	}

	// List trip packs
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String()+"/packs", nil)
	listW := httptest.NewRecorder()
	h.ListTripPacks(listW, listReq, trip.Id)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 for list trip packs, got %d", listW.Code)
	}

	var packs []api.TripPackWithDetails
	if err := json.NewDecoder(listW.Body).Decode(&packs); err != nil {
		t.Fatalf("decode packs list: %v", err)
	}

	if len(packs) != 1 {
		t.Errorf("expected 1 pack, got %d", len(packs))
	}
	if packs[0].PackId != packID {
		t.Errorf("expected pack ID %s, got %s", packID, packs[0].PackId)
	}

	// Remove pack from trip
	removeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/trips/"+trip.Id.String()+"/packs/"+packID.String(), nil)
	removeW := httptest.NewRecorder()
	h.RemoveTripPack(removeW, removeReq, trip.Id, packID)

	if removeW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for remove pack, got %d", removeW.Code)
	}

	// Verify removal
	listReq2 := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String()+"/packs", nil)
	listW2 := httptest.NewRecorder()
	h.ListTripPacks(listW2, listReq2, trip.Id)

	if listW2.Code != http.StatusOK {
		t.Fatalf("expected 200 for list trip packs, got %d", listW2.Code)
	}

	var packsAfter []api.TripPackWithDetails
	if err := json.NewDecoder(listW2.Body).Decode(&packsAfter); err != nil {
		t.Fatalf("decode packs list: %v", err)
	}

	if len(packsAfter) != 0 {
		t.Errorf("expected 0 packs after removal, got %d", len(packsAfter))
	}
}

func TestTripsHandlerIntegrationTripItems(t *testing.T) {
	h, dbConn := newContainerizedTripsHandler(t)
	defer func() { _ = dbConn.Close() }()

	_, _, itemID, _, _ := insertTripTestData(t, dbConn)
	trip := mustCreateTrip(t, h)

	// Add item to trip
	notes := "Test item notes"
	addItemReqBody := api.TripItemCreate{
		ItemId:      itemID,
		Quantity:    2.0,
		CarryStatus: api.TripItemCreateCarryStatusPacked,
		Notes:       &notes,
	}
	addItemReqBytes, _ := json.Marshal(addItemReqBody)
	addItemReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/items", bytes.NewReader(addItemReqBytes))
	addItemW := httptest.NewRecorder()
	h.AddTripItem(addItemW, addItemReq, trip.Id)

	if addItemW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add item, got %d: %s", addItemW.Code, addItemW.Body.String())
	}

	// List trip items
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String()+"/items", nil)
	listW := httptest.NewRecorder()
	h.ListTripItems(listW, listReq, trip.Id)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 for list trip items, got %d", listW.Code)
	}

	var items []api.TripItemWithDetails
	if err := json.NewDecoder(listW.Body).Decode(&items); err != nil {
		t.Fatalf("decode items list: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
	if items[0].Quantity != 2.0 {
		t.Errorf("expected quantity 2, got %f", items[0].Quantity)
	}

	// Update trip item
	newQuantity := float32(3.0)
	updateItemReqBody := api.TripItemUpdate{Quantity: &newQuantity}
	updateItemReqBytes, _ := json.Marshal(updateItemReqBody)
	updateItemReq := httptest.NewRequest(http.MethodPatch, "/api/v1/trips/"+trip.Id.String()+"/items/"+itemID.String(), bytes.NewReader(updateItemReqBytes))
	updateItemW := httptest.NewRecorder()
	h.UpdateTripItem(updateItemW, updateItemReq, trip.Id, itemID)

	if updateItemW.Code != http.StatusOK {
		t.Fatalf("expected 200 for update item, got %d: %s", updateItemW.Code, updateItemW.Body.String())
	}

	var updatedItem api.TripItemWithDetails
	if err := json.NewDecoder(updateItemW.Body).Decode(&updatedItem); err != nil {
		t.Fatalf("decode updated item: %v", err)
	}

	if updatedItem.Quantity != newQuantity {
		t.Errorf("expected quantity %f, got %f", newQuantity, updatedItem.Quantity)
	}

	// Remove item from trip
	removeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/trips/"+trip.Id.String()+"/items/"+itemID.String(), nil)
	removeW := httptest.NewRecorder()
	h.RemoveTripItem(removeW, removeReq, trip.Id, itemID)

	if removeW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for remove item, got %d", removeW.Code)
	}
}

func TestTripsHandlerIntegrationTripSets(t *testing.T) {
	h, dbConn := newContainerizedTripsHandler(t)
	defer func() { _ = dbConn.Close() }()

	_, _, _, setID, _ := insertTripTestData(t, dbConn)
	trip := mustCreateTrip(t, h)

	// Add set to trip
	addSetReqBody := api.TripSetCreate{SetId: setID}
	addSetReqBytes, _ := json.Marshal(addSetReqBody)
	addSetReq := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+trip.Id.String()+"/sets", bytes.NewReader(addSetReqBytes))
	addSetW := httptest.NewRecorder()
	h.AddTripSet(addSetW, addSetReq, trip.Id)

	if addSetW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for add set, got %d: %s", addSetW.Code, addSetW.Body.String())
	}

	// List trip sets
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String()+"/sets", nil)
	listW := httptest.NewRecorder()
	h.ListTripSets(listW, listReq, trip.Id)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 for list trip sets, got %d", listW.Code)
	}

	var sets []api.TripSetWithDetails
	if err := json.NewDecoder(listW.Body).Decode(&sets); err != nil {
		t.Fatalf("decode sets list: %v", err)
	}

	if len(sets) != 1 {
		t.Errorf("expected 1 set, got %d", len(sets))
	}

	// Remove set from trip
	removeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/trips/"+trip.Id.String()+"/sets/"+setID.String(), nil)
	removeW := httptest.NewRecorder()
	h.RemoveTripSet(removeW, removeReq, trip.Id, setID)

	if removeW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for remove set, got %d", removeW.Code)
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

	// List trip persons
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+trip.Id.String()+"/persons", nil)
	listW := httptest.NewRecorder()
	h.ListTripPersons(listW, listReq, trip.Id)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 for list trip persons, got %d", listW.Code)
	}

	var persons []api.TripPersonWithDetails
	if err := json.NewDecoder(listW.Body).Decode(&persons); err != nil {
		t.Fatalf("decode persons list: %v", err)
	}

	if len(persons) != 1 {
		t.Errorf("expected 1 person, got %d", len(persons))
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
