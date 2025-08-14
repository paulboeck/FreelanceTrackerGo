package models

import (
	"testing"

	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientAdapter_Insert(t *testing.T) {
	// Setup test database (using MySQL testcontainer for consistent testing)
	testDB := testutil.SetupTestMySQL(t)
	defer testDB.Cleanup(t)

	// Create adapter instance
	adapter := NewClientAdapter(testDB.DB)

	t.Run("successful insert", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		name := "Test Client"
		id, err := adapter.Insert(name)
		
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
		
		id, err := adapter.Insert("")
		
		// Should succeed at database level (validation happens at handler level)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
	})
}

func TestClientAdapter_Get(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestMySQL(t)
	defer testDB.Cleanup(t)

	// Create adapter instance
	adapter := NewClientAdapter(testDB.DB)

	t.Run("get existing client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Insert a test client
		expectedName := "Test Client"
		id := testDB.InsertTestClient(t, expectedName)
		
		// Get the client using adapter
		client, err := adapter.Get(id)
		
		require.NoError(t, err)
		assert.Equal(t, id, client.ID)
		assert.Equal(t, expectedName, client.Name)
		assert.False(t, client.Created.IsZero())
		assert.False(t, client.Updated.IsZero())
	})

	t.Run("get non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		client, err := adapter.Get(999)
		
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		assert.Equal(t, Client{}, client)
	})
}

func TestClientAdapter_GetAll(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestMySQL(t)
	defer testDB.Cleanup(t)

	// Create adapter instance
	adapter := NewClientAdapter(testDB.DB)

	t.Run("get all with no clients", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		clients, err := adapter.GetAll()
		
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
		
		clients, err := adapter.GetAll()
		
		require.NoError(t, err)
		require.Len(t, clients, len(names))
		
		// Verify all clients are returned (SQLC orders by created_at DESC)
		// So we expect reverse order from insertion
		for i, client := range clients {
			expectedIndex := len(names) - 1 - i // Reverse order
			assert.Equal(t, expectedIDs[expectedIndex], client.ID)
			assert.Equal(t, names[expectedIndex], client.Name)
			assert.False(t, client.Created.IsZero())
			assert.False(t, client.Updated.IsZero())
		}
	})
}

func TestClientAdapter_Integration(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestMySQL(t)
	defer testDB.Cleanup(t)

	// Create adapter instance
	adapter := NewClientAdapter(testDB.DB)

	t.Run("full CRUD workflow with adapter", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// 1. Insert a new client
		clientName := "Integration Test Client"
		id, err := adapter.Insert(clientName)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// 2. Get the client
		client, err := adapter.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, client.ID)
		assert.Equal(t, clientName, client.Name)
		
		// 3. Verify it appears in GetAll
		clients, err := adapter.GetAll()
		require.NoError(t, err)
		require.Len(t, clients, 1)
		assert.Equal(t, client.ID, clients[0].ID)
		assert.Equal(t, client.Name, clients[0].Name)
	})
}

// TestInterface verifies that both implementations satisfy the same interface
func TestClientModelInterface(t *testing.T) {
	testDB := testutil.SetupTestMySQL(t)
	defer testDB.Cleanup(t)
	
	implementations := []struct {
		name string
		impl ClientModelInterface
	}{
		{"Original ClientModel", &ClientModel{DB: testDB.DB}},
		{"SQLC ClientAdapter", NewClientAdapter(testDB.DB)},
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