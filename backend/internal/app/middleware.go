package app

import (
	"net/http"
	"strings"

	"github.com/via-justa/overpacked-app/backend/internal/auth"
)

// maxBodyBytes caps the request body for general JSON endpoints to limit
// memory/storage abuse. The backup import endpoint streams large archives and
// installs its own (larger) limit, so it is exempt from this cap.
const maxBodyBytes = 8 << 20 // 8 MiB

// backupImportPath is exempted from maxBodyBytes; it sets its own limit.
const backupImportPath = "/api/v1/backup/import"

// limitBody caps the size of request bodies to defend against oversized or
// unbounded payloads (e.g. large image blobs on item create/update).
func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && r.URL.Path != backupImportPath {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// securityHeaders adds defense-in-depth response headers. The API is never
// framed and its bytes should never be content-sniffed.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// publicPaths are reachable without a valid access token.
var publicPaths = map[string]bool{
	"/health":              true,
	"/api/v1/auth/login":   true,
	"/api/v1/auth/refresh": true,
	"/api/v1/auth/logout":  true,
}

// requireAuth enforces a valid bearer access token on every route except publicPaths.
func requireAuth(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if publicPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("Authorization")
			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			if token == "" || authService.ValidateAccessToken(token) != nil {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
