package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func setupTimesheetTest(t *testing.T) (*TimesheetHandlers, *testutil.TestDatabase, int, int) {
	testDB := testutil.SetupTestSQLite(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create test client and project
	clientID := testDB.InsertTestClient(t, "Test Client")
	projectID := testDB.InsertTestProject(t, "Test Project", clientID)

	// Create handlers
	timesheetModel := models.NewTimesheetModel(testDB.DB)
	projectModel := models.NewProjectModel(testDB.DB)
	handlers := NewTimesheetHandlers(timesheetModel, projectModel)

	return handlers, testDB, userID, projectID
}

func TestTimesheetHandlers_GetTimesheet(t *testing.T) {
	handlers, testDB, userID, projectID := setupTimesheetTest(t)
	defer testDB.Cleanup(t)

	// Create test timesheet
	workDate := time.Now().Format("2006-01-02")
	timesheetID := testDB.InsertTestTimesheet(t, projectID, workDate, "5.0", "100.0", "Worked on feature")

	t.Run("get existing timesheet", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/timesheets/%d", timesheetID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "timesheets:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", timesheetID)},
		}

		handlers.GetTimesheet(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		timesheet, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(timesheetID), timesheet["id"])
		assert.Equal(t, "Worked on feature", timesheet["description"])
	})

	t.Run("get nonexistent timesheet", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/timesheets/99999", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "timesheets:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.GetTimesheet(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestTimesheetHandlers_CreateTimesheet(t *testing.T) {
	handlers, testDB, userID, projectID := setupTimesheetTest(t)
	defer testDB.Cleanup(t)

	t.Run("create valid timesheet", func(t *testing.T) {
		reqBody := TimesheetRequest{
			WorkDate:     time.Now().Format("2006-01-02"),
			HoursWorked:  7.5,
			HourlyRate:   125.00,
			Description:  "Backend development",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/timesheets", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "timesheets:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.CreateTimesheet(rr, req, ps)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		timesheet, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(7.5), timesheet["hoursWorked"])
		assert.Equal(t, "Backend development", timesheet["description"])
	})

	t.Run("create timesheet with invalid date format", func(t *testing.T) {
		reqBody := TimesheetRequest{
			WorkDate:     "2025/01/15",
			HoursWorked:  7.5,
			HourlyRate:   125.00,
			Description:  "Work",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/timesheets", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "timesheets:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.CreateTimesheet(rr, req, ps)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("create timesheet with negative hours", func(t *testing.T) {
		reqBody := TimesheetRequest{
			WorkDate:     time.Now().Format("2006-01-02"),
			HoursWorked:  -5.0,
			HourlyRate:   125.00,
			Description:  "Work",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/timesheets", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "timesheets:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.CreateTimesheet(rr, req, ps)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestTimesheetHandlers_UpdateTimesheet(t *testing.T) {
	handlers, testDB, userID, projectID := setupTimesheetTest(t)
	defer testDB.Cleanup(t)

	// Create test timesheet
	workDate := time.Now().Format("2006-01-02")
	timesheetID := testDB.InsertTestTimesheet(t, projectID, workDate, "5.0", "100.0", "Original work")

	t.Run("update existing timesheet", func(t *testing.T) {
		reqBody := TimesheetRequest{
			WorkDate:     workDate,
			HoursWorked:  8.0,
			HourlyRate:   150.00,
			Description:  "Updated work description",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/timesheets/%d", timesheetID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "timesheets:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", timesheetID)},
		}

		handlers.UpdateTimesheet(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		timesheet, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(8.0), timesheet["hoursWorked"])
		assert.Equal(t, "Updated work description", timesheet["description"])
	})

	t.Run("update nonexistent timesheet", func(t *testing.T) {
		reqBody := TimesheetRequest{
			WorkDate:     workDate,
			HoursWorked:  8.0,
			HourlyRate:   150.00,
			Description:  "Work",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/timesheets/99999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "timesheets:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.UpdateTimesheet(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestTimesheetHandlers_DeleteTimesheet(t *testing.T) {
	handlers, testDB, userID, projectID := setupTimesheetTest(t)
	defer testDB.Cleanup(t)

	// Create test timesheet
	workDate := time.Now().Format("2006-01-02")
	timesheetID := testDB.InsertTestTimesheet(t, projectID, workDate, "5.0", "100.0", "Work to delete")

	t.Run("delete existing timesheet", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/timesheets/%d", timesheetID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "timesheets:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", timesheetID)},
		}

		handlers.DeleteTimesheet(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data["message"], "deleted successfully")
	})

	t.Run("delete nonexistent timesheet", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/timesheets/99999", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "timesheets:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.DeleteTimesheet(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
