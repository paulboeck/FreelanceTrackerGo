package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// User represents a user in the system
type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
	Updated        time.Time
	DeletedAt      *time.Time
}

// UserModel wraps the generated SQLC Queries for user operations
type UserModel struct {
	queries *db.Queries
}

// NewUserModel creates a new UserModel
func NewUserModel(database *sql.DB) *UserModel {
	return &UserModel{
		queries: db.New(database),
	}
}

// Insert adds a new user to the database with bcrypt hashed password
func (u *UserModel) Insert(name, email, password string) (int, error) {
	ctx := context.Background()

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	params := db.InsertUserParams{
		Name:           name,
		Email:          email,
		HashedPassword: string(hashedPassword),
	}

	id, err := u.queries.InsertUser(ctx, params)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Authenticate checks if the provided email and password are correct
func (u *UserModel) Authenticate(email, password string) (int, error) {
	ctx := context.Background()

	user, err := u.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	// Compare the provided password with the hashed password
	hashedPassword, ok := user.HashedPassword.(string)
	if !ok {
		return 0, errors.New("invalid hashed password format")
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	return int(user.ID), nil
}

// Exists checks if a user with the given email exists
func (u *UserModel) Exists(email string) (bool, error) {
	ctx := context.Background()
	
	exists, err := u.queries.UserExists(ctx, email)
	if err != nil {
		return false, err
	}
	
	return exists != 0, nil
}

// Get retrieves a user by ID
func (u *UserModel) Get(id int) (User, error) {
	ctx := context.Background()
	row, err := u.queries.GetUserByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNoRecord
		}
		return User{}, err
	}

	var deletedAt *time.Time
	if row.DeletedAt != nil {
		if dt, ok := row.DeletedAt.(time.Time); ok {
			deletedAt = &dt
		}
	}

	var hashedPassword []byte
	if hpStr, ok := row.HashedPassword.(string); ok {
		hashedPassword = []byte(hpStr)
	}

	user := User{
		ID:             int(row.ID),
		Name:           row.Name,
		Email:          row.Email,
		HashedPassword: hashedPassword,
		Created:        row.CreatedAt,
		Updated:        row.UpdatedAt,
		DeletedAt:      deletedAt,
	}

	return user, nil
}

// Update modifies an existing user's name and email
func (u *UserModel) Update(id int, name, email string) error {
	ctx := context.Background()
	params := db.UpdateUserParams{
		ID:    int64(id),
		Name:  name,
		Email: email,
	}
	return u.queries.UpdateUser(ctx, params)
}

// Delete soft deletes a user by setting the deleted_at timestamp
func (u *UserModel) Delete(id int) error {
	ctx := context.Background()
	return u.queries.DeleteUser(ctx, int64(id))
}

// UserModelInterface defines the interface for user operations
type UserModelInterface interface {
	Insert(name, email, password string) (int, error)
	Authenticate(email, password string) (int, error)
	Exists(email string) (bool, error)
	Get(id int) (User, error)
	Update(id int, name, email string) error
	Delete(id int) error
}

// Ensure implementation satisfies the interface
var _ UserModelInterface = (*UserModel)(nil)