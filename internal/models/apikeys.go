package models

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
	"golang.org/x/crypto/bcrypt"
)

// APIKey represents an API key in the system
type APIKey struct {
	ID         int
	UserID     int
	Name       string
	KeyHash    string
	KeyPrefix  string
	Scopes     string
	LastUsedAt *time.Time
	Created    time.Time
	Updated    time.Time
	DeletedAt  *time.Time
}

// APIKeyModel wraps the generated SQLC Queries for API key operations
type APIKeyModel struct {
	queries *db.Queries
}

// NewAPIKeyModel creates a new APIKeyModel instance
func NewAPIKeyModel(database *sql.DB) *APIKeyModel {
	return &APIKeyModel{
		queries: db.New(database),
	}
}

// Valid scope patterns
var validScopes = map[string]bool{
	"*":               true, // Full access
	"clients:read":    true,
	"clients:write":   true,
	"clients:*":       true,
	"projects:read":   true,
	"projects:write":  true,
	"projects:*":      true,
	"timesheets:read": true,
	"timesheets:write": true,
	"timesheets:*":    true,
	"invoices:read":   true,
	"invoices:write":  true,
	"invoices:*":      true,
	"reports:read":    true,
	"reports:*":       true,
	"settings:read":   true,
	"settings:write":  true,
	"settings:*":      true,
	"apikeys:read":    true,
	"apikeys:write":   true,
	"apikeys:*":       true,
}

// ValidateScopes checks if all scopes are valid
func ValidateScopes(scopes string) error {
	if scopes == "" {
		return errors.New("scopes cannot be empty")
	}

	scopeList := strings.Fields(scopes) // Split by whitespace
	if len(scopeList) == 0 {
		return errors.New("at least one scope is required")
	}

	for _, scope := range scopeList {
		if !validScopes[scope] {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}

	return nil
}

// HasScope checks if the API key has a specific scope
func HasScope(keyScopes, requiredScope string) bool {
	scopes := strings.Fields(keyScopes)

	// Check for wildcard
	for _, scope := range scopes {
		if scope == "*" {
			return true
		}
		if scope == requiredScope {
			return true
		}
		// Check for resource wildcard (e.g., "clients:*" matches "clients:read")
		if strings.HasSuffix(scope, ":*") {
			resource := strings.TrimSuffix(scope, ":*")
			if strings.HasPrefix(requiredScope, resource+":") {
				return true
			}
		}
	}

	return false
}

// HasAllScopes checks if the API key has all required scopes
func HasAllScopes(keyScopes string, requiredScopes ...string) bool {
	for _, required := range requiredScopes {
		if !HasScope(keyScopes, required) {
			return false
		}
	}
	return true
}

// GenerateAPIKey generates a new API key with format ftk_<32_random_chars>
func GenerateAPIKey() (string, error) {
	// Generate 24 random bytes (will be 32 chars in base64)
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to URL-safe base64 without padding
	token := base64.RawURLEncoding.EncodeToString(bytes)
	return "ftk_" + token, nil
}

// HashAPIKey hashes an API key using bcrypt
func HashAPIKey(key string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(key), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CompareAPIKeyHash compares a plain text key with a bcrypt hash
func CompareAPIKeyHash(hash, key string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(key))
	return err == nil
}

// Insert creates a new API key and returns the generated key string and the key ID
func (m *APIKeyModel) Insert(userID int, name, scopes string) (string, int, error) {
	ctx := context.Background()

	// Validate scopes
	if err := ValidateScopes(scopes); err != nil {
		return "", 0, err
	}

	// Generate API key
	key, err := GenerateAPIKey()
	if err != nil {
		return "", 0, err
	}

	// Hash the key
	hash, err := HashAPIKey(key)
	if err != nil {
		return "", 0, err
	}

	// Extract prefix (first 8 characters after "ftk_")
	prefix := key[:12] // "ftk_" + first 8 chars

	// Insert into database
	id, err := m.queries.InsertAPIKey(ctx, db.InsertAPIKeyParams{
		UserID:    int64(userID),
		Name:      name,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Scopes:    scopes,
	})
	if err != nil {
		return "", 0, err
	}

	// Return the plaintext key (only time it's shown) and the ID
	return key, int(id), nil
}

// GetByHash retrieves an API key by its hash
func (m *APIKeyModel) GetByHash(hash string) (*APIKey, error) {
	ctx := context.Background()

	row, err := m.queries.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return &APIKey{
		ID:         int(row.ID),
		UserID:     int(row.UserID),
		Name:       row.Name,
		KeyHash:    row.KeyHash,
		KeyPrefix:  row.KeyPrefix,
		Scopes:     row.Scopes,
		LastUsedAt: nilTimeToPointer(row.LastUsedAt),
		Created:    row.Created,
		Updated:    row.Updated,
		DeletedAt:  nilTimeToPointer(row.DeletedAt),
	}, nil
}

// Authenticate verifies an API key and returns the associated API key record
func (m *APIKeyModel) Authenticate(key string) (*APIKey, error) {
	ctx := context.Background()

	// Validate format
	if len(key) < 12 || !strings.HasPrefix(key, "ftk_") {
		return nil, errors.New("invalid API key format")
	}

	// Extract prefix (first 12 characters: "ftk_" + 8 chars)
	prefix := key[:12]

	// Get all keys with this prefix (should be very few, usually 1)
	rows, err := m.queries.GetAPIKeysByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// Check each key's hash
	for _, row := range rows {
		if CompareAPIKeyHash(row.KeyHash, key) {
			// Found a match
			return &APIKey{
				ID:         int(row.ID),
				UserID:     int(row.UserID),
				Name:       row.Name,
				KeyHash:    row.KeyHash,
				KeyPrefix:  row.KeyPrefix,
				Scopes:     row.Scopes,
				LastUsedAt: nilTimeToPointer(row.LastUsedAt),
				Created:    row.Created,
				Updated:    row.Updated,
				DeletedAt:  nilTimeToPointer(row.DeletedAt),
			}, nil
		}
	}

	// No matching key found
	return nil, errors.New("invalid API key")
}

// UpdateLastUsed updates the last_used_at timestamp for an API key
func (m *APIKeyModel) UpdateLastUsed(id int) error {
	ctx := context.Background()
	return m.queries.UpdateAPIKeyLastUsed(ctx, int64(id))
}

// GetByUserID returns all API keys for a user
func (m *APIKeyModel) GetByUserID(userID int) ([]*APIKey, error) {
	ctx := context.Background()

	rows, err := m.queries.GetAPIKeysByUserID(ctx, int64(userID))
	if err != nil {
		return nil, err
	}

	keys := make([]*APIKey, len(rows))
	for i, row := range rows {
		keys[i] = &APIKey{
			ID:         int(row.ID),
			UserID:     int(row.UserID),
			Name:       row.Name,
			KeyHash:    row.KeyHash,
			KeyPrefix:  row.KeyPrefix,
			Scopes:     row.Scopes,
			LastUsedAt: nilTimeToPointer(row.LastUsedAt),
			Created:    row.Created,
			Updated:    row.Updated,
			DeletedAt:  nilTimeToPointer(row.DeletedAt),
		}
	}

	return keys, nil
}

// Update updates an API key's name and scopes
func (m *APIKeyModel) Update(id int, name, scopes string) error {
	ctx := context.Background()

	// Validate scopes
	if err := ValidateScopes(scopes); err != nil {
		return err
	}

	return m.queries.UpdateAPIKey(ctx, db.UpdateAPIKeyParams{
		Name:   name,
		Scopes: scopes,
		ID:     int64(id),
	})
}

// Delete soft deletes an API key
func (m *APIKeyModel) Delete(id int) error {
	ctx := context.Background()
	return m.queries.SoftDeleteAPIKey(ctx, int64(id))
}

// GetByID retrieves an API key by ID
func (m *APIKeyModel) GetByID(id int) (*APIKey, error) {
	ctx := context.Background()

	row, err := m.queries.GetAPIKeyByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return &APIKey{
		ID:         int(row.ID),
		UserID:     int(row.UserID),
		Name:       row.Name,
		KeyHash:    row.KeyHash,
		KeyPrefix:  row.KeyPrefix,
		Scopes:     row.Scopes,
		LastUsedAt: nilTimeToPointer(row.LastUsedAt),
		Created:    row.Created,
		Updated:    row.Updated,
		DeletedAt:  nilTimeToPointer(row.DeletedAt),
	}, nil
}

// nilTimeToPointer converts sql.NullTime to *time.Time
func nilTimeToPointer(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}
