package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))

	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave, app.authenticate)
	protected := dynamic.Append(app.requireAuthentication)

	// Public routes (no authentication required)
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protected routes (authentication required)
	mux.Handle("GET /{$}", protected.ThenFunc(app.clientsList))
	mux.Handle("GET /clients", protected.ThenFunc(app.clientsList))
	mux.Handle("GET /projects", protected.ThenFunc(app.projectsList))
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
	mux.Handle("GET /settings", protected.ThenFunc(app.settingsView))
	mux.Handle("GET /settings/edit", protected.ThenFunc(app.settingsEdit))
	mux.Handle("POST /settings/edit", protected.ThenFunc(app.settingsEditPost))
	mux.Handle("GET /user/logout", protected.ThenFunc(app.userLogout))

	standardChain := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standardChain.Then(mux)
}
