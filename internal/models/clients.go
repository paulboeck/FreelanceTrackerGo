package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// Client represents a client in the system.
// This struct defines the shape of client data throughout the application.
type Client struct {
	ID                      int
	Name                    string
	Email                   string
	// *string is a pointer to a string - it can be nil (NULL in database) or point to a string value
	// Pointers are used for optional fields that may not have a value
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
	DeletedAt               *time.Time // Pointer allows nil (not deleted) or a timestamp (soft deleted)
}

// ClientModel wraps the generated SQLC Queries for client operations.
// This is the data access layer - it handles all database operations for clients.
type ClientModel struct {
	queries *db.Queries // Pointer to SQLC-generated database query methods
}

// NewClientModel creates a new ClientModel instance.
// This is a constructor function (Go doesn't have constructors like other languages).
func NewClientModel(database *sql.DB) *ClientModel {
	// &ClientModel{} creates a new struct and returns a pointer to it
	return &ClientModel{
		queries: db.New(database), // Initialize the SQLC queries with the database connection
	}
}

// Insert adds a new client to the database and returns its ID.
// This is a method on *ClientModel (the receiver is 'c').
func (c *ClientModel) Insert(name, email string, phone, address1, address2, address3, city, state, zipCode *string, hourlyRate float64, notes, additionalInfo, additionalInfo2, billTo *string, includeAddressOnInvoice bool, invoiceCCEmail, invoiceCCDescription, universityAffiliation *string) (int, error) {
	// context.Background() creates an empty context - it's used for passing request-scoped values and cancellation
	ctx := context.Background()

	// Create a params struct to pass to the SQLC-generated query
	// This groups all the parameters together in a type-safe way
	params := db.InsertClientParams{
		Name:                    name,
		Email:                   email,
		Phone:                   convertStringPtr(phone),           // Convert *string to sql.NullString
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

	// Call the SQLC-generated InsertClient method to execute the SQL query
	id, err := c.queries.InsertClient(ctx, params)
	if err != nil {
		return 0, err
	}
	// Convert int64 to int and return the new client's ID
	return int(id), nil
}

// convertStringPtr converts a *string to sql.NullString.
// sql.NullString is the database/sql way to represent nullable strings.
// This helper function is used throughout the model to convert between Go pointers and SQL nulls.
func convertStringPtr(s *string) sql.NullString {
	if s == nil {
		// If pointer is nil, return a NullString with Valid=false (represents NULL in database)
		return sql.NullString{Valid: false}
	}
	// If pointer has a value, dereference it (*s) and return a valid NullString
	return sql.NullString{String: *s, Valid: true}
}

// convertNullString converts sql.NullString to *string.
// This is the reverse of convertStringPtr - converts from database format to Go format.
func convertNullString(ns sql.NullString) *string {
	if !ns.Valid {
		// If the database value is NULL, return nil pointer
		return nil
	}
	// If there's a value, return a pointer to it
	// &ns.String takes the address of the string field, creating a pointer
	return &ns.String
}

// Get retrieves a single client by ID from the database.
func (c *ClientModel) Get(id int) (Client, error) {
	ctx := context.Background()
	// int64(id) converts int to int64 (type conversion) - database IDs are typically int64
	row, err := c.queries.GetClient(ctx, int64(id))
	if err != nil {
		// errors.Is checks if the error matches a specific error type
		if errors.Is(err, sql.ErrNoRows) {
			// If no rows were found, return our custom ErrNoRecord error
			return Client{}, ErrNoRecord
		}
		// For other errors, return the Client zero value (empty struct) and the error
		return Client{}, err
	}

	// Handle the DeletedAt field - it comes from database as interface{} so needs type conversion
	var deletedAt *time.Time
	if row.DeletedAt != nil {
		// Type assertion: check if the value is a time.Time
		// dt gets the value, ok is true if the conversion succeeded
		if dt, ok := row.DeletedAt.(time.Time); ok {
			deletedAt = &dt // Take the address to get a pointer
		}
	}

	// Build a Client struct from the database row
	// This converts from the SQLC-generated types to our application types
	client := Client{
		ID:                      int(row.ID),                      // Convert int64 to int
		Name:                    row.Name,
		Email:                   row.Email,
		Phone:                   convertNullString(row.Phone),     // Convert sql.NullString to *string
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

// GetAll retrieves all non-deleted clients from the database.
func (c *ClientModel) GetAll() ([]Client, error) {
	ctx := context.Background()
	rows, err := c.queries.GetAllClients(ctx)
	if err != nil {
		return nil, err
	}

	// make() creates a slice with a specific length and capacity
	// []Client is a slice (dynamic array) of Client structs
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

// Update modifies an existing client in the database.
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

// Delete soft deletes a client by setting the deleted_at timestamp.
// Soft delete means the record stays in the database but is marked as deleted.
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

// SearchWithPagination searches clients by name or email with pagination
func (c *ClientModel) SearchWithPagination(searchTerm string, limit, offset int64) ([]Client, error) {
	ctx := context.Background()
	searchPattern := "%" + searchTerm + "%"
	rows, err := c.queries.SearchClientsWithPagination(ctx, db.SearchClientsWithPaginationParams{
		Name:   searchPattern,
		Email:  searchPattern,
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

// SearchCount returns the count of clients matching the search term
func (c *ClientModel) SearchCount(searchTerm string) (int64, error) {
	ctx := context.Background()
	searchPattern := "%" + searchTerm + "%"
	return c.queries.SearchClientsCount(ctx, db.SearchClientsCountParams{
		Name:  searchPattern,
		Email: searchPattern,
	})
}

// ClientModelInterface defines the interface for client operations.
// An interface in Go defines a contract - any type that implements these methods satisfies the interface.
// This allows for dependency injection and easier testing (can create mock implementations).
type ClientModelInterface interface {
	Insert(name, email string, phone, address1, address2, address3, city, state, zipCode *string, hourlyRate float64, notes, additionalInfo, additionalInfo2, billTo *string, includeAddressOnInvoice bool, invoiceCCEmail, invoiceCCDescription, universityAffiliation *string) (int, error)
	Get(id int) (Client, error)
	GetAll() ([]Client, error)
	GetWithPagination(limit, offset int64) ([]Client, error)
	GetCount() (int64, error)
	SearchWithPagination(searchTerm string, limit, offset int64) ([]Client, error)
	SearchCount(searchTerm string) (int64, error)
	Update(id int, name, email string, phone, address1, address2, address3, city, state, zipCode *string, hourlyRate float64, notes, additionalInfo, additionalInfo2, billTo *string, includeAddressOnInvoice bool, invoiceCCEmail, invoiceCCDescription, universityAffiliation *string) error
	Delete(id int) error
}

// Ensure implementation satisfies the interface at compile time.
// This is a compile-time check - if ClientModel doesn't implement all methods, the code won't compile.
// The underscore (_) discards the value; we only care about the type check.
var _ ClientModelInterface = (*ClientModel)(nil)
