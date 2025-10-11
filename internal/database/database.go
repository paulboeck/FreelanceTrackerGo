package database

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	// The underscore (_) import is a "blank import" - it imports the package only for its side effects
	// Here, it registers the SQLite driver with the database/sql package without directly using it
	_ "modernc.org/sqlite"
)

// OpenDB opens a SQLite database connection and verifies it works.
func OpenDB(dsn string) (*sql.DB, error) {
	// sql.Open doesn't actually open a connection - it just validates the DSN
	// The actual connection happens later when you use the database
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		// %w wraps the error, preserving the original error while adding context
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ping actually attempts to connect to the database to verify it's accessible
	if err := db.Ping(); err != nil {
		db.Close() // Clean up if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations runs database migrations using goose for SQLite.
// Migrations are SQL files that update the database schema (add tables, columns, etc.)
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Set the SQL dialect so goose knows what syntax to use
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Run all pending migrations (migrations that haven't been applied yet)
	// goose.Up applies all migrations in order
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
