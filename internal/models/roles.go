package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

type Role struct {
	ID          int
	Name        string
	Description string
	Created     time.Time
	Updated     time.Time
	DeletedAt   *time.Time
}

// RoleModel wraps the generated SQLC Queries for role operations
type RoleModel struct {
	queries *db.Queries
}

func (m *RoleModel) Get(id int) (Role, error) {
	ctx := context.Background()

	dbRole, err := m.queries.GetRoleByID(ctx, int64(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return Role{}, ErrNoRecord
		}
		return Role{}, err
	}

	return convertDBRole(dbRole), nil
}

func (m *RoleModel) GetByName(name string) (Role, error) {
	ctx := context.Background()

	dbRole, err := m.queries.GetRoleByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return Role{}, ErrNoRecord
		}
		return Role{}, err
	}

	return convertDBRole(dbRole), nil
}

func (m *RoleModel) GetAll() ([]Role, error) {
	ctx := context.Background()

	dbRoles, err := m.queries.GetAllRoles(ctx)
	if err != nil {
		return nil, err
	}

	roles := make([]Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = convertDBRole(dbRole)
	}

	return roles, nil
}

func (m *RoleModel) GetUserRoles(userID int) ([]Role, error) {
	ctx := context.Background()

	dbRoles, err := m.queries.GetUserRoles(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	roles := make([]Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = convertDBRole(dbRole)
	}

	return roles, nil
}

func (m *RoleModel) AssignToUser(userID, roleID int) error {
	ctx := context.Background()

	return m.queries.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID: int64(userID),
		RoleID: int64(roleID),
	})
}

func (m *RoleModel) RemoveFromUser(userID, roleID int) error {
	ctx := context.Background()

	return m.queries.RemoveRoleFromUser(ctx, db.RemoveRoleFromUserParams{
		UserID: int64(userID),
		RoleID: int64(roleID),
	})
}

func (m *RoleModel) RemoveAllFromUser(userID int) error {
	ctx := context.Background()

	return m.queries.RemoveAllRolesFromUser(ctx, int64(userID))
}

// Helper function to convert db.Role to models.Role
func convertDBRole(dbRole db.Role) Role {
	role := Role{
		ID:          int(dbRole.ID),
		Name:        dbRole.Name,
		Description: "",
		Created:     dbRole.CreatedAt,
		Updated:     dbRole.UpdatedAt,
	}

	if dbRole.Description.Valid {
		role.Description = dbRole.Description.String
	}

	if dbRole.DeletedAt != nil {
		if deletedAt, ok := dbRole.DeletedAt.(time.Time); ok {
			role.DeletedAt = &deletedAt
		}
	}

	return role
}

// RoleModelInterface defines the interface for role operations
type RoleModelInterface interface {
	Get(id int) (Role, error)
	GetByName(name string) (Role, error)
	GetAll() ([]Role, error)
	GetUserRoles(userID int) ([]Role, error)
	AssignToUser(userID, roleID int) error
	RemoveFromUser(userID, roleID int) error
	RemoveAllFromUser(userID int) error
}

// Ensure implementation satisfies the interface
var _ RoleModelInterface = (*RoleModel)(nil)

// NewRoleModel creates a new RoleModel
func NewRoleModel(database *sql.DB) *RoleModel {
	return &RoleModel{
		queries: db.New(database),
	}
}
