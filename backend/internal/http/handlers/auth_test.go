package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/via-justa/overpacked-app/backend/internal/auth"
)

func newTestAuthHandler(t *testing.T) *AuthHandler {
	t.Helper()
	svc, err := auth.NewService("admin", "pw123", "test-secret")
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return NewAuthHandler(svc)
}

func TestAuthHandlerLoginSuccess(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t)
	body := []byte(`{"username":"admin","password":"pw123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var payload authResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.AccessToken == "" || payload.RefreshToken == "" {
		t.Fatal("expected non-empty tokens in response")
	}
	if payload.TokenType != "Bearer" {
		t.Fatalf("expected token_type Bearer, got %q", payload.TokenType)
	}
}

func TestAuthHandlerLoginValidationAndCredentials(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t)

	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{"username":`)))
	badW := httptest.NewRecorder()
	h.Login(badW, badReq)
	if badW.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", badW.Result().StatusCode)
	}

	wrongReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{"username":"admin","password":"wrong"}`)))
	wrongW := httptest.NewRecorder()
	h.Login(wrongW, wrongReq)
	if wrongW.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid credentials, got %d", wrongW.Result().StatusCode)
	}
}

func TestAuthHandlerRefresh(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{"username":"admin","password":"pw123"}`)))
	loginW := httptest.NewRecorder()
	h.Login(loginW, loginReq)

	var loginPayload authResponse
	if err := json.NewDecoder(loginW.Result().Body).Decode(&loginPayload); err != nil {
		t.Fatalf("decode login response: %v", err)
	}

	refreshBody, err := json.Marshal(map[string]string{"refresh_token": loginPayload.RefreshToken})
	if err != nil {
		t.Fatalf("marshal refresh body: %v", err)
	}
	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refreshBody))
	refreshW := httptest.NewRecorder()
	h.Refresh(refreshW, refreshReq)
	if refreshW.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for valid refresh, got %d", refreshW.Result().StatusCode)
	}

	invalidReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader([]byte(`{"refresh_token":"bad-token"}`)))
	invalidW := httptest.NewRecorder()
	h.Refresh(invalidW, invalidReq)
	if invalidW.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid refresh token, got %d", invalidW.Result().StatusCode)
	}
}

func TestAuthHandlerLogout(t *testing.T) {
	t.Parallel()

	h := newTestAuthHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", http.NoBody)
	req.Header.Set("Authorization", "Bearer some-token")
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Result().StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Result().StatusCode)
	}
}
