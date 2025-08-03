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

	standardChain := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standardChain.Then(mux)
}
