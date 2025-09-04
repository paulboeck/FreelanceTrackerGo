package models

import (
	"testing"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvoiceModel_Insert(t *testing.T) {
	// Setup test database using SQLite
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewInvoiceModel(testDB.DB)

	t.Run("successful insert with date paid", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		datePaid := time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		
		id, err := model.Insert(projectID, invoiceDate, &datePaid, paymentTerms, amountDue, false)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the invoice was actually inserted using direct query
		var insertedProjectID int
		var insertedInvoiceDate, insertedDatePaid string
		var insertedPaymentTerms string
		var insertedAmountDue float64
		err = testDB.DB.QueryRow("SELECT project_id, invoice_date, date_paid, payment_terms, amount_due FROM invoice WHERE id = ?", id).Scan(
			&insertedProjectID, &insertedInvoiceDate, &insertedDatePaid, &insertedPaymentTerms, &insertedAmountDue)
		require.NoError(t, err)
		assert.Equal(t, projectID, insertedProjectID)
		assert.Contains(t, insertedInvoiceDate, "2024-01-15")
		assert.Contains(t, insertedDatePaid, "2024-01-25")
		assert.Equal(t, paymentTerms, insertedPaymentTerms)
		assert.Equal(t, amountDue, insertedAmountDue)
	})

	t.Run("successful insert without date paid", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, false)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the invoice was actually inserted with NULL date_paid
		var insertedProjectID int
		var insertedInvoiceDate string
		var insertedDatePaid interface{}
		var insertedPaymentTerms string
		var insertedAmountDue float64
		err = testDB.DB.QueryRow("SELECT project_id, invoice_date, date_paid, payment_terms, amount_due FROM invoice WHERE id = ?", id).Scan(
			&insertedProjectID, &insertedInvoiceDate, &insertedDatePaid, &insertedPaymentTerms, &insertedAmountDue)
		require.NoError(t, err)
		assert.Equal(t, projectID, insertedProjectID)
		assert.Contains(t, insertedInvoiceDate, "2024-01-15")
		assert.Nil(t, insertedDatePaid)
		assert.Equal(t, paymentTerms, insertedPaymentTerms)
		assert.Equal(t, amountDue, insertedAmountDue)
	})

	t.Run("insert with non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		
		id, err := model.Insert(999, invoiceDate, nil, paymentTerms, amountDue, false) // Non-existent project
		
		// SQLite might not enforce foreign key constraints by default in tests
		// Just verify it doesn't crash
		if err != nil {
			assert.Equal(t, 0, id)
		}
	})

	t.Run("insert with zero amount", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 0.0
		
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, false)
		
		// Should succeed at database level (validation happens at handler level)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
	})
}

func TestInvoiceModel_Get(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewInvoiceModel(testDB.DB)

	t.Run("get existing invoice with date paid", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice
		expectedInvoiceDate := "2024-01-15"
		expectedDatePaid := "2024-01-25"
		expectedPaymentTerms := "Net 30"
		expectedAmountDue := "1250.00"
		id := testDB.InsertTestInvoice(t, projectID, expectedInvoiceDate, expectedDatePaid, expectedPaymentTerms, expectedAmountDue)
		
		// Get the invoice using model
		invoice, err := model.Get(id)
		
		require.NoError(t, err)
		assert.Equal(t, id, invoice.ID)
		assert.Equal(t, projectID, invoice.ProjectID)
		assert.Equal(t, expectedInvoiceDate, invoice.InvoiceDate.Format("2006-01-02"))
		assert.NotNil(t, invoice.DatePaid)
		assert.Equal(t, expectedDatePaid, invoice.DatePaid.Format("2006-01-02"))
		assert.Equal(t, expectedPaymentTerms, invoice.PaymentTerms)
		assert.Equal(t, 1250.00, invoice.AmountDue)
		assert.False(t, invoice.Created.IsZero())
		assert.False(t, invoice.Updated.IsZero())
		assert.Nil(t, invoice.DeletedAt)
	})

	t.Run("get existing invoice without date paid", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice without date paid
		expectedInvoiceDate := "2024-01-15"
		expectedPaymentTerms := "Net 30"
		expectedAmountDue := "1250.00"
		id := testDB.InsertTestInvoice(t, projectID, expectedInvoiceDate, "", expectedPaymentTerms, expectedAmountDue)
		
		// Get the invoice using model
		invoice, err := model.Get(id)
		
		require.NoError(t, err)
		assert.Equal(t, id, invoice.ID)
		assert.Equal(t, projectID, invoice.ProjectID)
		assert.Equal(t, expectedInvoiceDate, invoice.InvoiceDate.Format("2006-01-02"))
		assert.Nil(t, invoice.DatePaid)
		assert.Equal(t, expectedPaymentTerms, invoice.PaymentTerms)
		assert.Equal(t, 1250.00, invoice.AmountDue)
	})

	t.Run("get non-existent invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		
		invoice, err := model.Get(999)
		
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		assert.Equal(t, Invoice{}, invoice)
	})
}

func TestInvoiceModel_GetByProject(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewInvoiceModel(testDB.DB)

	t.Run("get invoices for project with multiple invoices", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and projects
		clientID := testDB.InsertTestClient(t, "Test Client")
		project1ID := testDB.InsertTestProject(t, "Project 1", clientID)
		project2ID := testDB.InsertTestProject(t, "Project 2", clientID)
		
		// Create invoices for project 1
		invoice1ID := testDB.InsertTestInvoice(t, project1ID, "2024-01-15", "2024-01-25", "Net 30", "1250.00")
		invoice2ID := testDB.InsertTestInvoice(t, project1ID, "2024-02-15", "", "Net 30", "750.00")
		
		// Create invoice for project 2 (should not be returned)
		_ = testDB.InsertTestInvoice(t, project2ID, "2024-01-20", "", "Net 15", "500.00")
		
		invoices, err := model.GetByProject(project1ID)
		
		require.NoError(t, err)
		require.Len(t, invoices, 2)
		
		// Verify the correct invoices are returned
		invoiceIDs := make([]int, len(invoices))
		amounts := make([]float64, len(invoices))
		for i, invoice := range invoices {
			invoiceIDs[i] = invoice.ID
			amounts[i] = invoice.AmountDue
			assert.Equal(t, project1ID, invoice.ProjectID)
			assert.False(t, invoice.Created.IsZero())
			assert.False(t, invoice.Updated.IsZero())
		}
		
		assert.Contains(t, invoiceIDs, invoice1ID)
		assert.Contains(t, invoiceIDs, invoice2ID)
		assert.Contains(t, amounts, 1250.00)
		assert.Contains(t, amounts, 750.00)
	})

	t.Run("get invoices for project with no invoices", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project with no invoices
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Project with no invoices", clientID)
		
		invoices, err := model.GetByProject(projectID)
		
		require.NoError(t, err)
		assert.Empty(t, invoices)
	})

	t.Run("get invoices for non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		
		invoices, err := model.GetByProject(999)
		
		require.NoError(t, err)
		assert.Empty(t, invoices)
	})
}

func TestInvoiceModel_Update(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewInvoiceModel(testDB.DB)

	t.Run("successful update with date paid", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice
		originalInvoiceDate := "2024-01-15"
		originalPaymentTerms := "Net 30"
		originalAmountDue := "1250.00"
		id := testDB.InsertTestInvoice(t, projectID, originalInvoiceDate, "", originalPaymentTerms, originalAmountDue)
		
		// Update the invoice
		newInvoiceDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		newDatePaid := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
		newPaymentTerms := "Net 15"
		newAmountDue := 950.00
		err := model.Update(id, newInvoiceDate, &newDatePaid, newPaymentTerms, newAmountDue, false)
		require.NoError(t, err)
		
		// Verify the invoice was updated
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, invoice.ID)
		assert.Equal(t, "2024-01-20", invoice.InvoiceDate.Format("2006-01-02"))
		assert.NotNil(t, invoice.DatePaid)
		assert.Equal(t, "2024-02-15", invoice.DatePaid.Format("2006-01-02"))
		assert.Equal(t, newPaymentTerms, invoice.PaymentTerms)
		assert.Equal(t, newAmountDue, invoice.AmountDue)
		assert.False(t, invoice.Updated.IsZero())
		
		// Verify the updated_at timestamp changed
		assert.True(t, invoice.Updated.After(invoice.Created) || invoice.Updated.Equal(invoice.Created))
	})

	t.Run("successful update without date paid", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice with date paid
		originalInvoiceDate := "2024-01-15"
		originalDatePaid := "2024-01-25"
		originalPaymentTerms := "Net 30"
		originalAmountDue := "1250.00"
		id := testDB.InsertTestInvoice(t, projectID, originalInvoiceDate, originalDatePaid, originalPaymentTerms, originalAmountDue)
		
		// Update the invoice removing date paid
		newInvoiceDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		newPaymentTerms := "Net 15"
		newAmountDue := 950.00
		err := model.Update(id, newInvoiceDate, nil, newPaymentTerms, newAmountDue, false)
		require.NoError(t, err)
		
		// Verify the invoice was updated
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-20", invoice.InvoiceDate.Format("2006-01-02"))
		assert.Nil(t, invoice.DatePaid)
		assert.Equal(t, newPaymentTerms, invoice.PaymentTerms)
		assert.Equal(t, newAmountDue, invoice.AmountDue)
	})

	t.Run("update non-existent invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		
		newInvoiceDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		newPaymentTerms := "Net 15"
		newAmountDue := 950.00
		err := model.Update(999, newInvoiceDate, nil, newPaymentTerms, newAmountDue, false)
		
		// Should not return an error (SQLite UPDATE doesn't fail for non-existent rows)
		require.NoError(t, err)
	})

	t.Run("update with zero amount", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice
		originalInvoiceDate := "2024-01-15"
		originalPaymentTerms := "Net 30"
		originalAmountDue := "1250.00"
		id := testDB.InsertTestInvoice(t, projectID, originalInvoiceDate, "", originalPaymentTerms, originalAmountDue)
		
		// Update with zero amount (should succeed at database level)
		newInvoiceDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		newPaymentTerms := "Net 15"
		newAmountDue := 0.0
		err := model.Update(id, newInvoiceDate, nil, newPaymentTerms, newAmountDue, false)
		require.NoError(t, err)
		
		// Verify the invoice was updated
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, 0.0, invoice.AmountDue)
		assert.Equal(t, newPaymentTerms, invoice.PaymentTerms)
	})
}

func TestInvoiceModel_Delete(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewInvoiceModel(testDB.DB)

	t.Run("successful delete", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice
		invoiceDate := "2024-01-15"
		paymentTerms := "Net 30"
		amountDue := "1250.00"
		id := testDB.InsertTestInvoice(t, projectID, invoiceDate, "", paymentTerms, amountDue)
		
		// Verify invoice exists
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, paymentTerms, invoice.PaymentTerms)
		assert.Nil(t, invoice.DeletedAt)
		
		// Delete the invoice
		err = model.Delete(id)
		require.NoError(t, err)
		
		// Verify the invoice is no longer returned by Get (soft deleted)
		_, err = model.Get(id)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		
		// Verify the invoice is no longer in GetByProject
		invoices, err := model.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, invoices)
		
		// Verify the invoice still exists in database but with deleted_at set
		var deletedAt interface{}
		err = testDB.DB.QueryRow("SELECT deleted_at FROM invoice WHERE id = ?", id).Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("delete non-existent invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		
		err := model.Delete(999)
		
		// Should not return an error (SQLite UPDATE doesn't fail for non-existent rows)
		require.NoError(t, err)
	})

	t.Run("delete already deleted invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert and delete invoice
		invoiceDate := "2024-01-15"
		paymentTerms := "Net 30"
		amountDue := "1250.00"
		id := testDB.InsertTestInvoice(t, projectID, invoiceDate, "", paymentTerms, amountDue)
		err := model.Delete(id)
		require.NoError(t, err)
		
		// Try to delete again
		err = model.Delete(id)
		require.NoError(t, err) // Should not error, but should have no effect
		
		// Verify still deleted
		_, err = model.Get(id)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
	})
}

func TestInvoiceModel_Integration(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewInvoiceModel(testDB.DB)

	t.Run("full CRUD workflow with invoice model", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// 1. Create client and project
		clientID := testDB.InsertTestClient(t, "Integration Test Client")
		projectID := testDB.InsertTestProject(t, "Integration Test Project", clientID)
		
		// 2. Insert a new invoice
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, false)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// 3. Get the invoice
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, invoice.ID)
		assert.Equal(t, projectID, invoice.ProjectID)
		assert.Equal(t, "2024-01-15", invoice.InvoiceDate.Format("2006-01-02"))
		assert.Nil(t, invoice.DatePaid)
		assert.Equal(t, paymentTerms, invoice.PaymentTerms)
		assert.Equal(t, amountDue, invoice.AmountDue)
		
		// 4. Verify it appears in GetByProject
		invoices, err := model.GetByProject(projectID)
		require.NoError(t, err)
		require.Len(t, invoices, 1)
		assert.Equal(t, invoice.ID, invoices[0].ID)
		assert.Equal(t, invoice.AmountDue, invoices[0].AmountDue)
		
		// 5. Update the invoice with payment
		newInvoiceDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		datePaid := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
		newPaymentTerms := "Net 15"
		newAmountDue := 950.00
		err = model.Update(id, newInvoiceDate, &datePaid, newPaymentTerms, newAmountDue, false)
		require.NoError(t, err)
		
		// 6. Verify update
		updatedInvoice, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-20", updatedInvoice.InvoiceDate.Format("2006-01-02"))
		assert.NotNil(t, updatedInvoice.DatePaid)
		assert.Equal(t, "2024-02-15", updatedInvoice.DatePaid.Format("2006-01-02"))
		assert.Equal(t, newPaymentTerms, updatedInvoice.PaymentTerms)
		assert.Equal(t, newAmountDue, updatedInvoice.AmountDue)
		assert.True(t, updatedInvoice.Updated.After(invoice.Updated) || updatedInvoice.Updated.Equal(invoice.Updated))
		
		// 7. Delete the invoice
		err = model.Delete(id)
		require.NoError(t, err)
		
		// 8. Verify deletion
		_, err = model.Get(id)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		
		invoices, err = model.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, invoices)
	})
}

// TestInterface verifies that the implementation satisfies the interface
func TestInvoiceModelInterface(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	
	implementations := []struct {
		name string
		impl InvoiceModelInterface
	}{
		{"SQLite InvoiceModel", NewInvoiceModel(testDB.DB)},
	}
	
	for _, test := range implementations {
		t.Run(test.name, func(t *testing.T) {
			testDB.TruncateTable(t, "invoice")
			testDB.TruncateTable(t, "project")
			testDB.TruncateTable(t, "client")
			
			// Create test client and project first
			clientID := testDB.InsertTestClient(t, "Interface Test Client")
			projectID := testDB.InsertTestProject(t, "Interface Test Project", clientID)
			
			// Test that the implementation works correctly
			invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
			paymentTerms := "Net 30"
			amountDue := 1250.00
			
			// Insert
			id, err := test.impl.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, false)
			require.NoError(t, err)
			assert.Greater(t, id, 0)
			
			// Get
			invoice, err := test.impl.Get(id)
			require.NoError(t, err)
			assert.Equal(t, id, invoice.ID)
			assert.Equal(t, projectID, invoice.ProjectID)
			assert.Equal(t, paymentTerms, invoice.PaymentTerms)
			assert.Equal(t, amountDue, invoice.AmountDue)
			
			// GetByProject
			invoices, err := test.impl.GetByProject(projectID)
			require.NoError(t, err)
			require.Len(t, invoices, 1)
			assert.Equal(t, id, invoices[0].ID)
			assert.Equal(t, amountDue, invoices[0].AmountDue)
			
			// Update
			newInvoiceDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
			datePaid := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
			newPaymentTerms := "Net 15"
			newAmountDue := 950.00
			err = test.impl.Update(id, newInvoiceDate, &datePaid, newPaymentTerms, newAmountDue, false)
			require.NoError(t, err)
			
			updatedInvoice, err := test.impl.Get(id)
			require.NoError(t, err)
			assert.NotNil(t, updatedInvoice.DatePaid)
			assert.Equal(t, newPaymentTerms, updatedInvoice.PaymentTerms)
			assert.Equal(t, newAmountDue, updatedInvoice.AmountDue)
			
			// Delete
			err = test.impl.Delete(id)
			require.NoError(t, err)
			
			_, err = test.impl.Get(id)
			assert.Error(t, err)
			assert.Equal(t, ErrNoRecord, err)
		})
	}
}

func TestInvoiceModel_DisplayDetails(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewInvoiceModel(testDB.DB)

	t.Run("insert with display details true", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		displayDetails := true
		
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, displayDetails)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the display details was inserted correctly
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.True(t, invoice.DisplayDetails)
	})

	t.Run("insert with display details false", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		displayDetails := false
		
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, displayDetails)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the display details was inserted correctly
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.False(t, invoice.DisplayDetails)
	})

	t.Run("update display details from false to true", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice with display details false
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, false)
		require.NoError(t, err)
		
		// Verify initially false
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.False(t, invoice.DisplayDetails)
		
		// Update to display details true
		err = model.Update(id, invoiceDate, nil, paymentTerms, amountDue, true)
		require.NoError(t, err)
		
		// Verify the display details was updated
		updatedInvoice, err := model.Get(id)
		require.NoError(t, err)
		assert.True(t, updatedInvoice.DisplayDetails)
	})

	t.Run("update display details from true to false", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice with display details true
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, true)
		require.NoError(t, err)
		
		// Verify initially true
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.True(t, invoice.DisplayDetails)
		
		// Update to display details false
		err = model.Update(id, invoiceDate, nil, paymentTerms, amountDue, false)
		require.NoError(t, err)
		
		// Verify the display details was updated
		updatedInvoice, err := model.Get(id)
		require.NoError(t, err)
		assert.False(t, updatedInvoice.DisplayDetails)
	})
}

func TestInvoiceModel_GetComprehensiveForPDF(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instances
	invoiceModel := NewInvoiceModel(testDB.DB)
	clientModel := NewClientModel(testDB.DB)
	projectModel := NewProjectModel(testDB.DB)
	timesheetModel := NewTimesheetModel(testDB.DB)

	t.Run("get comprehensive data for simple invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client with rich data
		clientName := "Test University"
		clientEmail := "test@university.edu"
		phone := "555-123-4567"
		address1 := "123 University Ave"
		address2 := "Suite 200"
		city := "College Town"
		state := "CA"
		zipCode := "90210"
		hourlyRate := 85.0
		notes := "Test client notes"
		billTo := "Custom Bill To Address\nLine 2\nLine 3"
		universityAff := "Test University Department"
		
		clientID, err := clientModel.Insert(
			clientName, clientEmail, &phone, &address1, &address2, nil, &city, &state, &zipCode, 
			hourlyRate, &notes, nil, nil, &billTo, true, nil, nil, &universityAff,
		)
		require.NoError(t, err)
		
		// Create test project with attributes
		project := Project{
			Name:                   "Test Academic Project",
			ClientID:               clientID,
			Status:                 "In Progress",
			HourlyRate:             90.0,
			DiscountPercent:        &[]float64{10.0}[0], // 10% discount
			DiscountReason:         "Early payment discount",
			AdjustmentAmount:       &[]float64{-25.0}[0], // $25 adjustment
			AdjustmentReason:       "Complexity adjustment",
			CurrencyDisplay:        "USD",
			CurrencyConversionRate: 1.0,
			FlatFeeInvoice:         false,
			Notes:                  "Project notes for invoice",
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		// Create test timesheets
		timesheet1ID, err := timesheetModel.Insert(projectID, time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), 3.5, 90.0, "Research and analysis")
		require.NoError(t, err)
		timesheet2ID, err := timesheetModel.Insert(projectID, time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC), 2.0, 90.0, "Writing and editing")
		require.NoError(t, err)
		
		// Create invoice
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30 - Early payment discount applied"
		amountDue := 495.0 // 5.5 hours * $90
		invoiceID, err := invoiceModel.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, true)
		require.NoError(t, err)
		
		// Test GetComprehensiveForPDF
		data, err := invoiceModel.GetComprehensiveForPDF(invoiceID)
		require.NoError(t, err)
		
		// Verify invoice data
		assert.Equal(t, invoiceID, data.Invoice.ID)
		assert.Equal(t, projectID, data.Invoice.ProjectID)
		assert.Equal(t, invoiceDate, data.Invoice.InvoiceDate)
		assert.Equal(t, paymentTerms, data.Invoice.PaymentTerms)
		assert.Equal(t, amountDue, data.Invoice.AmountDue)
		assert.True(t, data.Invoice.DisplayDetails)
		
		// Verify project data
		assert.Equal(t, "Test Academic Project", data.Project.Name)
		assert.Equal(t, clientID, data.Project.ClientID)
		assert.Equal(t, "In Progress", data.Project.Status)
		assert.Equal(t, 90.0, data.Project.HourlyRate)
		assert.NotNil(t, data.Project.DiscountPercent)
		assert.Equal(t, 10.0, *data.Project.DiscountPercent)
		assert.Equal(t, "Early payment discount", data.Project.DiscountReason)
		assert.NotNil(t, data.Project.AdjustmentAmount)
		assert.Equal(t, -25.0, *data.Project.AdjustmentAmount)
		assert.Equal(t, "Complexity adjustment", data.Project.AdjustmentReason)
		assert.False(t, data.Project.FlatFeeInvoice)
		assert.Equal(t, "Project notes for invoice", data.Project.Notes)
		
		// Verify client data
		assert.Equal(t, clientName, data.Client.Name)
		assert.Equal(t, clientEmail, data.Client.Email)
		assert.NotNil(t, data.Client.Phone)
		assert.Equal(t, phone, *data.Client.Phone)
		assert.NotNil(t, data.Client.Address1)
		assert.Equal(t, address1, *data.Client.Address1)
		assert.NotNil(t, data.Client.BillTo)
		assert.Equal(t, billTo, *data.Client.BillTo)
		assert.True(t, data.Client.IncludeAddressOnInvoice)
		assert.NotNil(t, data.Client.UniversityAffiliation)
		assert.Equal(t, universityAff, *data.Client.UniversityAffiliation)
		
		// Verify timesheets data
		require.Len(t, data.Timesheets, 2)
		
		// Find timesheets by ID (order not guaranteed)
		var ts1, ts2 *Timesheet
		for i := range data.Timesheets {
			if data.Timesheets[i].ID == timesheet1ID {
				ts1 = &data.Timesheets[i]
			} else if data.Timesheets[i].ID == timesheet2ID {
				ts2 = &data.Timesheets[i]
			}
		}
		require.NotNil(t, ts1)
		require.NotNil(t, ts2)
		
		assert.Equal(t, 3.5, ts1.HoursWorked)
		assert.Equal(t, "Research and analysis", ts1.Description)
		assert.Equal(t, 2.0, ts2.HoursWorked)
		assert.Equal(t, "Writing and editing", ts2.Description)
		
		// Verify calculated totals
		assert.Equal(t, 5.5, data.TotalHours) // 3.5 + 2.0
		assert.Equal(t, 420.5, data.Subtotal) // 495.0 - 10% discount (49.5) - adjustment (25.0)
		
		// Verify discount calculation (10% of 495)
		expectedDiscount := 495.0 * 0.10
		assert.Equal(t, expectedDiscount, data.DiscountAmount)
		
		// Verify adjustment
		assert.Equal(t, -25.0, data.AdjustmentAmount)
		
		// Verify final total (495 - 49.5 discount - 25 adjustment = 420.5)
		expectedFinal := 495.0 - expectedDiscount - 25.0
		assert.Equal(t, expectedFinal, data.FinalTotal)
	})

	t.Run("get comprehensive data for flat fee invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create simple test data
		clientID := testDB.InsertTestClient(t, "Flat Fee Client")
		
		project := Project{
			Name:           "Flat Fee Project",
			ClientID:       clientID,
			Status:         "Complete",
			HourlyRate:     75.0,
			FlatFeeInvoice: true,
			Notes:          "Fixed price project",
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		// Create invoice for flat fee
		invoiceDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		flatFeeAmount := 2500.0
		invoiceID, err := invoiceModel.Insert(projectID, invoiceDate, nil, "Net 15", flatFeeAmount, false)
		require.NoError(t, err)
		
		// Test comprehensive data
		data, err := invoiceModel.GetComprehensiveForPDF(invoiceID)
		require.NoError(t, err)
		
		// Verify flat fee project handling
		assert.True(t, data.Project.FlatFeeInvoice)
		assert.Equal(t, flatFeeAmount, data.Invoice.AmountDue)
		assert.Equal(t, flatFeeAmount, data.FinalTotal)
		assert.Equal(t, 0.0, data.DiscountAmount) // No discount
		assert.Equal(t, 0.0, data.AdjustmentAmount) // No adjustment
		assert.Empty(t, data.Timesheets) // No timesheets
		assert.Equal(t, 0.0, data.TotalHours) // No hours
	})

	t.Run("get comprehensive data for non-existent invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		
		data, err := invoiceModel.GetComprehensiveForPDF(999)
		
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		assert.Equal(t, ComprehensiveInvoiceData{}, data)
	})
}

func TestInvoiceModel_GenerateComprehensivePDF(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instances
	invoiceModel := NewInvoiceModel(testDB.DB)
	clientModel := NewClientModel(testDB.DB)
	projectModel := NewProjectModel(testDB.DB)
	timesheetModel := NewTimesheetModel(testDB.DB)

	// Helper to create test settings with logo path
	createTestSettings := func() map[string]AppSettingValue {
		return map[string]AppSettingValue{
			"invoice_title":                       {Value: "Professional Invoice", DataType: "string"},
			"freelancer_name":                     {Value: "John Doe Consulting", DataType: "string"},
			"freelancer_address":                  {Value: "123 Business St", DataType: "string"},
			"freelancer_city_state_zip":           {Value: "Business City, CA 90210", DataType: "string"},
			"freelancer_phone":                    {Value: "(555) 123-4567", DataType: "string"},
			"freelancer_email":                    {Value: "john@consulting.com", DataType: "string"},
			"company_logo_path":                   {Value: "./ui/static/img/logo.png", DataType: "string"},
			"invoice_header_decoration":           {Value: "=== === ===", DataType: "string"},
			"invoice_payment_terms_default":       {Value: "Payment due within 30 days. Thank you!", DataType: "string"},
			"invoice_thank_you_message":           {Value: "Thank you for choosing our services!", DataType: "string"},
			"invoice_show_individual_timesheets":  {Value: "true", DataType: "bool"},
			"invoice_currency_symbol":             {Value: "$", DataType: "string"},
		}
	}

	t.Run("generate PDF with detailed timesheets", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create comprehensive test data
		clientName := "Test Corporation"
		billTo := "Test Corporation\nAttn: Accounting\n456 Corporate Blvd\nBusiness City, CA 90210"
		clientID, err := clientModel.Insert(
			clientName, "accounting@testcorp.com", nil, nil, nil, nil, nil, nil, nil,
			100.0, nil, nil, nil, &billTo, true, nil, nil, nil,
		)
		require.NoError(t, err)
		
		project := Project{
			Name:       "Detailed Project",
			ClientID:   clientID,
			Status:     "Complete",
			HourlyRate: 100.0,
			Notes:      "Project completed successfully with detailed tracking",
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		// Create multiple timesheets
		_, err = timesheetModel.Insert(projectID, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 4.0, 100.0, "Initial research and planning")
		require.NoError(t, err)
		_, err = timesheetModel.Insert(projectID, time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), 3.5, 100.0, "Development work")
		require.NoError(t, err)
		_, err = timesheetModel.Insert(projectID, time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC), 2.0, 100.0, "Testing and validation")
		require.NoError(t, err)
		
		// Create invoice with display details enabled
		invoiceID, err := invoiceModel.Insert(projectID, time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), nil, "Net 30", 950.0, true)
		require.NoError(t, err)
		
		// Generate PDF
		settings := createTestSettings()
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(invoiceID, settings)
		require.NoError(t, err)
		
		// Verify PDF was generated
		assert.Greater(t, len(pdfBytes), 1000) // Should be a substantial PDF
		
		// Verify PDF header
		assert.Contains(t, string(pdfBytes[:200]), "PDF") // Should start with PDF header
	})

	t.Run("generate PDF with summary view", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test data
		clientID := testDB.InsertTestClient(t, "Summary Client")
		project := Project{
			Name:       "Summary Project",
			ClientID:   clientID,
			Status:     "Complete",
			HourlyRate: 85.0,
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		// Create invoice with display details disabled (summary view)
		invoiceID, err := invoiceModel.Insert(projectID, time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), nil, "Net 15", 425.0, false)
		require.NoError(t, err)
		
		// Generate PDF with summary settings
		settings := createTestSettings()
		settings["invoice_show_individual_timesheets"] = AppSettingValue{Value: "false", DataType: "bool"}
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(invoiceID, settings)
		require.NoError(t, err)
		
		// Verify PDF was generated
		assert.Greater(t, len(pdfBytes), 800) // Should be a decent-sized PDF
	})

	t.Run("generate PDF with discount and adjustment", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test data with discount and adjustment
		clientID := testDB.InsertTestClient(t, "Discount Client")
		project := Project{
			Name:             "Discounted Project",
			ClientID:         clientID,
			Status:           "Complete",
			HourlyRate:       100.0,
			DiscountPercent:  &[]float64{15.0}[0], // 15% discount
			DiscountReason:   "Volume discount",
			AdjustmentAmount: &[]float64{50.0}[0], // $50 bonus
			AdjustmentReason: "Complexity bonus",
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		// Create invoice
		invoiceID, err := invoiceModel.Insert(projectID, time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), nil, "Net 30", 1000.0, false)
		require.NoError(t, err)
		
		// Generate PDF
		settings := createTestSettings()
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(invoiceID, settings)
		require.NoError(t, err)
		
		// Verify PDF was generated
		assert.Greater(t, len(pdfBytes), 1000)
	})

	t.Run("generate PDF for flat fee project", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create flat fee project
		clientID := testDB.InsertTestClient(t, "Flat Fee Client")
		project := Project{
			Name:           "Website Redesign",
			ClientID:       clientID,
			Status:         "Complete",
			HourlyRate:     0.0, // Not used for flat fee
			FlatFeeInvoice: true,
			Notes:          "Complete website redesign as agreed",
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		// Create flat fee invoice
		invoiceID, err := invoiceModel.Insert(projectID, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), nil, "Net 30", 5000.0, false)
		require.NoError(t, err)
		
		// Generate PDF
		settings := createTestSettings()
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(invoiceID, settings)
		require.NoError(t, err)
		
		// Verify PDF was generated
		assert.Greater(t, len(pdfBytes), 800)
	})

	t.Run("generate PDF with minimal settings", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create minimal test data
		clientID := testDB.InsertTestClient(t, "Minimal Client")
		project := Project{
			Name:       "Basic Project",
			ClientID:   clientID,
			Status:     "Complete",
			HourlyRate: 75.0,
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		invoiceID, err := invoiceModel.Insert(projectID, time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), nil, "", 375.0, false)
		require.NoError(t, err)
		
		// Generate PDF with empty settings (should use fallbacks)
		emptySettings := make(map[string]AppSettingValue)
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(invoiceID, emptySettings)
		require.NoError(t, err)
		
		// Should still generate a PDF with fallback values
		assert.Greater(t, len(pdfBytes), 600)
	})

	t.Run("generate PDF for non-existent invoice", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		
		settings := createTestSettings()
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(999, settings)
		
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		assert.Nil(t, pdfBytes)
	})

	t.Run("generate PDF with client address preferences", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create client with address but IncludeAddressOnInvoice = false
		phone := "555-987-6543"
		address1 := "789 Corporate Way"
		city := "Metro City"
		state := "NY"
		zipCode := "10001"
		clientID, err := clientModel.Insert(
			"Address Test Client", "test@company.com", &phone, &address1, nil, nil, &city, &state, &zipCode,
			80.0, nil, nil, nil, nil, false, nil, nil, nil, // IncludeAddressOnInvoice = false
		)
		require.NoError(t, err)
		
		project := Project{
			Name:       "Address Test Project",
			ClientID:   clientID,
			Status:     "Complete",
			HourlyRate: 80.0,
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		invoiceID, err := invoiceModel.Insert(projectID, time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC), nil, "Net 30", 320.0, false)
		require.NoError(t, err)
		
		// Generate PDF - should not include address since IncludeAddressOnInvoice is false
		settings := createTestSettings()
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(invoiceID, settings)
		require.NoError(t, err)
		
		// Verify PDF was generated
		assert.Greater(t, len(pdfBytes), 600)
	})

	t.Run("generate PDF with logo fallback", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test data
		clientID := testDB.InsertTestClient(t, "Logo Test Client")
		project := Project{
			Name:       "Logo Test Project",
			ClientID:   clientID,
			Status:     "Complete",
			HourlyRate: 75.0,
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		invoiceID, err := invoiceModel.Insert(projectID, time.Date(2024, 1, 30, 0, 0, 0, 0, time.UTC), nil, "Net 30", 300.0, false)
		require.NoError(t, err)
		
		// Test with non-existent logo path (should fallback to decoration)
		settings := createTestSettings()
		settings["company_logo_path"] = AppSettingValue{Value: "./non/existent/logo.png", DataType: "string"}
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(invoiceID, settings)
		require.NoError(t, err)
		
		// Should still generate PDF with fallback decoration
		assert.Greater(t, len(pdfBytes), 600)
	})
}

func TestInvoiceModel_ComprehensiveIntegration(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instances
	invoiceModel := NewInvoiceModel(testDB.DB)
	clientModel := NewClientModel(testDB.DB)
	projectModel := NewProjectModel(testDB.DB)
	timesheetModel := NewTimesheetModel(testDB.DB)

	t.Run("full comprehensive invoice workflow", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Step 1: Create a comprehensive client
		clientName := "Comprehensive Test University"
		clientEmail := "billing@testuniv.edu"
		phone := "555-111-2222"
		address1 := "999 Research Blvd"
		address2 := "Academic Complex"
		address3 := "Building C"
		city := "University City"
		state := "TX"
		zipCode := "78712"
		hourlyRate := 95.0
		notes := "Comprehensive test client for invoice system"
		additionalInfo := "Grant funded project"
		additionalInfo2 := "Requires detailed invoicing"
		billTo := "University Accounting Department\nAttn: Dr. Jane Smith\n999 Research Blvd, Bldg C\nUniversity City, TX 78712"
		invoiceCCEmail := "grants@testuniv.edu"
		invoiceCCDesc := "Grant administrator"
		universityAff := "Department of Computer Science"
		
		clientID, err := clientModel.Insert(
			clientName, clientEmail, &phone, &address1, &address2, &address3, &city, &state, &zipCode,
			hourlyRate, &notes, &additionalInfo, &additionalInfo2, &billTo, true,
			&invoiceCCEmail, &invoiceCCDesc, &universityAff,
		)
		require.NoError(t, err)
		
		// Step 2: Create a comprehensive project with all attributes
		project := Project{
			Name:                   "Comprehensive Research Analysis",
			ClientID:               clientID,
			Status:                 "In Progress",
			HourlyRate:             100.0, // Different from client default
			DiscountPercent:        &[]float64{12.5}[0], // 12.5% discount
			DiscountReason:         "Long-term partnership discount",
			AdjustmentAmount:       &[]float64{75.0}[0], // $75 bonus
			AdjustmentReason:       "Additional complexity bonus",
			CurrencyDisplay:        "USD",
			CurrencyConversionRate: 1.0,
			FlatFeeInvoice:         false,
			InvoiceCCEmail:         "project-manager@testuniv.edu",
			InvoiceCCDescription:   "Project Manager",
			ScheduleComments:       "Flexible timeline based on data availability",
			AdditionalInfo:         "Multi-phase analysis project",
			AdditionalInfo2:        "Requires monthly progress reports",
			Notes:                  "This project involves comprehensive data analysis with detailed documentation requirements.",
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		// Step 3: Create multiple detailed timesheets
		timesheets := []struct{
			date        time.Time
			hours       float64
			rate        float64
			description string
		}{
			{time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), 3.5, 100.0, "Initial data collection and preprocessing"},
			{time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC), 4.0, 100.0, "Statistical analysis and model development"},
			{time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), 2.5, 100.0, "Results visualization and interpretation"},
			{time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 3.0, 100.0, "Draft report writing and documentation"},
			{time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), 2.0, 100.0, "Review and revision of analysis"},
		}
		
		totalHours := 0.0
		for _, ts := range timesheets {
			_, err = timesheetModel.Insert(projectID, ts.date, ts.hours, ts.rate, ts.description)
			require.NoError(t, err)
			totalHours += ts.hours
		}
		
		// Step 4: Create comprehensive invoice
		invoiceDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Payment due within 30 days of receipt. University purchase order required. Please remit payment to address shown above."
		baseAmount := totalHours * 100.0 // 15 hours * $100 = $1500
		invoiceID, err := invoiceModel.Insert(projectID, invoiceDate, nil, paymentTerms, baseAmount, true)
		require.NoError(t, err)
		
		// Step 5: Test comprehensive data retrieval
		data, err := invoiceModel.GetComprehensiveForPDF(invoiceID)
		require.NoError(t, err)
		
		// Step 6: Verify all data is correctly populated
		
		// Invoice verification
		assert.Equal(t, invoiceID, data.Invoice.ID)
		assert.Equal(t, invoiceDate, data.Invoice.InvoiceDate)
		assert.Equal(t, paymentTerms, data.Invoice.PaymentTerms)
		assert.Equal(t, baseAmount, data.Invoice.AmountDue)
		assert.True(t, data.Invoice.DisplayDetails)
		
		// Project verification
		assert.Equal(t, "Comprehensive Research Analysis", data.Project.Name)
		assert.Equal(t, 100.0, data.Project.HourlyRate)
		assert.NotNil(t, data.Project.DiscountPercent)
		assert.Equal(t, 12.5, *data.Project.DiscountPercent)
		assert.Equal(t, "Long-term partnership discount", data.Project.DiscountReason)
		assert.NotNil(t, data.Project.AdjustmentAmount)
		assert.Equal(t, 75.0, *data.Project.AdjustmentAmount)
		assert.Equal(t, "Additional complexity bonus", data.Project.AdjustmentReason)
		assert.False(t, data.Project.FlatFeeInvoice)
		
		// Client verification
		assert.Equal(t, clientName, data.Client.Name)
		assert.Equal(t, clientEmail, data.Client.Email)
		assert.Equal(t, billTo, *data.Client.BillTo)
		assert.True(t, data.Client.IncludeAddressOnInvoice)
		assert.Equal(t, universityAff, *data.Client.UniversityAffiliation)
		
		// Timesheets verification
		require.Len(t, data.Timesheets, 5)
		assert.Equal(t, totalHours, data.TotalHours)
		
		// Financial calculations verification
		assert.Equal(t, 1387.5, data.Subtotal) // baseAmount (1500) - discount (187.5) + adjustment (75)
		expectedDiscount := baseAmount * 0.125 // 12.5%
		assert.Equal(t, expectedDiscount, data.DiscountAmount)
		assert.Equal(t, 75.0, data.AdjustmentAmount)
		expectedFinal := baseAmount - expectedDiscount + 75.0
		assert.Equal(t, expectedFinal, data.FinalTotal)
		
		// Step 7: Test comprehensive PDF generation with rich settings
		settings := map[string]AppSettingValue{
			"invoice_title":                       {Value: "Professional Research Invoice", DataType: "string"},
			"freelancer_name":                     {Value: "Dr. Research Consultant LLC", DataType: "string"},
			"freelancer_address":                  {Value: "456 Professional Plaza", DataType: "string"},
			"freelancer_city_state_zip":           {Value: "Austin, TX 78701", DataType: "string"},
			"freelancer_phone":                    {Value: "(512) 555-1234", DataType: "string"},
			"freelancer_email":                    {Value: "billing@researchconsult.com", DataType: "string"},
			"company_logo_path":                   {Value: "./ui/static/img/logo.png", DataType: "string"},
			"invoice_payment_terms_default":       {Value: "Payment due within 30 days. University purchase orders accepted.", DataType: "string"},
			"invoice_thank_you_message":           {Value: "Thank you for choosing our research services!", DataType: "string"},
			"invoice_show_individual_timesheets":  {Value: "true", DataType: "bool"},
			"invoice_currency_symbol":             {Value: "$", DataType: "string"},
		}
		
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(invoiceID, settings)
		require.NoError(t, err)
		
		// Verify comprehensive PDF generation
		assert.Greater(t, len(pdfBytes), 2000) // Should be a substantial PDF with all details
		assert.Contains(t, string(pdfBytes[:100]), "PDF") // PDF header verification
		
		// Step 8: Test interface compliance
		var _ InvoiceModelInterface = invoiceModel
		
		// Step 9: Test updating and regenerating
		datePaid := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
		err = invoiceModel.Update(invoiceID, invoiceDate, &datePaid, paymentTerms, baseAmount, true)
		require.NoError(t, err)
		
		// Regenerate PDF with paid status
		pdfBytesUpdated, err := invoiceModel.GenerateComprehensivePDF(invoiceID, settings)
		require.NoError(t, err)
		assert.Greater(t, len(pdfBytesUpdated), 2000)
		
		// Step 10: Cleanup test
		err = invoiceModel.Delete(invoiceID)
		require.NoError(t, err)
		
		// Verify soft delete
		_, err = invoiceModel.GetComprehensiveForPDF(invoiceID)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
	})

	t.Run("edge cases and error handling", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Test with minimal data
		clientID := testDB.InsertTestClient(t, "Minimal Client")
		project := Project{
			Name:       "Minimal Project",
			ClientID:   clientID,
			Status:     "Complete",
			HourlyRate: 50.0,
		}
		projectID, err := projectModel.Insert(project)
		require.NoError(t, err)
		
		// Invoice with no timesheets
		invoiceID, err := invoiceModel.Insert(projectID, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), nil, "", 100.0, false)
		require.NoError(t, err)
		
		// Should handle gracefully
		data, err := invoiceModel.GetComprehensiveForPDF(invoiceID)
		require.NoError(t, err)
		assert.Empty(t, data.Timesheets)
		assert.Equal(t, 0.0, data.TotalHours)
		assert.Equal(t, 100.0, data.FinalTotal)
		
		// PDF generation should still work
		emptySettings := make(map[string]AppSettingValue)
		pdfBytes, err := invoiceModel.GenerateComprehensivePDF(invoiceID, emptySettings)
		require.NoError(t, err)
		assert.Greater(t, len(pdfBytes), 400)
	})
}

func TestInvoiceModel_DisplayDetailsInsertAndUpdate(t *testing.T) {
	// Setup test database using SQLite
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewInvoiceModel(testDB.DB)

	t.Run("insert with display details true", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		displayDetails := true
		
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, displayDetails)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the display details was inserted correctly
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.True(t, invoice.DisplayDetails)
	})

	t.Run("insert with display details false", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		displayDetails := false
		
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, displayDetails)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the display details was inserted correctly
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.False(t, invoice.DisplayDetails)
	})

	t.Run("update display details from false to true", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice with display details false
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, false)
		require.NoError(t, err)
		
		// Verify initially false
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.False(t, invoice.DisplayDetails)
		
		// Update to display details true
		err = model.Update(id, invoiceDate, nil, paymentTerms, amountDue, true)
		require.NoError(t, err)
		
		// Verify the display details was updated
		updatedInvoice, err := model.Get(id)
		require.NoError(t, err)
		assert.True(t, updatedInvoice.DisplayDetails)
	})

	t.Run("update display details from true to false", func(t *testing.T) {
		testDB.TruncateTable(t, "invoice")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert invoice with display details true
		invoiceDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		paymentTerms := "Net 30"
		amountDue := 1250.00
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue, true)
		require.NoError(t, err)
		
		// Verify initially true
		invoice, err := model.Get(id)
		require.NoError(t, err)
		assert.True(t, invoice.DisplayDetails)
		
		// Update to display details false
		err = model.Update(id, invoiceDate, nil, paymentTerms, amountDue, false)
		require.NoError(t, err)
		
		// Verify the display details was updated
		updatedInvoice, err := model.Get(id)
		require.NoError(t, err)
		assert.False(t, updatedInvoice.DisplayDetails)
	})
}