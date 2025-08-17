package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

// TestDatabase represents a test database instance
type TestDatabase struct {
	DB       *sql.DB
	filePath string
}

// SetupTestSQLite creates a test SQLite database in memory
func SetupTestSQLite(t *testing.T) *TestDatabase {
	// Create a temporary file for the test database
	tempDir := t.TempDir()
	dbFile := filepath.Join(tempDir, "test.db")

	// Open SQLite database
	db, err := sql.Open("sqlite", dbFile)
	require.NoError(t, err)

	// Test the connection
	err = db.Ping()
	require.NoError(t, err)

	// Create the table schema
	err = createSchema(db)
	require.NoError(t, err)

	return &TestDatabase{
		DB:       db,
		filePath: dbFile,
	}
}

// Cleanup cleans up the test database
func (td *TestDatabase) Cleanup(t *testing.T) {
	if td.DB != nil {
		td.DB.Close()
	}
	if td.filePath != "" {
		os.Remove(td.filePath)
	}
}

// createSchema creates the necessary tables for testing
func createSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS client (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL
		);
		
		CREATE TABLE IF NOT EXISTS project (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			client_id INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			FOREIGN KEY (client_id) REFERENCES client(id)
		);
		
		CREATE TABLE IF NOT EXISTS timesheet (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			work_date DATE NOT NULL,
			hours_worked DECIMAL(5,2) NOT NULL,
			description VARCHAR(255),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			FOREIGN KEY (project_id) REFERENCES project(id)
		);
	`
	
	_, err := db.Exec(schema)
	return err
}

// TruncateTable truncates the specified table for test cleanup
func (td *TestDatabase) TruncateTable(t *testing.T, tableName string) {
	_, err := td.DB.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
	require.NoError(t, err)
}

// InsertTestClient inserts a test client and returns its ID
func (td *TestDatabase) InsertTestClient(t *testing.T, name string) int {
	result, err := td.DB.Exec("INSERT INTO client (name) VALUES (?)", name)
	require.NoError(t, err)
	
	id, err := result.LastInsertId()
	require.NoError(t, err)
	
	return int(id)
}