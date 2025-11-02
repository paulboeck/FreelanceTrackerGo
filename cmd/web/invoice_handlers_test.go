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

// TestInvoiceCreate tests the invoice create form display
func TestInvoiceCreate(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("display invoice create form", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup: Create client, project, and timesheet for amount calculation
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProjectWithRate(t, clientID, "Test Project", 100.0)
		testDB.InsertTestTimesheetWithTime(t, projectID, time.Now(), 5.0, 100.0, "Work")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/project/invoice/create/%d", projectID), nil)
		req.SetPathValue("id", strconv.Itoa(projectID))
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceCreate))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Form should be pre-populated with calculated amount (5 hours * $100 = $500)
		assert.Contains(t, body, "500.00")
		assert.Contains(t, body, time.Now().Format("2006-01-02"))
	})

	t.Run("invoice create for non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")

		req := httptest.NewRequest(http.MethodGet, "/project/invoice/create/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceCreate))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invoice create form with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/project/invoice/create/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceCreate))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestInvoiceCreatePost tests creating invoices
func TestInvoiceCreatePost(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful invoice creation", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		testDB.InsertTestTimesheet(t, projectID, "2024-01-15", "5.0", "100.00", "Work")

		// Create invoice
		form := url.Values{}
		form.Add("invoice_date", "2024-01-20")
		form.Add("payment_terms", "Net 30")
		form.Add("amount_due", "500.00")
		form.Add("display_details", "true")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/create/%d", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceCreatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, fmt.Sprintf("/project/view/%d", projectID))

		// Verify the invoice was created
		invoices, err := app.invoices.GetByProject(projectID)
		require.NoError(t, err)
		require.Len(t, invoices, 1)
		assert.Equal(t, 500.0, invoices[0].AmountDue)
		assert.Equal(t, "Net 30", invoices[0].PaymentTerms)
		assert.True(t, invoices[0].DisplayDetails)
		assert.Nil(t, invoices[0].DatePaid)
	})

	t.Run("validation error - empty invoice date", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)

		// Try to create with empty invoice date
		form := url.Values{}
		form.Add("invoice_date", "")
		form.Add("payment_terms", "Net 30")
		form.Add("amount_due", "500.00")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/create/%d", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceCreatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

		// Verify no invoice was created
		invoices, err := app.invoices.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, invoices)
	})

	t.Run("validation error - empty amount due", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)

		// Try to create with empty amount
		form := url.Values{}
		form.Add("invoice_date", "2024-01-20")
		form.Add("payment_terms", "Net 30")
		form.Add("amount_due", "")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/create/%d", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceCreatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

		// Verify no invoice was created
		invoices, err := app.invoices.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, invoices)
	})

	t.Run("validation error - invalid amount format", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)

		// Try to create with invalid amount
		form := url.Values{}
		form.Add("invoice_date", "2024-01-20")
		form.Add("payment_terms", "Net 30")
		form.Add("amount_due", "invalid")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/create/%d", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceCreatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("create invoice for non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")

		form := url.Values{}
		form.Add("invoice_date", "2024-01-20")
		form.Add("payment_terms", "Net 30")
		form.Add("amount_due", "500.00")

		req := httptest.NewRequest(http.MethodPost, "/project/invoice/create/999", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "999")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceCreatePost))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestInvoiceUpdate tests the invoice update form display
func TestInvoiceUpdate(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("display invoice update form", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		invoiceID := testDB.InsertTestInvoice(t, projectID, "2024-01-20", "", "Net 30", "500.00")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/project/invoice/update/%d", invoiceID), nil)
		req.SetPathValue("id", strconv.Itoa(invoiceID))
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceUpdate))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Form should be pre-populated with current values
		assert.Contains(t, body, "2024-01-20")
		assert.Contains(t, body, "Net 30")
		assert.Contains(t, body, "500.00")
	})

	t.Run("invoice update form for non-existent invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")

		req := httptest.NewRequest(http.MethodGet, "/project/invoice/update/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceUpdate))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invoice update form with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/project/invoice/update/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceUpdate))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestInvoiceUpdatePost tests updating invoices
func TestInvoiceUpdatePost(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful invoice update", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		invoiceID := testDB.InsertTestInvoice(t, projectID, "2024-01-20", "", "Net 30", "500.00")

		// Update invoice
		form := url.Values{}
		form.Add("invoice_date", "2024-01-25")
		form.Add("payment_terms", "Net 60")
		form.Add("amount_due", "750.00")
		form.Add("display_details", "true")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/update/%d", invoiceID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(invoiceID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, fmt.Sprintf("/project/view/%d", projectID))

		// Verify the invoice was updated
		invoice, err := app.invoices.Get(invoiceID)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-25", invoice.InvoiceDate.Format("2006-01-02"))
		assert.Equal(t, "Net 60", invoice.PaymentTerms)
		assert.Equal(t, 750.0, invoice.AmountDue)
		assert.True(t, invoice.DisplayDetails)
	})

	t.Run("validation error - empty invoice date", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		invoiceID := testDB.InsertTestInvoice(t, projectID, "2024-01-20", "", "Net 30", "500.00")

		// Try to update with empty date
		form := url.Values{}
		form.Add("invoice_date", "")
		form.Add("payment_terms", "Net 30")
		form.Add("amount_due", "500.00")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/update/%d", invoiceID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(invoiceID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

		// Verify invoice was NOT updated
		invoice, err := app.invoices.Get(invoiceID)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-20", invoice.InvoiceDate.Format("2006-01-02"))
	})

	t.Run("update non-existent invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")

		form := url.Values{}
		form.Add("invoice_date", "2024-01-20")
		form.Add("payment_terms", "Net 30")
		form.Add("amount_due", "500.00")

		req := httptest.NewRequest(http.MethodPost, "/project/invoice/update/999", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "999")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceUpdatePost))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update with invalid ID", func(t *testing.T) {
		form := url.Values{}
		form.Add("invoice_date", "2024-01-20")
		form.Add("payment_terms", "Net 30")
		form.Add("amount_due", "500.00")

		req := httptest.NewRequest(http.MethodPost, "/project/invoice/update/invalid", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "invalid")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceUpdatePost))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestInvoiceDelete tests the invoice delete operation
func TestInvoiceDeleteHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful invoice delete", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Setup
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		invoiceID := testDB.InsertTestInvoice(t, projectID, "2024-01-20", "", "Net 30", "500.00")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/delete/%d", invoiceID), nil)
		req.SetPathValue("id", strconv.Itoa(invoiceID))
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceDelete))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, fmt.Sprintf("/project/view/%d", projectID))

		// Verify the invoice was soft deleted
		invoices, err := app.invoices.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, invoices)
	})

	t.Run("delete non-existent invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")

		req := httptest.NewRequest(http.MethodPost, "/project/invoice/delete/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceDelete))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/project/invoice/delete/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceDelete))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestInvoiceWorkflow tests the complete invoice lifecycle
func TestInvoiceWorkflow(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("complete invoice workflow", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// 1. Setup: Create client, project, timesheet
		clientID := testDB.InsertTestClient(t, "Workflow Client")
		projectID := testDB.InsertTestProjectWithRate(t, clientID, "Workflow Project", 100.0)
		testDB.InsertTestTimesheetWithTime(t, projectID, time.Now(), 5.0, 100.0, "Work")

		// 2. Create invoice
		form := url.Values{}
		form.Add("invoice_date", time.Now().Format("2006-01-02"))
		form.Add("payment_terms", "Net 30")
		form.Add("amount_due", "500.00")
		form.Add("display_details", "true")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/create/%d", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceCreatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// 3. Verify invoice was created
		invoices, err := app.invoices.GetByProject(projectID)
		require.NoError(t, err)
		require.Len(t, invoices, 1)
		invoiceID := invoices[0].ID
		assert.Equal(t, 500.0, invoices[0].AmountDue)
		assert.Equal(t, "Net 30", invoices[0].PaymentTerms)

		// 4. Update the invoice
		form = url.Values{}
		form.Add("invoice_date", time.Now().Format("2006-01-02"))
		form.Add("payment_terms", "Net 60")
		form.Add("amount_due", "600.00")
		form.Add("display_details", "false")

		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/update/%d", invoiceID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(invoiceID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// 5. Verify invoice was updated
		invoice, err := app.invoices.Get(invoiceID)
		require.NoError(t, err)
		assert.Equal(t, 600.0, invoice.AmountDue)
		assert.Equal(t, "Net 60", invoice.PaymentTerms)
		assert.False(t, invoice.DisplayDetails)

		// 6. Delete the invoice
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/invoice/delete/%d", invoiceID), nil)
		req.SetPathValue("id", strconv.Itoa(invoiceID))
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.invoiceDelete))
		handler.ServeHTTP(rr, req)

		// Should redirect to project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// 7. Verify invoice was deleted
		invoices, err = app.invoices.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, invoices)
	})
}
