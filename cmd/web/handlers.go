package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

const NAME_LENGTH = 255

type clientCreateForm struct {
	Name                string `form:"name"`
	validator.Validator `form:"-"`
}

// home handles http requests to the root URl of the project
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

// clientView handles a GET request to the for a specific client ID,
// queries the database for that client, and passes the result to be rendered
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

// clientCreate handles a GET request which returns an empty client detail form
func (app *application) clientCreate(res http.ResponseWriter, req *http.Request) {
	data := app.newTemplateData(req)
	data.Form = clientCreateForm{}
	app.render(res, req, http.StatusOK, "client_create.html", data)
}

// clientCreatePost handles a POST request with client form data which is then
// validated and used to insert a new client into the database
func (app *application) clientCreatePost(res http.ResponseWriter, req *http.Request) {
	var form clientCreateForm
	err := app.decodePostForm(req, &form)
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	err = app.formDecoder.Decode(&form, req.PostForm)
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "Name is required")
	form.CheckField(validator.MaxChars(form.Name, NAME_LENGTH), "name", fmt.Sprintf("Name must be shorter than %d characters", NAME_LENGTH))

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
