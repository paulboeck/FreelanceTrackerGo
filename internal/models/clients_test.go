package models

import (
	"testing"

	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientModel_Insert(t *testing.T) {
	// Setup test database using SQLite
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewClientModel(testDB.DB)

	t.Run("successful insert", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		name := "Test Client"
		id, err := model.Insert(name)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the client was actually inserted using direct query
		var insertedName string
		err = testDB.DB.QueryRow("SELECT name FROM client WHERE id = ?", id).Scan(&insertedName)
		require.NoError(t, err)
		assert.Equal(t, name, insertedName)
	})

	t.Run("insert empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		id, err := model.Insert("")
		
		// Should succeed at database level (validation happens at handler level)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
	})
}

func TestClientModel_Get(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewClientModel(testDB.DB)

	t.Run("get existing client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Insert a test client
		expectedName := "Test Client"
		id := testDB.InsertTestClient(t, expectedName)
		
		// Get the client using model
		client, err := model.Get(id)
		
		require.NoError(t, err)
		assert.Equal(t, id, client.ID)
		assert.Equal(t, expectedName, client.Name)
		assert.False(t, client.Created.IsZero())
		assert.False(t, client.Updated.IsZero())
	})

	t.Run("get non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		client, err := model.Get(999)
		
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		assert.Equal(t, Client{}, client)
	})
}

func TestClientModel_GetAll(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewClientModel(testDB.DB)

	t.Run("get all with no clients", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		clients, err := model.GetAll()
		
		require.NoError(t, err)
		assert.Empty(t, clients)
	})

	t.Run("get all with multiple clients", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Insert multiple clients
		names := []string{"Client A", "Client B", "Client C"}
		expectedIDs := make([]int, len(names))
		
		for i, name := range names {
			expectedIDs[i] = testDB.InsertTestClient(t, name)
		}
		
		clients, err := model.GetAll()
		
		require.NoError(t, err)
		require.Len(t, clients, len(names))
		
		// Verify all clients are returned (SQLC orders by created_at DESC)
		// So we expect reverse order from insertion
		clientMap := make(map[int]string)
		for _, client := range clients {
			clientMap[client.ID] = client.Name
			assert.False(t, client.Created.IsZero())
			assert.False(t, client.Updated.IsZero())
		}
		
		// Verify all expected clients exist regardless of order
		for i, expectedID := range expectedIDs {
			assert.Equal(t, names[i], clientMap[expectedID])
		}
	})
}

func TestClientModel_Integration(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewClientModel(testDB.DB)

	t.Run("full CRUD workflow with model", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// 1. Insert a new client
		clientName := "Integration Test Client"
		id, err := model.Insert(clientName)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// 2. Get the client
		client, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, client.ID)
		assert.Equal(t, clientName, client.Name)
		
		// 3. Verify it appears in GetAll
		clients, err := model.GetAll()
		require.NoError(t, err)
		require.Len(t, clients, 1)
		assert.Equal(t, client.ID, clients[0].ID)
		assert.Equal(t, client.Name, clients[0].Name)
	})
}

// TestInterface verifies that both implementations satisfy the same interface
func TestClientModelInterface(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	
	implementations := []struct {
		name string
		impl ClientModelInterface
	}{
		{"SQLite ClientModel", NewClientModel(testDB.DB)},
	}
	
	for _, test := range implementations {
		t.Run(test.name, func(t *testing.T) {
			testDB.TruncateTable(t, "client")
			
			// Test that both implementations work identically
			name := "Interface Test Client"
			
			// Insert
			id, err := test.impl.Insert(name)
			require.NoError(t, err)
			assert.Greater(t, id, 0)
			
			// Get
			client, err := test.impl.Get(id)
			require.NoError(t, err)
			assert.Equal(t, id, client.ID)
			assert.Equal(t, name, client.Name)
			
			// GetAll
			clients, err := test.impl.GetAll()
			require.NoError(t, err)
			require.Len(t, clients, 1)
			assert.Equal(t, id, clients[0].ID)
			assert.Equal(t, name, clients[0].Name)
		})
	}
}

func TestClientModel_Update(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewClientModel(testDB.DB)

	t.Run("successful update", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Insert a test client
		originalName := "Original Client"
		id := testDB.InsertTestClient(t, originalName)
		
		// Update the client
		newName := "Updated Client"
		err := model.Update(id, newName)
		require.NoError(t, err)
		
		// Verify the client was updated
		client, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, client.ID)
		assert.Equal(t, newName, client.Name)
		assert.False(t, client.Updated.IsZero())
		
		// Verify the updated_at timestamp changed (it should be after creation)
		assert.True(t, client.Updated.After(client.Created) || client.Updated.Equal(client.Created))
	})

	t.Run("update non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		err := model.Update(999, "New Name")
		
		// Should not return an error (MySQL UPDATE doesn't fail for non-existent rows)
		require.NoError(t, err)
		
		// Verify no client exists with this name
		clients, err := model.GetAll()
		require.NoError(t, err)
		assert.Empty(t, clients)
	})

	t.Run("update with empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Insert a test client
		originalName := "Original Client"
		id := testDB.InsertTestClient(t, originalName)
		
		// Update with empty name (should succeed at database level)
		err := model.Update(id, "")
		require.NoError(t, err)
		
		// Verify the client was updated
		client, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "", client.Name)
	})
}

// Update the interface test to include the Update method
func TestClientModelInterface_Update(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	
	implementations := []struct {
		name string
		impl ClientModelInterface
	}{
		{"SQLite ClientModel", NewClientModel(testDB.DB)},
	}
	
	for _, test := range implementations {
		t.Run(test.name, func(t *testing.T) {
			testDB.TruncateTable(t, "client")
			
			// Test that both implementations work identically for Update
			originalName := "Interface Test Client"
			
			// Insert
			id, err := test.impl.Insert(originalName)
			require.NoError(t, err)
			assert.Greater(t, id, 0)
			
			// Update
			newName := "Updated Interface Test Client"
			err = test.impl.Update(id, newName)
			require.NoError(t, err)
			
			// Get and verify update
			client, err := test.impl.Get(id)
			require.NoError(t, err)
			assert.Equal(t, id, client.ID)
			assert.Equal(t, newName, client.Name)
			
			// Verify in GetAll
			clients, err := test.impl.GetAll()
			require.NoError(t, err)
			require.Len(t, clients, 1)
			assert.Equal(t, id, clients[0].ID)
			assert.Equal(t, newName, clients[0].Name)
		})
	}
}

func TestClientModel_Delete(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewClientModel(testDB.DB)

	t.Run("successful delete", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Insert a test client
		originalName := "Client to Delete"
		id := testDB.InsertTestClient(t, originalName)
		
		// Verify client exists
		client, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, originalName, client.Name)
		assert.Nil(t, client.DeletedAt)
		
		// Delete the client
		err = model.Delete(id)
		require.NoError(t, err)
		
		// Verify the client is no longer returned by Get (soft deleted)
		_, err = model.Get(id)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		
		// Verify the client is no longer in GetAll
		clients, err := model.GetAll()
		require.NoError(t, err)
		assert.Empty(t, clients)
		
		// Verify the client still exists in database but with deleted_at set
		var deletedAt interface{}
		err = testDB.DB.QueryRow("SELECT deleted_at FROM client WHERE id = ?", id).Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("delete non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		err := model.Delete(999)
		
		// Should not return an error (SQLite UPDATE doesn't fail for non-existent rows)
		require.NoError(t, err)
	})

	t.Run("delete already deleted client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Insert and delete a client
		originalName := "Already Deleted Client"
		id := testDB.InsertTestClient(t, originalName)
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

func TestClientModel_SoftDeleteIntegration(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewClientModel(testDB.DB)

	t.Run("soft delete excludes clients from queries", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Insert multiple clients
		activeClient := testDB.InsertTestClient(t, "Active Client")
		deletedClient := testDB.InsertTestClient(t, "Deleted Client")
		anotherActiveClient := testDB.InsertTestClient(t, "Another Active Client")
		
		// Delete one client
		err := model.Delete(deletedClient)
		require.NoError(t, err)
		
		// Verify GetAll only returns active clients
		clients, err := model.GetAll()
		require.NoError(t, err)
		require.Len(t, clients, 2)
		
		// Verify the correct clients are returned
		clientIDs := make([]int, len(clients))
		for i, client := range clients {
			clientIDs[i] = client.ID
		}
		assert.Contains(t, clientIDs, activeClient)
		assert.Contains(t, clientIDs, anotherActiveClient)
		assert.NotContains(t, clientIDs, deletedClient)
		
		// Verify Get returns active clients
		activeClientResult, err := model.Get(activeClient)
		require.NoError(t, err)
		assert.Equal(t, "Active Client", activeClientResult.Name)
		
		// Verify Get doesn't return deleted client
		_, err = model.Get(deletedClient)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
	})
}