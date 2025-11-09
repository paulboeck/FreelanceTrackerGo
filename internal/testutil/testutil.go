package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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
			email TEXT NOT NULL DEFAULT '',
			phone TEXT,
			address1 TEXT,
			address2 TEXT,
			address3 TEXT,
			city TEXT,
			state TEXT,
			zip_code TEXT,
			hourly_rate DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			notes TEXT,
			additional_info TEXT,
			additional_info2 TEXT,
			bill_to TEXT,
			include_address_on_invoice BOOLEAN NOT NULL DEFAULT 1,
			invoice_cc_email TEXT,
			invoice_cc_description TEXT,
			university_affiliation TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL
		);
		
		CREATE TABLE IF NOT EXISTS project (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			client_id INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'Estimating',
			hourly_rate REAL NOT NULL DEFAULT 0.00,
			deadline TEXT,
			scheduled_start TEXT,
			invoice_cc_email TEXT,
			invoice_cc_description TEXT,
			schedule_comments TEXT,
			additional_info TEXT,
			additional_info2 TEXT,
			discount_percent REAL,
			discount_reason TEXT,
			adjustment_amount REAL,
			adjustment_reason TEXT,
			currency_display TEXT NOT NULL DEFAULT 'USD',
			currency_conversion_rate REAL NOT NULL DEFAULT 1.00000,
			flat_fee_invoice INTEGER NOT NULL DEFAULT 0,
			notes TEXT,
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
			hourly_rate REAL NOT NULL DEFAULT 0.00,
			description VARCHAR(255),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			FOREIGN KEY (project_id) REFERENCES project(id)
		);
		
		CREATE TABLE IF NOT EXISTS invoice (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			invoice_num INTEGER,
			project_id INTEGER NOT NULL,
			invoice_date DATE NOT NULL,
			date_paid DATE NULL,
			payment_terms TEXT NULL,
			amount_due DECIMAL(10,2) NOT NULL,
			display_details BOOLEAN NOT NULL DEFAULT false,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			FOREIGN KEY (project_id) REFERENCES project(id)
		);
		
		CREATE TABLE IF NOT EXISTS setting (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			data_type TEXT NOT NULL CHECK (data_type IN ('string', 'int', 'float', 'decimal', 'bool')),
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS user (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			hashed_password TEXT(60) NOT NULL,
			require_password_change INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL
		);
		
		CREATE INDEX IF NOT EXISTS idx_user_email ON user(email);
		CREATE INDEX IF NOT EXISTS idx_user_deleted_at ON user(deleted_at);
		
		CREATE TABLE IF NOT EXISTS sessions (
			token CHAR(43) PRIMARY KEY,
			data BLOB NOT NULL,
			expiry TIMESTAMP(6) NOT NULL
		);
		
		CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions (expiry);

		CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES user(id),
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL UNIQUE,
			key_prefix TEXT NOT NULL,
			scopes TEXT NOT NULL,
			last_used_at TIMESTAMP,
			created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
		CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
		CREATE INDEX IF NOT EXISTS idx_api_keys_deleted_at ON api_keys(deleted_at);

		INSERT OR IGNORE INTO setting (key, value, data_type, description) VALUES
			('default_hourly_rate', '85.00', 'decimal', 'Default hourly rate for new projects'),
			('invoice_title', 'Invoice for Academic Editing', 'string', 'Title displayed on generated invoices'),
			('freelancer_name', 'Your Name Here', 'string', 'Freelancer name for invoices'),
			('freelancer_address', 'Your Address', 'string', 'Freelancer address for invoices'),
			('freelancer_phone', 'Your Phone', 'string', 'Freelancer phone for invoices'),
			('freelancer_email', 'your.email@example.com', 'string', 'Freelancer email for invoices'),
			('email_enabled', 'false', 'bool', 'Enable email notifications'),
			('smtp_host', '', 'string', 'SMTP server hostname'),
			('smtp_port', '587', 'int', 'SMTP server port'),
			('smtp_username', '', 'string', 'SMTP username'),
			('smtp_password', '', 'string', 'SMTP password (encrypted)'),
			('smtp_from_name', 'Freelance Tracker', 'string', 'Email from name'),
			('smtp_use_tls', 'true', 'bool', 'Use TLS for SMTP connection');
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
	result, err := td.DB.Exec("INSERT INTO client (name, email, hourly_rate) VALUES (?, ?, ?)",
		name, "test@example.com", 50.00)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestClientWithDefaults inserts a test client with default values for all fields and returns its ID
func (td *TestDatabase) InsertTestClientWithDefaults(t *testing.T, name string, hourlyRate float64, additionalInfo, additionalInfo2, invoiceCCEmail, invoiceCCDescription string) int {
	result, err := td.DB.Exec(`INSERT INTO client (name, email, hourly_rate, additional_info, additional_info2, invoice_cc_email, invoice_cc_description) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		name, "test@example.com", hourlyRate, additionalInfo, additionalInfo2, invoiceCCEmail, invoiceCCDescription)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestProject inserts a test project and returns its ID
func (td *TestDatabase) InsertTestProject(t *testing.T, name string, clientID int) int {
	result, err := td.DB.Exec(`INSERT INTO project (name, client_id, status, hourly_rate, currency_display, currency_conversion_rate, flat_fee_invoice) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`, name, clientID, "Estimating", 50.0, "USD", 1.0, 0)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestTimesheet inserts a test timesheet and returns its ID
func (td *TestDatabase) InsertTestTimesheet(t *testing.T, projectID int, workDate, hoursWorked, hourlyRate, description string) int {
	result, err := td.DB.Exec("INSERT INTO timesheet (project_id, work_date, hours_worked, hourly_rate, description) VALUES (?, ?, ?, ?, ?)",
		projectID, workDate, hoursWorked, hourlyRate, description)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestInvoice inserts a test invoice and returns its ID
func (td *TestDatabase) InsertTestInvoice(t *testing.T, projectID int, invoiceDate, datePaid, paymentTerms, amountDue string) int {
	var datePaidParam interface{}
	if datePaid != "" {
		datePaidParam = datePaid
	}

	result, err := td.DB.Exec("INSERT INTO invoice (project_id, invoice_date, date_paid, payment_terms, amount_due, display_details) VALUES (?, ?, ?, ?, ?, ?)",
		projectID, invoiceDate, datePaidParam, paymentTerms, amountDue, false)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestProjectWithRate inserts a test project with specified hourly rate and returns its ID
func (td *TestDatabase) InsertTestProjectWithRate(t *testing.T, clientID int, name string, hourlyRate float64) int {
	result, err := td.DB.Exec(`INSERT INTO project (name, client_id, status, hourly_rate, currency_display, currency_conversion_rate, flat_fee_invoice)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, name, clientID, "Estimating", hourlyRate, "USD", 1.0, 0)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestProjectWithDiscount inserts a test project with discount and returns its ID
func (td *TestDatabase) InsertTestProjectWithDiscount(t *testing.T, clientID int, name string, hourlyRate float64, discountPercent float64, discountReason string) int {
	result, err := td.DB.Exec(`INSERT INTO project (name, client_id, status, hourly_rate, discount_percent, discount_reason, currency_display, currency_conversion_rate, flat_fee_invoice)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, name, clientID, "Estimating", hourlyRate, discountPercent, discountReason, "USD", 1.0, 0)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestTimesheetWithTime inserts a test timesheet with time.Time and float64 parameters and returns its ID
func (td *TestDatabase) InsertTestTimesheetWithTime(t *testing.T, projectID int, workDate time.Time, hoursWorked float64, hourlyRate float64, description string) int {
	result, err := td.DB.Exec("INSERT INTO timesheet (project_id, work_date, hours_worked, hourly_rate, description) VALUES (?, ?, ?, ?, ?)",
		projectID, workDate.Format("2006-01-02"), hoursWorked, hourlyRate, description)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestInvoiceWithTime inserts a test invoice with time.Time parameters and returns its ID
func (td *TestDatabase) InsertTestInvoiceWithTime(t *testing.T, projectID int, invoiceDate time.Time, datePaid *time.Time, paymentTerms string, amountDue float64, displayDetails bool) int {
	var datePaidParam interface{}
	if datePaid != nil {
		datePaidParam = datePaid.Format("2006-01-02")
	}

	result, err := td.DB.Exec("INSERT INTO invoice (project_id, invoice_date, date_paid, payment_terms, amount_due, display_details) VALUES (?, ?, ?, ?, ?, ?)",
		projectID, invoiceDate.Format("2006-01-02"), datePaidParam, paymentTerms, amountDue, displayDetails)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestUser inserts a test user and returns its ID
// The password is hashed using bcrypt before insertion
func (td *TestDatabase) InsertTestUser(t *testing.T, name, email, hashedPassword string) int {
	result, err := td.DB.Exec("INSERT INTO user (name, email, hashed_password) VALUES (?, ?, ?)",
		name, email, hashedPassword)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}

// InsertTestAPIKey inserts a test API key and returns its ID
func (td *TestDatabase) InsertTestAPIKey(t *testing.T, userID int, name, keyHash, keyPrefix, scopes string) int {
	result, err := td.DB.Exec("INSERT INTO api_keys (user_id, name, key_hash, key_prefix, scopes) VALUES (?, ?, ?, ?, ?)",
		userID, name, keyHash, keyPrefix, scopes)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return int(id)
}
