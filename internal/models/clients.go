package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// Client represents a client in the system
type Client struct {
	ID                      int
	Name                    string
	Email                   string
	Phone                   *string
	Address1                *string
	Address2                *string
	Address3                *string
	City                    *string
	State                   *string
	ZipCode                 *string
	HourlyRate              float64
	Notes                   *string
	AdditionalInfo          *string
	AdditionalInfo2         *string
	BillTo                  *string
	IncludeAddressOnInvoice bool
	InvoiceCCEmail          *string
	InvoiceCCDescription    *string
	UniversityAffiliation   *string
	Updated                 time.Time
	Created                 time.Time
	DeletedAt               *time.Time
}

// ClientModel wraps the generated SQLC Queries for client operations
type ClientModel struct {
	queries *db.Queries
}

// NewClientModel creates a new ClientModel
func NewClientModel(database *sql.DB) *ClientModel {
	return &ClientModel{
		queries: db.New(database),
	}
}

// Insert adds a new client to the database and returns its ID
func (c *ClientModel) Insert(name, email string, phone, address1, address2, address3, city, state, zipCode *string, hourlyRate float64, notes, additionalInfo, additionalInfo2, billTo *string, includeAddressOnInvoice bool, invoiceCCEmail, invoiceCCDescription, universityAffiliation *string) (int, error) {
	ctx := context.Background()
	
	params := db.InsertClientParams{
		Name:                    name,
		Email:                   email,
		Phone:                   convertStringPtr(phone),
		Address1:                convertStringPtr(address1),
		Address2:                convertStringPtr(address2),
		Address3:                convertStringPtr(address3),
		City:                    convertStringPtr(city),
		State:                   convertStringPtr(state),
		ZipCode:                 convertStringPtr(zipCode),
		HourlyRate:              hourlyRate,
		Notes:                   convertStringPtr(notes),
		AdditionalInfo:          convertStringPtr(additionalInfo),
		AdditionalInfo2:         convertStringPtr(additionalInfo2),
		BillTo:                  convertStringPtr(billTo),
		IncludeAddressOnInvoice: includeAddressOnInvoice,
		InvoiceCcEmail:          convertStringPtr(invoiceCCEmail),
		InvoiceCcDescription:    convertStringPtr(invoiceCCDescription),
		UniversityAffiliation:   convertStringPtr(universityAffiliation),
	}
	
	id, err := c.queries.InsertClient(ctx, params)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Helper function to convert *string to sql.NullString
func convertStringPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// Helper function to convert sql.NullString to *string
func convertNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// Get retrieves a client by ID
func (c *ClientModel) Get(id int) (Client, error) {
	ctx := context.Background()
	row, err := c.queries.GetClient(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Client{}, ErrNoRecord
		}
		return Client{}, err
	}

	var deletedAt *time.Time
	if row.DeletedAt != nil {
		if dt, ok := row.DeletedAt.(time.Time); ok {
			deletedAt = &dt
		}
	}

	client := Client{
		ID:                      int(row.ID),
		Name:                    row.Name,
		Email:                   row.Email,
		Phone:                   convertNullString(row.Phone),
		Address1:                convertNullString(row.Address1),
		Address2:                convertNullString(row.Address2),
		Address3:                convertNullString(row.Address3),
		City:                    convertNullString(row.City),
		State:                   convertNullString(row.State),
		ZipCode:                 convertNullString(row.ZipCode),
		HourlyRate:              row.HourlyRate,
		Notes:                   convertNullString(row.Notes),
		AdditionalInfo:          convertNullString(row.AdditionalInfo),
		AdditionalInfo2:         convertNullString(row.AdditionalInfo2),
		BillTo:                  convertNullString(row.BillTo),
		IncludeAddressOnInvoice: row.IncludeAddressOnInvoice,
		InvoiceCCEmail:          convertNullString(row.InvoiceCcEmail),
		InvoiceCCDescription:    convertNullString(row.InvoiceCcDescription),
		UniversityAffiliation:   convertNullString(row.UniversityAffiliation),
		Updated:                 row.UpdatedAt,
		Created:                 row.CreatedAt,
		DeletedAt:               deletedAt,
	}

	return client, nil
}

// GetAll retrieves all clients from the database
func (c *ClientModel) GetAll() ([]Client, error) {
	ctx := context.Background()
	rows, err := c.queries.GetAllClients(ctx)
	if err != nil {
		return nil, err
	}

	clients := make([]Client, len(rows))
	for i, row := range rows {
		var deletedAt *time.Time
		if row.DeletedAt != nil {
			if dt, ok := row.DeletedAt.(time.Time); ok {
				deletedAt = &dt
			}
		}

		clients[i] = Client{
			ID:                      int(row.ID),
			Name:                    row.Name,
			Email:                   row.Email,
			Phone:                   convertNullString(row.Phone),
			Address1:                convertNullString(row.Address1),
			Address2:                convertNullString(row.Address2),
			Address3:                convertNullString(row.Address3),
			City:                    convertNullString(row.City),
			State:                   convertNullString(row.State),
			ZipCode:                 convertNullString(row.ZipCode),
			HourlyRate:              row.HourlyRate,
			Notes:                   convertNullString(row.Notes),
			AdditionalInfo:          convertNullString(row.AdditionalInfo),
			AdditionalInfo2:         convertNullString(row.AdditionalInfo2),
			BillTo:                  convertNullString(row.BillTo),
			IncludeAddressOnInvoice: row.IncludeAddressOnInvoice,
			InvoiceCCEmail:          convertNullString(row.InvoiceCcEmail),
			InvoiceCCDescription:    convertNullString(row.InvoiceCcDescription),
			UniversityAffiliation:   convertNullString(row.UniversityAffiliation),
			Updated:                 row.UpdatedAt,
			Created:                 row.CreatedAt,
			DeletedAt:               deletedAt,
		}
	}

	return clients, nil
}

// Update modifies an existing client in the database
func (c *ClientModel) Update(id int, name, email string, phone, address1, address2, address3, city, state, zipCode *string, hourlyRate float64, notes, additionalInfo, additionalInfo2, billTo *string, includeAddressOnInvoice bool, invoiceCCEmail, invoiceCCDescription, universityAffiliation *string) error {
	ctx := context.Background()
	params := db.UpdateClientParams{
		ID:                      int64(id),
		Name:                    name,
		Email:                   email,
		Phone:                   convertStringPtr(phone),
		Address1:                convertStringPtr(address1),
		Address2:                convertStringPtr(address2),
		Address3:                convertStringPtr(address3),
		City:                    convertStringPtr(city),
		State:                   convertStringPtr(state),
		ZipCode:                 convertStringPtr(zipCode),
		HourlyRate:              hourlyRate,
		Notes:                   convertStringPtr(notes),
		AdditionalInfo:          convertStringPtr(additionalInfo),
		AdditionalInfo2:         convertStringPtr(additionalInfo2),
		BillTo:                  convertStringPtr(billTo),
		IncludeAddressOnInvoice: includeAddressOnInvoice,
		InvoiceCcEmail:          convertStringPtr(invoiceCCEmail),
		InvoiceCcDescription:    convertStringPtr(invoiceCCDescription),
		UniversityAffiliation:   convertStringPtr(universityAffiliation),
	}
	return c.queries.UpdateClient(ctx, params)
}

// Delete soft deletes a client by setting the deleted_at timestamp
func (c *ClientModel) Delete(id int) error {
	ctx := context.Background()
	return c.queries.DeleteClient(ctx, int64(id))
}

// GetWithPagination retrieves clients with pagination
func (c *ClientModel) GetWithPagination(limit, offset int64) ([]Client, error) {
	ctx := context.Background()
	rows, err := c.queries.GetClientsWithPagination(ctx, db.GetClientsWithPaginationParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	var clients []Client
	for _, row := range rows {
		var deletedAt *time.Time
		if row.DeletedAt != nil {
			if timeStr, ok := row.DeletedAt.(string); ok {
				parsedTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
				if err == nil {
					deletedAt = &parsedTime
				}
			}
		}

		clients = append(clients, Client{
			ID:                      int(row.ID),
			Name:                    row.Name,
			Email:                   row.Email,
			Phone:                   convertNullString(row.Phone),
			Address1:                convertNullString(row.Address1),
			Address2:                convertNullString(row.Address2),
			Address3:                convertNullString(row.Address3),
			City:                    convertNullString(row.City),
			State:                   convertNullString(row.State),
			ZipCode:                 convertNullString(row.ZipCode),
			HourlyRate:              row.HourlyRate,
			Notes:                   convertNullString(row.Notes),
			AdditionalInfo:          convertNullString(row.AdditionalInfo),
			AdditionalInfo2:         convertNullString(row.AdditionalInfo2),
			BillTo:                  convertNullString(row.BillTo),
			IncludeAddressOnInvoice: row.IncludeAddressOnInvoice,
			InvoiceCCEmail:          convertNullString(row.InvoiceCcEmail),
			InvoiceCCDescription:    convertNullString(row.InvoiceCcDescription),
			UniversityAffiliation:   convertNullString(row.UniversityAffiliation),
			Updated:                 row.UpdatedAt,
			Created:                 row.CreatedAt,
			DeletedAt:               deletedAt,
		})
	}

	return clients, nil
}

// GetCount returns the total count of non-deleted clients
func (c *ClientModel) GetCount() (int64, error) {
	ctx := context.Background()
	return c.queries.GetClientsCount(ctx)
}

// ClientModelInterface defines the interface for client operations
type ClientModelInterface interface {
	Insert(name, email string, phone, address1, address2, address3, city, state, zipCode *string, hourlyRate float64, notes, additionalInfo, additionalInfo2, billTo *string, includeAddressOnInvoice bool, invoiceCCEmail, invoiceCCDescription, universityAffiliation *string) (int, error)
	Get(id int) (Client, error)
	GetAll() ([]Client, error)
	GetWithPagination(limit, offset int64) ([]Client, error)
	GetCount() (int64, error)
	Update(id int, name, email string, phone, address1, address2, address3, city, state, zipCode *string, hourlyRate float64, notes, additionalInfo, additionalInfo2, billTo *string, includeAddressOnInvoice bool, invoiceCCEmail, invoiceCCDescription, universityAffiliation *string) error
	Delete(id int) error
}

// Ensure implementation satisfies the interface
var _ ClientModelInterface = (*ClientModel)(nil)