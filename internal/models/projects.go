package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// Project represents a project in the system
type Project struct {
	ID        int
	Name      string
	ClientID  int
	Updated   time.Time
	Created   time.Time
	DeletedAt *time.Time
}

// ProjectModel wraps the generated SQLC Queries for project operations
type ProjectModel struct {
	queries *db.Queries
}

// NewProjectModel creates a new ProjectModel
func NewProjectModel(database *sql.DB) *ProjectModel {
	return &ProjectModel{
		queries: db.New(database),
	}
}

// Insert adds a new project to the database and returns its ID
func (p *ProjectModel) Insert(name string, clientID int) (int, error) {
	ctx := context.Background()
	params := db.InsertProjectParams{
		Name:     name,
		ClientID: int64(clientID),
	}
	id, err := p.queries.InsertProject(ctx, params)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Get retrieves a project by ID
func (p *ProjectModel) Get(id int) (Project, error) {
	ctx := context.Background()
	row, err := p.queries.GetProject(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrNoRecord
		}
		return Project{}, err
	}

	var deletedAt *time.Time
	if row.DeletedAt != nil {
		if dt, ok := row.DeletedAt.(time.Time); ok {
			deletedAt = &dt
		}
	}

	project := Project{
		ID:        int(row.ID),
		Name:      row.Name,
		ClientID:  int(row.ClientID),
		Updated:   row.UpdatedAt,
		Created:   row.CreatedAt,
		DeletedAt: deletedAt,
	}

	return project, nil
}

// GetByClient retrieves all projects for a specific client
func (p *ProjectModel) GetByClient(clientID int) ([]Project, error) {
	ctx := context.Background()
	rows, err := p.queries.GetProjectsByClient(ctx, int64(clientID))
	if err != nil {
		return nil, err
	}

	projects := make([]Project, len(rows))
	for i, row := range rows {
		var deletedAt *time.Time
		if row.DeletedAt != nil {
			if dt, ok := row.DeletedAt.(time.Time); ok {
				deletedAt = &dt
			}
		}

		projects[i] = Project{
			ID:        int(row.ID),
			Name:      row.Name,
			ClientID:  int(row.ClientID),
			Updated:   row.UpdatedAt,
			Created:   row.CreatedAt,
			DeletedAt: deletedAt,
		}
	}

	return projects, nil
}

// Update modifies an existing project in the database
func (p *ProjectModel) Update(id int, name string) error {
	ctx := context.Background()
	params := db.UpdateProjectParams{
		ID:   int64(id),
		Name: name,
	}
	return p.queries.UpdateProject(ctx, params)
}

// Delete soft deletes a project by setting the deleted_at timestamp
func (p *ProjectModel) Delete(id int) error {
	ctx := context.Background()
	return p.queries.DeleteProject(ctx, int64(id))
}

// ProjectModelInterface defines the interface for project operations
type ProjectModelInterface interface {
	Insert(name string, clientID int) (int, error)
	Get(id int) (Project, error)
	GetByClient(clientID int) ([]Project, error)
	Update(id int, name string) error
	Delete(id int) error
}

// Ensure implementation satisfies the interface
var _ ProjectModelInterface = (*ProjectModel)(nil)