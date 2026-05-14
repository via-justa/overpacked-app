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

func TestManufacturersHandlerCreateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewManufacturersHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/manufacturers", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.CreateManufacturer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestManufacturersHandlerUpdateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewManufacturersHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/manufacturers/"+uuid.NewString(), bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.UpdateManufacturer(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

var (
	manufacturersMigrationsOnce sync.Once
	manufacturersMigrationsErr  error
)

func newContainerizedManufacturersHandler(t *testing.T) (*ManufacturersHandler, *sql.DB) {
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

	manufacturersMigrationsOnce.Do(func() {
		manufacturersMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if manufacturersMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", manufacturersMigrationsErr)
	}

	if _, err := dbConn.ExecContext(context.Background(), "TRUNCATE TABLE manufacturers RESTART IDENTITY CASCADE"); err != nil {
		_ = dbConn.Close()
		t.Fatalf("truncate manufacturers: %v", err)
	}

	return NewManufacturersHandler(store.New(dbConn)), dbConn
}

func TestManufacturersHandlerIntegrationCRUD(t *testing.T) {
	h, dbConn := newContainerizedManufacturersHandler(t)
	defer func() { _ = dbConn.Close() }()

	createBody := []byte(`{"name":"Acme","website":"https://acme.test"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/manufacturers", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	h.CreateManufacturer(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create manufacturer, got %d", createW.Code)
	}

	var created api.Manufacturer
	if err := json.NewDecoder(createW.Body).Decode(&created); err != nil {
		t.Fatalf("decode create manufacturer response: %v", err)
	}
	if created.Name != "Acme" {
		t.Fatalf("expected created manufacturer name Acme, got %q", created.Name)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/manufacturers/"+created.Id.String(), http.NoBody)
	getW := httptest.NewRecorder()
	h.GetManufacturer(getW, getReq, types.UUID(created.Id))
	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 from get manufacturer, got %d", getW.Code)
	}

	newName := "Acme Updated"
	updateBody, err := json.Marshal(api.ManufacturerUpdate{Name: &newName})
	if err != nil {
		t.Fatalf("marshal update manufacturer body: %v", err)
	}
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/manufacturers/"+created.Id.String(), bytes.NewReader(updateBody))
	updateW := httptest.NewRecorder()
	h.UpdateManufacturer(updateW, updateReq, types.UUID(created.Id))
	if updateW.Code != http.StatusOK {
		t.Fatalf("expected 200 from update manufacturer, got %d", updateW.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/manufacturers", http.NoBody)
	listW := httptest.NewRecorder()
	h.ListManufacturers(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 from list manufacturers, got %d", listW.Code)
	}

	var list []api.Manufacturer
	if err := json.NewDecoder(listW.Body).Decode(&list); err != nil {
		t.Fatalf("decode manufacturers list: %v", err)
	}
	if len(list) != 1 || list[0].Name != "Acme Updated" {
		t.Fatalf("expected one updated manufacturer, got %+v", list)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/manufacturers/"+created.Id.String(), http.NoBody)
	deleteW := httptest.NewRecorder()
	h.DeleteManufacturer(deleteW, deleteReq, types.UUID(created.Id))
	if deleteW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from delete manufacturer, got %d", deleteW.Code)
	}

	getMissingReq := httptest.NewRequest(http.MethodGet, "/api/v1/manufacturers/"+created.Id.String(), http.NoBody)
	getMissingW := httptest.NewRecorder()
	h.GetManufacturer(getMissingW, getMissingReq, types.UUID(created.Id))
	if getMissingW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for deleted manufacturer, got %d", getMissingW.Code)
	}
}
