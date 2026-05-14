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

var (
	personsMigrationsOnce sync.Once
	personsMigrationsErr  error
)

func newContainerizedPersonsHandler(t *testing.T) (*PersonsHandler, *sql.DB) {
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

	personsMigrationsOnce.Do(func() {
		personsMigrationsErr = migrations.Run(context.Background(), dbConn, "up", nil)
	})
	if personsMigrationsErr != nil {
		_ = dbConn.Close()
		t.Fatalf("run migrations: %v", personsMigrationsErr)
	}

	if _, err := dbConn.ExecContext(context.Background(), "TRUNCATE TABLE packs, persons RESTART IDENTITY CASCADE"); err != nil {
		_ = dbConn.Close()
		t.Fatalf("truncate tables: %v", err)
	}

	st := store.New(dbConn)
	return NewPersonsHandler(st), dbConn
}

func TestPersonsHandlerIntegrationCreateGetDelete(t *testing.T) {
	h, dbConn := newContainerizedPersonsHandler(t)
	defer func() { _ = dbConn.Close() }()

	createReqBody := []byte(`{"name":"Alice","gender":"female","body_weight_grams":62000}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/persons", bytes.NewReader(createReqBody))
	createW := httptest.NewRecorder()

	h.CreatePerson(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create person, got %d", createW.Code)
	}

	var created api.Person
	if err := json.NewDecoder(createW.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Name != "Alice" {
		t.Fatalf("expected created person name Alice, got %q", created.Name)
	}

	getW := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/persons/"+created.Id.String(), http.NoBody)
	h.GetPerson(getW, getReq, types.UUID(created.Id))
	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 from get person, got %d", getW.Code)
	}

	deleteW := httptest.NewRecorder()
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/persons/"+created.Id.String(), http.NoBody)
	h.DeletePerson(deleteW, deleteReq, types.UUID(created.Id))
	if deleteW.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from delete person, got %d", deleteW.Code)
	}

	getAfterDeleteW := httptest.NewRecorder()
	getAfterDeleteReq := httptest.NewRequest(http.MethodGet, "/api/v1/persons/"+created.Id.String(), http.NoBody)
	h.GetPerson(getAfterDeleteW, getAfterDeleteReq, types.UUID(created.Id))
	if getAfterDeleteW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when getting deleted person, got %d", getAfterDeleteW.Code)
	}
}

func TestPersonsHandlerIntegrationList(t *testing.T) {
	h, dbConn := newContainerizedPersonsHandler(t)
	defer func() { _ = dbConn.Close() }()

	for _, name := range []string{"Alice", "Bob"} {
		body := []byte(`{"name":"` + name + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/persons", bytes.NewReader(body))
		w := httptest.NewRecorder()
		h.CreatePerson(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 creating %s, got %d", name, w.Code)
		}
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/persons", http.NoBody)
	listW := httptest.NewRecorder()
	h.ListPersons(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 from list persons, got %d", listW.Code)
	}

	var persons []api.Person
	if err := json.NewDecoder(listW.Body).Decode(&persons); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(persons) != 2 {
		t.Fatalf("expected 2 persons in list, got %d", len(persons))
	}

	seen := map[string]bool{}
	for _, p := range persons {
		seen[p.Name] = true
	}
	if !seen["Alice"] || !seen["Bob"] {
		t.Fatalf("expected list to contain Alice and Bob, got %+v", seen)
	}
}

func TestPersonsHandlerIntegrationUpdate(t *testing.T) {
	h, dbConn := newContainerizedPersonsHandler(t)
	defer func() { _ = dbConn.Close() }()

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/persons", bytes.NewReader([]byte(`{"name":"Before"}`)))
	createW := httptest.NewRecorder()
	h.CreatePerson(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create person, got %d", createW.Code)
	}

	var created api.Person
	if err := json.NewDecoder(createW.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	newName := "After"
	updateReqBody, err := json.Marshal(api.PersonUpdate{Name: &newName})
	if err != nil {
		t.Fatalf("marshal update request: %v", err)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/persons/"+created.Id.String(), bytes.NewReader(updateReqBody))
	updateW := httptest.NewRecorder()
	h.UpdatePerson(updateW, updateReq, types.UUID(created.Id))
	if updateW.Code != http.StatusOK {
		t.Fatalf("expected 200 from update person, got %d", updateW.Code)
	}

	var updated api.Person
	if err := json.NewDecoder(updateW.Body).Decode(&updated); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updated.Name != "After" {
		t.Fatalf("expected updated person name After, got %q", updated.Name)
	}

	notFoundID := types.UUID(uuid.New())
	notFoundReq := httptest.NewRequest(http.MethodPatch, "/api/v1/persons/"+notFoundID.String(), bytes.NewReader(updateReqBody))
	notFoundW := httptest.NewRecorder()
	h.UpdatePerson(notFoundW, notFoundReq, notFoundID)
	if notFoundW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 updating missing person, got %d", notFoundW.Code)
	}
}
