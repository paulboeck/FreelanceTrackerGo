package main

import (
	"net/http"

	"github.com/justinas/alice"
)

// routes sets up all the HTTP routes for the application and returns the router.
func (app *application) routes() http.Handler {
	// Create a new ServeMux (HTTP request multiplexer/router)
	// It matches incoming request URLs to registered patterns and calls the appropriate handler
	mux := http.NewServeMux()

	// Set up a file server to serve static files (CSS, JavaScript, images)
	// http.Dir converts a string to a file system path
	fileServer := http.FileServer(http.Dir("./ui/static"))

	// Register the file server to handle requests starting with /static/
	// StripPrefix removes "/static" from the URL before passing it to the file server
	// So "/static/css/style.css" becomes "css/style.css" when looking in the ./ui/static directory
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// Create middleware chains using alice (makes chaining middleware cleaner)
	// dynamic chain: runs on all dynamic routes (not static files)
	dynamic := alice.New(app.sessionManager.LoadAndSave, app.authenticate)
	// protected chain: adds authentication requirement and password change check to the dynamic chain
	protected := dynamic.Append(app.requireAuthentication, app.requirePasswordChange)

	// Public routes (no authentication required)
	// dynamic.ThenFunc wraps the handler function with the dynamic middleware chain
	// The route pattern "GET /user/login" specifies both the HTTP method and path
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	// Password change route (authenticated but bypasses password change check for this specific route)
	mux.Handle("GET /user/password/change", dynamic.Append(app.requireAuthentication).ThenFunc(app.userPasswordChange))
	mux.Handle("POST /user/password/change", dynamic.Append(app.requireAuthentication).ThenFunc(app.userPasswordChangePost))

	// Protected routes with permission checks
	// {$} means exact match - only "/" not "/something"
	// {id} is a path parameter - it captures the value from the URL (e.g., /client/view/42 -> id=42)

	// Client routes
	mux.Handle("GET /{$}", protected.Append(app.requirePermission("clients.list")).ThenFunc(app.clientsList))
	mux.Handle("GET /clients", protected.Append(app.requirePermission("clients.list")).ThenFunc(app.clientsList))
	mux.Handle("GET /client/view/{id}", protected.Append(app.requirePermission("clients.view")).ThenFunc(app.clientView))
	mux.Handle("GET /client/create", protected.Append(app.requirePermission("clients.create")).ThenFunc(app.clientCreate))
	mux.Handle("POST /client/create", protected.Append(app.requirePermission("clients.create")).ThenFunc(app.clientCreatePost))
	mux.Handle("GET /client/update/{id}", protected.Append(app.requirePermission("clients.edit")).ThenFunc(app.clientUpdate))
	mux.Handle("POST /client/update/{id}", protected.Append(app.requirePermission("clients.edit")).ThenFunc(app.clientUpdatePost))
	mux.Handle("POST /client/delete/{id}", protected.Append(app.requirePermission("clients.delete")).ThenFunc(app.clientDelete))
	mux.Handle("GET /api/client/{id}/hourlyrate", protected.Append(app.requirePermission("clients.view")).ThenFunc(app.clientHourlyRateAPI))

	// Project routes
	mux.Handle("GET /projects", protected.Append(app.requirePermission("projects.list")).ThenFunc(app.projectsList))
	mux.Handle("GET /project/view/{id}", protected.Append(app.requirePermission("projects.view")).ThenFunc(app.projectView))
	mux.Handle("GET /project/create", protected.Append(app.requirePermission("projects.create")).ThenFunc(app.projectCreateGeneral))
	mux.Handle("POST /project/create", protected.Append(app.requirePermission("projects.create")).ThenFunc(app.projectCreateGeneralPost))
	mux.Handle("GET /client/{id}/project/create", protected.Append(app.requirePermission("projects.create")).ThenFunc(app.projectCreate))
	mux.Handle("POST /client/{id}/project/create", protected.Append(app.requirePermission("projects.create")).ThenFunc(app.projectCreatePost))
	mux.Handle("GET /project/update/{id}", protected.Append(app.requirePermission("projects.edit")).ThenFunc(app.projectUpdate))
	mux.Handle("POST /project/update/{id}", protected.Append(app.requirePermission("projects.edit")).ThenFunc(app.projectUpdatePost))
	mux.Handle("POST /project/delete/{id}", protected.Append(app.requirePermission("projects.delete")).ThenFunc(app.projectDelete))

	// Timesheet routes
	mux.Handle("GET /project/{id}/timesheet/create", protected.Append(app.requirePermission("timesheets.create")).ThenFunc(app.timesheetCreate))
	mux.Handle("POST /project/{id}/timesheet/create", protected.Append(app.requirePermission("timesheets.create")).ThenFunc(app.timesheetCreatePost))
	mux.Handle("GET /timesheet/update/{id}", protected.Append(app.requirePermission("timesheets.edit")).ThenFunc(app.timesheetUpdate))
	mux.Handle("POST /timesheet/update/{id}", protected.Append(app.requirePermission("timesheets.edit")).ThenFunc(app.timesheetUpdatePost))
	mux.Handle("POST /timesheet/delete/{id}", protected.Append(app.requirePermission("timesheets.delete")).ThenFunc(app.timesheetDelete))

	// Invoice routes
	mux.Handle("GET /project/{id}/invoice/create", protected.Append(app.requirePermission("invoices.create")).ThenFunc(app.invoiceCreate))
	mux.Handle("POST /project/{id}/invoice/create", protected.Append(app.requirePermission("invoices.create")).ThenFunc(app.invoiceCreatePost))
	mux.Handle("GET /invoice/update/{id}", protected.Append(app.requirePermission("invoices.edit")).ThenFunc(app.invoiceUpdate))
	mux.Handle("POST /invoice/update/{id}", protected.Append(app.requirePermission("invoices.edit")).ThenFunc(app.invoiceUpdatePost))
	mux.Handle("POST /invoice/delete/{id}", protected.Append(app.requirePermission("invoices.delete")).ThenFunc(app.invoiceDelete))
	mux.Handle("GET /invoice/print/{id}", protected.Append(app.requirePermission("invoices.print")).ThenFunc(app.invoicePrint))
	mux.Handle("POST /invoice/email/{id}", protected.Append(app.requirePermission("invoices.email")).ThenFunc(app.invoiceEmail))

	// Report routes
	mux.Handle("GET /reports/income", protected.Append(app.requirePermission("reports.view")).ThenFunc(app.incomeReport))
	mux.Handle("POST /reports/income", protected.Append(app.requirePermission("reports.view")).ThenFunc(app.incomeReportPost))

	// Settings routes
	mux.Handle("GET /settings", protected.Append(app.requirePermission("settings.view")).ThenFunc(app.settingsView))
	mux.Handle("GET /settings/edit", protected.Append(app.requirePermission("settings.edit")).ThenFunc(app.settingsEdit))
	mux.Handle("POST /settings/edit", protected.Append(app.requirePermission("settings.edit")).ThenFunc(app.settingsEditPost))

	// User management routes
	mux.Handle("GET /users", protected.Append(app.requirePermission("users.list")).ThenFunc(app.usersList))
	mux.Handle("GET /user/create", protected.Append(app.requirePermission("users.create")).ThenFunc(app.userSignup))
	mux.Handle("POST /user/create", protected.Append(app.requirePermission("users.create")).ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/edit/{id}", protected.Append(app.requirePermission("users.edit")).ThenFunc(app.userEdit))
	mux.Handle("POST /user/edit/{id}", protected.Append(app.requirePermission("users.edit")).ThenFunc(app.userEditPost))
	mux.Handle("POST /user/delete/{id}", protected.Append(app.requirePermission("users.delete")).ThenFunc(app.userDelete))

	// Logout route (no special permission required, just authentication)
	mux.Handle("GET /user/logout", protected.ThenFunc(app.userLogout))

	// REST API routes (separate router with JWT/API key authentication)
	// The API routes are handled by their own router with different authentication
	apiRouter := app.apiRoutes(app.apiKeys)
	mux.Handle("/api/", http.StripPrefix("", apiRouter))

	// Create the standard middleware chain that wraps ALL routes
	// Order matters: recoverPanic runs first (outermost), then logRequest, then commonHeaders
	standardChain := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	// Then() wraps the mux with the standard middleware chain
	return standardChain.Then(mux)
}
