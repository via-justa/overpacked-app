package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/via-justa/overpacked-app/backend/internal/auth"
	"github.com/via-justa/overpacked-app/backend/internal/http/handlers"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	authService, err := auth.NewService("admin", "pw123", "test-secret")
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	authHandler := handlers.NewAuthHandler(authService)

	return setupRoutes(authHandler, store.New(nil), "pw123")
}

func TestRoutesHealth(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Body.String(); got != "ok" {
		t.Fatalf("expected body ok, got %q", got)
	}
}

func TestRoutesAuthLogin(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	body := []byte(`{"username":"admin","password":"pw123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("decode login payload: %v", err)
	}
	if payload["access_token"] == "" || payload["refresh_token"] == "" {
		t.Fatal("expected non-empty access_token and refresh_token")
	}
}

func TestRoutesAuthRefreshRouteWired(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader([]byte(`{"refresh_token":"bad"}`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 from refresh handler, got %d", w.Code)
	}
}

func TestRoutesAuthLogoutRouteWired(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from logout handler, got %d", w.Code)
	}
}

func TestRoutesPersonsInvalidUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
	}{
		{name: "get", method: http.MethodGet},
		{name: "patch", method: http.MethodPatch},
		{name: "delete", method: http.MethodDelete},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newTestRouter(t)
			req := httptest.NewRequest(tt.method, "/api/v1/persons/not-a-uuid", http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for invalid person id, got %d", w.Code)
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("expected content-type application/json, got %q", got)
			}
			if got := w.Body.String(); got != "{\"error\":\"invalid person id\"}" {
				t.Fatalf("expected invalid person id error body, got %q", got)
			}
		})
	}
}

func TestRoutesPersonsCreateInvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/persons", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid create person body, got %d", w.Code)
	}
}

func TestRoutesPersonsMethodNotAllowed(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/persons", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for unsupported method, got %d", w.Code)
	}
}

func TestRoutesSettingsUpdateInvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewReader([]byte(`{"weight_unit":`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid settings body, got %d", w.Code)
	}
}

func TestRoutesManufacturersInvalidUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
	}{
		{name: "get", method: http.MethodGet},
		{name: "patch", method: http.MethodPatch},
		{name: "delete", method: http.MethodDelete},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newTestRouter(t)
			req := httptest.NewRequest(tt.method, "/api/v1/manufacturers/not-a-uuid", http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for invalid manufacturer id, got %d", w.Code)
			}
		})
	}
}

func TestRoutesManufacturersCreateInvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/manufacturers", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid manufacturer create body, got %d", w.Code)
	}
}

func TestRoutesItemsInvalidUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
	}{
		{name: "get", method: http.MethodGet},
		{name: "patch", method: http.MethodPatch},
		{name: "delete", method: http.MethodDelete},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newTestRouter(t)
			req := httptest.NewRequest(tt.method, "/api/v1/items/not-a-uuid", http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for invalid item id, got %d", w.Code)
			}
		})
	}
}

func TestRoutesItemsCreateInvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid item create body, got %d", w.Code)
	}
}

func TestRoutesSetsInvalidUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{name: "get set", path: "/api/v1/sets/not-a-uuid"},
		{name: "get set items", path: "/api/v1/sets/not-a-uuid/items"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newTestRouter(t)
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for invalid set id, got %d", w.Code)
			}
		})
	}
}

func TestRoutesSetsCreateInvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sets", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid set create body, got %d", w.Code)
	}
}

func TestRoutesPacksInvalidUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{name: "get pack", path: "/api/v1/packs/not-a-uuid"},
		{name: "get pack items", path: "/api/v1/packs/not-a-uuid/items"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newTestRouter(t)
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for invalid pack id, got %d", w.Code)
			}
		})
	}
}

func TestRoutesPacksCreateInvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/packs", bytes.NewReader([]byte(`{"name":`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid pack create body, got %d", w.Code)
	}
}
