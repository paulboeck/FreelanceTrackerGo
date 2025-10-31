package main

import (
	"fmt"
	"net/http"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

// commonHeaders is middleware that adds security headers to all HTTP responses.
// Middleware in Go wraps an http.Handler and returns a new http.Handler.
// The pattern is: take a handler, do something before/after it, then call the next handler.
func commonHeaders(next http.Handler) http.Handler {
	// http.HandlerFunc converts a function into an http.Handler
	// This function is a closure - it "closes over" the 'next' variable from the outer function
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers before calling the next handler
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// logRequest is middleware that logs each incoming HTTP request.
// It's a method on *application so it can access app.logger.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract request details to log
		var (
			ip     = r.RemoteAddr       // Client's IP address
			proto  = r.Proto            // HTTP protocol version (e.g., "HTTP/1.1")
			method = r.Method           // HTTP method (GET, POST, etc.)
			uri    = r.URL.RequestURI() // Requested URI path
		)

		// Log the request details
		app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}

// recoverPanic is middleware that recovers from panics and returns a 500 error.
// Without this, a panic would crash the entire server.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer schedules this function to run when the outer function returns
		// This is Go's mechanism for handling panics (like try/catch in other languages)
		defer func() {
			// recover() captures a panic and returns the panic value (or nil if no panic)
			pv := recover()

			if pv != nil {
				// If there was a panic, close the connection and send an error response
				w.Header().Set("Connection", "close")
				// fmt.Errorf with %v converts the panic value to an error
				app.serverError(w, r, fmt.Errorf("%v", pv))
			}
		}()
		// Call the next handler (if it panics, defer above will catch it)
		next.ServeHTTP(w, r)
	})
}

// requireAuthentication is middleware that redirects unauthenticated users to the login page.
// It protects routes that require a logged-in user.
func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is authenticated
		if !app.isAuthenticated(r) {
			// Redirect to login page with HTTP 303 See Other status
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return // Stop processing - don't call next handler
		}

		// Tell browsers not to cache authenticated pages (security measure)
		w.Header().Add("Cache-Control", "no-store")

		// User is authenticated, continue to the next handler
		next.ServeHTTP(w, r)
	})
}

// authenticate is middleware that checks if a user is logged in and adds their info to the context.
// It doesn't block unauthenticated users - it just enriches the context with user data if available.
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get the authenticated user ID from the session
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			// No authenticated user, continue without setting context
			next.ServeHTTP(w, r)
			return
		}

		// Check if the user still exists in the database
		_, err := app.users.Get(id)
		if err != nil {
			// If the error is NOT "no record found", it's a server error
			if err != models.ErrNoRecord {
				app.serverError(w, r, err)
				return
			}
			// User doesn't exist anymore, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// User exists, add authentication status and user ID to the request context
		ctx := contextSetUser(r.Context(), id)

		// Load user's permissions and add to context
		permissions, err := app.permissions.GetUserPermissionNames(id)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		ctx = contextSetPermissions(ctx, permissions)

		// WithContext creates a shallow copy of the request with the new context
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// requirePermission is middleware factory that creates middleware to check for a specific permission.
// It returns a middleware function that checks if the authenticated user has the required permission.
func (app *application) requirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the user is authenticated
			if !app.isAuthenticated(r) {
				http.Redirect(w, r, "/user/login", http.StatusSeeOther)
				return
			}

			// Check if the user has the required permission
			if !app.hasPermission(r, permission) {
				// User is authenticated but doesn't have permission - show forbidden error
				app.clientError(w, http.StatusForbidden)
				return
			}

			// User has permission, continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// requirePasswordChange is middleware that redirects users who need to change their password.
// This ensures users with require_password_change flag set must change their password before accessing the app.
func (app *application) requirePasswordChange(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check for authenticated users
		if !app.isAuthenticated(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Get user ID from context
		userID := getUserID(r)
		if userID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Get user from database
		user, err := app.users.Get(userID)
		if err != nil {
			if err != models.ErrNoRecord {
				app.serverError(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// If user requires password change, redirect to password change page
		// But allow access to the password change page itself and logout
		if user.RequirePasswordChange {
			// Allow access to password change page and logout
			if r.URL.Path != "/user/password/change" && r.URL.Path != "/user/logout" {
				http.Redirect(w, r, "/user/password/change", http.StatusSeeOther)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
