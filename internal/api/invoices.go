package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/email"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

// InvoiceHandlers handles invoice API endpoints
type InvoiceHandlers struct {
	invoices *models.InvoiceModel
	projects *models.ProjectModel
	clients  *models.ClientModel
	settings *models.AppSettingModel
}

// NewInvoiceHandlers creates a new InvoiceHandlers instance
func NewInvoiceHandlers(invoices *models.InvoiceModel, projects *models.ProjectModel, clients *models.ClientModel, settings *models.AppSettingModel) *InvoiceHandlers {
	return &InvoiceHandlers{
		invoices: invoices,
		projects: projects,
		clients:  clients,
		settings: settings,
	}
}

// InvoiceResponse represents an invoice in API responses
type InvoiceResponse struct {
	ID             int      `json:"id"`
	InvoiceNum     int      `json:"invoiceNum"`
	ProjectID      int      `json:"projectId"`
	InvoiceDate    string   `json:"invoiceDate"`
	DatePaid       *string  `json:"datePaid"`
	PaymentTerms   string   `json:"paymentTerms"`
	AmountDue      float64  `json:"amountDue"`
	DisplayDetails bool     `json:"displayDetails"`
	Created        string   `json:"created"`
	Updated        string   `json:"updated"`
}

// InvoiceRequest represents the request body for creating/updating invoices
type InvoiceRequest struct {
	InvoiceDate    string  `json:"invoiceDate"`
	DatePaid       *string `json:"datePaid"`
	PaymentTerms   string  `json:"paymentTerms"`
	AmountDue      float64 `json:"amountDue"`
	DisplayDetails bool    `json:"displayDetails"`
}

// EmailInvoiceRequest represents the request body for emailing an invoice
type EmailInvoiceRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
}

// GetInvoice returns a single invoice by ID
// GET /api/v1/invoices/:id
func (h *InvoiceHandlers) GetInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid invoice ID")
		return
	}

	invoice, err := h.invoices.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Invoice not found")
			return
		}
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, invoiceToResponse(&invoice), nil)
}

// CreateInvoice creates a new invoice for a project
// POST /api/v1/projects/:id/invoices
func (h *InvoiceHandlers) CreateInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	projectID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid project ID")
		return
	}

	// Check if project exists
	_, err = h.projects.Get(projectID)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Project not found")
			return
		}
		ErrorInternal(w)
		return
	}

	var req InvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.InvoiceDate), "invoiceDate", "Invoice date is required")
	v.CheckField(req.AmountDue >= 0, "amountDue", "Amount due must be non-negative")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Parse invoice date
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		v.AddFieldError("invoiceDate", "Invalid date format (use YYYY-MM-DD)")
		ErrorValidation(w, *v)
		return
	}

	// Parse date paid if provided
	var datePaid *time.Time
	if req.DatePaid != nil {
		t, err := time.Parse("2006-01-02", *req.DatePaid)
		if err != nil {
			v.AddFieldError("datePaid", "Invalid date format (use YYYY-MM-DD)")
			ErrorValidation(w, *v)
			return
		}
		datePaid = &t
	}

	// Insert invoice
	id, err := h.invoices.Insert(projectID, invoiceDate, datePaid, req.PaymentTerms, req.AmountDue, req.DisplayDetails)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return created invoice
	createdInvoice, err := h.invoices.Get(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusCreated, invoiceToResponse(&createdInvoice), nil)
}

// UpdateInvoice updates an existing invoice
// PUT /api/v1/invoices/:id
func (h *InvoiceHandlers) UpdateInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid invoice ID")
		return
	}

	// Check if invoice exists
	_, err = h.invoices.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Invoice not found")
			return
		}
		ErrorInternal(w)
		return
	}

	var req InvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.InvoiceDate), "invoiceDate", "Invoice date is required")
	v.CheckField(req.AmountDue >= 0, "amountDue", "Amount due must be non-negative")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Parse invoice date
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		v.AddFieldError("invoiceDate", "Invalid date format (use YYYY-MM-DD)")
		ErrorValidation(w, *v)
		return
	}

	// Parse date paid if provided
	var datePaid *time.Time
	if req.DatePaid != nil {
		t, err := time.Parse("2006-01-02", *req.DatePaid)
		if err != nil {
			v.AddFieldError("datePaid", "Invalid date format (use YYYY-MM-DD)")
			ErrorValidation(w, *v)
			return
		}
		datePaid = &t
	}

	// Update invoice
	err = h.invoices.Update(id, invoiceDate, datePaid, req.PaymentTerms, req.AmountDue, req.DisplayDetails)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return updated invoice
	updatedInvoice, err := h.invoices.Get(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, invoiceToResponse(&updatedInvoice), nil)
}

// DeleteInvoice soft deletes an invoice
// DELETE /api/v1/invoices/:id
func (h *InvoiceHandlers) DeleteInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid invoice ID")
		return
	}

	// Check if invoice exists
	_, err = h.invoices.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Invoice not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Delete invoice
	err = h.invoices.Delete(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Invoice deleted successfully",
	}, nil)
}

// GenerateInvoicePDF generates a PDF for an invoice
// GET /api/v1/invoices/:id/pdf
func (h *InvoiceHandlers) GenerateInvoicePDF(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid invoice ID")
		return
	}

	// Check if invoice exists
	invoice, err := h.invoices.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Invoice not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Get all settings for PDF generation
	allSettings, err := h.settings.GetAll()
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Generate PDF
	pdfBytes, err := h.invoices.GenerateComprehensivePDF(id, allSettings)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoice_%d.pdf", invoice.InvoiceNum))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// EmailInvoice sends an invoice via email
// POST /api/v1/invoices/:id/email
func (h *InvoiceHandlers) EmailInvoice(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid invoice ID")
		return
	}

	// Check if invoice exists
	invoice, err := h.invoices.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Invoice not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Get project and client
	project, err := h.projects.Get(invoice.ProjectID)
	if err != nil {
		ErrorInternal(w)
		return
	}

	client, err := h.clients.Get(project.ClientID)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Parse request body (optional - allows custom recipient/subject/body)
	var req EmailInvoiceRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	// Set default recipient to client email
	recipient := client.Email
	if req.To != "" {
		recipient = req.To
	}

	// Validate recipient
	if recipient == "" {
		ErrorBadRequest(w, "Client does not have an email address configured")
		return
	}

	// Get freelancer name for the email
	freelancerName, err := h.settings.GetString("freelancer_name")
	if err != nil {
		freelancerName = "FreelanceTracker"
	}

	// Create email subject
	subject := req.Subject
	if subject == "" {
		subject = fmt.Sprintf("Invoice #%d for %s", invoice.InvoiceNum, project.Name)
	}

	// Create email body
	body := req.Body
	if body == "" {
		body = fmt.Sprintf(`Dear %s,

Please find attached invoice #%d for the project "%s".

Invoice Details:
- Amount Due: $%.2f
- Payment Terms: %s
- Invoice Date: %s

If you have any questions about this invoice, please don't hesitate to contact me.

Best regards,
%s`,
			client.Name,
			invoice.InvoiceNum,
			project.Name,
			invoice.AmountDue,
			invoice.PaymentTerms,
			invoice.InvoiceDate.Format("January 2, 2006"),
			freelancerName)
	}

	// Generate PDF attachment
	allSettings, err := h.settings.GetAll()
	if err != nil {
		ErrorInternal(w)
		return
	}

	pdfBytes, err := h.invoices.GenerateComprehensivePDF(id, allSettings)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Create email service from settings
	emailService, err := email.NewServiceFromSettings(h.settings)
	if err != nil {
		ErrorBadRequest(w, fmt.Sprintf("Failed to initialize email service: %v", err))
		return
	}

	// Create PDF attachment
	pdfAttachment := email.Attachment{
		Filename: fmt.Sprintf("invoice_%d.pdf", invoice.InvoiceNum),
		Data:     pdfBytes,
		MimeType: "application/pdf",
	}

	// Collect recipients (main + CC)
	recipients := []string{recipient}
	if project.InvoiceCCEmail != "" {
		recipients = append(recipients, project.InvoiceCCEmail)
	}

	// Create email
	emailMsg := email.Email{
		To:          recipients,
		Subject:     subject,
		Body:        body,
		IsHTML:      false,
		Attachments: []email.Attachment{pdfAttachment},
	}

	// Send email
	err = emailService.Send(emailMsg)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Invoice emailed successfully",
		"to":      recipient,
	}, nil)
}

// invoiceToResponse converts a models.Invoice to an InvoiceResponse
func invoiceToResponse(i *models.Invoice) InvoiceResponse {
	var datePaid *string
	if i.DatePaid != nil {
		d := i.DatePaid.Format("2006-01-02")
		datePaid = &d
	}

	return InvoiceResponse{
		ID:             i.ID,
		InvoiceNum:     i.InvoiceNum,
		ProjectID:      i.ProjectID,
		InvoiceDate:    i.InvoiceDate.Format("2006-01-02"),
		DatePaid:       datePaid,
		PaymentTerms:   i.PaymentTerms,
		AmountDue:      i.AmountDue,
		DisplayDetails: i.DisplayDetails,
		Created:        i.Created.Format("2006-01-02T15:04:05Z"),
		Updated:        i.Updated.Format("2006-01-02T15:04:05Z"),
	}
}
