package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))

	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /client/view/{id}", dynamic.ThenFunc(app.clientView))
	mux.Handle("GET /client/create", dynamic.ThenFunc(app.clientCreate))
	mux.Handle("POST /client/create", dynamic.ThenFunc(app.clientCreatePost))
	mux.Handle("GET /client/update/{id}", dynamic.ThenFunc(app.clientUpdate))
	mux.Handle("POST /client/update/{id}", dynamic.ThenFunc(app.clientUpdatePost))
	mux.Handle("POST /client/delete/{id}", dynamic.ThenFunc(app.clientDelete))
	mux.Handle("GET /client/{id}/project/create", dynamic.ThenFunc(app.projectCreate))
	mux.Handle("POST /client/{id}/project/create", dynamic.ThenFunc(app.projectCreatePost))
	mux.Handle("GET /project/view/{id}", dynamic.ThenFunc(app.projectView))
	mux.Handle("GET /project/update/{id}", dynamic.ThenFunc(app.projectUpdate))
	mux.Handle("POST /project/update/{id}", dynamic.ThenFunc(app.projectUpdatePost))
	mux.Handle("POST /project/delete/{id}", dynamic.ThenFunc(app.projectDelete))
	mux.Handle("GET /project/{id}/timesheet/create", dynamic.ThenFunc(app.timesheetCreate))
	mux.Handle("POST /project/{id}/timesheet/create", dynamic.ThenFunc(app.timesheetCreatePost))
	mux.Handle("GET /timesheet/update/{id}", dynamic.ThenFunc(app.timesheetUpdate))
	mux.Handle("POST /timesheet/update/{id}", dynamic.ThenFunc(app.timesheetUpdatePost))
	mux.Handle("POST /timesheet/delete/{id}", dynamic.ThenFunc(app.timesheetDelete))
	mux.Handle("GET /project/{id}/invoice/create", dynamic.ThenFunc(app.invoiceCreate))
	mux.Handle("POST /project/{id}/invoice/create", dynamic.ThenFunc(app.invoiceCreatePost))
	mux.Handle("GET /invoice/update/{id}", dynamic.ThenFunc(app.invoiceUpdate))
	mux.Handle("POST /invoice/update/{id}", dynamic.ThenFunc(app.invoiceUpdatePost))
	mux.Handle("POST /invoice/delete/{id}", dynamic.ThenFunc(app.invoiceDelete))
	mux.Handle("GET /invoice/print/{id}", dynamic.ThenFunc(app.invoicePrint))
	mux.Handle("GET /settings", dynamic.ThenFunc(app.settingsView))
	mux.Handle("GET /settings/edit", dynamic.ThenFunc(app.settingsEdit))
	mux.Handle("POST /settings/edit", dynamic.ThenFunc(app.settingsEditPost))

	standardChain := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standardChain.Then(mux)
}
