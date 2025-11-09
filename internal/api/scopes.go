package api

import (
	"context"
	"net/http"
)

// Scope-related context keys
type contextKey string

const (
	contextKeyAPIKeyID     contextKey = "api_key_id"
	contextKeyUserID       contextKey = "user_id"
	contextKeyScopes       contextKey = "scopes"
	contextKeyAPIKeyName   contextKey = "api_key_name"
)

// GetAPIKeyID retrieves the API key ID from the request context
func GetAPIKeyID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(contextKeyAPIKeyID).(int)
	return id, ok
}

// GetUserID retrieves the user ID from the request context
func GetUserID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(contextKeyUserID).(int)
	return id, ok
}

// GetScopes retrieves the scopes from the request context
func GetScopes(r *http.Request) (string, bool) {
	scopes, ok := r.Context().Value(contextKeyScopes).(string)
	return scopes, ok
}

// GetAPIKeyName retrieves the API key name from the request context
func GetAPIKeyName(r *http.Request) (string, bool) {
	name, ok := r.Context().Value(contextKeyAPIKeyName).(string)
	return name, ok
}

// SetAPIKeyContext sets API key information in the request context
func SetAPIKeyContext(ctx context.Context, keyID, userID int, scopes, name string) context.Context {
	ctx = context.WithValue(ctx, contextKeyAPIKeyID, keyID)
	ctx = context.WithValue(ctx, contextKeyUserID, userID)
	ctx = context.WithValue(ctx, contextKeyScopes, scopes)
	ctx = context.WithValue(ctx, contextKeyAPIKeyName, name)
	return ctx
}

// HasScope checks if the request has a specific scope
func HasScope(r *http.Request, scope string) bool {
	scopes, ok := GetScopes(r)
	if !ok {
		return false
	}
	return hasScopeString(scopes, scope)
}

// HasAllScopes checks if the request has all specified scopes
func HasAllScopes(r *http.Request, requiredScopes ...string) bool {
	scopes, ok := GetScopes(r)
	if !ok {
		return false
	}

	for _, required := range requiredScopes {
		if !hasScopeString(scopes, required) {
			return false
		}
	}
	return true
}

// hasScopeString checks if a scope string contains a specific scope
// This duplicates the logic from models.HasScope but avoids circular dependency
func hasScopeString(keyScopes, requiredScope string) bool {
	// Split scopes by whitespace
	scopes := splitScopes(keyScopes)

	// Check for wildcard
	for _, scope := range scopes {
		if scope == "*" {
			return true
		}
		if scope == requiredScope {
			return true
		}
		// Check for resource wildcard (e.g., "clients:*" matches "clients:read")
		if len(scope) > 2 && scope[len(scope)-2:] == ":*" {
			resource := scope[:len(scope)-2]
			if len(requiredScope) > len(resource)+1 && requiredScope[:len(resource)+1] == resource+":" {
				return true
			}
		}
	}

	return false
}

// splitScopes splits a space-separated scope string into a slice
func splitScopes(scopes string) []string {
	if scopes == "" {
		return nil
	}

	var result []string
	current := ""
	for _, ch := range scopes {
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}

	return result
}

// RequireScopes creates a middleware that requires specific scopes
func RequireScopes(scopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !HasAllScopes(r, scopes...) {
				ErrorForbidden(w, "Required scopes: "+join(scopes, ", "))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// join concatenates strings with a separator
func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
