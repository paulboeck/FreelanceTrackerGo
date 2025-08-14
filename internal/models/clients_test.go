package models

import (
	"testing"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientModel_Insert(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestMySQL(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := &ClientModel{DB: testDB.DB}

	t.Run("successful insert", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		name := "Test Client"
		id, err := model.Insert(name)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the client was actually inserted
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

	t.Run("insert very long name", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Create a name longer than 255 characters
		longName := ""
		for i := 0; i < 260; i++ {
			longName += "a"
		}
		
		_, err := model.Insert(longName)
		
		// Should fail due to database constraint
		assert.Error(t, err)
	})
}

func TestClientModel_Get(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestMySQL(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := &ClientModel{DB: testDB.DB}

	t.Run("get existing client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		// Insert a test client
		expectedName := "Test Client"
		id := testDB.InsertTestClient(t, expectedName)
		
		// Get the client
		client, err := model.Get(id)
		
		require.NoError(t, err)
		assert.Equal(t, id, client.ID)
		assert.Equal(t, expectedName, client.Name)
		assert.False(t, client.Created.IsZero())
		assert.False(t, client.Updated.IsZero())
		
		// Verify timestamps are reasonable (within last minute)
		now := time.Now()
		assert.True(t, client.Created.Before(now) || client.Created.Equal(now))
		assert.True(t, client.Updated.Before(now) || client.Updated.Equal(now))
	})

	t.Run("get non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		client, err := model.Get(999)
		
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		assert.Equal(t, Client{}, client)
	})

	t.Run("get with invalid ID", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		client, err := model.Get(-1)
		
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		assert.Equal(t, Client{}, client)
	})
}

func TestClientModel_GetAll(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestMySQL(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := &ClientModel{DB: testDB.DB}

	t.Run("get all with no clients", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		clients, err := model.GetAll()
		
		require.NoError(t, err)
		assert.Empty(t, clients)
	})

	t.Run("get all with single client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		
		expectedName := "Single Client"
		id := testDB.InsertTestClient(t, expectedName)
		
		clients, err := model.GetAll()
		
		require.NoError(t, err)
		require.Len(t, clients, 1)
		
		client := clients[0]
		assert.Equal(t, id, client.ID)
		assert.Equal(t, expectedName, client.Name)
		assert.False(t, client.Created.IsZero())
		assert.False(t, client.Updated.IsZero())
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
		
		// Verify all clients are returned (order might vary)
		clientMap := make(map[int]string)
		for _, client := range clients {
			clientMap[client.ID] = client.Name
			assert.False(t, client.Created.IsZero())
			assert.False(t, client.Updated.IsZero())
		}
		
		for i, expectedID := range expectedIDs {
			assert.Equal(t, names[i], clientMap[expectedID])
		}
	})
}

func TestClientModel_Integration(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestMySQL(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := &ClientModel{DB: testDB.DB}

	t.Run("full CRUD workflow", func(t *testing.T) {
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
		
		// 4. Insert another client
		secondName := "Second Client"
		secondID, err := model.Insert(secondName)
		require.NoError(t, err)
		
		// 5. Verify GetAll returns both
		clients, err = model.GetAll()
		require.NoError(t, err)
		require.Len(t, clients, 2)
		
		// Verify both clients exist
		clientNames := make(map[int]string)
		for _, c := range clients {
			clientNames[c.ID] = c.Name
		}
		assert.Equal(t, clientName, clientNames[id])
		assert.Equal(t, secondName, clientNames[secondID])
	})
}

// Benchmark tests to establish performance baseline
func BenchmarkClientModel_Insert(b *testing.B) {
	testDB := testutil.SetupTestMySQL(&testing.T{})
	defer testDB.Cleanup(&testing.T{})
	
	model := &ClientModel{DB: testDB.DB}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.Insert("Benchmark Client")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientModel_Get(b *testing.B) {
	testDB := testutil.SetupTestMySQL(&testing.T{})
	defer testDB.Cleanup(&testing.T{})
	
	model := &ClientModel{DB: testDB.DB}
	
	// Insert a test client
	id, err := model.Insert("Benchmark Client")
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.Get(id)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientModel_GetAll(b *testing.B) {
	testDB := testutil.SetupTestMySQL(&testing.T{})
	defer testDB.Cleanup(&testing.T{})
	
	model := &ClientModel{DB: testDB.DB}
	
	// Insert several test clients
	for i := 0; i < 10; i++ {
		_, err := model.Insert("Benchmark Client")
		if err != nil {
			b.Fatal(err)
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.GetAll()
		if err != nil {
			b.Fatal(err)
		}
	}
}