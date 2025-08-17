package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

const NAME_LENGTH = 255

type clientForm struct {
	Name                string `form:"name"`
	validator.Validator `form:"-"`
}

type projectForm struct {
	Name                string `form:"name"`
	validator.Validator `form:"-"`
}

type timesheetForm struct {
	WorkDate            string  `form:"work_date"`
	HoursWorked         string  `form:"hours_worked"`
	Description         string  `form:"description"`
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

	// Get projects for this client
	projects, err := app.projects.GetByClient(id)
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	data := app.newTemplateData(req)
	data.Client = &client
	data.Projects = projects

	app.render(res, req, http.StatusOK, "client.html", data)
}

// projectView handles a GET request to view a specific project ID,
// queries the database for that project and its client, and passes the result to be rendered
func (app *application) projectView(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	project, err := app.projects.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get the client for this project
	client, err := app.clients.Get(project.ClientID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get timesheets for this project
	timesheets, err := app.timesheets.GetByProject(id)
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	data := app.newTemplateData(req)
	data.Project = &project
	data.Client = &client
	data.Timesheets = timesheets

	app.render(res, req, http.StatusOK, "project.html", data)
}

// clientCreate handles a GET request which returns an empty client detail form
func (app *application) clientCreate(res http.ResponseWriter, req *http.Request) {
	data := app.newTemplateData(req)
	data.Form = clientForm{}
	app.render(res, req, http.StatusOK, "client_create.html", data)
}

// clientCreatePost handles a POST request with client form data which is then
// validated and used to insert a new client into the database
func (app *application) clientCreatePost(res http.ResponseWriter, req *http.Request) {
	var form clientForm
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

// clientUpdate handles a GET request which returns a client update form pre-populated with client data
func (app *application) clientUpdate(res http.ResponseWriter, req *http.Request) {
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
	data.Form = clientForm{
		Name: client.Name,
	}
	data.Client = &client
	app.render(res, req, http.StatusOK, "client_create.html", data)
}

// clientUpdatePost handles a POST request with client form data which is then
// validated and used to update an existing client in the database
func (app *application) clientUpdatePost(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	var form clientForm
	err = app.decodePostForm(req, &form)
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
		data.Form = form
		data.Client = &client
		app.render(res, req, http.StatusUnprocessableEntity, "client_create.html", data)
		return
	}

	// Check if client exists before updating
	_, err = app.clients.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	err = app.clients.Update(id, form.Name)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	http.Redirect(res, req, fmt.Sprintf("/client/view/%d", id), http.StatusSeeOther)
}

// clientDelete handles a POST request to soft delete a client
func (app *application) clientDelete(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	// Check if client exists before deleting
	_, err = app.clients.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	err = app.clients.Delete(id)
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	// Redirect to home page after successful deletion
	http.Redirect(res, req, "/", http.StatusSeeOther)
}

// projectCreate handles a GET request which returns an empty project creation form
func (app *application) projectCreate(res http.ResponseWriter, req *http.Request) {
	clientID, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || clientID < 0 {
		http.NotFound(res, req)
		return
	}

	// Check if client exists
	client, err := app.clients.Get(clientID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	data := app.newTemplateData(req)
	data.Form = projectForm{}
	data.Client = &client
	app.render(res, req, http.StatusOK, "project_create.html", data)
}

// projectCreatePost handles a POST request with project form data which is then
// validated and used to insert a new project into the database
func (app *application) projectCreatePost(res http.ResponseWriter, req *http.Request) {
	clientID, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || clientID < 0 {
		http.NotFound(res, req)
		return
	}

	// Check if client exists
	client, err := app.clients.Get(clientID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	var form projectForm
	err = app.decodePostForm(req, &form)
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
		data.Client = &client
		app.render(res, req, http.StatusUnprocessableEntity, "project_create.html", data)
		return
	}

	_, err = app.projects.Insert(form.Name, clientID)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	http.Redirect(res, req, fmt.Sprintf("/client/view/%d", clientID), http.StatusSeeOther)
}

// projectUpdate handles a GET request which returns a project update form pre-populated with project data
func (app *application) projectUpdate(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	project, err := app.projects.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get the client for context
	client, err := app.clients.Get(project.ClientID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	data := app.newTemplateData(req)
	data.Form = projectForm{
		Name: project.Name,
	}
	data.Client = &client
	app.render(res, req, http.StatusOK, "project_create.html", data)
}

// projectUpdatePost handles a POST request with project form data which is then
// validated and used to update an existing project in the database
func (app *application) projectUpdatePost(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	// Get the project to ensure it exists and get the client ID
	project, err := app.projects.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	var form projectForm
	err = app.decodePostForm(req, &form)
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
		client, err := app.clients.Get(project.ClientID)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				http.NotFound(res, req)
			} else {
				app.serverError(res, req, err)
			}
			return
		}
		data := app.newTemplateData(req)
		data.Form = form
		data.Client = &client
		app.render(res, req, http.StatusUnprocessableEntity, "project_create.html", data)
		return
	}

	err = app.projects.Update(id, form.Name)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	http.Redirect(res, req, fmt.Sprintf("/client/view/%d", project.ClientID), http.StatusSeeOther)
}

// projectDelete handles a POST request to soft delete a project
func (app *application) projectDelete(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	// Check if project exists before deleting and get client ID for redirect
	project, err := app.projects.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	err = app.projects.Delete(id)
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	// Redirect to client view page after successful deletion
	http.Redirect(res, req, fmt.Sprintf("/client/view/%d", project.ClientID), http.StatusSeeOther)
}

// timesheetCreate handles a GET request which returns an empty timesheet creation form
func (app *application) timesheetCreate(res http.ResponseWriter, req *http.Request) {
	projectID, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || projectID < 0 {
		http.NotFound(res, req)
		return
	}

	// Check if project exists
	project, err := app.projects.Get(projectID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get the client for context
	client, err := app.clients.Get(project.ClientID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	data := app.newTemplateData(req)
	data.Form = timesheetForm{}
	data.Project = &project
	data.Client = &client
	app.render(res, req, http.StatusOK, "timesheet_create.html", data)
}

// timesheetCreatePost handles a POST request with timesheet form data which is then
// validated and used to insert a new timesheet into the database
func (app *application) timesheetCreatePost(res http.ResponseWriter, req *http.Request) {
	projectID, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || projectID < 0 {
		http.NotFound(res, req)
		return
	}

	// Check if project exists
	project, err := app.projects.Get(projectID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get the client for context
	client, err := app.clients.Get(project.ClientID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	var form timesheetForm
	err = app.decodePostForm(req, &form)
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	err = app.formDecoder.Decode(&form, req.PostForm)
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.WorkDate), "work_date", "Work date is required")
	form.CheckField(validator.NotBlank(form.HoursWorked), "hours_worked", "Hours worked is required")
	form.CheckField(validator.MaxChars(form.Description, NAME_LENGTH), "description", fmt.Sprintf("Description must be shorter than %d characters", NAME_LENGTH))

	// Parse and validate work date
	var workDate time.Time
	if form.Valid() {
		workDate, err = time.Parse("2006-01-02", form.WorkDate)
		if err != nil {
			form.AddFieldError("work_date", "Work date must be in YYYY-MM-DD format")
		}
	}

	// Parse and validate hours worked
	var hoursWorked float64
	if form.Valid() {
		hoursWorked, err = strconv.ParseFloat(form.HoursWorked, 64)
		if err != nil || hoursWorked < 0 {
			form.AddFieldError("hours_worked", "Hours worked must be a positive number")
		}
	}

	if !form.Valid() {
		data := app.newTemplateData(req)
		data.Form = form
		data.Project = &project
		data.Client = &client
		app.render(res, req, http.StatusUnprocessableEntity, "timesheet_create.html", data)
		return
	}

	_, err = app.timesheets.Insert(projectID, workDate, hoursWorked, form.Description)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	http.Redirect(res, req, fmt.Sprintf("/project/view/%d", projectID), http.StatusSeeOther)
}

// timesheetUpdate handles a GET request which returns a timesheet update form pre-populated with timesheet data
func (app *application) timesheetUpdate(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	timesheet, err := app.timesheets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get the project for context
	project, err := app.projects.Get(timesheet.ProjectID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get the client for context
	client, err := app.clients.Get(project.ClientID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	data := app.newTemplateData(req)
	data.Form = timesheetForm{
		WorkDate:    timesheet.WorkDate.Format("2006-01-02"),
		HoursWorked: fmt.Sprintf("%.2f", timesheet.HoursWorked),
		Description: timesheet.Description,
	}
	data.Project = &project
	data.Client = &client
	app.render(res, req, http.StatusOK, "timesheet_create.html", data)
}

// timesheetUpdatePost handles a POST request with timesheet form data which is then
// validated and used to update an existing timesheet in the database
func (app *application) timesheetUpdatePost(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	// Get the timesheet to ensure it exists and get the project ID
	timesheet, err := app.timesheets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get project and client for context
	project, err := app.projects.Get(timesheet.ProjectID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	client, err := app.clients.Get(project.ClientID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	var form timesheetForm
	err = app.decodePostForm(req, &form)
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	err = app.formDecoder.Decode(&form, req.PostForm)
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.WorkDate), "work_date", "Work date is required")
	form.CheckField(validator.NotBlank(form.HoursWorked), "hours_worked", "Hours worked is required")
	form.CheckField(validator.MaxChars(form.Description, NAME_LENGTH), "description", fmt.Sprintf("Description must be shorter than %d characters", NAME_LENGTH))

	// Parse and validate work date
	var workDate time.Time
	if form.Valid() {
		workDate, err = time.Parse("2006-01-02", form.WorkDate)
		if err != nil {
			form.AddFieldError("work_date", "Work date must be in YYYY-MM-DD format")
		}
	}

	// Parse and validate hours worked
	var hoursWorked float64
	if form.Valid() {
		hoursWorked, err = strconv.ParseFloat(form.HoursWorked, 64)
		if err != nil || hoursWorked < 0 {
			form.AddFieldError("hours_worked", "Hours worked must be a positive number")
		}
	}

	if !form.Valid() {
		data := app.newTemplateData(req)
		data.Form = form
		data.Project = &project
		data.Client = &client
		app.render(res, req, http.StatusUnprocessableEntity, "timesheet_create.html", data)
		return
	}

	err = app.timesheets.Update(id, workDate, hoursWorked, form.Description)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	http.Redirect(res, req, fmt.Sprintf("/project/view/%d", timesheet.ProjectID), http.StatusSeeOther)
}

// timesheetDelete handles a POST request to soft delete a timesheet
func (app *application) timesheetDelete(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	// Check if timesheet exists before deleting and get project ID for redirect
	timesheet, err := app.timesheets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	err = app.timesheets.Delete(id)
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	// Redirect to project view page after successful deletion
	http.Redirect(res, req, fmt.Sprintf("/project/view/%d", timesheet.ProjectID), http.StatusSeeOther)
}
