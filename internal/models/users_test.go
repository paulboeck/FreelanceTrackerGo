package models

import (
	"testing"

	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserModel_Insert(t *testing.T) {
	// Setup test database using SQLite
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewUserModel(testDB.DB)

	t.Run("successful insert", func(t *testing.T) {
		testDB.TruncateTable(t, "users")

		name := "Test User"
		email := "test@example.com"
		password := "testpassword123"

		id, err := model.Insert(name, email, password)

		require.NoError(t, err)
		assert.Greater(t, id, 0)

		// Verify the user was actually inserted using direct query
		var insertedName, insertedEmail string
		err = testDB.DB.QueryRow("SELECT name, email FROM users WHERE id = ?", id).Scan(&insertedName, &insertedEmail)
		require.NoError(t, err)
		assert.Equal(t, name, insertedName)
		assert.Equal(t, email, insertedEmail)
	})
}

func TestUserModel_Authenticate(t *testing.T) {
	// Setup test database using SQLite
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewUserModel(testDB.DB)

	t.Run("successful authentication", func(t *testing.T) {
		testDB.TruncateTable(t, "users")

		name := "Auth User"
		email := "auth@example.com"
		password := "authpassword123"

		// Insert user
		insertedID, err := model.Insert(name, email, password)
		require.NoError(t, err)

		// Test correct authentication
		authID, err := model.Authenticate(email, password)
		require.NoError(t, err)
		assert.Equal(t, insertedID, authID)
	})

	t.Run("invalid password", func(t *testing.T) {
		testDB.TruncateTable(t, "users")

		name := "Auth User"
		email := "auth2@example.com"
		password := "authpassword123"

		// Insert user
		_, err := model.Insert(name, email, password)
		require.NoError(t, err)

		// Test incorrect password
		_, err = model.Authenticate(email, "wrongpassword")
		assert.Equal(t, ErrInvalidCredentials, err)
	})

	t.Run("nonexistent user", func(t *testing.T) {
		testDB.TruncateTable(t, "users")

		// Test non-existent user
		_, err := model.Authenticate("nonexistent@example.com", "password")
		assert.Equal(t, ErrInvalidCredentials, err)
	})
}

func TestUserModel_Exists(t *testing.T) {
	// Setup test database using SQLite
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewUserModel(testDB.DB)

	t.Run("user does not exist", func(t *testing.T) {
		testDB.TruncateTable(t, "users")

		email := "nonexistent@example.com"

		exists, err := model.Exists(email)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("user exists", func(t *testing.T) {
		testDB.TruncateTable(t, "users")

		name := "Existing User"
		email := "existing@example.com"
		password := "password123"

		// Insert user
		_, err := model.Insert(name, email, password)
		require.NoError(t, err)

		// Test existing user
		exists, err := model.Exists(email)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}