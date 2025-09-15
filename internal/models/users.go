package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Updated        time.Time
	Created        time.Time
	DeletedAt      *time.Time
}

// UserModel wraps the generated SQLC Queries for user operations
type UserModel struct {
	queries *db.Queries
}

func (m *UserModel) Insert(name, email, password string) (int, error) {
	ctx := context.Background()

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return 0, err
	}

	// Insert the user
	id, err := m.queries.InsertUser(ctx, db.InsertUserParams{
		Name:           name,
		Email:          email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		// Check for duplicate email error
		if err.Error() == "UNIQUE constraint failed: user.email" {
			return 0, ErrDuplicateEmail
		}
		return 0, err
	}

	return int(id), nil
}

func (m *UserModel) Get(id int) (User, error) {
	ctx := context.Background()

	dbUser, err := m.queries.GetUserByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNoRecord
		}
		return User{}, err
	}

	// Convert HashedPassword from interface{} to []byte
	var hashedPassword []byte
	if hashedPasswordStr, ok := dbUser.HashedPassword.(string); ok {
		hashedPassword = []byte(hashedPasswordStr)
	} else if hashedPasswordBytes, ok := dbUser.HashedPassword.([]byte); ok {
		hashedPassword = hashedPasswordBytes
	} else {
		return User{}, errors.New("invalid hashed password type")
	}

	// Convert from db.User to models.User
	user := User{
		ID:             int(dbUser.ID),
		Name:           dbUser.Name,
		Email:          dbUser.Email,
		HashedPassword: hashedPassword,
		Updated:        dbUser.UpdatedAt,
		Created:        dbUser.CreatedAt,
	}

	if dbUser.DeletedAt != nil {
		if deletedAt, ok := dbUser.DeletedAt.(time.Time); ok {
			user.DeletedAt = &deletedAt
		}
	}

	return user, nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	ctx := context.Background()

	// Get user by email
	dbUser, err := m.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	// Verify password - convert string to []byte
	var hashedPassword []byte
	if hashedPasswordStr, ok := dbUser.HashedPassword.(string); ok {
		hashedPassword = []byte(hashedPasswordStr)
	} else if hashedPasswordBytes, ok := dbUser.HashedPassword.([]byte); ok {
		hashedPassword = hashedPasswordBytes
	} else {
		return 0, errors.New("invalid hashed password type")
	}
	
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return int(dbUser.ID), nil
}

func (m *UserModel) Exists(email string) (bool, error) {
	ctx := context.Background()

	exists, err := m.queries.UserExists(ctx, email)
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}

// UserModelInterface defines the interface for user operations
type UserModelInterface interface {
	Insert(name, email, password string) (int, error)
	Get(id int) (User, error)
	Authenticate(email, password string) (int, error)
	Exists(email string) (bool, error)
}

// Ensure implementation satisfies the interface
var _ UserModelInterface = (*UserModel)(nil)

// NewUserModel creates a new UserModel
func NewUserModel(database *sql.DB) *UserModel {
	return &UserModel{
		queries: db.New(database),
	}
}
