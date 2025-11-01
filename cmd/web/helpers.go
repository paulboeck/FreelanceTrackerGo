package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

// serverError logs a server error with stack trace and sends a 500 response to the client.
// This is a method on *application (receiver), meaning it can access app.logger and other app fields.
// In Go, you define methods by adding a receiver between 'func' and the function name.
func (app *application) serverError(resp http.ResponseWriter, req *http.Request, err error) {
	// var (...) declares multiple variables at once
	var (
		method = req.Method
		uri    = req.URL.RequestURI()
		trace  = string(debug.Stack()) // Get the stack trace to help debug where the error occurred
	)

	// Log the error with context (method, URI, and stack trace)
	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	// Send a generic 500 Internal Server Error response to the client
	http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError sends an error response for client mistakes (like 404 Not Found, 400 Bad Request).
func (app *application) clientError(resp http.ResponseWriter, status int) {
	http.Error(resp, http.StatusText(status), status)
}

// render executes an HTML template and writes it to the HTTP response.
// It uses a two-stage process: first render to a buffer, then write to response.
// This way, if rendering fails, we can send an error without partial HTML being sent.
func (app *application) render(resp http.ResponseWriter, req *http.Request, status int, page string, data templateData) {
	// Try to get the compiled template from the cache
	// The second value (ok) is a boolean that's true if the key exists in the map
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template page %s does not exist", page)
		app.serverError(resp, req, err)
		return
	}

	// new(bytes.Buffer) creates a new buffer to write the rendered template into
	// We render to a buffer first so if there's an error, we don't send partial HTML
	buf := new(bytes.Buffer)

	// ExecuteTemplate renders the template with the provided data
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(resp, req, err)
	}

	// Set the HTTP status code (e.g., 200 OK, 201 Created)
	resp.WriteHeader(status)

	// Write the buffered content to the HTTP response
	// The underscore (_) discards the first return value (bytes written) since we don't need it
	_, err = buf.WriteTo(resp)
	if err != nil {
		app.serverError(resp, req, err)
	}
}

// newTemplateData creates a templateData struct with common data needed by all templates.
func (app *application) newTemplateData(req *http.Request) templateData {
	isAuth := app.isAuthenticated(req)
	perms := getUserPermissions(req)

	// Get current user if authenticated
	var currentUser *models.User
	if isAuth {
		userID := getUserID(req)
		if userID != 0 {
			user, err := app.users.Get(userID)
			if err == nil {
				currentUser = &user
			}
		}
	}

	// Safely get flash message - only if session context exists
	var flash string
	defer func() {
		if r := recover(); r != nil {
			// If session context doesn't exist (e.g., in tests), ignore and continue
			flash = ""
		}
	}()
	flash = app.sessionManager.PopString(req.Context(), "flash")

	return templateData{
		CurrentYear: time.Now().Year(),

		// Use the PopString() method to retrieve the value for the "flash" key.
		// PopString() also deletes the key and value from the session data, so it
		// behaves like a one-time fetch. If there is no matching key in the session
		// data this will return the empty string.
		Flash: flash,

		// Check if the user is authenticated
		IsAuthenticated: isAuth,

		// Current authenticated user
		User: currentUser,

		// Load user permissions for template conditionals
		Permissions: perms,
	}
}

// decodePostForm parses form data from an HTTP POST request into a Go struct.
// The 'any' type (interface{}) means it accepts any type - the actual struct is determined at runtime.
// dst is the destination where the form data will be decoded into.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	// ParseForm parses the raw query from the URL and updates r.PostForm
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Decode the form data into the destination struct
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// Check if the error is an InvalidDecoderError (programmer error, not user error)
		var invalidDecoderError *form.InvalidDecoderError
		// errors.As checks if err is of the specified type and assigns it if so
		if errors.As(err, &invalidDecoderError) {
			// panic() stops normal execution and begins panicking (use for unrecoverable errors)
			panic(err)
		}
		return err
	}
	return nil

}

// isAuthenticated returns true if the current request is from an authenticated user.
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

// contextKey is a custom type for context keys to avoid collisions with other packages.
// Using a custom type prevents conflicts if other packages use string keys.
type contextKey string

// Context keys for storing request-scoped values
const (
	isAuthenticatedContextKey = contextKey("isAuthenticated")
	userIDContextKey          = contextKey("userID")
	permissionsContextKey     = contextKey("permissions")
)

// contextSetUser adds the user authentication status and ID to the request context.
// context.Context is used to pass request-scoped values through the call chain.
// This is Go's way of passing data through middleware and handlers without global variables.
func contextSetUser(ctx context.Context, userID int) context.Context {
	// Set authentication status
	ctx = context.WithValue(ctx, isAuthenticatedContextKey, userID != 0)
	// Set user ID
	ctx = context.WithValue(ctx, userIDContextKey, userID)
	return ctx
}

// contextSetPermissions adds user permissions to the request context.
func contextSetPermissions(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, permissionsContextKey, permissions)
}

// getUserID returns the authenticated user's ID from the request context.
func getUserID(r *http.Request) int {
	userID, ok := r.Context().Value(userIDContextKey).(int)
	if !ok {
		return 0
	}
	return userID
}

// getUserPermissions returns the user's permissions from the request context.
func getUserPermissions(r *http.Request) []string {
	permissions, ok := r.Context().Value(permissionsContextKey).([]string)
	if !ok {
		return []string{}
	}
	return permissions
}

// hasPermission checks if the user has a specific permission.
func (app *application) hasPermission(r *http.Request, permission string) bool {
	permissions := getUserPermissions(r)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}
