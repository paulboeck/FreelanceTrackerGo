package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

type DatabaseType string

const (
	DatabaseTypeMySQL  DatabaseType = "mysql"
	DatabaseTypeSQLite DatabaseType = "sqlite"
)

type Config struct {
	Type DatabaseType
	DSN  string
}

// OpenDB opens a database connection based on the configuration
func OpenDB(config Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch config.Type {
	case DatabaseTypeMySQL:
		db, err = sql.Open("mysql", config.DSN)
	case DatabaseTypeSQLite:
		db, err = sql.Open("sqlite", config.DSN)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations runs database migrations using goose
func RunMigrations(db *sql.DB, dbType DatabaseType, migrationsDir string) error {
	var dialect string
	switch dbType {
	case DatabaseTypeMySQL:
		dialect = "mysql"
	case DatabaseTypeSQLite:
		dialect = "sqlite3"
	default:
		return fmt.Errorf("unsupported database type for migrations: %s", dbType)
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// GetDatabaseTypeFromEnv determines the database type from environment variables
func GetDatabaseTypeFromEnv() DatabaseType {
	dbType := os.Getenv("DATABASE_TYPE")
	switch dbType {
	case "mysql":
		return DatabaseTypeMySQL
	case "sqlite":
		return DatabaseTypeSQLite
	default:
		// Default to SQLite for new installations
		return DatabaseTypeSQLite
	}
}

// GetDefaultDSN returns the default DSN for the given database type
func GetDefaultDSN(dbType DatabaseType) string {
	switch dbType {
	case DatabaseTypeMySQL:
		return "root:root@/freelance_tracker?parseTime=true"
	case DatabaseTypeSQLite:
		return "./freelance_tracker.db"
	default:
		return "./freelance_tracker.db"
	}
}