package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
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
	"github.com/via-justa/overpacked-app/backend/internal/storage"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

func TestItemsHandlerCreateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewItemsHandler(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.CreateItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestItemsHandlerUpdateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewItemsHandler(nil, nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/items/"+uuid.NewString(), bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.UpdateItem(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

var (
	itemsMigrationsOnce sync.Once
	itemsMigrationsErr  error
)

func newContainerizedItemsHandler(t *testing.T) (*ItemsHandler, *sql.DB) {
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

	itemsMigrationsOnce.Do(func() {
		itemsMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if itemsMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", itemsMigrationsErr)
	}

	if _, err := dbConn.ExecContext(context.Background(), "TRUNCATE TABLE pack_items, set_items, items, manufacturers RESTART IDENTITY CASCADE"); err != nil {
		_ = dbConn.Close()
		t.Fatalf("truncate item-related tables: %v", err)
	}

	images, err := storage.New(t.TempDir())
	if err != nil {
		_ = dbConn.Close()
		t.Fatalf("init image storage: %v", err)
	}

	return NewItemsHandler(store.New(dbConn), images), dbConn
}

func insertManufacturer(t *testing.T, dbConn *sql.DB, name string) types.UUID {
	t.Helper()

	id := uuid.New()
	if _, err := dbConn.ExecContext(context.Background(), "INSERT INTO manufacturers (id, name) VALUES ($1, $2)", id, name); err != nil {
		t.Fatalf("insert manufacturer: %v", err)
	}
	return types.UUID(id)
}

func TestItemsHandlerIntegrationCRUD(t *testing.T) {
	h, dbConn := newContainerizedItemsHandler(t)
	defer func() { _ = dbConn.Close() }()

	manufacturerID := insertManufacturer(t, dbConn, "Acme")
	weight := float32(500)
	volume := float32(250)
	value := float32(149.99)
	createdConsumable := mustCreateConsumableItem(t, h, manufacturerID, weight, volume, value)
	mustGetItemOK(t, h, createdConsumable.Id)

	updatedName := "Fuel Canister Updated"
	updatedValue := float32(189.5)
	mustUpdateItemValue(t, h, createdConsumable.Id, updatedName, updatedValue)
	mustListContainUpdatedItem(t, h, updatedName, updatedValue)
	mustDeleteItem(t, h, createdConsumable.Id)
	mustGetItemNotFound(t, h, createdConsumable.Id)
}

func mustCreateConsumableItem(t *testing.T, h *ItemsHandler, manufacturerID types.UUID, weight, volume, value float32) api.Item {
	t.Helper()

	createBody, err := json.Marshal(api.ItemCreate{
		ManufacturerId: manufacturerID,
		Type:           "consumable",
		Name:           "Fuel Canister",
		IsActive:       true,
		Value:          &value,
		WeightGrams:    &weight,
		VolumeMl:       &volume,
	})
	if err != nil {
		t.Fatalf("marshal create item body: %v", err)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	h.CreateItem(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create item, got %d", createW.Code)
	}

	var created api.Item
	if err := json.NewDecoder(createW.Body).Decode(&created); err != nil {
		t.Fatalf("decode create item response: %v", err)
	}

	if created.Name != "Fuel Canister" {
		t.Fatalf("expected created item name Fuel Canister, got %q", created.Name)
	}
	if created.WeightGrams == nil || *created.WeightGrams != 500 {
		t.Fatalf("expected created item weight 500, got %+v", created.WeightGrams)
	}
	if created.Value == nil || *created.Value != value {
		t.Fatalf("expected created item value %v, got %+v", value, created.Value)
	}

	return created
}

func mustGetItemOK(t *testing.T, h *ItemsHandler, id types.UUID) {
	t.Helper()
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+id.String(), http.NoBody)
	getW := httptest.NewRecorder()
	h.GetItem(getW, getReq, id)
	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 from get item, got %d", getW.Code)
	}
}

func mustUpdateItemValue(t *testing.T, h *ItemsHandler, id types.UUID, name string, value float32) {
	t.Helper()
	updateBody, err := json.Marshal(api.ItemUpdate{Name: &name, Value: &value})
	if err != nil {
		t.Fatalf("marshal update item body: %v", err)
	}
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/items/"+id.String(), bytes.NewReader(updateBody))
	updateW := httptest.NewRecorder()
	h.UpdateItem(updateW, updateReq, id)
	if updateW.Code != http.StatusOK {
		t.Fatalf("expected 200 from update item, got %d", updateW.Code)
	}
}

func mustListContainUpdatedItem(t *testing.T, h *ItemsHandler, expectedName string, expectedValue float32) {
	t.Helper()
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/items", http.NoBody)
	listW := httptest.NewRecorder()
	h.ListItems(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 from list items, got %d", listW.Code)
	}

	var list []api.Item
	if err := json.NewDecoder(listW.Body).Decode(&list); err != nil {
		t.Fatalf("decode items list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected one updated item, got %+v", list)
	}

	if list[0].Name != expectedName {
		t.Fatalf("expected updated item name, got %q", list[0].Name)
	}
	if list[0].Value == nil || *list[0].Value != expectedValue {
		t.Fatalf("expected updated item value %v, got %+v", expectedValue, list[0].Value)
	}
}

func mustDeleteItem(t *testing.T, h *ItemsHandler, id types.UUID) {
	t.Helper()
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+id.String(), http.NoBody)
	deleteW := httptest.NewRecorder()
	h.DeleteItem(deleteW, deleteReq, id)
	if deleteW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from delete item, got %d", deleteW.Code)
	}
}

func mustGetItemNotFound(t *testing.T, h *ItemsHandler, id types.UUID) {
	t.Helper()
	getMissingReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+id.String(), http.NoBody)
	getMissingW := httptest.NewRecorder()
	h.GetItem(getMissingW, getMissingReq, id)
	if getMissingW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for deleted item, got %d", getMissingW.Code)
	}
}

func TestItemsHandlerIntegrationImageFlow(t *testing.T) {
	h, dbConn := newContainerizedItemsHandler(t)
	defer func() { _ = dbConn.Close() }()

	manufacturerID := insertManufacturer(t, dbConn, "Acme")
	weight := float32(500)
	volume := float32(250)
	value := float32(149.99)
	item := mustCreateConsumableItem(t, h, manufacturerID, weight, volume, value)

	png := pngBytes(t, 16, 9)

	// Upload an image and confirm the item now exposes an image_url.
	uploaded := mustUploadItemImage(t, h, item.Id, png)
	if uploaded.ImageUrl == nil || *uploaded.ImageUrl == "" {
		t.Fatalf("expected image_url after upload, got %+v", uploaded.ImageUrl)
	}
	if uploaded.ImageWidthPx == nil || *uploaded.ImageWidthPx != 16 {
		t.Fatalf("expected image width 16, got %+v", uploaded.ImageWidthPx)
	}

	// The image endpoint streams back the exact bytes.
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+item.Id.String()+"/image", http.NoBody)
	getW := httptest.NewRecorder()
	h.GetItemImage(getW, getReq, item.Id)
	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 from get image, got %d", getW.Code)
	}
	if !bytes.Equal(getW.Body.Bytes(), png) {
		t.Fatalf("served image bytes differ from uploaded bytes")
	}
	if ct := getW.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected image/png content type, got %q", ct)
	}

	// Delete the image; the endpoint then reports no image.
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+item.Id.String()+"/image", http.NoBody)
	delW := httptest.NewRecorder()
	h.DeleteItemImage(delW, delReq, item.Id)
	if delW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from delete image, got %d", delW.Code)
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+item.Id.String()+"/image", http.NoBody)
	missingW := httptest.NewRecorder()
	h.GetItemImage(missingW, missingReq, item.Id)
	if missingW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after image deletion, got %d", missingW.Code)
	}
}

func mustUploadItemImage(t *testing.T, h *ItemsHandler, id types.UUID, data []byte) api.Item {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "item.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write image part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/items/"+id.String()+"/image", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	h.UploadItemImage(w, req, id)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from upload image, got %d", w.Code)
	}

	var updated api.Item
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("decode upload image response: %v", err)
	}
	return updated
}

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}
