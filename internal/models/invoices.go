package models

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// Invoice represents an invoice in the system
type Invoice struct {
	ID           int
	ProjectID    int
	InvoiceDate  time.Time
	DatePaid     *time.Time
	PaymentTerms string
	AmountDue    float64
	Updated      time.Time
	Created      time.Time
	DeletedAt    *time.Time
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
func (i *InvoiceModel) Insert(projectID int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64) (int, error) {
	ctx := context.Background()
	
	var datePaidPtr interface{}
	if datePaid != nil {
		datePaidPtr = *datePaid
	}
	
	params := db.InsertInvoiceParams{
		ProjectID:    int64(projectID),
		InvoiceDate:  invoiceDate,
		DatePaid:     datePaidPtr,
		PaymentTerms: paymentTerms,
		AmountDue:    amountDue,
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
		ID:           int(row.ID),
		ProjectID:    int(row.ProjectID),
		InvoiceDate:  row.InvoiceDate,
		DatePaid:     datePaid,
		PaymentTerms: row.PaymentTerms,
		AmountDue:    row.AmountDue,
		Updated:      row.UpdatedAt,
		Created:      row.CreatedAt,
		DeletedAt:    deletedAt,
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
			ID:           int(row.ID),
			ProjectID:    int(row.ProjectID),
			InvoiceDate:  row.InvoiceDate,
			DatePaid:     datePaid,
			PaymentTerms: row.PaymentTerms,
			AmountDue:    row.AmountDue,
			Updated:      row.UpdatedAt,
			Created:      row.CreatedAt,
			DeletedAt:    deletedAt,
		}
	}

	return invoices, nil
}

// Update modifies an existing invoice in the database
func (i *InvoiceModel) Update(id int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64) error {
	ctx := context.Background()
	
	var datePaidPtr interface{}
	if datePaid != nil {
		datePaidPtr = *datePaid
	}
	
	params := db.UpdateInvoiceParams{
		ID:           int64(id),
		InvoiceDate:  invoiceDate,
		DatePaid:     datePaidPtr,
		PaymentTerms: paymentTerms,
		AmountDue:    amountDue,
	}
	return i.queries.UpdateInvoice(ctx, params)
}

// Delete soft deletes an invoice by setting the deleted_at timestamp
func (i *InvoiceModel) Delete(id int) error {
	ctx := context.Background()
	return i.queries.DeleteInvoice(ctx, int64(id))
}

// InvoiceData represents extended invoice data for PDF generation
type InvoiceData struct {
	Invoice     Invoice
	ProjectName string
	ClientName  string
	Timesheets  []Timesheet
}

// GetForPDF retrieves invoice data with related information for PDF generation
func (i *InvoiceModel) GetForPDF(id int) (InvoiceData, error) {
	ctx := context.Background()
	
	// Get invoice with client and project names
	row, err := i.queries.GetInvoiceForPDF(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InvoiceData{}, ErrNoRecord
		}
		return InvoiceData{}, err
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
		ID:           int(row.ID),
		ProjectID:    int(row.ProjectID),
		InvoiceDate:  row.InvoiceDate,
		DatePaid:     datePaid,
		PaymentTerms: row.PaymentTerms,
		AmountDue:    row.AmountDue,
		Updated:      row.UpdatedAt,
		Created:      row.CreatedAt,
		DeletedAt:    deletedAt,
	}

	// Get timesheets for the project
	timesheetRows, err := i.queries.GetTimesheetsByProject(ctx, int64(row.ProjectID))
	if err != nil {
		return InvoiceData{}, err
	}

	timesheets := make([]Timesheet, len(timesheetRows))
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
			Description: description,
			Updated:     tsRow.UpdatedAt,
			Created:     tsRow.CreatedAt,
			DeletedAt:   tsDeletedAt,
		}
	}

	return InvoiceData{
		Invoice:     invoice,
		ProjectName: row.ProjectName,
		ClientName:  row.ClientName,
		Timesheets:  timesheets,
	}, nil
}

// GeneratePDF generates a PDF invoice based on the example format
func (i *InvoiceModel) GeneratePDF(id int) ([]byte, error) {
	data, err := i.GetForPDF(id)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)

	// Add simple decorative element (like the wheat symbol in example)
	pdf.SetFont("Arial", "", 14)
	pdf.Cell(0, 8, "🌾 🌾 🌾")
	pdf.Ln(12)

	// Title - centered and larger
	pdf.SetFont("Arial", "B", 18)
	titleWidth := pdf.GetStringWidth("Invoice for Academic Editing")
	pdf.SetX((210 - titleWidth) / 2) // Center on page
	pdf.Cell(titleWidth, 10, "Invoice for Academic Editing")
	pdf.Ln(20)

	// Date and Invoice Number with proper spacing
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 8, fmt.Sprintf("Date: %s", data.Invoice.InvoiceDate.Format("Jan. 2, 2006")))
	pdf.SetX(120) // Position invoice number on right
	pdf.Cell(70, 8, fmt.Sprintf("Invoice No: %d", data.Invoice.ID))
	pdf.Ln(20)

	// Two column layout for Invoiced To and Pay To
	leftColX := 20.0
	rightColX := 110.0
	
	// Section headers
	pdf.SetFont("Arial", "B", 12)
	pdf.SetX(leftColX)
	pdf.Cell(80, 8, "Invoiced To:")
	pdf.SetX(rightColX)
	pdf.Cell(80, 8, "Pay To:")
	pdf.Ln(8)

	// Left column - Client info
	pdf.SetFont("Arial", "", 11)
	pdf.SetX(leftColX)
	pdf.Cell(80, 6, fmt.Sprintf("Client: %s", data.ClientName))
	
	// Right column - Freelancer info (placeholder data)
	pdf.SetX(rightColX)
	pdf.Cell(80, 6, "Your Name Here")
	pdf.Ln(6)
	
	pdf.SetX(leftColX)
	pdf.Cell(80, 6, "Purchase Order Info") // Placeholder
	pdf.SetX(rightColX)
	pdf.Cell(80, 6, "Your Street Address")
	pdf.Ln(6)
	
	pdf.SetX(rightColX)
	pdf.Cell(80, 6, "Your City, State ZIP")
	pdf.Ln(6)
	
	pdf.SetX(rightColX)
	pdf.Cell(80, 6, "Your Phone Number")
	pdf.Ln(6)
	
	pdf.SetX(rightColX)
	pdf.Cell(80, 6, "your.email@example.com")
	pdf.Ln(25)

	// Table with borders and proper formatting
	
	// Table header with background and borders
	pdf.SetFillColor(240, 240, 240) // Light gray background
	pdf.SetFont("Arial", "B", 11)
	
	// Header row
	pdf.CellFormat(80, 10, "DESCRIPTION", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 10, "HOURS", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 10, "RATE", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 10, "TOTAL", "1", 1, "C", true, 0, "")

	// Calculate totals
	totalHours := 0.0
	for _, ts := range data.Timesheets {
		totalHours += ts.HoursWorked
	}
	
	var hourlyRate float64
	if totalHours > 0 {
		hourlyRate = data.Invoice.AmountDue / totalHours
	}

	// Table content row
	pdf.SetFillColor(255, 255, 255) // White background
	pdf.SetFont("Arial", "", 11)
	
	description := fmt.Sprintf("%s: %s", data.ProjectName, data.Invoice.InvoiceDate.Format("January-2006"))
	pdf.CellFormat(80, 10, description, "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 10, fmt.Sprintf("%.2f", totalHours), "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("USD $%.2f", hourlyRate), "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 10, fmt.Sprintf("USD $%.2f", data.Invoice.AmountDue), "1", 1, "C", true, 0, "")

	pdf.Ln(10)

	// Total row - right aligned
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(135, 10, "Total")
	pdf.CellFormat(35, 10, fmt.Sprintf("USD $%.2f", data.Invoice.AmountDue), "", 1, "C", false, 0, "")
	
	pdf.Ln(15)

	// Notes section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Notes:")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	// Split payment terms into multiple lines if needed
	pdf.MultiCell(0, 5, data.Invoice.PaymentTerms, "", "L", false)
	pdf.Ln(10)

	// Thank you message - centered and italic
	pdf.SetFont("Arial", "I", 12)
	thankYouText := "Thank you for your business!"
	thankYouWidth := pdf.GetStringWidth(thankYouText)
	pdf.SetX((210 - thankYouWidth) / 2)
	pdf.Cell(thankYouWidth, 8, thankYouText)

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// InvoiceModelInterface defines the interface for invoice operations
type InvoiceModelInterface interface {
	Insert(projectID int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64) (int, error)
	Get(id int) (Invoice, error)
	GetByProject(projectID int) ([]Invoice, error)
	Update(id int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64) error
	Delete(id int) error
	GetForPDF(id int) (InvoiceData, error)
	GeneratePDF(id int) ([]byte, error)
}

// Ensure implementation satisfies the interface
var _ InvoiceModelInterface = (*InvoiceModel)(nil)