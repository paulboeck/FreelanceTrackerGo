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
	// protected chain: adds authentication requirement to the dynamic chain
	protected := dynamic.Append(app.requireAuthentication)

	// Public routes (no authentication required)
	// dynamic.ThenFunc wraps the handler function with the dynamic middleware chain
	// The route pattern "GET /user/signup" specifies both the HTTP method and path
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protected routes (authentication required)
	// These use the protected chain which includes requireAuthentication middleware
	// {$} means exact match - only "/" not "/something"
	mux.Handle("GET /{$}", protected.ThenFunc(app.clientsList))
	mux.Handle("GET /clients", protected.ThenFunc(app.clientsList))
	mux.Handle("GET /projects", protected.ThenFunc(app.projectsList))
	mux.Handle("GET /project/create", protected.ThenFunc(app.projectCreateGeneral))
	mux.Handle("POST /project/create", protected.ThenFunc(app.projectCreateGeneralPost))
	// {id} is a path parameter - it captures the value from the URL (e.g., /client/view/42 -> id=42)
	mux.Handle("GET /client/view/{id}", protected.ThenFunc(app.clientView))
	mux.Handle("GET /client/create", protected.ThenFunc(app.clientCreate))
	mux.Handle("POST /client/create", protected.ThenFunc(app.clientCreatePost))
	mux.Handle("GET /client/update/{id}", protected.ThenFunc(app.clientUpdate))
	mux.Handle("POST /client/update/{id}", protected.ThenFunc(app.clientUpdatePost))
	mux.Handle("POST /client/delete/{id}", protected.ThenFunc(app.clientDelete))
	mux.Handle("GET /client/{id}/project/create", protected.ThenFunc(app.projectCreate))
	mux.Handle("POST /client/{id}/project/create", protected.ThenFunc(app.projectCreatePost))
	mux.Handle("GET /project/view/{id}", protected.ThenFunc(app.projectView))
	mux.Handle("GET /project/update/{id}", protected.ThenFunc(app.projectUpdate))
	mux.Handle("POST /project/update/{id}", protected.ThenFunc(app.projectUpdatePost))
	mux.Handle("POST /project/delete/{id}", protected.ThenFunc(app.projectDelete))
	mux.Handle("GET /project/{id}/timesheet/create", protected.ThenFunc(app.timesheetCreate))
	mux.Handle("POST /project/{id}/timesheet/create", protected.ThenFunc(app.timesheetCreatePost))
	mux.Handle("GET /timesheet/update/{id}", protected.ThenFunc(app.timesheetUpdate))
	mux.Handle("POST /timesheet/update/{id}", protected.ThenFunc(app.timesheetUpdatePost))
	mux.Handle("POST /timesheet/delete/{id}", protected.ThenFunc(app.timesheetDelete))
	mux.Handle("GET /project/{id}/invoice/create", protected.ThenFunc(app.invoiceCreate))
	mux.Handle("POST /project/{id}/invoice/create", protected.ThenFunc(app.invoiceCreatePost))
	mux.Handle("GET /invoice/update/{id}", protected.ThenFunc(app.invoiceUpdate))
	mux.Handle("POST /invoice/update/{id}", protected.ThenFunc(app.invoiceUpdatePost))
	mux.Handle("POST /invoice/delete/{id}", protected.ThenFunc(app.invoiceDelete))
	mux.Handle("GET /invoice/print/{id}", protected.ThenFunc(app.invoicePrint))
	mux.Handle("POST /invoice/email/{id}", protected.ThenFunc(app.invoiceEmail))
	mux.Handle("GET /reports/income", protected.ThenFunc(app.incomeReport))
	mux.Handle("POST /reports/income", protected.ThenFunc(app.incomeReportPost))
	mux.Handle("GET /settings", protected.ThenFunc(app.settingsView))
	mux.Handle("GET /settings/edit", protected.ThenFunc(app.settingsEdit))
	mux.Handle("POST /settings/edit", protected.ThenFunc(app.settingsEditPost))
	mux.Handle("GET /api/client/{id}/hourlyrate", protected.ThenFunc(app.clientHourlyRateAPI))
	mux.Handle("GET /user/logout", protected.ThenFunc(app.userLogout))

	// Create the standard middleware chain that wraps ALL routes
	// Order matters: recoverPanic runs first (outermost), then logRequest, then commonHeaders
	standardChain := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	// Then() wraps the mux with the standard middleware chain
	return standardChain.Then(mux)
}
