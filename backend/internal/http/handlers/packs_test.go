package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

func TestPacksHandlerCreateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewPacksHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/packs", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.CreatePack(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestPacksHandlerUpdateInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewPacksHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/packs/"+uuid.NewString(), bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	h.UpdatePack(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestPacksHandlerAddPackItemInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewPacksHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/packs/"+uuid.NewString()+"/items", bytes.NewReader([]byte(`{"item_id":`)))
	w := httptest.NewRecorder()

	h.AddPackItem(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestPacksHandlerUpdatePackItemInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewPacksHandler(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/packs/"+uuid.NewString()+"/items/"+uuid.NewString(), bytes.NewReader([]byte(`{"quantity":`)))
	w := httptest.NewRecorder()

	h.UpdatePackItem(w, req, types.UUID(uuid.New()), types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}

func TestPacksHandlerAddPackSetInvalidBody(t *testing.T) {
	t.Parallel()

	h := NewPacksHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/packs/"+uuid.NewString()+"/sets", bytes.NewReader([]byte(`{"set_id":`)))
	w := httptest.NewRecorder()

	h.AddPackSet(w, req, types.UUID(uuid.New()))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid request body, got %d", w.Code)
	}
}
