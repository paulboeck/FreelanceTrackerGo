package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

type Permission struct {
	ID          int
	Name        string
	Description string
	Created     time.Time
	Updated     time.Time
	DeletedAt   *time.Time
}

// PermissionModel wraps the generated SQLC Queries for permission operations
type PermissionModel struct {
	queries *db.Queries
}

func (m *PermissionModel) Get(id int) (Permission, error) {
	ctx := context.Background()

	dbPerm, err := m.queries.GetPermissionByID(ctx, int64(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return Permission{}, ErrNoRecord
		}
		return Permission{}, err
	}

	return convertDBPermission(dbPerm), nil
}

func (m *PermissionModel) GetByName(name string) (Permission, error) {
	ctx := context.Background()

	dbPerm, err := m.queries.GetPermissionByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return Permission{}, ErrNoRecord
		}
		return Permission{}, err
	}

	return convertDBPermission(dbPerm), nil
}

func (m *PermissionModel) GetAll() ([]Permission, error) {
	ctx := context.Background()

	dbPerms, err := m.queries.GetAllPermissions(ctx)
	if err != nil {
		return nil, err
	}

	perms := make([]Permission, len(dbPerms))
	for i, dbPerm := range dbPerms {
		perms[i] = convertDBPermission(dbPerm)
	}

	return perms, nil
}

func (m *PermissionModel) GetRolePermissions(roleID int) ([]Permission, error) {
	ctx := context.Background()

	dbPerms, err := m.queries.GetRolePermissions(ctx, int64(roleID))
	if err != nil {
		return nil, err
	}

	perms := make([]Permission, len(dbPerms))
	for i, dbPerm := range dbPerms {
		perms[i] = convertDBPermission(dbPerm)
	}

	return perms, nil
}

func (m *PermissionModel) GetUserPermissions(userID int) ([]Permission, error) {
	ctx := context.Background()

	dbPerms, err := m.queries.GetUserPermissions(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	perms := make([]Permission, len(dbPerms))
	for i, dbPerm := range dbPerms {
		perms[i] = convertDBPermission(dbPerm)
	}

	return perms, nil
}

// GetUserPermissionNames returns a slice of permission names for a user
func (m *PermissionModel) GetUserPermissionNames(userID int) ([]string, error) {
	perms, err := m.GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(perms))
	for i, perm := range perms {
		names[i] = perm.Name
	}

	return names, nil
}

// UserHasPermission checks if a user has a specific permission
func (m *PermissionModel) UserHasPermission(userID int, permissionName string) (bool, error) {
	perms, err := m.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, perm := range perms {
		if perm.Name == permissionName {
			return true, nil
		}
	}

	return false, nil
}

// Helper function to convert db.Permission to models.Permission
func convertDBPermission(dbPerm db.Permission) Permission {
	perm := Permission{
		ID:          int(dbPerm.ID),
		Name:        dbPerm.Name,
		Description: "",
		Created:     dbPerm.CreatedAt,
		Updated:     dbPerm.UpdatedAt,
	}

	if dbPerm.Description.Valid {
		perm.Description = dbPerm.Description.String
	}

	if dbPerm.DeletedAt != nil {
		if deletedAt, ok := dbPerm.DeletedAt.(time.Time); ok {
			perm.DeletedAt = &deletedAt
		}
	}

	return perm
}

// PermissionModelInterface defines the interface for permission operations
type PermissionModelInterface interface {
	Get(id int) (Permission, error)
	GetByName(name string) (Permission, error)
	GetAll() ([]Permission, error)
	GetRolePermissions(roleID int) ([]Permission, error)
	GetUserPermissions(userID int) ([]Permission, error)
	GetUserPermissionNames(userID int) ([]string, error)
	UserHasPermission(userID int, permissionName string) (bool, error)
}

// Ensure implementation satisfies the interface
var _ PermissionModelInterface = (*PermissionModel)(nil)

// NewPermissionModel creates a new PermissionModel
func NewPermissionModel(database *sql.DB) *PermissionModel {
	return &PermissionModel{
		queries: db.New(database),
	}
}
