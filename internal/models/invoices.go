package models

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// Invoice represents an invoice in the system
type Invoice struct {
	ID             int
	ProjectID      int
	InvoiceDate    time.Time
	DatePaid       *time.Time
	PaymentTerms   string
	AmountDue      float64
	DisplayDetails bool
	Updated        time.Time
	Created        time.Time
	DeletedAt      *time.Time
}

// InvoiceModel wraps the generated SQLC Queries for invoice operations
type InvoiceModel struct {
	queries *db.Queries
}

// NewInvoiceModel creates a new InvoiceModel
func NewInvoiceModel(database *sql.DB) *InvoiceModel {
	return &InvoiceModel{
		queries: db.New(database),
	}
}

// Insert adds a new invoice to the database and returns its ID
func (i *InvoiceModel) Insert(projectID int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64, displayDetails bool) (int, error) {
	ctx := context.Background()

	var datePaidPtr interface{}
	if datePaid != nil {
		datePaidPtr = *datePaid
	}

	params := db.InsertInvoiceParams{
		ProjectID:      int64(projectID),
		InvoiceDate:    invoiceDate,
		DatePaid:       datePaidPtr,
		PaymentTerms:   paymentTerms,
		AmountDue:      amountDue,
		DisplayDetails: displayDetails,
	}
	id, err := i.queries.InsertInvoice(ctx, params)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Get retrieves an invoice by ID
func (i *InvoiceModel) Get(id int) (Invoice, error) {
	ctx := context.Background()
	row, err := i.queries.GetInvoice(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invoice{}, ErrNoRecord
		}
		return Invoice{}, err
	}

	var deletedAt *time.Time
	if row.DeletedAt != nil {
		if dt, ok := row.DeletedAt.(time.Time); ok {
			deletedAt = &dt
		}
	}

	var datePaid *time.Time
	if row.DatePaid != nil {
		if dp, ok := row.DatePaid.(time.Time); ok {
			datePaid = &dp
		}
	}

	invoice := Invoice{
		ID:             int(row.ID),
		ProjectID:      int(row.ProjectID),
		InvoiceDate:    row.InvoiceDate,
		DatePaid:       datePaid,
		PaymentTerms:   row.PaymentTerms,
		AmountDue:      row.AmountDue,
		DisplayDetails: row.DisplayDetails,
		Updated:        row.UpdatedAt,
		Created:        row.CreatedAt,
		DeletedAt:      deletedAt,
	}

	return invoice, nil
}

// GetByProject retrieves all invoices for a specific project
func (i *InvoiceModel) GetByProject(projectID int) ([]Invoice, error) {
	ctx := context.Background()
	rows, err := i.queries.GetInvoicesByProject(ctx, int64(projectID))
	if err != nil {
		return nil, err
	}

	invoices := make([]Invoice, len(rows))
	for j, row := range rows {
		var deletedAt *time.Time
		if row.DeletedAt != nil {
			if dt, ok := row.DeletedAt.(time.Time); ok {
				deletedAt = &dt
			}
		}

		var datePaid *time.Time
		if row.DatePaid != nil {
			if dp, ok := row.DatePaid.(time.Time); ok {
				datePaid = &dp
			}
		}

		invoices[j] = Invoice{
			ID:             int(row.ID),
			ProjectID:      int(row.ProjectID),
			InvoiceDate:    row.InvoiceDate,
			DatePaid:       datePaid,
			PaymentTerms:   row.PaymentTerms,
			AmountDue:      row.AmountDue,
			DisplayDetails: row.DisplayDetails,
			Updated:        row.UpdatedAt,
			Created:        row.CreatedAt,
			DeletedAt:      deletedAt,
		}
	}

	return invoices, nil
}

// Update modifies an existing invoice in the database
func (i *InvoiceModel) Update(id int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64, displayDetails bool) error {
	ctx := context.Background()

	var datePaidPtr interface{}
	if datePaid != nil {
		datePaidPtr = *datePaid
	}

	params := db.UpdateInvoiceParams{
		ID:             int64(id),
		InvoiceDate:    invoiceDate,
		DatePaid:       datePaidPtr,
		PaymentTerms:   paymentTerms,
		AmountDue:      amountDue,
		DisplayDetails: displayDetails,
	}
	return i.queries.UpdateInvoice(ctx, params)
}

// Delete soft deletes an invoice by setting the deleted_at timestamp
func (i *InvoiceModel) Delete(id int) error {
	ctx := context.Background()
	return i.queries.DeleteInvoice(ctx, int64(id))
}

// ComprehensiveInvoiceData represents complete invoice data with all related information for professional PDF generation
type ComprehensiveInvoiceData struct {
	Invoice          Invoice
	Project          Project
	Client           Client
	Timesheets       []Timesheet
	TotalHours       float64
	Subtotal         float64
	DiscountAmount   float64
	AdjustmentAmount float64
	FinalTotal       float64
}

// InvoiceTemplateData represents the data structure for HTML template rendering
type InvoiceTemplateData struct {
	Invoice          Invoice
	Project          Project
	Client           Client
	Timesheets       []Timesheet
	TotalHours       float64
	AvgRate          float64
	Subtotal         float64
	DiscountAmount   float64
	AdjustmentAmount float64
	FinalTotal       float64
	Settings         InvoiceTemplateSettings
}

// InvoiceTemplateSettings represents settings for the HTML template
type InvoiceTemplateSettings struct {
	InvoiceTitle             string
	CompanyLogoPath          string
	CompanyLogoDataURL       string // Base64 data URL for embedding in HTML
	FreelancerName           string
	FreelancerAddress        string
	FreelancerCityStateZip   string
	FreelancerPhone          string
	FreelancerEmail          string
	CurrencySymbol           string
	ShowIndividualTimesheets bool
	DefaultPaymentTerms      string
	ThankYouMessage          string
}

// GetComprehensiveForPDF retrieves comprehensive invoice data with all related information for professional PDF generation
func (i *InvoiceModel) GetComprehensiveForPDF(id int) (ComprehensiveInvoiceData, error) {
	ctx := context.Background()

	// Get invoice with comprehensive client and project data
	// Note: This will use the new GetInvoiceComprehensiveForPDF query once SQLC is regenerated
	// For now, we'll use the existing GetInvoiceForPDF and fetch additional data
	row, err := i.queries.GetInvoiceForPDF(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ComprehensiveInvoiceData{}, ErrNoRecord
		}
		return ComprehensiveInvoiceData{}, err
	}

	// Convert to Invoice struct
	var deletedAt *time.Time
	if row.DeletedAt != nil {
		if dt, ok := row.DeletedAt.(time.Time); ok {
			deletedAt = &dt
		}
	}

	var datePaid *time.Time
	if row.DatePaid != nil {
		if dp, ok := row.DatePaid.(time.Time); ok {
			datePaid = &dp
		}
	}

	invoice := Invoice{
		ID:             int(row.ID),
		ProjectID:      int(row.ProjectID),
		InvoiceDate:    row.InvoiceDate,
		DatePaid:       datePaid,
		PaymentTerms:   row.PaymentTerms,
		AmountDue:      row.AmountDue,
		DisplayDetails: row.DisplayDetails,
		Updated:        row.UpdatedAt,
		Created:        row.CreatedAt,
		DeletedAt:      deletedAt,
	}

	// TODO: Once SQLC is regenerated, we can get comprehensive client and project data in one query
	// For now, fetch them separately using existing models

	// Get project details
	projectModel := &ProjectModel{queries: i.queries}
	project, err := projectModel.Get(int(row.ProjectID))
	if err != nil {
		return ComprehensiveInvoiceData{}, fmt.Errorf("failed to get project: %w", err)
	}

	// Get client details
	clientModel := &ClientModel{queries: i.queries}
	client, err := clientModel.Get(project.ClientID)
	if err != nil {
		return ComprehensiveInvoiceData{}, fmt.Errorf("failed to get client: %w", err)
	}

	// Get timesheets for the project
	timesheetRows, err := i.queries.GetTimesheetsByProject(ctx, int64(row.ProjectID))
	if err != nil {
		return ComprehensiveInvoiceData{}, fmt.Errorf("failed to get timesheets: %w", err)
	}

	timesheets := make([]Timesheet, len(timesheetRows))
	totalHours := 0.0
	for j, tsRow := range timesheetRows {
		var tsDeletedAt *time.Time
		if tsRow.DeletedAt != nil {
			if dt, ok := tsRow.DeletedAt.(time.Time); ok {
				tsDeletedAt = &dt
			}
		}

		description := ""
		if tsRow.Description.Valid {
			description = tsRow.Description.String
		}

		timesheets[j] = Timesheet{
			ID:          int(tsRow.ID),
			ProjectID:   int(tsRow.ProjectID),
			WorkDate:    tsRow.WorkDate,
			HoursWorked: tsRow.HoursWorked,
			HourlyRate:  tsRow.HourlyRate,
			Description: description,
			Updated:     tsRow.UpdatedAt,
			Created:     tsRow.CreatedAt,
			DeletedAt:   tsDeletedAt,
		}

		totalHours += tsRow.HoursWorked
	}

	// Calculate amounts
	subtotal := invoice.AmountDue
	discountAmount := 0.0
	adjustmentAmountValue := 0.0

	// Apply project-level discount if applicable
	if project.DiscountPercent != nil && *project.DiscountPercent > 0 {
		discountAmount = subtotal * (*project.DiscountPercent / 100.0)
		subtotal -= discountAmount
	}

	// Apply project-level adjustment if applicable
	if project.AdjustmentAmount != nil {
		adjustmentAmountValue = *project.AdjustmentAmount
		subtotal += adjustmentAmountValue
	}

	return ComprehensiveInvoiceData{
		Invoice:          invoice,
		Project:          project,
		Client:           client,
		Timesheets:       timesheets,
		TotalHours:       totalHours,
		Subtotal:         subtotal,
		DiscountAmount:   discountAmount,
		AdjustmentAmount: adjustmentAmountValue,
		FinalTotal:       subtotal, // After discounts and adjustments
	}, nil
}

// GenerateComprehensivePDF generates a professional PDF invoice using chromedp HTML template
func (i *InvoiceModel) GenerateComprehensivePDF(id int, settings map[string]AppSettingValue) ([]byte, error) {
	// Use the new HTML-based PDF generation
	return i.GenerateHTMLPDF(id, settings)
}

// getLogoDataURL reads the logo file and converts it to a base64 data URL
func getLogoDataURL(logoPath string) (string, error) {
	if logoPath == "" {
		return "", nil
	}

	// Get absolute path from project root
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	fullPath := filepath.Join(projectRoot, logoPath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", nil // Return empty string if file doesn't exist
	}

	// Read the image file
	imageData, err := os.ReadFile(fullPath)
	if err != nil {
		return "", nil // Return empty string on error rather than failing
	}

	// Determine MIME type based on file extension
	ext := filepath.Ext(fullPath)
	var mimeType string
	switch strings.ToLower(ext) {
	case ".png":
		mimeType = "image/png"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".svg":
		mimeType = "image/svg+xml"
	case ".gif":
		mimeType = "image/gif"
	default:
		mimeType = "image/png" // Default to PNG
	}

	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data), nil
}

// GenerateHTMLPDF generates a PDF invoice using chromedp with HTML template
func (i *InvoiceModel) GenerateHTMLPDF(id int, settings map[string]AppSettingValue) ([]byte, error) {
	data, err := i.GetComprehensiveForPDF(id)
	if err != nil {
		return nil, err
	}

	// Helper to get setting value with fallback
	getSetting := func(key, fallback string) string {
		if setting, exists := settings[key]; exists {
			return setting.AsString()
		}
		return fallback
	}

	// Helper to get boolean setting with fallback
	getBoolSetting := func(key string, fallback bool) bool {
		if setting, exists := settings[key]; exists {
			if val, err := setting.AsBool(); err == nil {
				return val
			}
		}
		return fallback
	}

	// Calculate average rate
	avgRate := data.Project.HourlyRate
	if data.TotalHours > 0 && !data.Project.FlatFeeInvoice {
		avgRate = data.Invoice.AmountDue / data.TotalHours
	}

	// Prepare template data
	templateData := InvoiceTemplateData{
		Invoice:          data.Invoice,
		Project:          data.Project,
		Client:           data.Client,
		Timesheets:       data.Timesheets,
		TotalHours:       data.TotalHours,
		AvgRate:          avgRate,
		Subtotal:         data.Subtotal,
		DiscountAmount:   data.DiscountAmount,
		AdjustmentAmount: data.AdjustmentAmount,
		FinalTotal:       data.FinalTotal,
		Settings: InvoiceTemplateSettings{
			InvoiceTitle:             getSetting("invoice_title", "Invoice for Academic Editing"),
			CompanyLogoPath:          getSetting("company_logo_path", "./ui/static/img/logo.png"),
			CompanyLogoDataURL:       "", // Will be populated below
			FreelancerName:           getSetting("freelancer_name", "Your Name Here"),
			FreelancerAddress:        getSetting("freelancer_address", "Your Address"),
			FreelancerCityStateZip:   getSetting("freelancer_city_state_zip", "Your City, State ZIP"),
			FreelancerPhone:          getSetting("freelancer_phone", "Your Phone"),
			FreelancerEmail:          getSetting("freelancer_email", "your.email@example.com"),
			CurrencySymbol:           getSetting("invoice_currency_symbol", "$"),
			ShowIndividualTimesheets: getBoolSetting("invoice_show_individual_timesheets", true),
			DefaultPaymentTerms:      getSetting("invoice_payment_terms_default", "Payment is due within 30 days of receipt of this invoice."),
			ThankYouMessage:          getSetting("invoice_thank_you_message", "Thank you for your business!"),
		},
	}

	// Convert logo path to base64 data URL if it exists
	if logoDataURL, err := getLogoDataURL(templateData.Settings.CompanyLogoPath); err == nil && logoDataURL != "" {
		templateData.Settings.CompanyLogoDataURL = logoDataURL
	}

	// Create template with helper functions using embedded template
	tmpl := template.New("invoice")
	tmpl = tmpl.Funcs(template.FuncMap{
		"split": strings.Split,
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"isPositive": func(val float64) bool {
			return val > 0
		},
		"isNonZero": func(val float64) bool {
			return val != 0
		},
	})

	// Get the current file's directory to find project root
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename))) // Go up 3 levels from internal/models
	templatePath := filepath.Join(projectRoot, "ui", "html", "invoice.html")

	// Read template file
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err = tmpl.Parse(string(templateBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Render the HTML
	var htmlBuffer bytes.Buffer
	err = tmpl.Execute(&htmlBuffer, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Debug: Write HTML to file for inspection
	if os.Getenv("DEBUG_HTML") == "1" {
		os.WriteFile("/tmp/debug_invoice.html", htmlBuffer.Bytes(), 0644)
	}

	// Create context for chromedp
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Generate PDF using chromedp with temporary file approach
	var pdfBytes []byte

	// Write HTML to temporary file to avoid URL encoding issues
	tmpFile, err := os.CreateTemp("", "invoice_*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = tmpFile.Write(htmlBuffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Use file:// URL instead of data URI
	fileURL := "file://" + tmpFile.Name()

	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Give more time for rendering
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBytes, _, err = page.PrintToPDF().
				WithPrintBackground(true). // Enable background printing
				WithPaperWidth(8.27).      // A4 width in inches
				WithPaperHeight(11.7).     // A4 height in inches
				WithMarginTop(0.79).       // 20mm in inches
				WithMarginBottom(0.79).
				WithMarginLeft(0.79).
				WithMarginRight(0.79).
				WithDisplayHeaderFooter(false).
				WithScale(1.0). // Ensure proper scaling
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfBytes, nil
}

// InvoiceModelInterface defines the interface for invoice operations
type InvoiceModelInterface interface {
	Insert(projectID int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64, displayDetails bool) (int, error)
	Get(id int) (Invoice, error)
	GetByProject(projectID int) ([]Invoice, error)
	Update(id int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64, displayDetails bool) error
	Delete(id int) error
	GetComprehensiveForPDF(id int) (ComprehensiveInvoiceData, error)
	GenerateComprehensivePDF(id int, settings map[string]AppSettingValue) ([]byte, error)
	GenerateHTMLPDF(id int, settings map[string]AppSettingValue) ([]byte, error)
}

// Ensure implementation satisfies the interface
var _ InvoiceModelInterface = (*InvoiceModel)(nil)
