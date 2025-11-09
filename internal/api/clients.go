package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

// ClientHandlers handles client API endpoints
type ClientHandlers struct {
	clients  *models.ClientModel
	projects *models.ProjectModel
}

// NewClientHandlers creates a new ClientHandlers instance
func NewClientHandlers(clients *models.ClientModel, projects *models.ProjectModel) *ClientHandlers {
	return &ClientHandlers{
		clients:  clients,
		projects: projects,
	}
}

// ClientResponse represents a client in API responses
type ClientResponse struct {
	ID                      int     `json:"id"`
	Name                    string  `json:"name"`
	Email                   string  `json:"email"`
	Phone                   *string `json:"phone"`
	Address1                *string `json:"address1"`
	Address2                *string `json:"address2"`
	Address3                *string `json:"address3"`
	City                    *string `json:"city"`
	State                   *string `json:"state"`
	ZipCode                 *string `json:"zipCode"`
	HourlyRate              float64 `json:"hourlyRate"`
	Notes                   *string `json:"notes"`
	AdditionalInfo          *string `json:"additionalInfo"`
	AdditionalInfo2         *string `json:"additionalInfo2"`
	BillTo                  *string `json:"billTo"`
	IncludeAddressOnInvoice bool    `json:"includeAddressOnInvoice"`
	InvoiceCCEmail          *string `json:"invoiceCCEmail"`
	InvoiceCCDescription    *string `json:"invoiceCCDescription"`
	UniversityAffiliation   *string `json:"universityAffiliation"`
	Created                 string  `json:"created"`
	Updated                 string  `json:"updated"`
}

// ClientRequest represents the request body for creating/updating clients
type ClientRequest struct {
	Name                    string  `json:"name"`
	Email                   string  `json:"email"`
	Phone                   *string `json:"phone"`
	Address1                *string `json:"address1"`
	Address2                *string `json:"address2"`
	Address3                *string `json:"address3"`
	City                    *string `json:"city"`
	State                   *string `json:"state"`
	ZipCode                 *string `json:"zipCode"`
	HourlyRate              float64 `json:"hourlyRate"`
	Notes                   *string `json:"notes"`
	AdditionalInfo          *string `json:"additionalInfo"`
	AdditionalInfo2         *string `json:"additionalInfo2"`
	BillTo                  *string `json:"billTo"`
	IncludeAddressOnInvoice bool    `json:"includeAddressOnInvoice"`
	InvoiceCCEmail          *string `json:"invoiceCCEmail"`
	InvoiceCCDescription    *string `json:"invoiceCCDescription"`
	UniversityAffiliation   *string `json:"universityAffiliation"`
}

// ListClients returns a paginated list of clients
// GET /api/v1/clients?page=1&pageSize=20&search=term
func (h *ClientHandlers) ListClients(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	var clients []models.Client
	var total int64
	var err error

	// Calculate offset
	offset := int64((page - 1) * pageSize)

	if search != "" {
		// Search clients
		clients, err = h.clients.SearchWithPagination(search, int64(pageSize), offset)
		if err != nil {
			ErrorInternal(w)
			return
		}
		total, err = h.clients.SearchCount(search)
		if err != nil {
			ErrorInternal(w)
			return
		}
	} else {
		// Get all clients with pagination
		clients, err = h.clients.GetWithPagination(int64(pageSize), offset)
		if err != nil {
			ErrorInternal(w)
			return
		}
		total, err = h.clients.GetCount()
		if err != nil {
			ErrorInternal(w)
			return
		}
	}

	// Convert to response format
	response := make([]ClientResponse, len(clients))
	for i, client := range clients {
		response[i] = clientToResponse(&client)
	}

	WritePaginatedJSON(w, response, page, pageSize, int(total))
}

// GetClient returns a single client by ID
// GET /api/v1/clients/:id
func (h *ClientHandlers) GetClient(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid client ID")
		return
	}

	client, err := h.clients.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Client not found")
			return
		}
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, clientToResponse(&client), nil)
}

// CreateClient creates a new client
// POST /api/v1/clients
func (h *ClientHandlers) CreateClient(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req ClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.Name), "name", "Name is required")
	v.CheckField(validator.NotBlank(req.Email), "email", "Email is required")
	v.CheckField(validator.Matches(req.Email, validator.EmailRegex), "email", "Must be a valid email address")
	v.CheckField(req.HourlyRate >= 0, "hourlyRate", "Hourly rate must be non-negative")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Insert client
	id, err := h.clients.Insert(
		req.Name,
		req.Email,
		req.Phone,
		req.Address1,
		req.Address2,
		req.Address3,
		req.City,
		req.State,
		req.ZipCode,
		req.HourlyRate,
		req.Notes,
		req.AdditionalInfo,
		req.AdditionalInfo2,
		req.BillTo,
		req.IncludeAddressOnInvoice,
		req.InvoiceCCEmail,
		req.InvoiceCCDescription,
		req.UniversityAffiliation,
	)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return created client
	client, err := h.clients.Get(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusCreated, clientToResponse(&client), nil)
}

// UpdateClient updates an existing client
// PUT /api/v1/clients/:id
func (h *ClientHandlers) UpdateClient(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid client ID")
		return
	}

	// Check if client exists
	_, err = h.clients.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Client not found")
			return
		}
		ErrorInternal(w)
		return
	}

	var req ClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.Name), "name", "Name is required")
	v.CheckField(validator.NotBlank(req.Email), "email", "Email is required")
	v.CheckField(validator.Matches(req.Email, validator.EmailRegex), "email", "Must be a valid email address")
	v.CheckField(req.HourlyRate >= 0, "hourlyRate", "Hourly rate must be non-negative")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Update client
	err = h.clients.Update(
		id,
		req.Name,
		req.Email,
		req.Phone,
		req.Address1,
		req.Address2,
		req.Address3,
		req.City,
		req.State,
		req.ZipCode,
		req.HourlyRate,
		req.Notes,
		req.AdditionalInfo,
		req.AdditionalInfo2,
		req.BillTo,
		req.IncludeAddressOnInvoice,
		req.InvoiceCCEmail,
		req.InvoiceCCDescription,
		req.UniversityAffiliation,
	)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return updated client
	client, err := h.clients.Get(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, clientToResponse(&client), nil)
}

// DeleteClient soft deletes a client
// DELETE /api/v1/clients/:id
func (h *ClientHandlers) DeleteClient(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid client ID")
		return
	}

	// Check if client exists
	_, err = h.clients.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Client not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Delete client
	err = h.clients.Delete(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Client deleted successfully",
	}, nil)
}

// GetClientProjects returns all projects for a client
// GET /api/v1/clients/:id/projects
func (h *ClientHandlers) GetClientProjects(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid client ID")
		return
	}

	// Check if client exists
	_, err = h.clients.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Client not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Get projects for client
	projects, err := h.projects.GetByClient(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Convert to response format (defined in projects.go)
	response := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		response[i] = projectToResponse(&project)
	}

	WriteJSON(w, http.StatusOK, response, nil)
}

// GetClientHourlyRate returns the hourly rate for a client
// GET /api/v1/clients/:id/hourlyrate
func (h *ClientHandlers) GetClientHourlyRate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid client ID")
		return
	}

	client, err := h.clients.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Client not found")
			return
		}
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]float64{
		"hourlyRate": client.HourlyRate,
	}, nil)
}

// clientToResponse converts a models.Client to a ClientResponse
func clientToResponse(c *models.Client) ClientResponse {
	return ClientResponse{
		ID:                      c.ID,
		Name:                    c.Name,
		Email:                   c.Email,
		Phone:                   c.Phone,
		Address1:                c.Address1,
		Address2:                c.Address2,
		Address3:                c.Address3,
		City:                    c.City,
		State:                   c.State,
		ZipCode:                 c.ZipCode,
		HourlyRate:              c.HourlyRate,
		Notes:                   c.Notes,
		AdditionalInfo:          c.AdditionalInfo,
		AdditionalInfo2:         c.AdditionalInfo2,
		BillTo:                  c.BillTo,
		IncludeAddressOnInvoice: c.IncludeAddressOnInvoice,
		InvoiceCCEmail:          c.InvoiceCCEmail,
		InvoiceCCDescription:    c.InvoiceCCDescription,
		UniversityAffiliation:   c.UniversityAffiliation,
		Created:                 c.Created.Format("2006-01-02T15:04:05Z"),
		Updated:                 c.Updated.Format("2006-01-02T15:04:05Z"),
	}
}
