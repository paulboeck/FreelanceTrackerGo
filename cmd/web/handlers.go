package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

const NAME_LENGTH = 255

// use `form:"-"` so the go-playground form library will ignore that attribute
// when parsing a request and populating a form struct
type clientForm struct {
	Name                    string `form:"name"`
	Email                   string `form:"email"`
	Phone                   string `form:"phone"`
	Address1                string `form:"address1"`
	Address2                string `form:"address2"`
	Address3                string `form:"address3"`
	City                    string `form:"city"`
	State                   string `form:"state"`
	ZipCode                 string `form:"zip_code"`
	HourlyRate              string `form:"hourly_rate"`
	Notes                   string `form:"notes"`
	AdditionalInfo          string `form:"additional_info"`
	AdditionalInfo2         string `form:"additional_info2"`
	BillTo                  string `form:"bill_to"`
	IncludeAddressOnInvoice bool   `form:"include_address_on_invoice"`
	InvoiceCCEmail          string `form:"invoice_cc_email"`
	InvoiceCCDescription    string `form:"invoice_cc_description"`
	UniversityAffiliation   string `form:"university_affiliation"`
	validator.Validator     `form:"-"`
}

type projectForm struct {
	Name                   string `form:"name"`
	Status                 string `form:"status"`
	HourlyRate             string `form:"hourly_rate"`
	Deadline               string `form:"deadline"`
	ScheduledStart         string `form:"scheduled_start"`
	InvoiceCCEmail         string `form:"invoice_cc_email"`
	InvoiceCCDescription   string `form:"invoice_cc_description"`
	ScheduleComments       string `form:"schedule_comments"`
	AdditionalInfo         string `form:"additional_info"`
	AdditionalInfo2        string `form:"additional_info2"`
	DiscountPercent        string `form:"discount_percent"`
	DiscountReason         string `form:"discount_reason"`
	AdjustmentAmount       string `form:"adjustment_amount"`
	AdjustmentReason       string `form:"adjustment_reason"`
	CurrencyDisplay        string `form:"currency_display"`
	CurrencyConversionRate string `form:"currency_conversion_rate"`
	FlatFeeInvoice         bool   `form:"flat_fee_invoice"`
	Notes                  string `form:"notes"`
	validator.Validator    `form:"-"`
}

type timesheetForm struct {
	WorkDate            string `form:"work_date"`
	HoursWorked         string `form:"hours_worked"`
	HourlyRate          string `form:"hourly_rate"`
	Description         string `form:"description"`
	IsUpdate            bool   `form:"-"`
	validator.Validator `form:"-"`
}

type invoiceForm struct {
	InvoiceDate         string `form:"invoice_date"`
	DatePaid            string `form:"date_paid"`
	PaymentTerms        string `form:"payment_terms"`
	AmountDue           string `form:"amount_due"`
	DisplayDetails      bool   `form:"display_details"`
	validator.Validator `form:"-"`
}

type settingsForm struct {
	Settings            map[string]string `form:"-"`
	validator.Validator `form:"-"`
}

// home handles http requests to the root URl of the project
func (app *application) home(res http.ResponseWriter, req *http.Request) {
	// Get page size setting with fallback
	pageSize := 10 // Default fallback
	if pageSizeSetting, err := app.settings.GetString("list_page_size"); err == nil {
		if ps, err := strconv.Atoi(pageSizeSetting); err == nil && ps > 0 {
			pageSize = ps
		}
	}
	
	// Get current page from query parameter
	currentPage := 1
	if pageParam := req.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			currentPage = p
		}
	}
	
	// Calculate offset
	offset := int64((currentPage - 1) * pageSize)
	
	// Get paginated clients and total count
	clients, err := app.clients.GetWithPagination(int64(pageSize), offset)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	
	totalCount, err := app.clients.GetCount()
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	
	// Calculate pagination info
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	
	pagination := &paginationData{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		HasPrev:     currentPage > 1,
		HasNext:     currentPage < totalPages,
		PrevPage:    currentPage - 1,
		NextPage:    currentPage + 1,
		PageSize:    pageSize,
	}

	data := app.newTemplateData(req)
	data.Clients = clients
	data.Pagination = pagination

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

	// Get invoices for this project
	invoices, err := app.invoices.GetByProject(id)
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	data := app.newTemplateData(req)
	data.Project = &project
	data.Client = &client
	data.Timesheets = timesheets
	data.Invoices = invoices

	app.render(res, req, http.StatusOK, "project.html", data)
}

// clientCreate handles a GET request which returns an empty client detail form
func (app *application) clientCreate(res http.ResponseWriter, req *http.Request) {
	data := app.newTemplateData(req)
	data.Form = clientForm{
		IncludeAddressOnInvoice: true, // Default to checked
	}
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

	form.CheckField(validator.NotBlank(form.Name), "name", "Name is required")
	form.CheckField(validator.MaxChars(form.Name, NAME_LENGTH), "name", fmt.Sprintf("Name must be shorter than %d characters", NAME_LENGTH))
	
	form.CheckField(validator.NotBlank(form.Email), "email", "Email is required")
	form.CheckField(validator.Matches(strings.ToLower(form.Email), validator.EmailRegex), "email", "Email must be a valid email address")
	form.CheckField(validator.MaxChars(form.Email, NAME_LENGTH), "email", fmt.Sprintf("Email must be shorter than %d characters", NAME_LENGTH))

	// Parse hourly rate
	hourlyRate, err := strconv.ParseFloat(form.HourlyRate, 64)
	if form.HourlyRate == "" {
		form.CheckField(false, "hourly_rate", "Hourly rate is required")
	} else if err != nil {
		form.CheckField(false, "hourly_rate", "Hourly rate must be a valid number")
	} else {
		form.CheckField(hourlyRate >= 0, "hourly_rate", "Hourly rate must be 0 or greater")
	}

	// Validate optional email fields
	if form.InvoiceCCEmail != "" {
		form.CheckField(validator.Matches(strings.ToLower(form.InvoiceCCEmail), validator.EmailRegex), "invoice_cc_email", "Invoice CC email must be a valid email address")
	}

	// Validate optional field lengths  
	form.CheckField(validator.MaxChars(form.Phone, NAME_LENGTH), "phone", fmt.Sprintf("Phone must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.Address1, NAME_LENGTH), "address1", fmt.Sprintf("Address 1 must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.Address2, NAME_LENGTH), "address2", fmt.Sprintf("Address 2 must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.Address3, NAME_LENGTH), "address3", fmt.Sprintf("Address 3 must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.City, NAME_LENGTH), "city", fmt.Sprintf("City must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.State, 50), "state", "State must be shorter than 50 characters")
	form.CheckField(validator.MaxChars(form.ZipCode, 20), "zip_code", "Zip code must be shorter than 20 characters")
	form.CheckField(validator.MaxChars(form.Notes, 2000), "notes", "Notes must be shorter than 2000 characters")
	form.CheckField(validator.MaxChars(form.AdditionalInfo, NAME_LENGTH), "additional_info", fmt.Sprintf("Additional info must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.AdditionalInfo2, NAME_LENGTH), "additional_info2", fmt.Sprintf("Additional info 2 must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.BillTo, NAME_LENGTH), "bill_to", fmt.Sprintf("Bill to must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.InvoiceCCDescription, 500), "invoice_cc_description", "Invoice CC description must be shorter than 500 characters")
	form.CheckField(validator.MaxChars(form.UniversityAffiliation, NAME_LENGTH), "university_affiliation", fmt.Sprintf("University affiliation must be shorter than %d characters", NAME_LENGTH))

	if !form.Valid() {
		data := app.newTemplateData(req)
		data.Form = form
		app.render(res, req, http.StatusUnprocessableEntity, "client_create.html", data)
		return
	}

	// Convert string fields to pointers for optional fields
	var phone, address1, address2, address3, city, state, zipCode, notes, additionalInfo, additionalInfo2, billTo, invoiceCCEmail, invoiceCCDescription, universityAffiliation *string
	
	if form.Phone != "" {
		phone = &form.Phone
	}
	if form.Address1 != "" {
		address1 = &form.Address1
	}
	if form.Address2 != "" {
		address2 = &form.Address2
	}
	if form.Address3 != "" {
		address3 = &form.Address3
	}
	if form.City != "" {
		city = &form.City
	}
	if form.State != "" {
		state = &form.State
	}
	if form.ZipCode != "" {
		zipCode = &form.ZipCode
	}
	if form.Notes != "" {
		notes = &form.Notes
	}
	if form.AdditionalInfo != "" {
		additionalInfo = &form.AdditionalInfo
	}
	if form.AdditionalInfo2 != "" {
		additionalInfo2 = &form.AdditionalInfo2
	}
	if form.BillTo != "" {
		billTo = &form.BillTo
	}
	if form.InvoiceCCEmail != "" {
		invoiceCCEmail = &form.InvoiceCCEmail
	}
	if form.InvoiceCCDescription != "" {
		invoiceCCDescription = &form.InvoiceCCDescription
	}
	if form.UniversityAffiliation != "" {
		universityAffiliation = &form.UniversityAffiliation
	}

	id, err := app.clients.Insert(
		form.Name, 
		form.Email, 
		phone, 
		address1, 
		address2, 
		address3, 
		city, 
		state, 
		zipCode, 
		hourlyRate, 
		notes, 
		additionalInfo, 
		additionalInfo2, 
		billTo, 
		form.IncludeAddressOnInvoice, 
		invoiceCCEmail, 
		invoiceCCDescription, 
		universityAffiliation,
	)
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
		Name:                    client.Name,
		Email:                   client.Email,
		Phone:                   ptrToString(client.Phone),
		Address1:                ptrToString(client.Address1),
		Address2:                ptrToString(client.Address2),
		Address3:                ptrToString(client.Address3),
		City:                    ptrToString(client.City),
		State:                   ptrToString(client.State),
		ZipCode:                 ptrToString(client.ZipCode),
		HourlyRate:              fmt.Sprintf("%.2f", client.HourlyRate),
		Notes:                   ptrToString(client.Notes),
		AdditionalInfo:          ptrToString(client.AdditionalInfo),
		AdditionalInfo2:         ptrToString(client.AdditionalInfo2),
		BillTo:                  ptrToString(client.BillTo),
		IncludeAddressOnInvoice: client.IncludeAddressOnInvoice,
		InvoiceCCEmail:          ptrToString(client.InvoiceCCEmail),
		InvoiceCCDescription:    ptrToString(client.InvoiceCCDescription),
		UniversityAffiliation:   ptrToString(client.UniversityAffiliation),
	}
	data.Client = &client
	app.render(res, req, http.StatusOK, "client_create.html", data)
}

// Helper function to convert *string to string
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// formToProject converts a projectForm to a models.Project struct
func formToProject(form projectForm, clientID, projectID int) (models.Project, error) {
	// Parse dates
	var deadline *time.Time
	if form.Deadline != "" {
		if d, err := time.Parse("2006-01-02", form.Deadline); err == nil {
			deadline = &d
		}
	}
	
	var scheduledStart *time.Time
	if form.ScheduledStart != "" {
		if d, err := time.Parse("2006-01-02", form.ScheduledStart); err == nil {
			scheduledStart = &d
		}
	}
	
	// Parse hourly rate
	hourlyRate, err := strconv.ParseFloat(form.HourlyRate, 64)
	if err != nil {
		return models.Project{}, fmt.Errorf("invalid hourly rate: %w", err)
	}
	
	// Parse discount percent
	var discountPercent *float64
	if form.DiscountPercent != "" {
		if dp, err := strconv.ParseFloat(form.DiscountPercent, 64); err == nil {
			discountPercent = &dp
		}
	}
	
	// Parse adjustment amount
	var adjustmentAmount *float64
	if form.AdjustmentAmount != "" {
		if aa, err := strconv.ParseFloat(form.AdjustmentAmount, 64); err == nil {
			adjustmentAmount = &aa
		}
	}
	
	// Parse currency conversion rate
	currencyConversionRate := 1.0
	if form.CurrencyConversionRate != "" {
		if ccr, err := strconv.ParseFloat(form.CurrencyConversionRate, 64); err == nil {
			currencyConversionRate = ccr
		}
	}
	
	// Set defaults
	currencyDisplay := form.CurrencyDisplay
	if currencyDisplay == "" {
		currencyDisplay = "USD"
	}
	
	return models.Project{
		ID:                     projectID,
		Name:                   form.Name,
		ClientID:               clientID,
		Status:                 form.Status,
		HourlyRate:             hourlyRate,
		Deadline:               deadline,
		ScheduledStart:         scheduledStart,
		InvoiceCCEmail:         form.InvoiceCCEmail,
		InvoiceCCDescription:   form.InvoiceCCDescription,
		ScheduleComments:       form.ScheduleComments,
		AdditionalInfo:         form.AdditionalInfo,
		AdditionalInfo2:        form.AdditionalInfo2,
		DiscountPercent:        discountPercent,
		DiscountReason:         form.DiscountReason,
		AdjustmentAmount:       adjustmentAmount,
		AdjustmentReason:       form.AdjustmentReason,
		CurrencyDisplay:        currencyDisplay,
		CurrencyConversionRate: currencyConversionRate,
		FlatFeeInvoice:         form.FlatFeeInvoice,
		Notes:                  form.Notes,
	}, nil
}

// projectToForm converts a models.Project to a projectForm struct
func projectToForm(project models.Project) projectForm {
	// Helper to format dates
	formatDate := func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return t.Format("2006-01-02")
	}
	
	// Helper to format float pointers
	formatFloatPtr := func(f *float64) string {
		if f == nil {
			return ""
		}
		return fmt.Sprintf("%.4f", *f)
	}
	
	return projectForm{
		Name:                   project.Name,
		Status:                 project.Status,
		HourlyRate:             fmt.Sprintf("%.2f", project.HourlyRate),
		Deadline:               formatDate(project.Deadline),
		ScheduledStart:         formatDate(project.ScheduledStart),
		InvoiceCCEmail:         project.InvoiceCCEmail,
		InvoiceCCDescription:   project.InvoiceCCDescription,
		ScheduleComments:       project.ScheduleComments,
		AdditionalInfo:         project.AdditionalInfo,
		AdditionalInfo2:        project.AdditionalInfo2,
		DiscountPercent:        formatFloatPtr(project.DiscountPercent),
		DiscountReason:         project.DiscountReason,
		AdjustmentAmount:       formatFloatPtr(project.AdjustmentAmount),
		AdjustmentReason:       project.AdjustmentReason,
		CurrencyDisplay:        project.CurrencyDisplay,
		CurrencyConversionRate: fmt.Sprintf("%.5f", project.CurrencyConversionRate),
		FlatFeeInvoice:         project.FlatFeeInvoice,
		Notes:                  project.Notes,
	}
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

	form.CheckField(validator.NotBlank(form.Name), "name", "Name is required")
	form.CheckField(validator.MaxChars(form.Name, NAME_LENGTH), "name", fmt.Sprintf("Name must be shorter than %d characters", NAME_LENGTH))
	
	form.CheckField(validator.NotBlank(form.Email), "email", "Email is required")
	form.CheckField(validator.Matches(strings.ToLower(form.Email), validator.EmailRegex), "email", "Email must be a valid email address")
	form.CheckField(validator.MaxChars(form.Email, NAME_LENGTH), "email", fmt.Sprintf("Email must be shorter than %d characters", NAME_LENGTH))

	// Parse hourly rate
	hourlyRate, err := strconv.ParseFloat(form.HourlyRate, 64)
	if form.HourlyRate == "" {
		form.CheckField(false, "hourly_rate", "Hourly rate is required")
	} else if err != nil {
		form.CheckField(false, "hourly_rate", "Hourly rate must be a valid number")
	} else {
		form.CheckField(hourlyRate >= 0, "hourly_rate", "Hourly rate must be 0 or greater")
	}

	// Validate optional email fields
	if form.InvoiceCCEmail != "" {
		form.CheckField(validator.Matches(strings.ToLower(form.InvoiceCCEmail), validator.EmailRegex), "invoice_cc_email", "Invoice CC email must be a valid email address")
	}

	// Validate optional field lengths  
	form.CheckField(validator.MaxChars(form.Phone, NAME_LENGTH), "phone", fmt.Sprintf("Phone must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.Address1, NAME_LENGTH), "address1", fmt.Sprintf("Address 1 must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.Address2, NAME_LENGTH), "address2", fmt.Sprintf("Address 2 must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.Address3, NAME_LENGTH), "address3", fmt.Sprintf("Address 3 must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.City, NAME_LENGTH), "city", fmt.Sprintf("City must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.State, 50), "state", "State must be shorter than 50 characters")
	form.CheckField(validator.MaxChars(form.ZipCode, 20), "zip_code", "Zip code must be shorter than 20 characters")
	form.CheckField(validator.MaxChars(form.Notes, 2000), "notes", "Notes must be shorter than 2000 characters")
	form.CheckField(validator.MaxChars(form.AdditionalInfo, NAME_LENGTH), "additional_info", fmt.Sprintf("Additional info must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.AdditionalInfo2, NAME_LENGTH), "additional_info2", fmt.Sprintf("Additional info 2 must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.BillTo, NAME_LENGTH), "bill_to", fmt.Sprintf("Bill to must be shorter than %d characters", NAME_LENGTH))
	form.CheckField(validator.MaxChars(form.InvoiceCCDescription, 500), "invoice_cc_description", "Invoice CC description must be shorter than 500 characters")
	form.CheckField(validator.MaxChars(form.UniversityAffiliation, NAME_LENGTH), "university_affiliation", fmt.Sprintf("University affiliation must be shorter than %d characters", NAME_LENGTH))

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

	// Convert string fields to pointers for optional fields
	var phone, address1, address2, address3, city, state, zipCode, notes, additionalInfo, additionalInfo2, billTo, invoiceCCEmail, invoiceCCDescription, universityAffiliation *string
	
	if form.Phone != "" {
		phone = &form.Phone
	}
	if form.Address1 != "" {
		address1 = &form.Address1
	}
	if form.Address2 != "" {
		address2 = &form.Address2
	}
	if form.Address3 != "" {
		address3 = &form.Address3
	}
	if form.City != "" {
		city = &form.City
	}
	if form.State != "" {
		state = &form.State
	}
	if form.ZipCode != "" {
		zipCode = &form.ZipCode
	}
	if form.Notes != "" {
		notes = &form.Notes
	}
	if form.AdditionalInfo != "" {
		additionalInfo = &form.AdditionalInfo
	}
	if form.AdditionalInfo2 != "" {
		additionalInfo2 = &form.AdditionalInfo2
	}
	if form.BillTo != "" {
		billTo = &form.BillTo
	}
	if form.InvoiceCCEmail != "" {
		invoiceCCEmail = &form.InvoiceCCEmail
	}
	if form.InvoiceCCDescription != "" {
		invoiceCCDescription = &form.InvoiceCCDescription
	}
	if form.UniversityAffiliation != "" {
		universityAffiliation = &form.UniversityAffiliation
	}

	err = app.clients.Update(
		id, 
		form.Name, 
		form.Email, 
		phone, 
		address1, 
		address2, 
		address3, 
		city, 
		state, 
		zipCode, 
		hourlyRate, 
		notes, 
		additionalInfo, 
		additionalInfo2, 
		billTo, 
		form.IncludeAddressOnInvoice, 
		invoiceCCEmail, 
		invoiceCCDescription, 
		universityAffiliation,
	)
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
	data.Form = projectForm{
		Status:                 "Estimating", // Default status
		HourlyRate:             fmt.Sprintf("%.2f", client.HourlyRate), // Default from client
		InvoiceCCEmail:         ptrToString(client.InvoiceCCEmail),    // Default from client
		InvoiceCCDescription:   ptrToString(client.InvoiceCCDescription), // Default from client
		AdditionalInfo:         ptrToString(client.AdditionalInfo),   // Default from client
		AdditionalInfo2:        ptrToString(client.AdditionalInfo2),  // Default from client
		CurrencyDisplay:        "USD",        // Default currency
		CurrencyConversionRate: "1.00000",    // Default conversion rate
	}
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

	form.CheckField(validator.NotBlank(form.Name), "name", "Name is required")
	form.CheckField(validator.MaxChars(form.Name, NAME_LENGTH), "name", fmt.Sprintf("Name must be shorter than %d characters", NAME_LENGTH))
	
	form.CheckField(validator.NotBlank(form.Status), "status", "Status is required")
	form.CheckField(validator.NotBlank(form.HourlyRate), "hourly_rate", "Hourly rate is required")

	if !form.Valid() {
		data := app.newTemplateData(req)
		data.Form = form
		data.Client = &client
		app.render(res, req, http.StatusUnprocessableEntity, "project_create.html", data)
		return
	}

	// Convert form data to Project struct
	project, err := formToProject(form, clientID, 0)
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	_, err = app.projects.Insert(project)
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
	data.Form = projectToForm(project)
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

	form.CheckField(validator.NotBlank(form.Name), "name", "Name is required")
	form.CheckField(validator.MaxChars(form.Name, NAME_LENGTH), "name", fmt.Sprintf("Name must be shorter than %d characters", NAME_LENGTH))
	
	form.CheckField(validator.NotBlank(form.Status), "status", "Status is required")
	form.CheckField(validator.NotBlank(form.HourlyRate), "hourly_rate", "Hourly rate is required")

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

	// Convert form data to Project struct
	updatedProject, err := formToProject(form, project.ClientID, id)
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	err = app.projects.Update(updatedProject)
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
	data.Form = timesheetForm{
		WorkDate:   time.Now().Format("2006-01-02"),
		HourlyRate: fmt.Sprintf("%.2f", project.HourlyRate), // Default from project
	}
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

	form.CheckField(validator.NotBlank(form.WorkDate), "work_date", "Work date is required")
	form.CheckField(validator.NotBlank(form.HoursWorked), "hours_worked", "Hours worked is required")
	form.CheckField(validator.NotBlank(form.HourlyRate), "hourly_rate", "Hourly rate is required")
	form.CheckField(validator.NotBlank(form.Description), "description", "Description is required")
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

	// Parse and validate hourly rate
	var hourlyRate float64
	if form.Valid() {
		hourlyRate, err = strconv.ParseFloat(form.HourlyRate, 64)
		if err != nil || hourlyRate < 0 {
			form.AddFieldError("hourly_rate", "Hourly rate must be a positive number")
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

	_, err = app.timesheets.Insert(projectID, workDate, hoursWorked, hourlyRate, form.Description)
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
		HourlyRate:  fmt.Sprintf("%.2f", timesheet.HourlyRate),
		Description: timesheet.Description,
		IsUpdate:    true,
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

	form.CheckField(validator.NotBlank(form.WorkDate), "work_date", "Work date is required")
	form.CheckField(validator.NotBlank(form.HoursWorked), "hours_worked", "Hours worked is required")
	form.CheckField(validator.NotBlank(form.HourlyRate), "hourly_rate", "Hourly rate is required")
	form.CheckField(validator.NotBlank(form.Description), "description", "Description is required")
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

	// Parse and validate hourly rate
	var hourlyRate float64
	if form.Valid() {
		hourlyRate, err = strconv.ParseFloat(form.HourlyRate, 64)
		if err != nil || hourlyRate < 0 {
			form.AddFieldError("hourly_rate", "Hourly rate must be a positive number")
		}
	}

	if !form.Valid() {
		form.IsUpdate = true
		data := app.newTemplateData(req)
		data.Form = form
		data.Project = &project
		data.Client = &client
		app.render(res, req, http.StatusUnprocessableEntity, "timesheet_create.html", data)
		return
	}

	err = app.timesheets.Update(id, workDate, hoursWorked, hourlyRate, form.Description)
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

// invoiceCreate handles a GET request which returns an empty invoice creation form
func (app *application) invoiceCreate(res http.ResponseWriter, req *http.Request) {
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
	data.Form = invoiceForm{
		InvoiceDate: time.Now().Format("2006-01-02"),
	}
	data.Project = &project
	data.Client = &client
	app.render(res, req, http.StatusOK, "invoice_create.html", data)
}

// invoiceCreatePost handles a POST request with invoice form data which is then
// validated and used to insert a new invoice into the database
func (app *application) invoiceCreatePost(res http.ResponseWriter, req *http.Request) {
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

	var form invoiceForm
	err = app.decodePostForm(req, &form)
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.InvoiceDate), "invoice_date", "Invoice date is required")
	form.CheckField(validator.NotBlank(form.AmountDue), "amount_due", "Amount due is required")
	form.CheckField(validator.MaxChars(form.PaymentTerms, NAME_LENGTH), "payment_terms", fmt.Sprintf("Payment terms must be shorter than %d characters", NAME_LENGTH))

	// Parse and validate invoice date
	var invoiceDate time.Time
	if form.Valid() {
		invoiceDate, err = time.Parse("2006-01-02", form.InvoiceDate)
		if err != nil {
			form.AddFieldError("invoice_date", "Invoice date must be in YYYY-MM-DD format")
		}
	}

	// Parse and validate amount due
	var amountDue float64
	if form.Valid() {
		amountDue, err = strconv.ParseFloat(form.AmountDue, 64)
		if err != nil || amountDue < 0 {
			form.AddFieldError("amount_due", "Amount due must be a positive number")
		}
	}

	// Parse date paid if provided
	var datePaid *time.Time
	if form.Valid() && form.DatePaid != "" {
		parsedDatePaid, err := time.Parse("2006-01-02", form.DatePaid)
		if err != nil {
			form.AddFieldError("date_paid", "Date paid must be in YYYY-MM-DD format")
		} else {
			datePaid = &parsedDatePaid
		}
	}

	if !form.Valid() {
		data := app.newTemplateData(req)
		data.Form = form
		data.Project = &project
		data.Client = &client
		app.render(res, req, http.StatusUnprocessableEntity, "invoice_create.html", data)
		return
	}

	_, err = app.invoices.Insert(projectID, invoiceDate, datePaid, form.PaymentTerms, amountDue, form.DisplayDetails)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	http.Redirect(res, req, fmt.Sprintf("/project/view/%d", projectID), http.StatusSeeOther)
}

// invoiceUpdate handles a GET request which returns an invoice update form pre-populated with invoice data
func (app *application) invoiceUpdate(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	invoice, err := app.invoices.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get the project for context
	project, err := app.projects.Get(invoice.ProjectID)
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

	var datePaidStr string
	if invoice.DatePaid != nil {
		datePaidStr = invoice.DatePaid.Format("2006-01-02")
	}

	data := app.newTemplateData(req)
	data.Form = invoiceForm{
		InvoiceDate:    invoice.InvoiceDate.Format("2006-01-02"),
		DatePaid:       datePaidStr,
		PaymentTerms:   invoice.PaymentTerms,
		AmountDue:      fmt.Sprintf("%.2f", invoice.AmountDue),
		DisplayDetails: invoice.DisplayDetails,
	}
	data.Project = &project
	data.Client = &client
	app.render(res, req, http.StatusOK, "invoice_create.html", data)
}

// invoiceUpdatePost handles a POST request with invoice form data which is then
// validated and used to update an existing invoice in the database
func (app *application) invoiceUpdatePost(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	// Get the invoice to ensure it exists and get the project ID
	invoice, err := app.invoices.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Get project and client for context
	project, err := app.projects.Get(invoice.ProjectID)
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

	var form invoiceForm
	err = app.decodePostForm(req, &form)
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.InvoiceDate), "invoice_date", "Invoice date is required")
	form.CheckField(validator.NotBlank(form.AmountDue), "amount_due", "Amount due is required")
	form.CheckField(validator.MaxChars(form.PaymentTerms, NAME_LENGTH), "payment_terms", fmt.Sprintf("Payment terms must be shorter than %d characters", NAME_LENGTH))

	// Parse and validate invoice date
	var invoiceDate time.Time
	if form.Valid() {
		invoiceDate, err = time.Parse("2006-01-02", form.InvoiceDate)
		if err != nil {
			form.AddFieldError("invoice_date", "Invoice date must be in YYYY-MM-DD format")
		}
	}

	// Parse and validate amount due
	var amountDue float64
	if form.Valid() {
		amountDue, err = strconv.ParseFloat(form.AmountDue, 64)
		if err != nil || amountDue < 0 {
			form.AddFieldError("amount_due", "Amount due must be a positive number")
		}
	}

	// Parse date paid if provided
	var datePaid *time.Time
	if form.Valid() && form.DatePaid != "" {
		parsedDatePaid, err := time.Parse("2006-01-02", form.DatePaid)
		if err != nil {
			form.AddFieldError("date_paid", "Date paid must be in YYYY-MM-DD format")
		} else {
			datePaid = &parsedDatePaid
		}
	}

	if !form.Valid() {
		data := app.newTemplateData(req)
		data.Form = form
		data.Project = &project
		data.Client = &client
		app.render(res, req, http.StatusUnprocessableEntity, "invoice_create.html", data)
		return
	}

	err = app.invoices.Update(id, invoiceDate, datePaid, form.PaymentTerms, amountDue, form.DisplayDetails)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	http.Redirect(res, req, fmt.Sprintf("/project/view/%d", invoice.ProjectID), http.StatusSeeOther)
}

// invoiceDelete handles a POST request to soft delete an invoice
func (app *application) invoiceDelete(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	// Check if invoice exists before deleting and get project ID for redirect
	invoice, err := app.invoices.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	err = app.invoices.Delete(id)
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	// Redirect to project view page after successful deletion
	http.Redirect(res, req, fmt.Sprintf("/project/view/%d", invoice.ProjectID), http.StatusSeeOther)
}

// invoicePrint handles a GET request to generate and download an invoice PDF
func (app *application) invoicePrint(res http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.NotFound(res, req)
		return
	}

	// Get settings for PDF generation
	allSettings, err := app.settings.GetAll()
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	// Generate PDF with settings
	pdfBytes, err := app.invoices.GeneratePDFWithSettings(id, allSettings)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(res, req)
		} else {
			app.serverError(res, req, err)
		}
		return
	}

	// Set headers for PDF download
	res.Header().Set("Content-Type", "application/pdf")
	res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"invoice_%d.pdf\"", id))
	res.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Write PDF to response
	_, err = res.Write(pdfBytes)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
}

// settingsView handles a GET request to view all application settings
func (app *application) settingsView(res http.ResponseWriter, req *http.Request) {
	settings, err := app.settings.GetAllDetailed()
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	data := app.newTemplateData(req)
	data.Settings = settings

	app.render(res, req, http.StatusOK, "settings.html", data)
}

// settingsEdit handles a GET request to display the settings edit form
func (app *application) settingsEdit(res http.ResponseWriter, req *http.Request) {
	settings, err := app.settings.GetAllDetailed()
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	data := app.newTemplateData(req)
	data.Settings = settings
	data.Form = settingsForm{}

	app.render(res, req, http.StatusOK, "settings_edit.html", data)
}

// settingsEditPost handles a POST request to update application settings
func (app *application) settingsEditPost(res http.ResponseWriter, req *http.Request) {
	// Get current settings to know what to expect
	settings, err := app.settings.GetAllDetailed()
	if err != nil {
		app.serverError(res, req, err)
		return
	}

	// Parse form
	err = req.ParseForm()
	if err != nil {
		app.clientError(res, http.StatusBadRequest)
		return
	}

	var form settingsForm
	form.Settings = make(map[string]string)

	// Extract values from form for each setting
	for _, setting := range settings {
		if value := req.PostForm.Get(setting.Key); value != "" {
			form.Settings[setting.Key] = value
		}
	}

	// Validate each setting based on its data type
	for _, setting := range settings {
		value, exists := form.Settings[setting.Key]
		if !exists {
			form.AddFieldError(setting.Key, "This field is required")
			continue
		}

		switch setting.DataType {
		case "decimal", "float":
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				form.AddFieldError(setting.Key, "Must be a valid number")
			}
		case "int":
			if _, err := strconv.Atoi(value); err != nil {
				form.AddFieldError(setting.Key, "Must be a valid integer")
			}
		case "bool":
			if value != "true" && value != "false" {
				form.AddFieldError(setting.Key, "Must be true or false")
			}
		}
	}

	// If there are validation errors, redisplay the form
	if !form.Valid() {
		data := app.newTemplateData(req)
		data.Settings = settings
		data.Form = form
		app.render(res, req, http.StatusUnprocessableEntity, "settings_edit.html", data)
		return
	}

	// Update each setting value
	for _, setting := range settings {
		if newValue, exists := form.Settings[setting.Key]; exists {
			err = app.settings.UpdateValue(setting.Key, newValue)
			if err != nil {
				app.serverError(res, req, err)
				return
			}
		}
	}

	// Redirect to settings view
	http.Redirect(res, req, "/settings", http.StatusSeeOther)
}

// projectsList handles a GET request which displays all projects
func (app *application) projectsList(res http.ResponseWriter, req *http.Request) {
	// Get page size setting with fallback
	pageSize := 10 // Default fallback
	if pageSizeSetting, err := app.settings.GetString("list_page_size"); err == nil {
		if ps, err := strconv.Atoi(pageSizeSetting); err == nil && ps > 0 {
			pageSize = ps
		}
	}
	
	// Get current page from query parameter
	currentPage := 1
	if pageParam := req.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			currentPage = p
		}
	}
	
	// Calculate offset
	offset := int64((currentPage - 1) * pageSize)
	
	// Get paginated projects and total count
	projects, err := app.projects.GetWithPagination(int64(pageSize), offset)
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	
	totalCount, err := app.projects.GetCount()
	if err != nil {
		app.serverError(res, req, err)
		return
	}
	
	// Calculate pagination info
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	
	pagination := &paginationData{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		HasPrev:     currentPage > 1,
		HasNext:     currentPage < totalPages,
		PrevPage:    currentPage - 1,
		NextPage:    currentPage + 1,
		PageSize:    pageSize,
	}

	data := app.newTemplateData(req)
	data.ProjectsWithClient = projects
	data.Pagination = pagination
	app.render(res, req, http.StatusOK, "projects.html", data)
}
