package api

import (
	"net/http"
	"strings"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

// AuthMiddleware creates a middleware that authenticates requests using API keys
func AuthMiddleware(apiKeys *models.APIKeyModel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				ErrorUnauthorized(w, "Missing Authorization header")
				return
			}

			// Check for Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				ErrorUnauthorized(w, "Invalid Authorization header format. Expected: Bearer <token>")
				return
			}

			apiKey := parts[1]

			// Authenticate the API key
			key, err := apiKeys.Authenticate(apiKey)
			if err != nil {
				ErrorUnauthorized(w, MsgInvalidAPIKey)
				return
			}

			// Update last used timestamp (async, don't block on error)
			go func() {
				_ = apiKeys.UpdateLastUsed(key.ID)
			}()

			// Add API key info to request context
			ctx := SetAPIKeyContext(r.Context(), key.ID, key.UserID, key.Scopes, key.Name)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuthMiddleware creates a middleware that optionally authenticates requests
// If an API key is provided, it validates it, but doesn't require it
func OptionalAuthMiddleware(apiKeys *models.APIKeyModel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No auth header, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Check for Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				// Invalid format, but since this is optional, just continue
				next.ServeHTTP(w, r)
				return
			}

			apiKey := parts[1]

			// Authenticate the API key
			key, err := apiKeys.Authenticate(apiKey)
			if err != nil {
				// Invalid key, but since this is optional, just continue
				next.ServeHTTP(w, r)
				return
			}

			// Update last used timestamp (async, don't block on error)
			go func() {
				_ = apiKeys.UpdateLastUsed(key.ID)
			}()

			// Add API key info to request context
			ctx := SetAPIKeyContext(r.Context(), key.ID, key.UserID, key.Scopes, key.Name)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
