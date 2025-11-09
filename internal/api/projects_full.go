package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

// ProjectHandlersFull handles project API endpoints
type ProjectHandlersFull struct {
	projects   *models.ProjectModel
	timesheets *models.TimesheetModel
	invoices   *models.InvoiceModel
}

// NewProjectHandlersFull creates a new ProjectHandlersFull instance
func NewProjectHandlersFull(projects *models.ProjectModel, timesheets *models.TimesheetModel, invoices *models.InvoiceModel) *ProjectHandlersFull {
	return &ProjectHandlersFull{
		projects:   projects,
		timesheets: timesheets,
		invoices:   invoices,
	}
}

// ProjectResponseFull represents a project in API responses with full details
type ProjectResponseFull struct {
	ID                     int      `json:"id"`
	Name                   string   `json:"name"`
	ClientID               int      `json:"clientId"`
	ClientName             string   `json:"clientName,omitempty"`
	Status                 string   `json:"status"`
	HourlyRate             float64  `json:"hourlyRate"`
	Deadline               *string  `json:"deadline"`
	ScheduledStart         *string  `json:"scheduledStart"`
	InvoiceCCEmail         string   `json:"invoiceCCEmail"`
	InvoiceCCDescription   string   `json:"invoiceCCDescription"`
	ScheduleComments       string   `json:"scheduleComments"`
	AdditionalInfo         string   `json:"additionalInfo"`
	AdditionalInfo2        string   `json:"additionalInfo2"`
	DiscountPercent        *float64 `json:"discountPercent"`
	DiscountReason         string   `json:"discountReason"`
	AdjustmentAmount       *float64 `json:"adjustmentAmount"`
	AdjustmentReason       string   `json:"adjustmentReason"`
	CurrencyDisplay        string   `json:"currencyDisplay"`
	CurrencyConversionRate float64  `json:"currencyConversionRate"`
	FlatFeeInvoice         bool     `json:"flatFeeInvoice"`
	Notes                  string   `json:"notes"`
	Created                string   `json:"created"`
	Updated                string   `json:"updated"`
}

// ProjectRequest represents the request body for creating/updating projects
type ProjectRequest struct {
	Name                   string   `json:"name"`
	ClientID               int      `json:"clientId"`
	Status                 string   `json:"status"`
	HourlyRate             float64  `json:"hourlyRate"`
	Deadline               *string  `json:"deadline"`
	ScheduledStart         *string  `json:"scheduledStart"`
	InvoiceCCEmail         string   `json:"invoiceCCEmail"`
	InvoiceCCDescription   string   `json:"invoiceCCDescription"`
	ScheduleComments       string   `json:"scheduleComments"`
	AdditionalInfo         string   `json:"additionalInfo"`
	AdditionalInfo2        string   `json:"additionalInfo2"`
	DiscountPercent        *float64 `json:"discountPercent"`
	DiscountReason         string   `json:"discountReason"`
	AdjustmentAmount       *float64 `json:"adjustmentAmount"`
	AdjustmentReason       string   `json:"adjustmentReason"`
	CurrencyDisplay        string   `json:"currencyDisplay"`
	CurrencyConversionRate float64  `json:"currencyConversionRate"`
	FlatFeeInvoice         bool     `json:"flatFeeInvoice"`
	Notes                  string   `json:"notes"`
}

// UpdateStatusRequest represents the request to update project status
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// ListProjects returns a paginated list of projects
// GET /api/v1/projects?page=1&pageSize=20&search=term
func (h *ProjectHandlersFull) ListProjects(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	search := r.URL.Query().Get("search")

	var projects []models.ProjectWithClient
	var total int64
	var err error

	// Calculate offset
	offset := int64((page - 1) * pageSize)

	if search != "" {
		// Search projects
		projects, err = h.projects.SearchWithPagination(search, int64(pageSize), offset)
		if err != nil {
			ErrorInternal(w)
			return
		}
		total, err = h.projects.SearchCount(search)
		if err != nil {
			ErrorInternal(w)
			return
		}
	} else {
		// Get all projects with pagination
		projects, err = h.projects.GetWithPagination(int64(pageSize), offset)
		if err != nil {
			ErrorInternal(w)
			return
		}
		total, err = h.projects.GetCount()
		if err != nil {
			ErrorInternal(w)
			return
		}
	}

	// Convert to response format
	response := make([]ProjectResponseFull, len(projects))
	for i, project := range projects {
		response[i] = projectWithClientToResponse(&project)
	}

	WritePaginatedJSON(w, response, page, pageSize, int(total))
}

// GetProject returns a single project by ID
// GET /api/v1/projects/:id
func (h *ProjectHandlersFull) GetProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid project ID")
		return
	}

	project, err := h.projects.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Project not found")
			return
		}
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, projectToResponseFull(&project), nil)
}

// CreateProject creates a new project
// POST /api/v1/projects
func (h *ProjectHandlersFull) CreateProject(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.Name), "name", "Name is required")
	v.CheckField(req.ClientID > 0, "clientId", "Valid client ID is required")
	v.CheckField(validator.NotBlank(req.Status), "status", "Status is required")
	v.CheckField(req.HourlyRate >= 0, "hourlyRate", "Hourly rate must be non-negative")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Parse dates
	var deadline, scheduledStart *time.Time
	if req.Deadline != nil {
		t, err := time.Parse("2006-01-02", *req.Deadline)
		if err != nil {
			v.AddFieldError("deadline", "Invalid date format (use YYYY-MM-DD)")
			ErrorValidation(w, *v)
			return
		}
		deadline = &t
	}
	if req.ScheduledStart != nil {
		t, err := time.Parse("2006-01-02", *req.ScheduledStart)
		if err != nil {
			v.AddFieldError("scheduledStart", "Invalid date format (use YYYY-MM-DD)")
			ErrorValidation(w, *v)
			return
		}
		scheduledStart = &t
	}

	// Create project
	project := models.Project{
		Name:                   req.Name,
		ClientID:               req.ClientID,
		Status:                 req.Status,
		HourlyRate:             req.HourlyRate,
		Deadline:               deadline,
		ScheduledStart:         scheduledStart,
		InvoiceCCEmail:         req.InvoiceCCEmail,
		InvoiceCCDescription:   req.InvoiceCCDescription,
		ScheduleComments:       req.ScheduleComments,
		AdditionalInfo:         req.AdditionalInfo,
		AdditionalInfo2:        req.AdditionalInfo2,
		DiscountPercent:        req.DiscountPercent,
		DiscountReason:         req.DiscountReason,
		AdjustmentAmount:       req.AdjustmentAmount,
		AdjustmentReason:       req.AdjustmentReason,
		CurrencyDisplay:        req.CurrencyDisplay,
		CurrencyConversionRate: req.CurrencyConversionRate,
		FlatFeeInvoice:         req.FlatFeeInvoice,
		Notes:                  req.Notes,
	}

	id, err := h.projects.Insert(project)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return created project
	createdProject, err := h.projects.Get(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusCreated, projectToResponseFull(&createdProject), nil)
}

// UpdateProject updates an existing project
// PUT /api/v1/projects/:id
func (h *ProjectHandlersFull) UpdateProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid project ID")
		return
	}

	// Check if project exists
	existing, err := h.projects.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Project not found")
			return
		}
		ErrorInternal(w)
		return
	}

	var req ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.Name), "name", "Name is required")
	v.CheckField(req.ClientID > 0, "clientId", "Valid client ID is required")
	v.CheckField(validator.NotBlank(req.Status), "status", "Status is required")
	v.CheckField(req.HourlyRate >= 0, "hourlyRate", "Hourly rate must be non-negative")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Parse dates
	var deadline, scheduledStart *time.Time
	if req.Deadline != nil {
		t, err := time.Parse("2006-01-02", *req.Deadline)
		if err != nil {
			v.AddFieldError("deadline", "Invalid date format (use YYYY-MM-DD)")
			ErrorValidation(w, *v)
			return
		}
		deadline = &t
	}
	if req.ScheduledStart != nil {
		t, err := time.Parse("2006-01-02", *req.ScheduledStart)
		if err != nil {
			v.AddFieldError("scheduledStart", "Invalid date format (use YYYY-MM-DD)")
			ErrorValidation(w, *v)
			return
		}
		scheduledStart = &t
	}

	// Update project
	existing.Name = req.Name
	existing.ClientID = req.ClientID
	existing.Status = req.Status
	existing.HourlyRate = req.HourlyRate
	existing.Deadline = deadline
	existing.ScheduledStart = scheduledStart
	existing.InvoiceCCEmail = req.InvoiceCCEmail
	existing.InvoiceCCDescription = req.InvoiceCCDescription
	existing.ScheduleComments = req.ScheduleComments
	existing.AdditionalInfo = req.AdditionalInfo
	existing.AdditionalInfo2 = req.AdditionalInfo2
	existing.DiscountPercent = req.DiscountPercent
	existing.DiscountReason = req.DiscountReason
	existing.AdjustmentAmount = req.AdjustmentAmount
	existing.AdjustmentReason = req.AdjustmentReason
	existing.CurrencyDisplay = req.CurrencyDisplay
	existing.CurrencyConversionRate = req.CurrencyConversionRate
	existing.FlatFeeInvoice = req.FlatFeeInvoice
	existing.Notes = req.Notes

	err = h.projects.Update(existing)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return updated project
	updatedProject, err := h.projects.Get(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, projectToResponseFull(&updatedProject), nil)
}

// UpdateProjectStatus updates only the project status
// PATCH /api/v1/projects/:id/status
func (h *ProjectHandlersFull) UpdateProjectStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid project ID")
		return
	}

	// Check if project exists
	existing, err := h.projects.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Project not found")
			return
		}
		ErrorInternal(w)
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.Status), "status", "Status is required")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Update status
	existing.Status = req.Status
	err = h.projects.Update(existing)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Project status updated successfully",
		"status":  req.Status,
	}, nil)
}

// DeleteProject soft deletes a project
// DELETE /api/v1/projects/:id
func (h *ProjectHandlersFull) DeleteProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid project ID")
		return
	}

	// Check if project exists
	_, err = h.projects.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Project not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Delete project
	err = h.projects.Delete(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Project deleted successfully",
	}, nil)
}

// GetProjectTimesheets returns all timesheets for a project
// GET /api/v1/projects/:id/timesheets
func (h *ProjectHandlersFull) GetProjectTimesheets(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid project ID")
		return
	}

	// Check if project exists
	_, err = h.projects.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Project not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Get timesheets for project
	timesheets, err := h.timesheets.GetByProject(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Convert to response format (defined in timesheets.go)
	response := make([]TimesheetResponse, len(timesheets))
	for i, timesheet := range timesheets {
		response[i] = timesheetToResponse(&timesheet)
	}

	WriteJSON(w, http.StatusOK, response, nil)
}

// GetProjectInvoices returns all invoices for a project
// GET /api/v1/projects/:id/invoices
func (h *ProjectHandlersFull) GetProjectInvoices(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid project ID")
		return
	}

	// Check if project exists
	_, err = h.projects.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Project not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Get invoices for project
	invoices, err := h.invoices.GetByProject(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Convert to response format (defined in invoices.go)
	response := make([]InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		response[i] = invoiceToResponse(&invoice)
	}

	WriteJSON(w, http.StatusOK, response, nil)
}

// projectToResponseFull converts a models.Project to a ProjectResponseFull
func projectToResponseFull(p *models.Project) ProjectResponseFull {
	var deadline, scheduledStart *string
	if p.Deadline != nil {
		d := p.Deadline.Format("2006-01-02")
		deadline = &d
	}
	if p.ScheduledStart != nil {
		s := p.ScheduledStart.Format("2006-01-02")
		scheduledStart = &s
	}

	return ProjectResponseFull{
		ID:                     p.ID,
		Name:                   p.Name,
		ClientID:               p.ClientID,
		Status:                 p.Status,
		HourlyRate:             p.HourlyRate,
		Deadline:               deadline,
		ScheduledStart:         scheduledStart,
		InvoiceCCEmail:         p.InvoiceCCEmail,
		InvoiceCCDescription:   p.InvoiceCCDescription,
		ScheduleComments:       p.ScheduleComments,
		AdditionalInfo:         p.AdditionalInfo,
		AdditionalInfo2:        p.AdditionalInfo2,
		DiscountPercent:        p.DiscountPercent,
		DiscountReason:         p.DiscountReason,
		AdjustmentAmount:       p.AdjustmentAmount,
		AdjustmentReason:       p.AdjustmentReason,
		CurrencyDisplay:        p.CurrencyDisplay,
		CurrencyConversionRate: p.CurrencyConversionRate,
		FlatFeeInvoice:         p.FlatFeeInvoice,
		Notes:                  p.Notes,
		Created:                p.Created.Format("2006-01-02T15:04:05Z"),
		Updated:                p.Updated.Format("2006-01-02T15:04:05Z"),
	}
}

// projectWithClientToResponse converts a models.ProjectWithClient to a ProjectResponseFull
func projectWithClientToResponse(p *models.ProjectWithClient) ProjectResponseFull {
	var deadline, scheduledStart *string
	if p.Deadline != nil {
		d := p.Deadline.Format("2006-01-02")
		deadline = &d
	}
	if p.ScheduledStart != nil {
		s := p.ScheduledStart.Format("2006-01-02")
		scheduledStart = &s
	}

	return ProjectResponseFull{
		ID:                     p.ID,
		Name:                   p.Name,
		ClientID:               p.ClientID,
		ClientName:             p.ClientName,
		Status:                 p.Status,
		HourlyRate:             p.HourlyRate,
		Deadline:               deadline,
		ScheduledStart:         scheduledStart,
		InvoiceCCEmail:         p.InvoiceCCEmail,
		InvoiceCCDescription:   p.InvoiceCCDescription,
		ScheduleComments:       p.ScheduleComments,
		AdditionalInfo:         p.AdditionalInfo,
		AdditionalInfo2:        p.AdditionalInfo2,
		DiscountPercent:        p.DiscountPercent,
		DiscountReason:         p.DiscountReason,
		AdjustmentAmount:       p.AdjustmentAmount,
		AdjustmentReason:       p.AdjustmentReason,
		CurrencyDisplay:        p.CurrencyDisplay,
		CurrencyConversionRate: p.CurrencyConversionRate,
		FlatFeeInvoice:         p.FlatFeeInvoice,
		Notes:                  p.Notes,
		Created:                p.Created.Format("2006-01-02T15:04:05Z"),
		Updated:                p.Updated.Format("2006-01-02T15:04:05Z"),
	}
}
