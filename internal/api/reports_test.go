package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func setupReportsTest(t *testing.T) (*ReportsHandlers, *testutil.TestDatabase, int, int) {
	testDB := testutil.SetupTestSQLite(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create test client and project
	clientID := testDB.InsertTestClient(t, "Test Client")
	projectID := testDB.InsertTestProject(t, "Test Project", clientID)

	// Create handlers
	invoiceModel := models.NewInvoiceModel(testDB.DB)
	handlers := NewReportsHandlers(invoiceModel)

	return handlers, testDB, userID, projectID
}

func TestReportsHandlers_GetIncomeReport(t *testing.T) {
	handlers, testDB, userID, projectID := setupReportsTest(t)
	defer testDB.Cleanup(t)

	// Create test invoices for 2024 with different months
	year2024 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	datePaid1 := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
	datePaid2 := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	datePaid3 := time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)

	testDB.InsertTestInvoiceWithTime(t, projectID, year2024, &datePaid1, "Net 30", 1500.00, false)
	testDB.InsertTestInvoiceWithTime(t, projectID, year2024, &datePaid2, "Net 30", 2000.00, false)
	testDB.InsertTestInvoiceWithTime(t, projectID, year2024, &datePaid3, "Net 30", 1000.00, false)

	// Create unpaid invoice (should not be included)
	testDB.InsertTestInvoiceWithTime(t, projectID, year2024, nil, "Net 30", 5000.00, false)

	// Create invoice for different year (should not be included)
	year2023 := time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC)
	datePaid2023 := time.Date(2023, 12, 20, 0, 0, 0, 0, time.UTC)
	testDB.InsertTestInvoiceWithTime(t, projectID, year2023, &datePaid2023, "Net 30", 3000.00, false)

	t.Run("get income report for year with data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/income?year=2024", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "reports:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.GetIncomeReport(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		report, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		// Verify year
		assert.Equal(t, float64(2024), report["year"])

		// Verify total (1500 + 2000 + 1000 = 4500)
		assert.Equal(t, float64(4500), report["total"])

		// Verify months array
		months, ok := report["months"].([]interface{})
		require.True(t, ok)
		assert.Len(t, months, 12)

		// Check January (month 1) has 1500
		month1 := months[0].(map[string]interface{})
		assert.Equal(t, float64(1), month1["month"])
		assert.Equal(t, float64(1500), month1["amount"])

		// Check March (month 3) has 3000 (2000 + 1000)
		month3 := months[2].(map[string]interface{})
		assert.Equal(t, float64(3), month3["month"])
		assert.Equal(t, float64(3000), month3["amount"])

		// Check February (month 2) has 0
		month2 := months[1].(map[string]interface{})
		assert.Equal(t, float64(2), month2["month"])
		assert.Equal(t, float64(0), month2["amount"])
	})

	t.Run("get income report for year without data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/income?year=2025", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "reports:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.GetIncomeReport(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		report, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		// Verify year
		assert.Equal(t, float64(2025), report["year"])

		// Verify total is 0
		assert.Equal(t, float64(0), report["total"])

		// Verify all months have 0 amount
		months, ok := report["months"].([]interface{})
		require.True(t, ok)
		assert.Len(t, months, 12)

		for i, month := range months {
			m := month.(map[string]interface{})
			assert.Equal(t, float64(i+1), m["month"])
			assert.Equal(t, float64(0), m["amount"])
		}
	})

	t.Run("get income report without year parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/income", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "reports:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.GetIncomeReport(rr, req, nil)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "Year parameter is required")
	})

	t.Run("get income report with invalid year format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/income?year=invalid", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "reports:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.GetIncomeReport(rr, req, nil)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "Invalid year format")
	})

	t.Run("get income report with year below 1900", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/income?year=1899", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "reports:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.GetIncomeReport(rr, req, nil)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "Invalid year format")
	})

	t.Run("get income report with year above 2100", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/income?year=2101", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "reports:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.GetIncomeReport(rr, req, nil)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "Invalid year format")
	})

	t.Run("verify 2023 data separate from 2024", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/income?year=2023", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "reports:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.GetIncomeReport(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		report, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		// Verify year
		assert.Equal(t, float64(2023), report["year"])

		// Verify total is 3000 (only the 2023 invoice)
		assert.Equal(t, float64(3000), report["total"])

		// Check December (month 12) has 3000
		months, ok := report["months"].([]interface{})
		require.True(t, ok)
		month12 := months[11].(map[string]interface{})
		assert.Equal(t, float64(12), month12["month"])
		assert.Equal(t, float64(3000), month12["amount"])
	})
}
