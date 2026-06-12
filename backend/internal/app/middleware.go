package app

import (
	"net/http"
	"strings"

	"github.com/via-justa/overpacked-app/backend/internal/auth"
)

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
