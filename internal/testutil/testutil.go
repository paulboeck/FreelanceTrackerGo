package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/stretchr/testify/require"

	_ "github.com/go-sql-driver/mysql"
)

// TestDatabase represents a test database instance
type TestDatabase struct {
	DB        *sql.DB
	container *mysql.MySQLContainer
	DSN       string
}

// SetupTestMySQL creates a test MySQL database using testcontainers
func SetupTestMySQL(t *testing.T) *TestDatabase {
	ctx := context.Background()

	mysqlContainer, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0"),
		mysql.WithDatabase("freelance_tracker_test"),
		mysql.WithUsername("test"),
		mysql.WithPassword("test"),
	)
	require.NoError(t, err)

	// Get the mapped port
	port, err := mysqlContainer.MappedPort(ctx, "3306")
	require.NoError(t, err)

	// Get the host
	host, err := mysqlContainer.Host(ctx)
	require.NoError(t, err)

	// Build DSN
	dsn := fmt.Sprintf("test:test@tcp(%s:%s)/freelance_tracker_test?parseTime=true", host, port.Port())

	// Wait for the database to be ready
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)

	// Wait for database to be ready with retry
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
		if i == maxRetries-1 {
			require.NoError(t, err, "database failed to become ready")
		}
	}

	// Create the table schema
	err = createSchema(db)
	require.NoError(t, err)

	return &TestDatabase{
		DB:        db,
		container: mysqlContainer,
		DSN:       dsn,
	}
}

// Cleanup cleans up the test database
func (td *TestDatabase) Cleanup(t *testing.T) {
	if td.DB != nil {
		td.DB.Close()
	}
	if td.container != nil {
		ctx := context.Background()
		err := td.container.Terminate(ctx)
		if err != nil {
			log.Printf("Error terminating container: %v", err)
		}
	}
}

// createSchema creates the necessary tables for testing
func createSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS client (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);
	`
	
	_, err := db.Exec(schema)
	return err
}

// TruncateTable truncates the specified table for test cleanup
func (td *TestDatabase) TruncateTable(t *testing.T, tableName string) {
	_, err := td.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName))
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