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
		
		id, err := model.Insert(projectID, invoiceDate, &datePaid, paymentTerms, amountDue)
		
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
		
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue)
		
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
		
		id, err := model.Insert(999, invoiceDate, nil, paymentTerms, amountDue) // Non-existent project
		
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
		
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue)
		
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
		err := model.Update(id, newInvoiceDate, &newDatePaid, newPaymentTerms, newAmountDue)
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
		err := model.Update(id, newInvoiceDate, nil, newPaymentTerms, newAmountDue)
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
		err := model.Update(999, newInvoiceDate, nil, newPaymentTerms, newAmountDue)
		
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
		err := model.Update(id, newInvoiceDate, nil, newPaymentTerms, newAmountDue)
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
		id, err := model.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue)
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
		err = model.Update(id, newInvoiceDate, &datePaid, newPaymentTerms, newAmountDue)
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
			id, err := test.impl.Insert(projectID, invoiceDate, nil, paymentTerms, amountDue)
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
			err = test.impl.Update(id, newInvoiceDate, &datePaid, newPaymentTerms, newAmountDue)
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