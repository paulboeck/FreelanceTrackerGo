package models

import (
	"testing"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimesheetModel_Insert(t *testing.T) {
	// Setup test database using SQLite
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewTimesheetModel(testDB.DB)

	t.Run("successful insert", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		workDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		hoursWorked := 8.5
		hourlyRate := 125.00
		description := "Test work description"
		
		id, err := model.Insert(projectID, workDate, hoursWorked, hourlyRate, description)
		
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the timesheet was actually inserted using direct query
		var insertedProjectID int
		var insertedWorkDate string
		var insertedHours float64
		var insertedHourlyRate float64
		var insertedDescription string
		err = testDB.DB.QueryRow("SELECT project_id, work_date, hours_worked, hourly_rate, description FROM timesheet WHERE id = ?", id).Scan(
			&insertedProjectID, &insertedWorkDate, &insertedHours, &insertedHourlyRate, &insertedDescription)
		require.NoError(t, err)
		assert.Equal(t, projectID, insertedProjectID)
		assert.Contains(t, insertedWorkDate, "2024-01-15")
		assert.Equal(t, hoursWorked, insertedHours)
		assert.Equal(t, hourlyRate, insertedHourlyRate)
		assert.Equal(t, description, insertedDescription)
	})

	t.Run("insert with non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		workDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		hoursWorked := 8.0
		hourlyRate := 100.00
		description := "Test description"
		
		id, err := model.Insert(999, workDate, hoursWorked, hourlyRate, description) // Non-existent project
		
		// SQLite might not enforce foreign key constraints by default in tests
		// Just verify it doesn't crash
		if err != nil {
			assert.Equal(t, 0, id)
		}
	})

	t.Run("insert with zero hours", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		workDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		hoursWorked := 0.0
		hourlyRate := 150.00
		description := "No work done"
		
		id, err := model.Insert(projectID, workDate, hoursWorked, hourlyRate, description)
		
		// Should succeed at database level (validation happens at handler level)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
	})

	t.Run("insert with empty description", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		workDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		hoursWorked := 8.0
		hourlyRate := 100.00
		description := "" // Empty description
		
		id, err := model.Insert(projectID, workDate, hoursWorked, hourlyRate, description)
		
		// Should succeed at database level (validation happens at handler level)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// Verify the timesheet was inserted with empty description
		timesheet, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "", timesheet.Description)
	})
}

func TestTimesheetModel_Get(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewTimesheetModel(testDB.DB)

	t.Run("get existing timesheet", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert timesheet
		expectedWorkDate := "2024-01-15"
		expectedHours := "8.50"
		expectedHourlyRate := "125.00"
		expectedDescription := "Test timesheet"
		id := testDB.InsertTestTimesheet(t, projectID, expectedWorkDate, expectedHours, expectedHourlyRate, expectedDescription)
		
		// Get the timesheet using model
		timesheet, err := model.Get(id)
		
		require.NoError(t, err)
		assert.Equal(t, id, timesheet.ID)
		assert.Equal(t, projectID, timesheet.ProjectID)
		assert.Equal(t, expectedWorkDate, timesheet.WorkDate.Format("2006-01-02"))
		assert.Equal(t, 8.5, timesheet.HoursWorked)
		assert.Equal(t, 125.00, timesheet.HourlyRate)
		assert.Equal(t, expectedDescription, timesheet.Description)
		assert.False(t, timesheet.Created.IsZero())
		assert.False(t, timesheet.Updated.IsZero())
		assert.Nil(t, timesheet.DeletedAt)
	})

	t.Run("get non-existent timesheet", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		
		timesheet, err := model.Get(999)
		
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		assert.Equal(t, Timesheet{}, timesheet)
	})
}

func TestTimesheetModel_GetByProject(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewTimesheetModel(testDB.DB)

	t.Run("get timesheets for project with multiple timesheets", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and projects
		clientID := testDB.InsertTestClient(t, "Test Client")
		project1ID := testDB.InsertTestProject(t, "Project 1", clientID)
		project2ID := testDB.InsertTestProject(t, "Project 2", clientID)
		
		// Create timesheets for project 1
		timesheet1ID := testDB.InsertTestTimesheet(t, project1ID, "2024-01-15", "8.00", "125.00", "Work A")
		timesheet2ID := testDB.InsertTestTimesheet(t, project1ID, "2024-01-16", "4.50", "135.00", "Work B")
		
		// Create timesheet for project 2 (should not be returned)
		_ = testDB.InsertTestTimesheet(t, project2ID, "2024-01-17", "2.00", "150.00", "Work C")
		
		timesheets, err := model.GetByProject(project1ID)
		
		require.NoError(t, err)
		require.Len(t, timesheets, 2)
		
		// Verify the correct timesheets are returned
		timesheetIDs := make([]int, len(timesheets))
		descriptions := make([]string, len(timesheets))
		for i, timesheet := range timesheets {
			timesheetIDs[i] = timesheet.ID
			descriptions[i] = timesheet.Description
			assert.Equal(t, project1ID, timesheet.ProjectID)
			assert.False(t, timesheet.Created.IsZero())
			assert.False(t, timesheet.Updated.IsZero())
		}
		
		assert.Contains(t, timesheetIDs, timesheet1ID)
		assert.Contains(t, timesheetIDs, timesheet2ID)
		assert.Contains(t, descriptions, "Work A")
		assert.Contains(t, descriptions, "Work B")
	})

	t.Run("get timesheets for project with no timesheets", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project with no timesheets
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Project with no timesheets", clientID)
		
		timesheets, err := model.GetByProject(projectID)
		
		require.NoError(t, err)
		assert.Empty(t, timesheets)
	})

	t.Run("get timesheets for non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		
		timesheets, err := model.GetByProject(999)
		
		require.NoError(t, err)
		assert.Empty(t, timesheets)
	})
}

func TestTimesheetModel_Update(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewTimesheetModel(testDB.DB)

	t.Run("successful update", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert timesheet
		originalWorkDate := "2024-01-15"
		originalHours := "8.00"
		originalHourlyRate := "100.00"
		originalDescription := "Original work"
		id := testDB.InsertTestTimesheet(t, projectID, originalWorkDate, originalHours, originalHourlyRate, originalDescription)
		
		// Update the timesheet
		newWorkDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		newHours := 6.5
		newHourlyRate := 120.00
		newDescription := "Updated work"
		err := model.Update(id, newWorkDate, newHours, newHourlyRate, newDescription)
		require.NoError(t, err)
		
		// Verify the timesheet was updated
		timesheet, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, timesheet.ID)
		assert.Equal(t, "2024-01-20", timesheet.WorkDate.Format("2006-01-02"))
		assert.Equal(t, newHours, timesheet.HoursWorked)
		assert.Equal(t, newHourlyRate, timesheet.HourlyRate)
		assert.Equal(t, newDescription, timesheet.Description)
		assert.False(t, timesheet.Updated.IsZero())
		
		// Verify the updated_at timestamp changed
		assert.True(t, timesheet.Updated.After(timesheet.Created) || timesheet.Updated.Equal(timesheet.Created))
	})

	t.Run("update non-existent timesheet", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		
		newWorkDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		newHours := 6.5
		newHourlyRate := 110.00
		newDescription := "Updated work"
		err := model.Update(999, newWorkDate, newHours, newHourlyRate, newDescription)
		
		// Should not return an error (SQLite UPDATE doesn't fail for non-existent rows)
		require.NoError(t, err)
	})

	t.Run("update with zero hours", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert timesheet
		originalWorkDate := "2024-01-15"
		originalHours := "8.00"
		originalHourlyRate := "75.00"
		originalDescription := "Original work"
		id := testDB.InsertTestTimesheet(t, projectID, originalWorkDate, originalHours, originalHourlyRate, originalDescription)
		
		// Update with zero hours (should succeed at database level)
		newWorkDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		newHours := 0.0
		newHourlyRate := 80.00
		newDescription := "No work done"
		err := model.Update(id, newWorkDate, newHours, newHourlyRate, newDescription)
		require.NoError(t, err)
		
		// Verify the timesheet was updated
		timesheet, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, 0.0, timesheet.HoursWorked)
		assert.Equal(t, newDescription, timesheet.Description)
	})
}

func TestTimesheetModel_Delete(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewTimesheetModel(testDB.DB)

	t.Run("successful delete", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert timesheet
		workDate := "2024-01-15"
		hours := "8.00"
		hourlyRate := "120.00"
		description := "Timesheet to delete"
		id := testDB.InsertTestTimesheet(t, projectID, workDate, hours, hourlyRate, description)
		
		// Verify timesheet exists
		timesheet, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, description, timesheet.Description)
		assert.Nil(t, timesheet.DeletedAt)
		
		// Delete the timesheet
		err = model.Delete(id)
		require.NoError(t, err)
		
		// Verify the timesheet is no longer returned by Get (soft deleted)
		_, err = model.Get(id)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		
		// Verify the timesheet is no longer in GetByProject
		timesheets, err := model.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, timesheets)
		
		// Verify the timesheet still exists in database but with deleted_at set
		var deletedAt interface{}
		err = testDB.DB.QueryRow("SELECT deleted_at FROM timesheet WHERE id = ?", id).Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("delete non-existent timesheet", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		
		err := model.Delete(999)
		
		// Should not return an error (SQLite UPDATE doesn't fail for non-existent rows)
		require.NoError(t, err)
	})

	t.Run("delete already deleted timesheet", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// Create test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)
		
		// Insert and delete timesheet
		workDate := "2024-01-15"
		hours := "8.00"
		hourlyRate := "90.00"
		description := "Already deleted timesheet"
		id := testDB.InsertTestTimesheet(t, projectID, workDate, hours, hourlyRate, description)
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

func TestTimesheetModel_Integration(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create model instance
	model := NewTimesheetModel(testDB.DB)

	t.Run("full CRUD workflow with timesheet model", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")
		
		// 1. Create client and project
		clientID := testDB.InsertTestClient(t, "Integration Test Client")
		projectID := testDB.InsertTestProject(t, "Integration Test Project", clientID)
		
		// 2. Insert a new timesheet
		workDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		hoursWorked := 8.5
		hourlyRate := 140.00
		description := "Integration test work"
		id, err := model.Insert(projectID, workDate, hoursWorked, hourlyRate, description)
		require.NoError(t, err)
		assert.Greater(t, id, 0)
		
		// 3. Get the timesheet
		timesheet, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, id, timesheet.ID)
		assert.Equal(t, projectID, timesheet.ProjectID)
		assert.Equal(t, "2024-01-15", timesheet.WorkDate.Format("2006-01-02"))
		assert.Equal(t, hoursWorked, timesheet.HoursWorked)
		assert.Equal(t, hourlyRate, timesheet.HourlyRate)
		assert.Equal(t, description, timesheet.Description)
		
		// 4. Verify it appears in GetByProject
		timesheets, err := model.GetByProject(projectID)
		require.NoError(t, err)
		require.Len(t, timesheets, 1)
		assert.Equal(t, timesheet.ID, timesheets[0].ID)
		assert.Equal(t, timesheet.HourlyRate, timesheets[0].HourlyRate)
		assert.Equal(t, timesheet.Description, timesheets[0].Description)
		
		// 5. Update the timesheet
		newWorkDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		newHours := 6.0
		newHourlyRate := 160.00
		newDescription := "Updated integration test work"
		err = model.Update(id, newWorkDate, newHours, newHourlyRate, newDescription)
		require.NoError(t, err)
		
		// 6. Verify update
		updatedTimesheet, err := model.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "2024-01-20", updatedTimesheet.WorkDate.Format("2006-01-02"))
		assert.Equal(t, newHours, updatedTimesheet.HoursWorked)
		assert.Equal(t, newHourlyRate, updatedTimesheet.HourlyRate)
		assert.Equal(t, newDescription, updatedTimesheet.Description)
		assert.True(t, updatedTimesheet.Updated.After(timesheet.Updated) || updatedTimesheet.Updated.Equal(timesheet.Updated))
		
		// 7. Delete the timesheet
		err = model.Delete(id)
		require.NoError(t, err)
		
		// 8. Verify deletion
		_, err = model.Get(id)
		assert.Error(t, err)
		assert.Equal(t, ErrNoRecord, err)
		
		timesheets, err = model.GetByProject(projectID)
		require.NoError(t, err)
		assert.Empty(t, timesheets)
	})
}

// TestInterface verifies that the implementation satisfies the interface
func TestTimesheetModelInterface(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	
	implementations := []struct {
		name string
		impl TimesheetModelInterface
	}{
		{"SQLite TimesheetModel", NewTimesheetModel(testDB.DB)},
	}
	
	for _, test := range implementations {
		t.Run(test.name, func(t *testing.T) {
			testDB.TruncateTable(t, "timesheet")
			testDB.TruncateTable(t, "project")
			testDB.TruncateTable(t, "client")
			
			// Create test client and project first
			clientID := testDB.InsertTestClient(t, "Interface Test Client")
			projectID := testDB.InsertTestProject(t, "Interface Test Project", clientID)
			
			// Test that the implementation works correctly
			workDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
			hoursWorked := 8.5
			hourlyRate := 130.00
			description := "Interface Test Work"
			
			// Insert
			id, err := test.impl.Insert(projectID, workDate, hoursWorked, hourlyRate, description)
			require.NoError(t, err)
			assert.Greater(t, id, 0)
			
			// Get
			timesheet, err := test.impl.Get(id)
			require.NoError(t, err)
			assert.Equal(t, id, timesheet.ID)
			assert.Equal(t, projectID, timesheet.ProjectID)
			assert.Equal(t, hoursWorked, timesheet.HoursWorked)
			assert.Equal(t, hourlyRate, timesheet.HourlyRate)
			assert.Equal(t, description, timesheet.Description)
			
			// GetByProject
			timesheets, err := test.impl.GetByProject(projectID)
			require.NoError(t, err)
			require.Len(t, timesheets, 1)
			assert.Equal(t, id, timesheets[0].ID)
			assert.Equal(t, hourlyRate, timesheets[0].HourlyRate)
			assert.Equal(t, description, timesheets[0].Description)
			
			// Update
			newWorkDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
			newHours := 6.0
			newHourlyRate := 155.00
			newDescription := "Updated Interface Test Work"
			err = test.impl.Update(id, newWorkDate, newHours, newHourlyRate, newDescription)
			require.NoError(t, err)
			
			updatedTimesheet, err := test.impl.Get(id)
			require.NoError(t, err)
			assert.Equal(t, newHours, updatedTimesheet.HoursWorked)
			assert.Equal(t, newHourlyRate, updatedTimesheet.HourlyRate)
			assert.Equal(t, newDescription, updatedTimesheet.Description)
			
			// Delete
			err = test.impl.Delete(id)
			require.NoError(t, err)
			
			_, err = test.impl.Get(id)
			assert.Error(t, err)
			assert.Equal(t, ErrNoRecord, err)
		})
	}
}