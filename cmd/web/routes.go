package main

import (
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))

	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /client/view/{id}", app.clientView)
	mux.HandleFunc("GET /client/create", app.clientCreate)
	mux.HandleFunc("POST /client/create", app.clientCreatePost)
	mux.HandleFunc("GET /client/update/{id}", app.clientUpdate)
	mux.HandleFunc("POST /client/update/{id}", app.clientUpdatePost)
	mux.HandleFunc("POST /client/delete/{id}", app.clientDelete)
	mux.HandleFunc("GET /client/{id}/project/create", app.projectCreate)
	mux.HandleFunc("POST /client/{id}/project/create", app.projectCreatePost)
	mux.HandleFunc("GET /project/view/{id}", app.projectView)
	mux.HandleFunc("GET /project/update/{id}", app.projectUpdate)
	mux.HandleFunc("POST /project/update/{id}", app.projectUpdatePost)
	mux.HandleFunc("POST /project/delete/{id}", app.projectDelete)
	mux.HandleFunc("GET /project/{id}/timesheet/create", app.timesheetCreate)
	mux.HandleFunc("POST /project/{id}/timesheet/create", app.timesheetCreatePost)
	mux.HandleFunc("GET /timesheet/update/{id}", app.timesheetUpdate)
	mux.HandleFunc("POST /timesheet/update/{id}", app.timesheetUpdatePost)
	mux.HandleFunc("POST /timesheet/delete/{id}", app.timesheetDelete)
	mux.HandleFunc("GET /project/{id}/invoice/create", app.invoiceCreate)
	mux.HandleFunc("POST /project/{id}/invoice/create", app.invoiceCreatePost)
	mux.HandleFunc("GET /invoice/update/{id}", app.invoiceUpdate)
	mux.HandleFunc("POST /invoice/update/{id}", app.invoiceUpdatePost)
	mux.HandleFunc("POST /invoice/delete/{id}", app.invoiceDelete)
	mux.HandleFunc("GET /invoice/print/{id}", app.invoicePrint)
	mux.HandleFunc("GET /settings", app.settingsView)
	mux.HandleFunc("GET /settings/edit", app.settingsEdit)
	mux.HandleFunc("POST /settings/edit", app.settingsEditPost)

	standardChain := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standardChain.Then(mux)
}
