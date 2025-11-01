package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTimesheetUpdate tests the timesheet update form display
func TestTimesheetUpdate(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("display timesheet update form", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup: Create client, project, and timesheet
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		timesheetID := testDB.InsertTestTimesheet(t, projectID, "2024-01-15", "8.0", "100.00", "Initial work")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/project/timesheet/update/%d", timesheetID), nil)
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdate))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Form should be pre-populated with current values
		assert.Contains(t, body, "2024-01-15")
		assert.Contains(t, body, "8.00")
		assert.Contains(t, body, "100.00")
		assert.Contains(t, body, "Initial work")
	})

	t.Run("timesheet update form for non-existent timesheet", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")

		req := httptest.NewRequest(http.MethodGet, "/project/timesheet/update/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdate))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("timesheet update form with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/project/timesheet/update/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdate))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestTimesheetUpdatePost tests updating timesheet information
func TestTimesheetUpdatePost(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful timesheet update", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup: Create client, project, and timesheet
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		timesheetID := testDB.InsertTestTimesheet(t, projectID, "2024-01-15", "5.0", "100.00", "Original work")

		// Update the timesheet
		form := url.Values{}
		form.Add("work_date", "2024-01-16")
		form.Add("hours_worked", "8.0")
		form.Add("hourly_rate", "125.00")
		form.Add("description", "Updated work description")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/update/%d", timesheetID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, fmt.Sprintf("/project/view/%d", projectID))

		// Verify the timesheet was actually updated in the database
		timesheet, err := app.timesheets.Get(timesheetID)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-16", timesheet.WorkDate.Format("2006-01-02"))
		assert.Equal(t, 8.0, timesheet.HoursWorked)
		assert.Equal(t, 125.0, timesheet.HourlyRate)
		assert.Equal(t, "Updated work description", timesheet.Description)
	})

	t.Run("validation error - empty work date", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		timesheetID := testDB.InsertTestTimesheet(t, projectID, "2024-01-15", "5.0", "100.00", "Work")

		// Try to update with empty work date
		form := url.Values{}
		form.Add("work_date", "")
		form.Add("hours_worked", "5.0")
		form.Add("hourly_rate", "100.00")
		form.Add("description", "Work")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/update/%d", timesheetID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Work date is required")

		// Verify the timesheet was NOT updated
		timesheet, err := app.timesheets.Get(timesheetID)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-15", timesheet.WorkDate.Format("2006-01-02"))
	})

	t.Run("validation error - empty hours worked", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		timesheetID := testDB.InsertTestTimesheet(t, projectID, "2024-01-15", "5.0", "100.00", "Work")

		// Try to update with empty hours
		form := url.Values{}
		form.Add("work_date", "2024-01-15")
		form.Add("hours_worked", "")
		form.Add("hourly_rate", "100.00")
		form.Add("description", "Work")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/update/%d", timesheetID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Hours worked is required")

		// Verify the timesheet was NOT updated
		timesheet, err := app.timesheets.Get(timesheetID)
		require.NoError(t, err)
		assert.Equal(t, 5.0, timesheet.HoursWorked)
	})

	t.Run("validation error - invalid hours worked", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		timesheetID := testDB.InsertTestTimesheet(t, projectID, "2024-01-15", "5.0", "100.00", "Work")

		// Try to update with negative hours
		form := url.Values{}
		form.Add("work_date", "2024-01-15")
		form.Add("hours_worked", "-5.0")
		form.Add("hourly_rate", "100.00")
		form.Add("description", "Work")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/update/%d", timesheetID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Hours worked must be a positive number")
	})

	t.Run("validation error - empty description", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		timesheetID := testDB.InsertTestTimesheet(t, projectID, "2024-01-15", "5.0", "100.00", "Original description")

		// Try to update with empty description
		form := url.Values{}
		form.Add("work_date", "2024-01-15")
		form.Add("hours_worked", "5.0")
		form.Add("hourly_rate", "100.00")
		form.Add("description", "")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/update/%d", timesheetID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

		// Verify the timesheet was NOT updated
		timesheet, err := app.timesheets.Get(timesheetID)
		require.NoError(t, err)
		assert.Equal(t, "Original description", timesheet.Description)
	})

	t.Run("form value preservation on validation error", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		timesheetID := testDB.InsertTestTimesheet(t, projectID, "2024-01-15", "5.0", "100.00", "Original")

		// Submit form with invalid data
		form := url.Values{}
		form.Add("work_date", "2024-01-20")
		form.Add("hours_worked", "10.5")
		form.Add("hourly_rate", "150.00")
		form.Add("description", "") // Invalid - empty description

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/update/%d", timesheetID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

		// Verify the timesheet was NOT updated (still has original values)
		timesheet, err := app.timesheets.Get(timesheetID)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-15", timesheet.WorkDate.Format("2006-01-02"))
		assert.Equal(t, 5.0, timesheet.HoursWorked)
		assert.Equal(t, 100.0, timesheet.HourlyRate)
		assert.Equal(t, "Original", timesheet.Description)
	})

	t.Run("update non-existent timesheet", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")

		form := url.Values{}
		form.Add("work_date", "2024-01-15")
		form.Add("hours_worked", "5.0")
		form.Add("hourly_rate", "100.00")
		form.Add("description", "Work")

		req := httptest.NewRequest(http.MethodPost, "/project/timesheet/update/999", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "999")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update with invalid ID", func(t *testing.T) {
		form := url.Values{}
		form.Add("work_date", "2024-01-15")
		form.Add("hours_worked", "5.0")
		form.Add("hourly_rate", "100.00")
		form.Add("description", "Work")

		req := httptest.NewRequest(http.MethodPost, "/project/timesheet/update/invalid", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "invalid")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestTimesheetDelete tests the timesheet delete operation
func TestTimesheetDeleteHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful timesheet delete", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup: Create client, project, and timesheet
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		timesheetID := testDB.InsertTestTimesheet(t, projectID, "2024-01-15", "5.0", "100.00", "Work to delete")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/delete/%d", timesheetID), nil)
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetDelete))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, fmt.Sprintf("/project/view/%d", projectID))

		// Verify the timesheet was soft deleted (no longer appears in GetByProject)
		timesheets, err := app.timesheets.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, timesheets)
	})

	t.Run("delete non-existent timesheet", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")

		req := httptest.NewRequest(http.MethodPost, "/project/timesheet/delete/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetDelete))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/project/timesheet/delete/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetDelete))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestTimesheetWorkflow tests the complete timesheet lifecycle
func TestTimesheetWorkflow(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("complete timesheet workflow", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// 1. Setup: Create client and project
		clientID := testDB.InsertTestClient(t, "Workflow Client")
		projectID := testDB.InsertTestProject(t, "Workflow Project", clientID)

		// 2. Create a timesheet
		form := url.Values{}
		form.Add("work_date", time.Now().Format("2006-01-02"))
		form.Add("hours_worked", "8.0")
		form.Add("hourly_rate", "100.00")
		form.Add("description", "Initial timesheet")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/create/%d", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetCreatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// 3. Verify timesheet was created
		timesheets, err := app.timesheets.GetByProject(projectID)
		require.NoError(t, err)
		require.Len(t, timesheets, 1)
		timesheetID := timesheets[0].ID
		assert.Equal(t, 8.0, timesheets[0].HoursWorked)
		assert.Equal(t, "Initial timesheet", timesheets[0].Description)

		// 4. Update the timesheet
		form = url.Values{}
		form.Add("work_date", time.Now().Format("2006-01-02"))
		form.Add("hours_worked", "10.0")
		form.Add("hourly_rate", "100.00")
		form.Add("description", "Updated timesheet")

		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/update/%d", timesheetID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// 5. Verify timesheet was updated
		timesheet, err := app.timesheets.Get(timesheetID)
		require.NoError(t, err)
		assert.Equal(t, 10.0, timesheet.HoursWorked)
		assert.Equal(t, "Updated timesheet", timesheet.Description)

		// 6. Delete the timesheet
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/delete/%d", timesheetID), nil)
		req.SetPathValue("id", strconv.Itoa(timesheetID))
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetDelete))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// 7. Verify timesheet was deleted
		timesheets, err = app.timesheets.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, timesheets)
	})
}
