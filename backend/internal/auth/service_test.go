package auth

import (
	"errors"
	"testing"
)

func TestNewServiceValidation(t *testing.T) {
	t.Parallel()

	_, err := NewService("", "pass", "secret")
	if err == nil {
		t.Fatal("expected error when username is empty")
	}

	_, err = NewService("user", "", "secret")
	if err == nil {
		t.Fatal("expected error when password is empty")
	}

	_, err = NewService("user", "pass", "")
	if err == nil {
		t.Fatal("expected error when secret is empty")
	}
}

func TestLoginSuccessAndFailure(t *testing.T) {
	t.Parallel()

	svc, err := NewService("admin", "pw123", "test-secret")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	pair, err := svc.Login("admin", "pw123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}
	if pair.ExpiresIn <= 0 {
		t.Fatalf("expected positive expires_in, got %d", pair.ExpiresIn)
	}

	_, err = svc.Login("admin", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestRefresh(t *testing.T) {
	t.Parallel()

	svc, err := NewService("admin", "pw123", "test-secret")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	pair, err := svc.Login("admin", "pw123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	newPair, err := svc.Refresh(pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh with refresh token: %v", err)
	}
	if newPair.AccessToken == "" || newPair.RefreshToken == "" {
		t.Fatal("expected non-empty tokens after refresh")
	}

	_, err = svc.Refresh(pair.AccessToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for access token refresh, got %v", err)
	}

	_, err = svc.Refresh("not-a-token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for malformed token, got %v", err)
	}
}

func TestLogout(t *testing.T) {
	t.Parallel()

	svc, err := NewService("admin", "pw123", "test-secret")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if err := svc.Logout("whatever"); err != nil {
		t.Fatalf("logout: %v", err)
	}
}
