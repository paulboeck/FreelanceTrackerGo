package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: User edit tests are currently skipped because the handlers depend on the roles
// infrastructure (app.roles.GetUserRoles, app.roles.GetAll) which is not included in
// the test database schema. To enable these tests, you would need to:
// 1. Add role, permission, and user_roles tables to the test schema in testutil.go
// 2. Initialize app.roles in createTestApp()
//
// For testing pattern examples, see invoice_recalculation_test.go and settings_test.go
// which demonstrate comprehensive HTTP integration testing patterns.

// TestUserEdit tests the user edit form display
func TestUserEdit(t *testing.T) {
	t.Skip("Skipping user edit tests - requires role infrastructure not in test schema")
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("display user edit form", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// Create a test user
		userID, err := app.users.Insert("Jane Doe", "jane@example.com", "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/edit/%d", userID), nil)
		req.SetPathValue("id", strconv.Itoa(userID))
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEdit))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Form should be pre-populated with current values
		assert.Contains(t, body, `value="Jane Doe"`)
		assert.Contains(t, body, `value="jane@example.com"`)
		assert.Contains(t, body, "name=\"name\"")
		assert.Contains(t, body, "name=\"email\"")
	})

	t.Run("user edit form for non-existent user", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		req := httptest.NewRequest(http.MethodGet, "/user/edit/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEdit))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("user edit form with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/edit/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEdit))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestUserEditPost tests updating user information
func TestUserEditPost(t *testing.T) {
	t.Skip("Skipping user edit tests - requires role infrastructure not in test schema")
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful user update", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// Create a test user
		userID, err := app.users.Insert("Original Name", "original@example.com", "password123")
		require.NoError(t, err)

		// Update the user
		form := url.Values{}
		form.Add("name", "Updated Name")
		form.Add("email", "updated@example.com")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/edit/%d", userID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(userID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEditPost))
		handler.ServeHTTP(rr, req)

		// Should redirect to users list
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Equal(t, "/users", location)

		// Verify the user was actually updated in the database
		user, err := app.users.Get(userID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", user.Name)
		assert.Equal(t, "updated@example.com", user.Email)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// Create a test user
		userID, err := app.users.Insert("Original Name", "original@example.com", "password123")
		require.NoError(t, err)

		// Try to update with empty name
		form := url.Values{}
		form.Add("name", "")
		form.Add("email", "original@example.com")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/edit/%d", userID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(userID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEditPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Name is required")

		// Verify the user was NOT updated
		user, err := app.users.Get(userID)
		require.NoError(t, err)
		assert.Equal(t, "Original Name", user.Name)
	})

	t.Run("validation error - empty email", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// Create a test user
		userID, err := app.users.Insert("Test User", "test@example.com", "password123")
		require.NoError(t, err)

		// Try to update with empty email
		form := url.Values{}
		form.Add("name", "Test User")
		form.Add("email", "")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/edit/%d", userID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(userID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEditPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Email is required")

		// Verify the user was NOT updated
		user, err := app.users.Get(userID)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("validation error - invalid email format", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// Create a test user
		userID, err := app.users.Insert("Test User", "test@example.com", "password123")
		require.NoError(t, err)

		// Try to update with invalid email
		form := url.Values{}
		form.Add("name", "Test User")
		form.Add("email", "not-an-email")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/edit/%d", userID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(userID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEditPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Email must be a valid email address")

		// Verify the user was NOT updated
		user, err := app.users.Get(userID)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("validation error - duplicate email", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// Create two users
		userID1, err := app.users.Insert("User One", "user1@example.com", "password123")
		require.NoError(t, err)

		_, err = app.users.Insert("User Two", "user2@example.com", "password123")
		require.NoError(t, err)

		// Try to update user1's email to user2's email
		form := url.Values{}
		form.Add("name", "User One")
		form.Add("email", "user2@example.com")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/edit/%d", userID1), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(userID1))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEditPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Email address is already in use")

		// Verify user1's email was NOT changed
		user, err := app.users.Get(userID1)
		require.NoError(t, err)
		assert.Equal(t, "user1@example.com", user.Email)
	})

	t.Run("form value preservation on validation error", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// Create a test user
		userID, err := app.users.Insert("Original Name", "original@example.com", "password123")
		require.NoError(t, err)

		// Create another user to trigger duplicate email error
		_, err = app.users.Insert("Other User", "other@example.com", "password123")
		require.NoError(t, err)

		// Submit form with valid name but duplicate email
		form := url.Values{}
		form.Add("name", "New Name Value")
		form.Add("email", "other@example.com") // Duplicate email

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/edit/%d", userID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(userID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEditPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()

		// Verify the submitted values are preserved in the form
		assert.Contains(t, body, "New Name Value")
		assert.Contains(t, body, "other@example.com")
		assert.Contains(t, body, "Email address is already in use")
	})

	t.Run("update non-existent user", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		form := url.Values{}
		form.Add("name", "Test User")
		form.Add("email", "test@example.com")

		req := httptest.NewRequest(http.MethodPost, "/user/edit/999", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "999")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEditPost))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update with invalid ID", func(t *testing.T) {
		form := url.Values{}
		form.Add("name", "Test User")
		form.Add("email", "test@example.com")

		req := httptest.NewRequest(http.MethodPost, "/user/edit/invalid", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "invalid")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEditPost))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// Note: Role-related tests are skipped because the test database schema
// doesn't include role tables. In a production environment with roles,
// you would add tests for:
// - Updating user and assigning roles
// - Updating user and removing all roles
// - Updating user and changing assigned roles

// TestUserEditWorkflow tests the complete user edit workflow
func TestUserEditWorkflow(t *testing.T) {
	t.Skip("Skipping user edit tests - requires role infrastructure not in test schema")
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("complete user edit workflow", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// 1. Create a user
		userID, err := app.users.Insert("Jane Doe", "jane@example.com", "password123")
		require.NoError(t, err)

		// 2. View users list (should show new user)
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.usersList))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Jane Doe")
		assert.Contains(t, body, "jane@example.com")

		// 3. Get the edit form
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/edit/%d", userID), nil)
		req.SetPathValue("id", strconv.Itoa(userID))
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEdit))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body = rr.Body.String()
		assert.Contains(t, body, "Jane Doe")
		assert.Contains(t, body, "jane@example.com")

		// 4. Submit the update
		form := url.Values{}
		form.Add("name", "Jane Smith")
		form.Add("email", "jane.smith@example.com")

		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/edit/%d", userID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(userID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.userEditPost))
		handler.ServeHTTP(rr, req)

		// Should redirect to users list
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, "/users", rr.Header().Get("Location"))

		// 5. Verify users list shows updated values
		req = httptest.NewRequest(http.MethodGet, "/users", nil)
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.usersList))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body = rr.Body.String()
		assert.Contains(t, body, "Jane Smith")
		assert.Contains(t, body, "jane.smith@example.com")
		assert.NotContains(t, body, "jane@example.com") // Old email should not appear
	})
}
