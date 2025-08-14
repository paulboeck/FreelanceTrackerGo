package models

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// ClientAdapter wraps the generated SQLC Queries to provide the same interface as ClientModel
type ClientAdapter struct {
	queries *db.Queries
}

// NewClientAdapter creates a new ClientAdapter
func NewClientAdapter(database *sql.DB) *ClientAdapter {
	return &ClientAdapter{
		queries: db.New(database),
	}
}

// Insert adds a new client to the database and returns its ID
func (c *ClientAdapter) Insert(name string) (int, error) {
	ctx := context.Background()
	id, err := c.queries.InsertClient(ctx, name)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Get retrieves a client by ID
func (c *ClientAdapter) Get(id int) (Client, error) {
	ctx := context.Background()
	row, err := c.queries.GetClient(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Client{}, ErrNoRecord
		}
		return Client{}, err
	}

	client := Client{
		ID:      int(row.ID),
		Name:    row.Name,
		Updated: row.UpdatedAt,
		Created: row.CreatedAt,
	}

	return client, nil
}

// GetAll retrieves all clients from the database
func (c *ClientAdapter) GetAll() ([]Client, error) {
	ctx := context.Background()
	rows, err := c.queries.GetAllClients(ctx)
	if err != nil {
		return nil, err
	}

	clients := make([]Client, len(rows))
	for i, row := range rows {
		clients[i] = Client{
			ID:      int(row.ID),
			Name:    row.Name,
			Updated: row.UpdatedAt,
			Created: row.CreatedAt,
		}
	}

	return clients, nil
}

// ClientModelInterface defines the interface that both ClientModel and ClientAdapter implement
type ClientModelInterface interface {
	Insert(name string) (int, error)
	Get(id int) (Client, error)
	GetAll() ([]Client, error)
}

// Ensure both implementations satisfy the interface
var _ ClientModelInterface = (*ClientModel)(nil)
var _ ClientModelInterface = (*ClientAdapter)(nil)