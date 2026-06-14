package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/via-justa/overpacked-app/backend/internal/api"
)

func TestSearchHandlerSearchGlobalQueryTooShort(t *testing.T) {
	t.Parallel()

	h := NewSearchHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=a", http.NoBody)
	w := httptest.NewRecorder()

	h.SearchGlobal(w, req, api.SearchGlobalParams{Q: "a"})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for short query, got %d", w.Code)
	}
}

func TestSearchHandlerSearchGlobalQueryWhitespace(t *testing.T) {
	t.Parallel()

	h := NewSearchHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", http.NoBody)
	w := httptest.NewRecorder()

	h.SearchGlobal(w, req, api.SearchGlobalParams{Q: "   "})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for whitespace query, got %d", w.Code)
	}
}

func TestSearchHandlerSearchGlobalInvalidType(t *testing.T) {
	t.Parallel()

	h := NewSearchHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=tent&types=bogus", http.NoBody)
	w := httptest.NewRecorder()

	h.SearchGlobal(w, req, api.SearchGlobalParams{
		Q:     "tent",
		Types: &[]api.SearchEntityType{"bogus"},
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid type filter, got %d", w.Code)
	}
}
