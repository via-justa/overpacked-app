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
	if payload.AccessToken == "" {
		t.Fatal("expected non-empty access token in response")
	}
	if payload.TokenType != "Bearer" {
		t.Fatalf("expected token_type Bearer, got %q", payload.TokenType)
	}

	refresh := findCookie(res.Cookies(), refreshCookieName)
	if refresh == nil || refresh.Value == "" {
		t.Fatal("expected refresh token cookie to be set")
	}
	if !refresh.HttpOnly {
		t.Fatal("expected refresh cookie to be HttpOnly")
	}
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
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

	refreshCookie := findCookie(loginW.Result().Cookies(), refreshCookieName)
	if refreshCookie == nil {
		t.Fatal("expected refresh cookie from login")
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", http.NoBody)
	refreshReq.AddCookie(refreshCookie)
	refreshW := httptest.NewRecorder()
	h.Refresh(refreshW, refreshReq)
	if refreshW.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for valid refresh, got %d", refreshW.Result().StatusCode)
	}
	if rotated := findCookie(refreshW.Result().Cookies(), refreshCookieName); rotated == nil || rotated.Value == "" {
		t.Fatal("expected refresh to rotate the cookie")
	}

	invalidReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", http.NoBody)
	invalidReq.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "bad-token"})
	invalidW := httptest.NewRecorder()
	h.Refresh(invalidW, invalidReq)
	if invalidW.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid refresh token, got %d", invalidW.Result().StatusCode)
	}

	missingReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", http.NoBody)
	missingW := httptest.NewRecorder()
	h.Refresh(missingW, missingReq)
	if missingW.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 when refresh cookie is missing, got %d", missingW.Result().StatusCode)
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
