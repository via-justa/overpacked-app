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
	"github.com/via-justa/overpacked-app/backend/internal/domain"
	"github.com/via-justa/overpacked-app/backend/internal/migrations"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

func TestLabelsHandlerCreateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewLabelsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/labels", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.CreateLabel(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestLabelsHandlerUpdateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewLabelsHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/labels/"+uuid.NewString(), bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.UpdateLabel(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestLabelsHandlerAddItemLabelInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewLabelsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items/"+uuid.NewString()+"/labels", bytes.NewReader([]byte(`{"label_id":`)))
	w := httptest.NewRecorder()

	h.AddItemLabel(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

var (
	labelsMigrationsOnce sync.Once
	labelsMigrationsErr  error
)

func newContainerizedLabelsHandler(t *testing.T) (*LabelsHandler, *sql.DB) {
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

	labelsMigrationsOnce.Do(func() {
		labelsMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if labelsMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", labelsMigrationsErr)
	}

	if _, err := dbConn.ExecContext(context.Background(), "TRUNCATE TABLE item_labels, labels, items, manufacturers RESTART IDENTITY CASCADE"); err != nil {
		_ = dbConn.Close()
		t.Fatalf("truncate tables: %v", err)
	}

	return NewLabelsHandler(store.New(dbConn)), dbConn
}

func testCreateLabel(t *testing.T, h *LabelsHandler, name, color string) api.Label {
	t.Helper()
	body := []byte(`{"name":"` + name + `","color":"` + color + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/labels", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.CreateLabel(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create label, got %d: %s", w.Code, w.Body.String())
	}

	var created api.Label
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("decode created label: %v", err)
	}
	return created
}

func testGetLabel(t *testing.T, h *LabelsHandler, labelID types.UUID, expectedStatus int) *api.Label {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/labels/"+uuid.UUID(labelID).String(), nil)
	w := httptest.NewRecorder()
	h.GetLabel(w, req, labelID)

	if w.Code != expectedStatus {
		t.Fatalf("expected %d from get label, got %d", expectedStatus, w.Code)
	}

	if expectedStatus == http.StatusNotFound {
		return nil
	}

	var label api.Label
	if err := json.NewDecoder(w.Body).Decode(&label); err != nil {
		t.Fatalf("decode fetched label: %v", err)
	}
	return &label
}

func testListLabels(t *testing.T, h *LabelsHandler, expectedCount int) []api.Label {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/labels", nil)
	w := httptest.NewRecorder()
	h.ListLabels(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from list labels, got %d", w.Code)
	}

	var labels []api.Label
	if err := json.NewDecoder(w.Body).Decode(&labels); err != nil {
		t.Fatalf("decode labels list: %v", err)
	}

	if len(labels) != expectedCount {
		t.Fatalf("expected %d label(s), got %d", expectedCount, len(labels))
	}
	return labels
}

func testUpdateLabel(t *testing.T, h *LabelsHandler, labelID types.UUID, name, color string) api.Label {
	t.Helper()
	body := []byte(`{"name":"` + name + `","color":"` + color + `"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/labels/"+uuid.UUID(labelID).String(), bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.UpdateLabel(w, req, labelID)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from update label, got %d: %s", w.Code, w.Body.String())
	}

	var updated api.Label
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated label: %v", err)
	}
	return updated
}

func testDeleteLabel(t *testing.T, h *LabelsHandler, labelID types.UUID) {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/labels/"+uuid.UUID(labelID).String(), nil)
	w := httptest.NewRecorder()
	h.DeleteLabel(w, req, labelID)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from delete label, got %d", w.Code)
	}
}

func TestLabelsHandlerIntegrationCRUD(t *testing.T) {
	h, dbConn := newContainerizedLabelsHandler(t)
	defer func() { _ = dbConn.Close() }()

	// Create label
	created := testCreateLabel(t, h, "Ultralight", "#FF5733")
	if created.Name != "Ultralight" {
		t.Errorf("expected name 'Ultralight', got %s", created.Name)
	}
	if created.Color == nil || *created.Color != "#FF5733" {
		t.Errorf("expected color '#FF5733', got %v", created.Color)
	}

	// Get label
	fetched := testGetLabel(t, h, created.Id, http.StatusOK)
	if fetched.Name != "Ultralight" {
		t.Errorf("expected fetched name 'Ultralight', got %s", fetched.Name)
	}

	// List labels
	testListLabels(t, h, 1)

	// Update label
	updated := testUpdateLabel(t, h, created.Id, "Ultra-Lightweight", "#00FF00")
	if updated.Name != "Ultra-Lightweight" {
		t.Errorf("expected updated name 'Ultra-Lightweight', got %s", updated.Name)
	}
	if updated.Color == nil || *updated.Color != "#00FF00" {
		t.Errorf("expected updated color '#00FF00', got %v", updated.Color)
	}

	// Delete label
	testDeleteLabel(t, h, created.Id)

	// Verify deleted
	testGetLabel(t, h, created.Id, http.StatusNotFound)
}

func insertLabelTestData(ctx context.Context, dbConn *sql.DB) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	// Create manufacturer
	var mfrID uuid.UUID
	if err := dbConn.QueryRowContext(ctx, `INSERT INTO manufacturers (name) VALUES ($1) RETURNING id`, "TestMfr").Scan(&mfrID); err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	// Create item
	var itemID uuid.UUID
	if err := dbConn.QueryRowContext(ctx, `
		INSERT INTO items (manufacturer_id, type_id, name, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, mfrID, "shelter", "Test Tent", true).Scan(&itemID); err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	// Create label
	var labelID uuid.UUID
	if err := dbConn.QueryRowContext(ctx, `INSERT INTO labels (name, color) VALUES ($1, $2) RETURNING id`, "Ultralight", "#FF5733").Scan(&labelID); err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	return mfrID, itemID, labelID, nil
}

func TestLabelsHandlerIntegrationItemLabels(t *testing.T) {
	h, dbConn := newContainerizedLabelsHandler(t)
	defer func() { _ = dbConn.Close() }()

	_, itemID, labelID, err := insertLabelTestData(context.Background(), dbConn)
	if err != nil {
		t.Fatalf("insert test data: %v", err)
	}

	// List item labels (should be empty)
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+itemID.String()+"/labels", nil)
	listW := httptest.NewRecorder()
	h.ListItemLabels(listW, listReq, types.UUID(itemID))
	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 from list item labels, got %d", listW.Code)
	}

	var labels []api.Label
	if err := json.NewDecoder(listW.Body).Decode(&labels); err != nil {
		t.Fatalf("decode labels list: %v", err)
	}
	if len(labels) != 0 {
		t.Fatalf("expected 0 labels initially, got %d", len(labels))
	}

	// Add label to item
	addBody := []byte(`{"label_id":"` + labelID.String() + `"}`)
	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/items/"+itemID.String()+"/labels", bytes.NewReader(addBody))
	addW := httptest.NewRecorder()
	h.AddItemLabel(addW, addReq, types.UUID(itemID))
	if addW.Code != http.StatusCreated {
		t.Fatalf("expected 201 from add item label, got %d: %s", addW.Code, addW.Body.String())
	}

	// List item labels (should have 1)
	listReq2 := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+itemID.String()+"/labels", nil)
	listW2 := httptest.NewRecorder()
	h.ListItemLabels(listW2, listReq2, types.UUID(itemID))
	if listW2.Code != http.StatusOK {
		t.Fatalf("expected 200 from list item labels, got %d", listW2.Code)
	}

	var labels2 []api.Label
	if err := json.NewDecoder(listW2.Body).Decode(&labels2); err != nil {
		t.Fatalf("decode labels list: %v", err)
	}
	if len(labels2) != 1 {
		t.Fatalf("expected 1 label after adding, got %d", len(labels2))
	}
	if labels2[0].Name != "Ultralight" {
		t.Errorf("expected label name 'Ultralight', got %s", labels2[0].Name)
	}

	// Remove label from item
	removeReq := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+itemID.String()+"/labels/"+labelID.String(), nil)
	removeW := httptest.NewRecorder()
	h.RemoveItemLabel(removeW, removeReq, types.UUID(itemID), types.UUID(labelID))
	if removeW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from remove item label, got %d", removeW.Code)
	}

	// List item labels (should be empty again)
	listReq3 := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+itemID.String()+"/labels", nil)
	listW3 := httptest.NewRecorder()
	h.ListItemLabels(listW3, listReq3, types.UUID(itemID))
	if listW3.Code != http.StatusOK {
		t.Fatalf("expected 200 from list item labels, got %d", listW3.Code)
	}

	var labels3 []api.Label
	if err := json.NewDecoder(listW3.Body).Decode(&labels3); err != nil {
		t.Fatalf("decode labels list: %v", err)
	}
	if len(labels3) != 0 {
		t.Fatalf("expected 0 labels after removing, got %d", len(labels3))
	}
}

func TestLabelsHandlerIntegrationItemNotFound(t *testing.T) {
	h, dbConn := newContainerizedLabelsHandler(t)
	defer func() { _ = dbConn.Close() }()

	// Test list labels for non-existent item
	fakeItemID := uuid.New()
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+fakeItemID.String()+"/labels", nil)
	listW := httptest.NewRecorder()
	h.ListItemLabels(listW, listReq, types.UUID(fakeItemID))
	if listW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent item, got %d", listW.Code)
	}
}

func TestLabelsHandlerIntegrationLabelNotFoundOnAdd(t *testing.T) {
	h, dbConn := newContainerizedLabelsHandler(t)
	defer func() { _ = dbConn.Close() }()

	// Create item
	st := store.New(dbConn)
	mfr := &domain.Manufacturer{Name: "TestMfr"}
	if err := st.Manufacturers.Create(context.Background(), mfr); err != nil {
		t.Fatalf("create manufacturer: %v", err)
	}
	item := &domain.Item{
		Name:               "Test Item",
		TypeID:             "shelter",
		IsActive:           true,
		ManufacturerID:     mfr.ID,
		DefaultCarryStatus: domain.CarryStatusPacked,
		DefaultQuantity:    1,
	}
	if err := st.Items.Create(context.Background(), item); err != nil {
		t.Fatalf("create item: %v", err)
	}

	// Try to add non-existent label
	fakeLabelID := uuid.New()
	addBody := []byte(`{"label_id":"` + fakeLabelID.String() + `"}`)
	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/items/"+item.ID.String()+"/labels", bytes.NewReader(addBody))
	addW := httptest.NewRecorder()
	h.AddItemLabel(addW, addReq, types.UUID(item.ID))
	if addW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent label, got %d", addW.Code)
	}
}
