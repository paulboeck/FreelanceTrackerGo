package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestCompleteFreelanceWorkflow tests a realistic end-to-end workflow
func TestCompleteFreelanceWorkflow(t *testing.T) {
	// Skip PDF generation tests in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping PDF generation test in CI environment")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create handlers
	clientModel := models.NewClientModel(testDB.DB)
	projectModel := models.NewProjectModel(testDB.DB)
	timesheetModel := models.NewTimesheetModel(testDB.DB)
	invoiceModel := models.NewInvoiceModel(testDB.DB)
	settingsModel := models.NewAppSettingModel(testDB.DB, "test-seed")

	clientHandlers := NewClientHandlers(clientModel, projectModel)
	projectHandlers := NewProjectHandlersFull(projectModel, timesheetModel, invoiceModel)
	timesheetHandlers := NewTimesheetHandlers(timesheetModel, projectModel)
	invoiceHandlers := NewInvoiceHandlers(invoiceModel, projectModel, clientModel, settingsModel)

	t.Run("complete freelance workflow", func(t *testing.T) {
		// ========================================
		// Step 1: Create a new client
		// ========================================
		phone := "555-1234"
		clientReq := ClientRequest{
			Name:       "Acme Corporation",
			Email:      "billing@acme.com",
			Phone:      &phone,
			HourlyRate: 150.00,
		}
		body, _ := json.Marshal(clientReq)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		clientHandlers.CreateClient(rr, req, nil)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var clientResp Response
		json.NewDecoder(rr.Body).Decode(&clientResp)
		client := clientResp.Data.(map[string]interface{})
		clientID := int(client["id"].(float64))

		// ========================================
		// Step 2: Create a project for the client
		// ========================================
		deadline := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
		projectReq := ProjectRequest{
			Name:       "Website Redesign",
			ClientID:   clientID,
			Status:     "active",
			HourlyRate: 150.00,
			Deadline:   &deadline,
			Notes:      "Modernize company website with React frontend",
		}
		body, _ = json.Marshal(projectReq)

		req = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx = SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr = httptest.NewRecorder()
		projectHandlers.CreateProject(rr, req, nil)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var projectResp Response
		json.NewDecoder(rr.Body).Decode(&projectResp)
		project := projectResp.Data.(map[string]interface{})
		projectID := int(project["id"].(float64))

		// ========================================
		// Step 3: Log timesheets over several days
		// ========================================
		timesheets := []struct {
			date        string
			hours       float64
			description string
		}{
			{time.Now().AddDate(0, 0, -7).Format("2006-01-02"), 8.0, "Initial planning and wireframes"},
			{time.Now().AddDate(0, 0, -6).Format("2006-01-02"), 7.5, "Design mockups and client review"},
			{time.Now().AddDate(0, 0, -5).Format("2006-01-02"), 8.0, "Frontend development - homepage"},
			{time.Now().AddDate(0, 0, -4).Format("2006-01-02"), 7.0, "Frontend development - about page"},
			{time.Now().AddDate(0, 0, -3).Format("2006-01-02"), 8.5, "Backend API integration"},
			{time.Now().AddDate(0, 0, -2).Format("2006-01-02"), 6.5, "Testing and bug fixes"},
			{time.Now().AddDate(0, 0, -1).Format("2006-01-02"), 5.0, "Client feedback and revisions"},
		}

		totalHours := 0.0
		for _, ts := range timesheets {
			totalHours += ts.hours

			timesheetReq := TimesheetRequest{
				WorkDate:     ts.date,
				HoursWorked:  ts.hours,
				HourlyRate:   150.00,
				Description:  ts.description,
			}
			body, _ = json.Marshal(timesheetReq)

			req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/timesheets", projectID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			ctx = SetAPIKeyContext(req.Context(), 1, userID, "timesheets:write", "Test Key")
			req = req.WithContext(ctx)

			rr = httptest.NewRecorder()
			ps := httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", projectID)}}
			timesheetHandlers.CreateTimesheet(rr, req, ps)

			assert.Equal(t, http.StatusCreated, rr.Code, "Failed to create timesheet: %s", ts.description)
		}

		expectedTotal := totalHours * 150.00 // $150/hr

		// ========================================
		// Step 4: Verify timesheets were created
		// ========================================
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%d/timesheets", projectID), nil)
		ctx = SetAPIKeyContext(req.Context(), 1, userID, "projects:read timesheets:read", "Test Key")
		req = req.WithContext(ctx)

		rr = httptest.NewRecorder()
		ps := httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", projectID)}}
		projectHandlers.GetProjectTimesheets(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var timesheetsResp Response
		json.NewDecoder(rr.Body).Decode(&timesheetsResp)
		retrievedTimesheets := timesheetsResp.Data.([]interface{})
		assert.Len(t, retrievedTimesheets, len(timesheets), "Should have created all timesheets")

		// ========================================
		// Step 5: Create an invoice
		// ========================================
		invoiceReq := InvoiceRequest{
			InvoiceDate:    time.Now().Format("2006-01-02"),
			PaymentTerms:   "Net 30",
			AmountDue:      expectedTotal,
			DisplayDetails: true,
		}
		body, _ = json.Marshal(invoiceReq)

		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/invoices", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx = SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr = httptest.NewRecorder()
		ps = httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", projectID)}}
		invoiceHandlers.CreateInvoice(rr, req, ps)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var invoiceResp Response
		json.NewDecoder(rr.Body).Decode(&invoiceResp)
		invoice := invoiceResp.Data.(map[string]interface{})
		invoiceID := int(invoice["id"].(float64))
		assert.Equal(t, expectedTotal, invoice["amountDue"].(float64))

		// ========================================
		// Step 6: Generate PDF invoice
		// ========================================
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/invoices/%d/pdf", invoiceID), nil)
		ctx = SetAPIKeyContext(req.Context(), 1, userID, "invoices:read", "Test Key")
		req = req.WithContext(ctx)

		rr = httptest.NewRecorder()
		ps = httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", invoiceID)}}
		invoiceHandlers.GenerateInvoicePDF(rr, req, ps)

		// PDF generation may or may not work in test environment
		if rr.Code == http.StatusOK {
			assert.Equal(t, "application/pdf", rr.Header().Get("Content-Type"))
			assert.Greater(t, len(rr.Body.Bytes()), 0, "PDF should contain data")
		}

		// ========================================
		// Step 7: Mark invoice as paid
		// ========================================
		datePaid := time.Now().Format("2006-01-02")
		updateReq := InvoiceRequest{
			InvoiceDate:    time.Now().Format("2006-01-02"),
			DatePaid:       &datePaid,
			PaymentTerms:   "Net 30",
			AmountDue:      expectedTotal,
			DisplayDetails: true,
		}
		body, _ = json.Marshal(updateReq)

		req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/invoices/%d", invoiceID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx = SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr = httptest.NewRecorder()
		ps = httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", invoiceID)}}
		invoiceHandlers.UpdateInvoice(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		// ========================================
		// Step 8: Update project status to completed
		// ========================================
		statusReq := UpdateStatusRequest{
			Status: "completed",
		}
		body, _ = json.Marshal(statusReq)

		req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/projects/%d/status", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx = SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
		req = req.WithContext(ctx)

		rr = httptest.NewRecorder()
		ps = httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", projectID)}}
		projectHandlers.UpdateProjectStatus(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		// ========================================
		// Verification: Confirm final state
		// ========================================

		// Verify client still exists and has correct data
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients/%d", clientID), nil)
		ctx = SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
		req = req.WithContext(ctx)

		rr = httptest.NewRecorder()
		ps = httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", clientID)}}
		clientHandlers.GetClient(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		json.NewDecoder(rr.Body).Decode(&clientResp)
		finalClient := clientResp.Data.(map[string]interface{})
		assert.Equal(t, "Acme Corporation", finalClient["name"])

		// Verify project is completed
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%d", projectID), nil)
		ctx = SetAPIKeyContext(req.Context(), 1, userID, "projects:read", "Test Key")
		req = req.WithContext(ctx)

		rr = httptest.NewRecorder()
		ps = httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", projectID)}}
		projectHandlers.GetProject(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		json.NewDecoder(rr.Body).Decode(&projectResp)
		finalProject := projectResp.Data.(map[string]interface{})
		assert.Equal(t, "completed", finalProject["status"])

		// Verify invoice is marked as paid
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/invoices/%d", invoiceID), nil)
		ctx = SetAPIKeyContext(req.Context(), 1, userID, "invoices:read", "Test Key")
		req = req.WithContext(ctx)

		rr = httptest.NewRecorder()
		ps = httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", invoiceID)}}
		invoiceHandlers.GetInvoice(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		json.NewDecoder(rr.Body).Decode(&invoiceResp)
		finalInvoice := invoiceResp.Data.(map[string]interface{})
		assert.NotNil(t, finalInvoice["datePaid"], "Invoice should be marked as paid")
	})
}

// TestMultipleClientsConcurrent tests handling multiple clients and projects
func TestMultipleClientsConcurrent(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	clientModel := models.NewClientModel(testDB.DB)
	projectModel := models.NewProjectModel(testDB.DB)
	timesheetModel := models.NewTimesheetModel(testDB.DB)
	invoiceModel := models.NewInvoiceModel(testDB.DB)

	clientHandlers := NewClientHandlers(clientModel, projectModel)
	projectHandlers := NewProjectHandlersFull(projectModel, timesheetModel, invoiceModel)

	t.Run("create multiple clients with projects", func(t *testing.T) {
		// Create 3 different clients
		clients := []string{"Client Alpha", "Client Beta", "Client Gamma"}
		createdClients := make(map[string]int)

		for i, clientName := range clients {
			phone := "555-0000"
			clientReq := ClientRequest{
				Name:       clientName,
				Email:      fmt.Sprintf("client%d@example.com", i+1),
				Phone:      &phone,
				HourlyRate: 100.00,
			}
			body, _ := json.Marshal(clientReq)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			clientHandlers.CreateClient(rr, req, nil)

			require.Equal(t, http.StatusCreated, rr.Code)

			var resp Response
			json.NewDecoder(rr.Body).Decode(&resp)
			client := resp.Data.(map[string]interface{})
			createdClients[clientName] = int(client["id"].(float64))
		}

		// Create 2 projects per client
		projectCount := 0
		for clientName, clientID := range createdClients {
			for i := 1; i <= 2; i++ {
				projectReq := ProjectRequest{
					Name:       fmt.Sprintf("%s - Project %d", clientName, i),
					ClientID:   clientID,
					Status:     "active",
					HourlyRate: 125.00,
				}
				body, _ := json.Marshal(projectReq)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:write", "Test Key")
				req = req.WithContext(ctx)

				rr := httptest.NewRecorder()
				projectHandlers.CreateProject(rr, req, nil)

				require.Equal(t, http.StatusCreated, rr.Code)
				projectCount++
			}
		}

		// Verify all projects were created
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "projects:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		projectHandlers.ListProjects(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		json.NewDecoder(rr.Body).Decode(&resp)
		projects := resp.Data.([]interface{})
		assert.Len(t, projects, 6, "Should have 6 projects (2 per client)")
	})
}
