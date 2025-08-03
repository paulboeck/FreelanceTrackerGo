package main

import (
	"errors"
	"fmt"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
	"net/http"
	"strconv"
)

const NAME_LENGTH = 255

type clientCreateForm struct {
	Name                string `form:"name"`
	validator.Validator `form:"-"`
}

func (app *application) home(res http.ResponseWriter, req *http.Request) {
	clients, err := app.clients.GetAll()
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	data := app.newTemplateData(req)
	data.Clients = clients

	app.render(res, req, http.StatusOK, "home.html", data)
}

func (app *application) clientView(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	client, err := app.clients.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	data := app.newTemplateData(req)
	data.Client = client

	app.render(res, req, http.StatusOK, "client.html", data)
}

func (app *application) clientCreate(res http.ResponseWriter, req *http.Request) {
	data := app.newTemplateData(req)
	data.Form = clientCreateForm{}
	app.render(res, req, http.StatusOK, "client_create.html", data)
}

func (app *application) clientCreatePost(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	var form clientCreateForm

	err = app.formDecoder.Decode(&form, req.PostForm)
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "Name is required")
	form.CheckField(validator.MaxChars(form.Name, NAME_LENGTH), "name", fmt.Sprintf("Name must be shorter than %d characters"))

	if !form.Valid() {
		data := app.newTemplateData(req)
		data.Form = form
		app.render(res, req, http.StatusUnprocessableEntity, "client_create.html", data)
		return
	}

	id, err := app.clients.Insert(form.Name)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	http.Redirect(res, req, fmt.Sprintf("/client/view/%d", id), http.StatusSeeOther)
}
