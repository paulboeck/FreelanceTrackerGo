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

func setupProjectTest(t *testing.T) (*ProjectHandlersFull, *testutil.TestDatabase, int, int) {
	testDB := testutil.SetupTestSQLite(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create test client
	clientID := testDB.InsertTestClient(t, "Test Client")

	// Create handlers
	projectModel := models.NewProjectModel(testDB.DB)
	timesheetModel := models.NewTimesheetModel(testDB.DB)
	invoiceModel := models.NewInvoiceModel(testDB.DB)
	handlers := NewProjectHandlersFull(projectModel, timesheetModel, invoiceModel)

	return handlers, testDB, userID, clientID
}

func TestProjectHandlers_ListProjects(t *testing.T) {
	handlers, testDB, userID, clientID := setupProjectTest(t)
	defer testDB.Cleanup(t)

	// Create some test projects
	testDB.InsertTestProject(t, "Project 1", clientID)
	testDB.InsertTestProject(t, "Project 2", clientID)
	testDB.InsertTestProject(t, "Project 3", clientID)

	t.Run("list all projects", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.ListProjects(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		projects, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, projects, 3)

		// Verify metadata
		assert.NotNil(t, resp.Meta)
	})

	t.Run("list projects with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects?page=1&pageSize=2", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.ListProjects(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		projects, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, projects, 2)

		// Verify pagination metadata
		assert.NotNil(t, resp.Meta)
		assert.Equal(t, 1, resp.Meta.Page)
		assert.Equal(t, 2, resp.Meta.PageSize)
		assert.Equal(t, 3, resp.Meta.Total)
	})

	t.Run("search projects", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects?search=Project+1", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.ListProjects(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		projects, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, projects, 1)

		project := projects[0].(map[string]interface{})
		assert.Equal(t, "Project 1", project["name"])
	})
}

func TestProjectHandlers_GetProject(t *testing.T) {
	handlers, testDB, userID, clientID := setupProjectTest(t)
	defer testDB.Cleanup(t)

	// Create test project
	projectID := testDB.InsertTestProject(t, "Test Project", clientID)

	t.Run("get existing project", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%d", projectID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.GetProject(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		project, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Test Project", project["name"])
		assert.Equal(t, float64(projectID), project["id"])
	})

	t.Run("get nonexistent project", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/99999", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.GetProject(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("get project with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/invalid", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "invalid"},
		}

		handlers.GetProject(rr, req, ps)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestProjectHandlers_CreateProject(t *testing.T) {
	handlers, testDB, userID, clientID := setupProjectTest(t)
	defer testDB.Cleanup(t)

	t.Run("create valid project", func(t *testing.T) {
		deadline := "2025-12-31"
		reqBody := ProjectRequest{
			Name:       "New Project",
			ClientID:   clientID,
			Status:     "active",
			HourlyRate: 150.00,
			Deadline:   &deadline,
			Notes:      "Test project notes",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateProject(rr, req, nil)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		project, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "New Project", project["name"])
		assert.Equal(t, float64(150), project["hourlyRate"])
		assert.Equal(t, "active", project["status"])
	})

	t.Run("create project without name", func(t *testing.T) {
		reqBody := ProjectRequest{
			Name:       "",
			ClientID:   clientID,
			Status:     "active",
			HourlyRate: 100.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateProject(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Equal(t, ErrCodeValidation, resp.Error.Code)
	})

	t.Run("create project with negative hourly rate", func(t *testing.T) {
		reqBody := ProjectRequest{
			Name:       "Test Project",
			ClientID:   clientID,
			Status:     "active",
			HourlyRate: -50.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateProject(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("create project with invalid date format", func(t *testing.T) {
		invalidDate := "2025/12/31"
		reqBody := ProjectRequest{
			Name:       "Test Project",
			ClientID:   clientID,
			Status:     "active",
			HourlyRate: 100.00,
			Deadline:   &invalidDate,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateProject(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestProjectHandlers_UpdateProject(t *testing.T) {
	handlers, testDB, userID, clientID := setupProjectTest(t)
	defer testDB.Cleanup(t)

	// Create test project
	projectID := testDB.InsertTestProject(t, "Original Project", clientID)

	t.Run("update existing project", func(t *testing.T) {
		deadline := "2026-01-15"
		reqBody := ProjectRequest{
			Name:       "Updated Project",
			ClientID:   clientID,
			Status:     "completed",
			HourlyRate: 175.00,
			Deadline:   &deadline,
			Notes:      "Updated notes",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/projects/%d", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.UpdateProject(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		project, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Updated Project", project["name"])
		assert.Equal(t, "completed", project["status"])
		assert.Equal(t, float64(175), project["hourlyRate"])
	})

	t.Run("update nonexistent project", func(t *testing.T) {
		reqBody := ProjectRequest{
			Name:       "Updated Project",
			ClientID:   clientID,
			Status:     "active",
			HourlyRate: 100.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/99999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.UpdateProject(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestProjectHandlers_UpdateProjectStatus(t *testing.T) {
	handlers, testDB, userID, clientID := setupProjectTest(t)
	defer testDB.Cleanup(t)

	// Create test project
	projectID := testDB.InsertTestProject(t, "Status Test Project", clientID)

	t.Run("update project status", func(t *testing.T) {
		reqBody := UpdateStatusRequest{
			Status: "completed",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/projects/%d/status", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.UpdateProjectStatus(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "completed", data["status"])
	})

	t.Run("update status with empty status", func(t *testing.T) {
		reqBody := UpdateStatusRequest{
			Status: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/projects/%d/status", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.UpdateProjectStatus(rr, req, ps)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestProjectHandlers_DeleteProject(t *testing.T) {
	handlers, testDB, userID, clientID := setupProjectTest(t)
	defer testDB.Cleanup(t)

	// Create test project
	projectID := testDB.InsertTestProject(t, "Project to Delete", clientID)

	t.Run("delete existing project", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/projects/%d", projectID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.DeleteProject(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data["message"], "deleted successfully")
	})

	t.Run("delete nonexistent project", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/99999", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.DeleteProject(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestProjectHandlers_GetProjectTimesheets(t *testing.T) {
	handlers, testDB, userID, clientID := setupProjectTest(t)
	defer testDB.Cleanup(t)

	// Create test project
	projectID := testDB.InsertTestProject(t, "Timesheet Test Project", clientID)

	// Create some timesheets
	workDate := time.Now().Format("2006-01-02")
	testDB.InsertTestTimesheet(t, projectID, workDate, "5.0", "100.0", "Worked on feature A")
	testDB.InsertTestTimesheet(t, projectID, workDate, "3.5", "100.0", "Worked on feature B")

	t.Run("get project timesheets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%d/timesheets", projectID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:read timesheets:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.GetProjectTimesheets(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		timesheets, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, timesheets, 2)
	})
}

func TestProjectHandlers_GetProjectInvoices(t *testing.T) {
	handlers, testDB, userID, clientID := setupProjectTest(t)
	defer testDB.Cleanup(t)

	// Create test project
	projectID := testDB.InsertTestProject(t, "Invoice Test Project", clientID)

	// Create some invoices
	invoiceDate := time.Now().Format("2006-01-02")
	testDB.InsertTestInvoice(t, projectID, invoiceDate, "", "Net 30", "1500.00")
	testDB.InsertTestInvoice(t, projectID, invoiceDate, "", "Net 15", "2000.00")

	t.Run("get project invoices", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%d/invoices", projectID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:read invoices:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.GetProjectInvoices(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		invoices, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, invoices, 2)
	})
}
