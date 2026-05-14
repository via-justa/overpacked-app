package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

func TestSetsHandlerCreateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewSetsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sets", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.CreateSet(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestSetsHandlerUpdateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewSetsHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/sets/"+uuid.NewString(), bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.UpdateSet(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestSetsHandlerAddSetItemInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewSetsHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sets/"+uuid.NewString()+"/items", bytes.NewReader([]byte(`{"item_id":`)))
	w := httptest.NewRecorder()

	h.AddSetItem(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestSetsHandlerUpdateSetItemInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewSetsHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/sets/"+uuid.NewString()+"/items/"+uuid.NewString(), bytes.NewReader([]byte(`{"quantity":`)))
	w := httptest.NewRecorder()

	h.UpdateSetItem(w, req, types.UUID(uuid.New()), types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}
