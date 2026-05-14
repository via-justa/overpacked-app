package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

func TestPersonsHandlerCreatePersonInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewPersonsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/persons", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.CreatePerson(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestPersonsHandlerUpdatePersonInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewPersonsHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/persons/"+uuid.NewString(), bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.UpdatePerson(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}
