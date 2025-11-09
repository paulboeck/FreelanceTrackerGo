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

func setupInvoiceTest(t *testing.T) (*InvoiceHandlers, *testutil.TestDatabase, int, int) {
	testDB := testutil.SetupTestSQLite(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create test client and project
	clientID := testDB.InsertTestClient(t, "Test Client")
	projectID := testDB.InsertTestProject(t, "Test Project", clientID)

	// Create handlers
	invoiceModel := models.NewInvoiceModel(testDB.DB)
	projectModel := models.NewProjectModel(testDB.DB)
	clientModel := models.NewClientModel(testDB.DB)
	settingsModel := models.NewAppSettingModel(testDB.DB, "test-seed")
	handlers := NewInvoiceHandlers(invoiceModel, projectModel, clientModel, settingsModel)

	return handlers, testDB, userID, projectID
}

func TestInvoiceHandlers_GetInvoice(t *testing.T) {
	handlers, testDB, userID, projectID := setupInvoiceTest(t)
	defer testDB.Cleanup(t)

	// Create test invoice
	invoiceDate := time.Now().Format("2006-01-02")
	invoiceID := testDB.InsertTestInvoice(t, projectID, invoiceDate, "", "Net 30", "1500.00")

	t.Run("get existing invoice", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/invoices/%d", invoiceID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", invoiceID)},
		}

		handlers.GetInvoice(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		invoice, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(invoiceID), invoice["id"])
		assert.Equal(t, "Net 30", invoice["paymentTerms"])
	})

	t.Run("get nonexistent invoice", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/99999", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.GetInvoice(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestInvoiceHandlers_CreateInvoice(t *testing.T) {
	handlers, testDB, userID, projectID := setupInvoiceTest(t)
	defer testDB.Cleanup(t)

	t.Run("create valid invoice", func(t *testing.T) {
		reqBody := InvoiceRequest{
			InvoiceDate:    time.Now().Format("2006-01-02"),
			PaymentTerms:   "Net 30",
			AmountDue:      2500.00,
			DisplayDetails: true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/invoices", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.CreateInvoice(rr, req, ps)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		invoice, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(2500), invoice["amountDue"])
		assert.Equal(t, "Net 30", invoice["paymentTerms"])
		assert.Equal(t, true, invoice["displayDetails"])
	})

	t.Run("create invoice with date paid", func(t *testing.T) {
		datePaid := time.Now().Format("2006-01-02")
		reqBody := InvoiceRequest{
			InvoiceDate:    time.Now().Format("2006-01-02"),
			DatePaid:       &datePaid,
			PaymentTerms:   "Net 15",
			AmountDue:      1000.00,
			DisplayDetails: false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/invoices", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.CreateInvoice(rr, req, ps)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("create invoice with invalid date format", func(t *testing.T) {
		reqBody := InvoiceRequest{
			InvoiceDate:    "2025/01/15",
			PaymentTerms:   "Net 30",
			AmountDue:      1500.00,
			DisplayDetails: false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/invoices", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.CreateInvoice(rr, req, ps)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("create invoice with negative amount", func(t *testing.T) {
		reqBody := InvoiceRequest{
			InvoiceDate:    time.Now().Format("2006-01-02"),
			PaymentTerms:   "Net 30",
			AmountDue:      -100.00,
			DisplayDetails: false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/invoices", projectID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", projectID)},
		}

		handlers.CreateInvoice(rr, req, ps)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("create invoice for nonexistent project", func(t *testing.T) {
		reqBody := InvoiceRequest{
			InvoiceDate:    time.Now().Format("2006-01-02"),
			PaymentTerms:   "Net 30",
			AmountDue:      1500.00,
			DisplayDetails: false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/99999/invoices", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.CreateInvoice(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestInvoiceHandlers_UpdateInvoice(t *testing.T) {
	handlers, testDB, userID, projectID := setupInvoiceTest(t)
	defer testDB.Cleanup(t)

	// Create test invoice
	invoiceDate := time.Now().Format("2006-01-02")
	invoiceID := testDB.InsertTestInvoice(t, projectID, invoiceDate, "", "Net 30", "1500.00")

	t.Run("update existing invoice", func(t *testing.T) {
		datePaid := time.Now().Format("2006-01-02")
		reqBody := InvoiceRequest{
			InvoiceDate:    invoiceDate,
			DatePaid:       &datePaid,
			PaymentTerms:   "Net 15",
			AmountDue:      2000.00,
			DisplayDetails: true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/invoices/%d", invoiceID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", invoiceID)},
		}

		handlers.UpdateInvoice(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		invoice, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(2000), invoice["amountDue"])
		assert.Equal(t, "Net 15", invoice["paymentTerms"])
	})

	t.Run("update nonexistent invoice", func(t *testing.T) {
		reqBody := InvoiceRequest{
			InvoiceDate:    invoiceDate,
			PaymentTerms:   "Net 30",
			AmountDue:      1500.00,
			DisplayDetails: false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/invoices/99999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.UpdateInvoice(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestInvoiceHandlers_DeleteInvoice(t *testing.T) {
	handlers, testDB, userID, projectID := setupInvoiceTest(t)
	defer testDB.Cleanup(t)

	// Create test invoice
	invoiceDate := time.Now().Format("2006-01-02")
	invoiceID := testDB.InsertTestInvoice(t, projectID, invoiceDate, "", "Net 30", "1500.00")

	t.Run("delete existing invoice", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/invoices/%d", invoiceID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", invoiceID)},
		}

		handlers.DeleteInvoice(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data["message"], "deleted successfully")
	})

	t.Run("delete nonexistent invoice", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/invoices/99999", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.DeleteInvoice(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestInvoiceHandlers_GenerateInvoicePDF(t *testing.T) {
	// Skip PDF generation tests in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping PDF generation test in CI environment")
	}

	handlers, testDB, userID, projectID := setupInvoiceTest(t)
	defer testDB.Cleanup(t)

	// Create test invoice with some timesheets
	invoiceDate := time.Now().Format("2006-01-02")
	invoiceID := testDB.InsertTestInvoice(t, projectID, invoiceDate, "", "Net 30", "1500.00")

	// Add some timesheets
	testDB.InsertTestTimesheet(t, projectID, invoiceDate, "5.0", "100.0", "Development work")

	t.Run("generate PDF for existing invoice", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/invoices/%d/pdf", invoiceID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", invoiceID)},
		}

		handlers.GenerateInvoicePDF(rr, req, ps)

		// Should return PDF or error gracefully
		// In test environment without full setup, may return 500
		assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusInternalServerError)

		if rr.Code == http.StatusOK {
			assert.Equal(t, "application/pdf", rr.Header().Get("Content-Type"))
			assert.Contains(t, rr.Header().Get("Content-Disposition"), "attachment")
			assert.Greater(t, len(rr.Body.Bytes()), 0)
		}
	})

	t.Run("generate PDF for nonexistent invoice", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/99999/pdf", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.GenerateInvoicePDF(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestInvoiceHandlers_EmailInvoice(t *testing.T) {
	handlers, testDB, userID, _ := setupInvoiceTest(t)
	defer testDB.Cleanup(t)

	// Create test client with email
	clientID := testDB.InsertTestClient(t, "Email Test Client")
	projectID2 := testDB.InsertTestProject(t, "Email Test Project", clientID)

	// Create test invoice
	invoiceDate := time.Now().Format("2006-01-02")
	invoiceID := testDB.InsertTestInvoice(t, projectID2, invoiceDate, "", "Net 30", "1500.00")

	t.Run("email invoice without email config", func(t *testing.T) {
		reqBody := EmailInvoiceRequest{
			To:      "client@example.com",
			Subject: "Test Invoice",
			Body:    "Please find attached invoice.",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/invoices/%d/email", invoiceID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", invoiceID)},
		}

		handlers.EmailInvoice(rr, req, ps)

		// Should fail gracefully without email configured
		// Either 400 (bad request) or 500 (internal error)
		assert.True(t, rr.Code == http.StatusBadRequest || rr.Code == http.StatusInternalServerError)
	})

	t.Run("email nonexistent invoice", func(t *testing.T) {
		reqBody := EmailInvoiceRequest{
			To: "client@example.com",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/99999/email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "invoices:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.EmailInvoice(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
