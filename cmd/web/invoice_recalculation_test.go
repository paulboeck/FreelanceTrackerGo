package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInvoiceRecalculationOnTimesheetCreate verifies that invoice amounts are automatically
// updated when a new timesheet is created on a project with an unpaid invoice
func TestInvoiceRecalculationOnTimesheetCreate(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	// Setup: Create a client
	clientID := testDB.InsertTestClient(t, "Test Client")

	// Setup: Create a project
	projectID := testDB.InsertTestProjectWithRate(t, clientID, "Test Project", 100.0)

	// Setup: Create one timesheet (5 hours at $100/hr = $500)
	testDB.InsertTestTimesheetWithTime(t, projectID, time.Now(), 5.0, 100.0, "Initial work")

	// Setup: Create an unpaid invoice with the current amount
	invoiceID := testDB.InsertTestInvoiceWithTime(t, projectID, time.Now(), nil, "Net 30", 500.0, true)

	// Verify initial invoice amount
	invoice, err := app.invoices.Get(invoiceID)
	require.NoError(t, err)
	assert.Equal(t, 500.0, invoice.AmountDue, "Initial invoice amount should be $500")

	// ACTION: Create a new timesheet (3 hours at $100/hr = $300)
	// Total should now be $800
	form := url.Values{}
	form.Add("work_date", time.Now().Format("2006-01-02"))
	form.Add("hours_worked", "3.0")
	form.Add("hourly_rate", "100.00")
	form.Add("description", "Additional work")

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/create/%d", projectID), strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", fmt.Sprintf("%d", projectID))
	rr := httptest.NewRecorder()

	// Use session middleware to handle flash messages
	handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetCreatePost))
	handler.ServeHTTP(rr, req)

	// VERIFY: Should redirect after successful creation
	assert.Equal(t, http.StatusSeeOther, rr.Code, "Should redirect after timesheet creation")

	// VERIFY: Invoice amount should be automatically updated to $800
	invoice, err = app.invoices.Get(invoiceID)
	require.NoError(t, err)
	assert.Equal(t, 800.0, invoice.AmountDue, "Invoice amount should be updated to $800 after adding 3 hours")
}

// TestInvoiceRecalculationOnTimesheetUpdate verifies that invoice amounts are automatically
// updated when a timesheet is modified on a project with an unpaid invoice
func TestInvoiceRecalculationOnTimesheetUpdate(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	// Setup: Create a client
	clientID := testDB.InsertTestClient(t, "Test Client")

	// Setup: Create a project
	projectID := testDB.InsertTestProjectWithRate(t, clientID, "Test Project", 100.0)

	// Setup: Create two timesheets
	timesheetID := testDB.InsertTestTimesheetWithTime(t, projectID, time.Now(), 5.0, 100.0, "Initial work")
	testDB.InsertTestTimesheetWithTime(t, projectID, time.Now(), 3.0, 100.0, "More work")
	// Total: 8 hours = $800

	// Setup: Create an unpaid invoice
	invoiceID := testDB.InsertTestInvoiceWithTime(t, projectID, time.Now(), nil, "Net 30", 800.0, true)

	// Verify initial invoice amount
	invoice, err := app.invoices.Get(invoiceID)
	require.NoError(t, err)
	assert.Equal(t, 800.0, invoice.AmountDue, "Initial invoice amount should be $800")

	// ACTION: Update the first timesheet from 5 hours to 10 hours
	// Total should now be 13 hours = $1300
	form := url.Values{}
	form.Add("work_date", time.Now().Format("2006-01-02"))
	form.Add("hours_worked", "10.0")
	form.Add("hourly_rate", "100.00")
	form.Add("description", "Initial work (updated)")

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/update/%d", timesheetID), strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", fmt.Sprintf("%d", timesheetID))
	rr := httptest.NewRecorder()

	// Use session middleware to handle flash messages
	handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetUpdatePost))
	handler.ServeHTTP(rr, req)

	// VERIFY: Should redirect after successful update
	assert.Equal(t, http.StatusSeeOther, rr.Code, "Should redirect after timesheet update")

	// VERIFY: Invoice amount should be automatically updated to $1300
	invoice, err = app.invoices.Get(invoiceID)
	require.NoError(t, err)
	assert.Equal(t, 1300.0, invoice.AmountDue, "Invoice amount should be updated to $1300 after changing to 10 hours")
}

// TestInvoiceRecalculationOnTimesheetDelete verifies that invoice amounts are automatically
// updated when a timesheet is deleted from a project with an unpaid invoice
func TestInvoiceRecalculationOnTimesheetDelete(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	// Setup: Create a client
	clientID := testDB.InsertTestClient(t, "Test Client")

	// Setup: Create a project
	projectID := testDB.InsertTestProjectWithRate(t, clientID, "Test Project", 100.0)

	// Setup: Create two timesheets
	timesheetID := testDB.InsertTestTimesheetWithTime(t, projectID, time.Now(), 5.0, 100.0, "Initial work")
	testDB.InsertTestTimesheetWithTime(t, projectID, time.Now(), 3.0, 100.0, "More work")
	// Total: 8 hours = $800

	// Setup: Create an unpaid invoice
	invoiceID := testDB.InsertTestInvoiceWithTime(t, projectID, time.Now(), nil, "Net 30", 800.0, true)

	// Verify initial invoice amount
	invoice, err := app.invoices.Get(invoiceID)
	require.NoError(t, err)
	assert.Equal(t, 800.0, invoice.AmountDue, "Initial invoice amount should be $800")

	// ACTION: Delete the first timesheet (5 hours)
	// Total should now be 3 hours = $300
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/delete/%d", timesheetID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", timesheetID))
	rr := httptest.NewRecorder()

	// Use session middleware to handle flash messages
	handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetDelete))
	handler.ServeHTTP(rr, req)

	// VERIFY: Should redirect after successful deletion
	assert.Equal(t, http.StatusSeeOther, rr.Code, "Should redirect after timesheet deletion")

	// VERIFY: Invoice amount should be automatically updated to $300
	invoice, err = app.invoices.Get(invoiceID)
	require.NoError(t, err)
	assert.Equal(t, 300.0, invoice.AmountDue, "Invoice amount should be updated to $300 after deleting 5 hours")
}

// TestPaidInvoicesNotRecalculated verifies that paid invoices remain unchanged
// when timesheets are modified - only unpaid invoices should be updated
func TestPaidInvoicesNotRecalculated(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	// Setup: Create a client
	clientID := testDB.InsertTestClient(t, "Test Client")

	// Setup: Create a project
	projectID := testDB.InsertTestProjectWithRate(t, clientID, "Test Project", 100.0)

	// Setup: Create a timesheet (5 hours at $100/hr = $500)
	testDB.InsertTestTimesheetWithTime(t, projectID, time.Now(), 5.0, 100.0, "Initial work")

	// Setup: Create a PAID invoice
	paidDate := time.Now()
	paidInvoiceID := testDB.InsertTestInvoiceWithTime(t, projectID, time.Now(), &paidDate, "Net 30", 500.0, true)

	// Setup: Also create an UNPAID invoice
	unpaidInvoiceID := testDB.InsertTestInvoiceWithTime(t, projectID, time.Now(), nil, "Net 30", 500.0, true)

	// Verify initial amounts
	paidInvoice, err := app.invoices.Get(paidInvoiceID)
	require.NoError(t, err)
	assert.Equal(t, 500.0, paidInvoice.AmountDue, "Paid invoice initial amount should be $500")

	unpaidInvoice, err := app.invoices.Get(unpaidInvoiceID)
	require.NoError(t, err)
	assert.Equal(t, 500.0, unpaidInvoice.AmountDue, "Unpaid invoice initial amount should be $500")

	// ACTION: Add a new timesheet (3 hours at $100/hr = $300)
	// Total should now be $800
	form := url.Values{}
	form.Add("work_date", time.Now().Format("2006-01-02"))
	form.Add("hours_worked", "3.0")
	form.Add("hourly_rate", "100.00")
	form.Add("description", "Additional work")

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/create/%d", projectID), strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", fmt.Sprintf("%d", projectID))
	rr := httptest.NewRecorder()

	// Use session middleware to handle flash messages
	handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetCreatePost))
	handler.ServeHTTP(rr, req)

	// VERIFY: Should redirect after successful creation
	assert.Equal(t, http.StatusSeeOther, rr.Code, "Should redirect after timesheet creation")

	// VERIFY: Paid invoice amount should remain unchanged at $500
	paidInvoice, err = app.invoices.Get(paidInvoiceID)
	require.NoError(t, err)
	assert.Equal(t, 500.0, paidInvoice.AmountDue, "Paid invoice amount should remain $500 (not recalculated)")

	// VERIFY: Unpaid invoice amount should be updated to $800
	unpaidInvoice, err = app.invoices.Get(unpaidInvoiceID)
	require.NoError(t, err)
	assert.Equal(t, 800.0, unpaidInvoice.AmountDue, "Unpaid invoice amount should be updated to $800")
}

// TestInvoiceRecalculationWithDiscount verifies that discounts are properly applied
// when recalculating invoice amounts
func TestInvoiceRecalculationWithDiscount(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	// Setup: Create a client
	clientID := testDB.InsertTestClient(t, "Test Client")

	// Setup: Create a project with 10% discount
	projectID := testDB.InsertTestProjectWithDiscount(t, clientID, "Test Project", 100.0, 10.0, "Volume discount")

	// Setup: Create a timesheet (10 hours at $100/hr = $1000, with 10% discount = $900)
	testDB.InsertTestTimesheetWithTime(t, projectID, time.Now(), 10.0, 100.0, "Initial work")

	// Setup: Create an unpaid invoice
	invoiceID := testDB.InsertTestInvoiceWithTime(t, projectID, time.Now(), nil, "Net 30", 900.0, true)

	// Verify initial invoice amount
	invoice, err := app.invoices.Get(invoiceID)
	require.NoError(t, err)
	assert.Equal(t, 900.0, invoice.AmountDue, "Initial invoice amount should be $900 (with 10% discount)")

	// ACTION: Add another timesheet (5 hours at $100/hr)
	// Subtotal: 15 hours = $1500
	// With 10% discount: $1350
	form := url.Values{}
	form.Add("work_date", time.Now().Format("2006-01-02"))
	form.Add("hours_worked", "5.0")
	form.Add("hourly_rate", "100.00")
	form.Add("description", "Additional work")

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/timesheet/create/%d", projectID), strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", fmt.Sprintf("%d", projectID))
	rr := httptest.NewRecorder()

	// Use session middleware to handle flash messages
	handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetCreatePost))
	handler.ServeHTTP(rr, req)

	// VERIFY: Invoice amount should be $1350 (15 hours at $100/hr with 10% discount)
	invoice, err = app.invoices.Get(invoiceID)
	require.NoError(t, err)
	assert.Equal(t, 1350.0, invoice.AmountDue, "Invoice amount should be $1350 after discount")
}
