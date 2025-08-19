package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

// InvoiceModelInterface defines the interface for invoice operations
type InvoiceModelInterface interface {
	Insert(projectID int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64) (int, error)
	Get(id int) (Invoice, error)
	GetByProject(projectID int) ([]Invoice, error)
	Update(id int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64) error
	Delete(id int) error
}

// Ensure implementation satisfies the interface
var _ InvoiceModelInterface = (*InvoiceModel)(nil)