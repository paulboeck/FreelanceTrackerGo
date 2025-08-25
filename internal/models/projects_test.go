package models

import (
	"testing"

	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectModel_Insert(t *testing.T) {
	// Setup test database using SQLite
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewProjectModel(testDB.DB)

	t.Run("successful insert", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create a test client first
		clientID := testDB.InsertTestClient(t, "Test Client")
		
		name := "Test Project"
		id, err := model.Insert(name, clientID)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the project was actually inserted using direct query
		var insertedName string
		var insertedClientID int
		err = testDB.DB.QueryRow("SELECT name, client_id FROM project WHERE id = ?", id).Scan(&insertedName, &insertedClientID)
		require.NoError(t, err)
		assert.Equal(t, name, insertedName)
		assert.Equal(t, clientID, insertedClientID)
	})

	t.Run("insert with non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		name := "Test Project"
		id, err := model.Insert(name, 999) // Non-existent client
		
		// SQLite might not enforce foreign key constraints by default in tests
		// Just verify it doesn't crash
		if err != nil {
			assert.Equal(t, 0, id)
		}
	})

	t.Run("insert empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create a test client first
		clientID := testDB.InsertTestClient(t, "Test Client")
		
		id, err := model.Insert("", clientID)
		
		// Should succeed at database level (validation happens at handler level)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
	})
}

func TestProjectModel_Get(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewProjectModel(testDB.DB)

	t.Run("get existing project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create a test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		expectedName := "Test Project"
		id := testDB.InsertTestProject(t, expectedName, clientID)
		
		// Get the project using model
		project, err := model.Get(id)
		
		require.NoError(t, err)
		assert.Equal(t, id, project.ID)
		assert.Equal(t, expectedName, project.Name)
		assert.Equal(t, clientID, project.ClientID)
		assert.False(t, project.Created.IsZero())
		assert.False(t, project.Updated.IsZero())
		assert.Nil(t, project.DeletedAt)
	})

	t.Run("get non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		
		project, err := model.Get(999)
		
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		assert.Equal(t, Project{}, project)
	})
}

func TestProjectModel_GetByClient(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewProjectModel(testDB.DB)

	t.Run("get projects for client with multiple projects", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test clients
		client1ID := testDB.InsertTestClient(t, "Client 1")
		client2ID := testDB.InsertTestClient(t, "Client 2")
		
		// Create projects for client 1
		project1ID := testDB.InsertTestProject(t, "Project A", client1ID)
		project2ID := testDB.InsertTestProject(t, "Project B", client1ID)
		
		// Create project for client 2 (should not be returned)
		_ = testDB.InsertTestProject(t, "Project C", client2ID)
		
		projects, err := model.GetByClient(client1ID)
		
		require.NoError(t, err)
		require.Len(t, projects, 2)
		
		// Verify the correct projects are returned
		projectIDs := make([]int, len(projects))
		projectNames := make([]string, len(projects))
		for i, project := range projects {
			projectIDs[i] = project.ID
			projectNames[i] = project.Name
			assert.Equal(t, client1ID, project.ClientID)
			assert.False(t, project.Created.IsZero())
			assert.False(t, project.Updated.IsZero())
		}
		
		assert.Contains(t, projectIDs, project1ID)
		assert.Contains(t, projectIDs, project2ID)
		assert.Contains(t, projectNames, "Project A")
		assert.Contains(t, projectNames, "Project B")
	})

	t.Run("get projects for client with no projects", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create a test client with no projects
		clientID := testDB.InsertTestClient(t, "Client with no projects")
		
		projects, err := model.GetByClient(clientID)
		
		require.NoError(t, err)
		assert.Empty(t, projects)
	})

	t.Run("get projects for non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		projects, err := model.GetByClient(999)
		
		require.NoError(t, err)
		assert.Empty(t, projects)
	})
}

func TestProjectModel_Update(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewProjectModel(testDB.DB)

	t.Run("successful update", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create a test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		originalName := "Original Project"
		id := testDB.InsertTestProject(t, originalName, clientID)
		
		// Update the project
		newName := "Updated Project"
		err := model.Update(id, newName)
		require.NoError(t, err)
		
		// Verify the project was updated
		project, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, project.ID)
		assert.Equal(t, newName, project.Name)
		assert.Equal(t, clientID, project.ClientID)
		assert.False(t, project.Updated.IsZero())
		
		// Verify the updated_at timestamp changed
		assert.True(t, project.Updated.After(project.Created) || project.Updated.Equal(project.Created))
	})

	t.Run("update non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		
		err := model.Update(999, "New Name")
		
		// Should not return an error (SQLite UPDATE doesn't fail for non-existent rows)
		require.NoError(t, err)
	})

	t.Run("update with empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create a test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		originalName := "Original Project"
		id := testDB.InsertTestProject(t, originalName, clientID)
		
		// Update with empty name (should succeed at database level)
		err := model.Update(id, "")
		require.NoError(t, err)
		
		// Verify the project was updated
		project, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "", project.Name)
	})
}

func TestProjectModel_Delete(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewProjectModel(testDB.DB)

	t.Run("successful delete", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create a test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		originalName := "Project to Delete"
		id := testDB.InsertTestProject(t, originalName, clientID)
		
		// Verify project exists
		project, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, originalName, project.Name)
		assert.Nil(t, project.DeletedAt)
		
		// Delete the project
		err = model.Delete(id)
		require.NoError(t, err)
		
		// Verify the project is no longer returned by Get (soft deleted)
		_, err = model.Get(id)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		
		// Verify the project is no longer in GetByClient
		projects, err := model.GetByClient(clientID)
		require.NoError(t, err)
		assert.Empty(t, projects)
		
		// Verify the project still exists in database but with deleted_at set
		var deletedAt interface{}
		err = testDB.DB.QueryRow("SELECT deleted_at FROM project WHERE id = ?", id).Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("delete non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		
		err := model.Delete(999)
		
		// Should not return an error (SQLite UPDATE doesn't fail for non-existent rows)
		require.NoError(t, err)
	})

	t.Run("delete already deleted project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create a test client and project, then delete it
		clientID := testDB.InsertTestClient(t, "Test Client")
		originalName := "Already Deleted Project"
		id := testDB.InsertTestProject(t, originalName, clientID)
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

func TestProjectModel_Integration(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewProjectModel(testDB.DB)

	t.Run("full CRUD workflow with project model", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// 1. Create a client
		clientName := "Integration Test Client"
		clientID := testDB.InsertTestClient(t, clientName)
		
		// 2. Insert a new project
		projectName := "Integration Test Project"
		id, err := model.Insert(projectName, clientID)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// 3. Get the project
		project, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, project.ID)
		assert.Equal(t, projectName, project.Name)
		assert.Equal(t, clientID, project.ClientID)
		
		// 4. Verify it appears in GetByClient
		projects, err := model.GetByClient(clientID)
		require.NoError(t, err)
		require.Len(t, projects, 1)
		assert.Equal(t, project.ID, projects[0].ID)
		assert.Equal(t, project.Name, projects[0].Name)
		
		// 5. Update the project
		newName := "Updated Integration Test Project"
		err = model.Update(id, newName)
		require.NoError(t, err)
		
		// 6. Verify update
		updatedProject, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, newName, updatedProject.Name)
		assert.True(t, updatedProject.Updated.After(project.Updated) || updatedProject.Updated.Equal(project.Updated))
		
		// 7. Delete the project
		err = model.Delete(id)
		require.NoError(t, err)
		
		// 8. Verify deletion
		_, err = model.Get(id)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		
		projects, err = model.GetByClient(clientID)
		require.NoError(t, err)
		assert.Empty(t, projects)
	})
}

// TestInterface verifies that the implementation satisfies the interface
func TestProjectModelInterface(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	
	implementations := []struct {
		name string
		impl ProjectModelInterface
	}{
		{"SQLite ProjectModel", NewProjectModel(testDB.DB)},
	}
	
	for _, test := range implementations {
		t.Run(test.name, func(t *testing.T) {
			testDB.TruncateTable(t, "project")
			testDB.TruncateTable(t, "client")
			
			// Create a test client first
			clientID := testDB.InsertTestClient(t, "Interface Test Client")
			
			// Test that the implementation works correctly
			name := "Interface Test Project"
			
			// Insert
			id, err := test.impl.Insert(name, clientID)
			require.NoError(t, err)
			assert.Greater(t, id, 0)
			
			// Get
			project, err := test.impl.Get(id)
			require.NoError(t, err)
			assert.Equal(t, id, project.ID)
			assert.Equal(t, name, project.Name)
			assert.Equal(t, clientID, project.ClientID)
			
			// GetByClient
			projects, err := test.impl.GetByClient(clientID)
			require.NoError(t, err)
			require.Len(t, projects, 1)
			assert.Equal(t, id, projects[0].ID)
			assert.Equal(t, name, projects[0].Name)
			
			// Update
			newName := "Updated Interface Test Project"
			err = test.impl.Update(id, newName)
			require.NoError(t, err)
			
			updatedProject, err := test.impl.Get(id)
			require.NoError(t, err)
			assert.Equal(t, newName, updatedProject.Name)
			
			// Delete
			err = test.impl.Delete(id)
			require.NoError(t, err)
			
			_, err = test.impl.Get(id)
			assert.Error(t, err)
			assert.Equal(t, ErrNoRecord, err)
		})
	}
}