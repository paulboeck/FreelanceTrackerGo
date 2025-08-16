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
	ID        int
	Name      string
	Updated   time.Time
	Created   time.Time
	DeletedAt *time.Time
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
func (c *ClientModel) Insert(name string) (int, error) {
	ctx := context.Background()
	id, err := c.queries.InsertClient(ctx, name)
	if err != nil {
		return 0, err
	}
	return int(id), nil
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
		ID:        int(row.ID),
		Name:      row.Name,
		Updated:   row.UpdatedAt,
		Created:   row.CreatedAt,
		DeletedAt: deletedAt,
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
			ID:        int(row.ID),
			Name:      row.Name,
			Updated:   row.UpdatedAt,
			Created:   row.CreatedAt,
			DeletedAt: deletedAt,
		}
	}

	return clients, nil
}

// Update modifies an existing client in the database
func (c *ClientModel) Update(id int, name string) error {
	ctx := context.Background()
	params := db.UpdateClientParams{
		ID:   int64(id),
		Name: name,
	}
	return c.queries.UpdateClient(ctx, params)
}

// Delete soft deletes a client by setting the deleted_at timestamp
func (c *ClientModel) Delete(id int) error {
	ctx := context.Background()
	return c.queries.DeleteClient(ctx, int64(id))
}

// ClientModelInterface defines the interface for client operations
type ClientModelInterface interface {
	Insert(name string) (int, error)
	Get(id int) (Client, error)
	GetAll() ([]Client, error)
	Update(id int, name string) error
	Delete(id int) error
}

// Ensure implementation satisfies the interface
var _ ClientModelInterface = (*ClientModel)(nil)