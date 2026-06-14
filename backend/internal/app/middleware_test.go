package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/via-justa/overpacked-app/backend/internal/auth"
)

func newTestAuthService(t *testing.T) *auth.Service {
	t.Helper()
	svc, err := auth.NewService("admin", "pw123", "test-secret")
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return svc
}

func TestRequireAuthAllowsPublicPaths(t *testing.T) {
	t.Parallel()

	handler := requireAuth(newTestAuthService(t))(okHandler())

	for path := range publicPaths {
		req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected public path %q to pass, got %d", path, w.Code)
		}
	}
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	t.Parallel()

	handler := requireAuth(newTestAuthService(t))(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", http.NoBody)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", w.Code)
	}
}

func TestRequireAuthRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	handler := requireAuth(newTestAuthService(t))(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", http.NoBody)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with invalid token, got %d", w.Code)
	}
}

func TestRequireAuthAcceptsValidAccessToken(t *testing.T) {
	t.Parallel()

	svc := newTestAuthService(t)
	tokens, err := svc.Login("admin", "pw123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	handler := requireAuth(svc)(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid token, got %d", w.Code)
	}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
