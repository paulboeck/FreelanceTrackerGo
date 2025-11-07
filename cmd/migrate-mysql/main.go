package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

// Config holds database connection strings
type Config struct {
	MySQLDSN  string
	SQLiteDSN string
	DryRun    bool
	Verbose   bool
}

func main() {
	cfg := Config{}

	flag.StringVar(&cfg.MySQLDSN, "mysql", "", "MySQL DSN (e.g., user:password@tcp(localhost:3306)/dbname)")
	flag.StringVar(&cfg.SQLiteDSN, "sqlite", "./freelance_tracker.db", "SQLite database file path")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Perform a dry run without writing to SQLite")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	if cfg.MySQLDSN == "" {
		fmt.Println("MySQL to SQLite Migration Tool")
		fmt.Println("===============================")
		fmt.Println()
		fmt.Println("Usage: migrate-mysql -mysql <mysql-dsn> [-sqlite <sqlite-path>] [-dry-run] [-verbose]")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  migrate-mysql -mysql 'user:password@tcp(localhost:3306)/freelance_db' -sqlite ./freelance_tracker.db")
		fmt.Println()
		fmt.Println("MySQL DSN format:")
		fmt.Println("  user:password@tcp(host:port)/database?parseTime=true")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  -mysql string")
		fmt.Println("        MySQL DSN connection string (required)")
		fmt.Println("  -sqlite string")
		fmt.Println("        SQLite database file path (default: ./freelance_tracker.db)")
		fmt.Println("  -dry-run")
		fmt.Println("        Test migration without writing to SQLite")
		fmt.Println("  -verbose")
		fmt.Println("        Show detailed progress information")
		os.Exit(1)
	}

	// Ensure parseTime=true is in MySQL DSN
	if !strings.Contains(cfg.MySQLDSN, "parseTime=true") {
		if strings.Contains(cfg.MySQLDSN, "?") {
			cfg.MySQLDSN += "&parseTime=true"
		} else {
			cfg.MySQLDSN += "?parseTime=true"
		}
	}

	if err := migrate(cfg); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("\n✓ Migration completed successfully!")
}

// TableMapping defines MySQL to SQLite table name mappings
type TableMapping struct {
	MySQLTable     string
	SQLiteTable    string
	ColumnMapping  map[string]string      // MySQL column name -> SQLite column name
	ExcludeColumns []string               // MySQL columns to skip during migration
	DefaultValues  map[string]interface{} // SQLite column -> default value (for columns not in MySQL)
	ForeignKeys    map[string]string      // SQLite FK column -> parent SQLite table name (e.g., "client_id" -> "client")
	StatusColumn   string                 // MySQL column name that contains status_id (will be converted to status name string)
}

func migrate(cfg Config) error {
	// Connect to MySQL
	fmt.Println("Connecting to MySQL database...")
	mysqlDB, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer mysqlDB.Close()

	if err := mysqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}
	fmt.Println("✓ Connected to MySQL database")

	// Connect to SQLite (only if not dry run)
	var sqliteDB *sql.DB
	if !cfg.DryRun {
		fmt.Printf("Connecting to SQLite database: %s\n", cfg.SQLiteDSN)

		// Add query parameters to DSN for better concurrency handling
		dsn := cfg.SQLiteDSN
		if !strings.Contains(dsn, "?") {
			dsn += "?_busy_timeout=10000&_journal_mode=WAL&_synchronous=NORMAL"
		}

		sqliteDB, err = sql.Open("sqlite", dsn)
		if err != nil {
			return fmt.Errorf("failed to connect to SQLite: %w", err)
		}
		defer sqliteDB.Close()

		// Set connection pool to 1 to avoid "database is locked" errors
		sqliteDB.SetMaxOpenConns(1)

		if err := sqliteDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping SQLite: %w", err)
		}

		// Configure SQLite for better performance and concurrency
		pragmas := []string{
			"PRAGMA foreign_keys = OFF;",   // Disable FK constraints during migration
			"PRAGMA busy_timeout = 10000;", // Wait up to 10 seconds for locks
			"PRAGMA journal_mode = WAL;",   // Enable Write-Ahead Logging for better concurrency
			"PRAGMA synchronous = NORMAL;", // Balance between safety and performance
			"PRAGMA cache_size = -64000;",  // Use 64MB cache
			"PRAGMA temp_store = MEMORY;",  // Store temp tables in memory
		}

		for _, pragma := range pragmas {
			if _, err := sqliteDB.Exec(pragma); err != nil {
				fmt.Printf("  ⚠ Warning: %s failed: %v\n", pragma, err)
			}
		}

		fmt.Println("✓ Connected to SQLite database")
		fmt.Println("  ℹ Foreign key constraints disabled for migration")
		fmt.Println("  ℹ WAL mode enabled, busy timeout set to 10 seconds")
	} else {
		fmt.Println("ℹ DRY RUN MODE - No data will be written to SQLite")
	}

	// Define table mappings: MySQL table name -> SQLite table name
	// Migrate in order to respect foreign key relationships
	// Column mappings handle MySQL column names that differ from SQLite
	// ID mapping will be used to convert MySQL string IDs to SQLite integer IDs
	idMappings := make(map[string]map[string]int) // table name -> (MySQL ID -> SQLite ID)

	// Load project status mappings from MySQL
	// In MySQL: ftapp_project.status_id -> ftapp_projectstatus.id
	// In SQLite: project.status is a string
	statusMappings, err := loadStatusMappings(mysqlDB)
	if err != nil {
		fmt.Printf("  ⚠ Warning: Could not load status mappings: %v\n", err)
		fmt.Println("  Status values will not be converted (will use status_id as string)")
		statusMappings = make(map[string]string) // Empty map to avoid nil errors
	} else if cfg.Verbose {
		fmt.Printf("  Loaded %d status mappings\n", len(statusMappings))
	}

	tableMappings := []TableMapping{
		{
			MySQLTable:  "ftapp_client",
			SQLiteTable: "client",
			ColumnMapping: map[string]string{
				"created_date":       "created_at",
				"last_modified_date": "updated_at",
				"addr_1":             "address1",
				"addr_2":             "address2",
				"addr_3":             "address3",
				"zip":                "zip_code",
				"additional_info_2":  "additional_info2",
				"invoice_cc_desc":    "invoice_cc_description",
			},
			ExcludeColumns: []string{"imported_project_id", "first_name", "last_name", "client_type_id", "org_id", "imported_id"},
			ForeignKeys:    map[string]string{}, // No foreign keys
		},
		{
			MySQLTable:  "ftapp_project",
			SQLiteTable: "project",
			ColumnMapping: map[string]string{
				"status_id":          "status", // status_id (FK) -> status (string name)
				"invoice_cc":         "invoice_cc_email",
				"invoice_cc_desc":    "invoice_cc_description",
				"additional_info_2":  "additional_info2",
				"created_date":       "created_at",
				"last_modified_date": "updated_at",
			},
			ExcludeColumns: []string{"imported_id", "imported_project_id", "imported_client_id"},
			ForeignKeys: map[string]string{
				"client_id": "client", // project.client_id references client.id
			},
			StatusColumn: "status_id", // This column needs special handling to convert ID to name
		},
		{
			MySQLTable:  "ftapp_timesheet",
			SQLiteTable: "timesheet",
			ColumnMapping: map[string]string{
				"hours":              "hours_worked",
				"created_date":       "created_at",
				"last_modified_date": "updated_at",
			},
			ExcludeColumns: []string{"imported_project_id"},
			ForeignKeys: map[string]string{
				"project_id": "project", // timesheet.project_id references project.id
			},
		},
		{
			MySQLTable:  "ftapp_invoice",
			SQLiteTable: "invoice",
			ColumnMapping: map[string]string{
				"created_date":       "created_at",
				"last_modified_date": "updated_at",
			},
			ExcludeColumns: []string{"imported_project_id"},
			DefaultValues: map[string]interface{}{
				"display_details": false, // This column doesn't exist in MySQL
				"invoice_num":     nil,   // This column doesn't exist in MySQL
			},
			ForeignKeys: map[string]string{
				"project_id": "project", // invoice.project_id references project.id
			},
		},
	}

	for _, mapping := range tableMappings {
		if err := migrateTable(mysqlDB, sqliteDB, mapping, cfg, idMappings, statusMappings); err != nil {
			return fmt.Errorf("failed to migrate table %s -> %s: %w", mapping.MySQLTable, mapping.SQLiteTable, err)
		}
	}

	// Re-enable foreign key constraints after migration
	if !cfg.DryRun && sqliteDB != nil {
		_, err = sqliteDB.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			fmt.Printf("  ⚠ Warning: Failed to re-enable foreign keys: %v\n", err)
		} else {
			fmt.Println("\n✓ Foreign key constraints re-enabled")
		}
	}

	return nil
}

func migrateTable(mysql, sqlite *sql.DB, mapping TableMapping, cfg Config, idMappings map[string]map[string]int, statusMappings map[string]string) error {
	fmt.Printf("\nMigrating table: %s -> %s\n", mapping.MySQLTable, mapping.SQLiteTable)

	// Initialize ID mapping for this table
	if idMappings[mapping.SQLiteTable] == nil {
		idMappings[mapping.SQLiteTable] = make(map[string]int)
	}

	// Find the maximum existing ID in SQLite table (if not dry run)
	maxExistingID := 0
	if !cfg.DryRun && sqlite != nil {
		var maxID sql.NullInt64
		err := sqlite.QueryRow(fmt.Sprintf("SELECT MAX(id) FROM %s", mapping.SQLiteTable)).Scan(&maxID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get max ID from SQLite: %w", err)
		}
		if maxID.Valid {
			maxExistingID = int(maxID.Int64)
			fmt.Printf("  ⚠ SQLite table already contains data (max ID: %d). New IDs will start from %d\n", maxExistingID, maxExistingID+1)

			// Check for existing row count
			var existingCount int
			if err := sqlite.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", mapping.SQLiteTable)).Scan(&existingCount); err == nil {
				fmt.Printf("  ⚠ Existing rows in SQLite: %d. Migration will append new data.\n", existingCount)
			}
		}
	}

	// Check if table exists in MySQL
	var exists bool
	err := mysql.QueryRow("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", mapping.MySQLTable).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !exists {
		fmt.Printf("  ⚠ Table %s does not exist in MySQL, skipping\n", mapping.MySQLTable)
		return nil
	}

	// Count rows in MySQL
	var count int
	if err := mysql.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", mapping.MySQLTable)).Scan(&count); err != nil {
		return fmt.Errorf("failed to count rows: %w", err)
	}
	fmt.Printf("  Found %d rows in MySQL table %s\n", count, mapping.MySQLTable)

	if count == 0 {
		fmt.Printf("  ℹ No data to migrate\n")
		return nil
	}

	// Get column names from MySQL
	allMysqlColumns, err := getColumns(mysql, mapping.MySQLTable)
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Find the ID column index in the original MySQL columns
	idColumnIndex := -1
	for i, col := range allMysqlColumns {
		if col == "id" {
			idColumnIndex = i
			break
		}
	}
	if idColumnIndex == -1 {
		return fmt.Errorf("could not find 'id' column in table %s", mapping.MySQLTable)
	}

	// Filter out excluded columns (but keep ID for now - we'll handle it specially)
	var mysqlColumns []string
	var columnIndices []int // Track which indices from original columns to keep
	for i, col := range allMysqlColumns {
		excluded := false
		for _, excludeCol := range mapping.ExcludeColumns {
			if col == excludeCol {
				excluded = true
				break
			}
		}
		if !excluded {
			mysqlColumns = append(mysqlColumns, col)
			columnIndices = append(columnIndices, i)
		}
	}

	// Map MySQL column names to SQLite column names
	sqliteColumns := make([]string, len(mysqlColumns))
	for i, mysqlCol := range mysqlColumns {
		if sqliteCol, ok := mapping.ColumnMapping[mysqlCol]; ok {
			sqliteColumns[i] = sqliteCol
		} else {
			// If no mapping exists, use the MySQL column name as-is
			sqliteColumns[i] = mysqlCol
		}
	}

	// Track which columns are foreign keys (using the SQLite column names)
	// Map: filtered column index -> parent table name
	fkColumns := make(map[int]string)
	for i, sqliteCol := range sqliteColumns {
		if parentTable, isFk := mapping.ForeignKeys[sqliteCol]; isFk {
			fkColumns[i] = parentTable
		}
	}

	// Track which column is the status column (needs status_id -> status_name conversion)
	statusColumnIndex := -1
	if mapping.StatusColumn != "" {
		for i, mysqlCol := range mysqlColumns {
			if mysqlCol == mapping.StatusColumn {
				statusColumnIndex = i
				break
			}
		}
	}

	// Find where the ID column is in our filtered columns list
	idColumnIndexInFiltered := -1
	for i, idx := range columnIndices {
		if idx == idColumnIndex {
			idColumnIndexInFiltered = i
			break
		}
	}

	// Add columns for default values (columns that exist in SQLite but not in MySQL)
	var defaultValueColumns []string
	var defaultValues []interface{}
	for sqliteCol, defaultVal := range mapping.DefaultValues {
		defaultValueColumns = append(defaultValueColumns, sqliteCol)
		defaultValues = append(defaultValues, defaultVal)
	}
	sqliteColumns = append(sqliteColumns, defaultValueColumns...)

	if cfg.Verbose {
		if len(mapping.ExcludeColumns) > 0 {
			fmt.Printf("  Excluded columns: %v\n", mapping.ExcludeColumns)
		}
		fmt.Printf("  MySQL columns: %v\n", mysqlColumns)
		fmt.Printf("  SQLite columns: %v\n", sqliteColumns)

		// Show explicit column mappings for debugging
		if len(mapping.ColumnMapping) > 0 {
			fmt.Println("  Column mappings applied:")
			for mysqlCol, sqliteCol := range mapping.ColumnMapping {
				// Check if this MySQL column exists in the filtered list
				found := false
				for _, col := range mysqlColumns {
					if col == mysqlCol {
						found = true
						break
					}
				}
				if found {
					fmt.Printf("    %s → %s ✓\n", mysqlCol, sqliteCol)
				} else {
					fmt.Printf("    %s → %s ⚠ (MySQL column not found)\n", mysqlCol, sqliteCol)
				}
			}
		}

		if len(defaultValueColumns) > 0 {
			fmt.Printf("  Default value columns: %v\n", defaultValueColumns)
		}
	}

	// Read all data from MySQL
	query := fmt.Sprintf("SELECT * FROM %s", mapping.MySQLTable)
	rows, err := mysql.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query MySQL: %w", err)
	}
	defer rows.Close()

	// Get column types
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("failed to get column types: %w", err)
	}

	inserted := 0
	skipped := 0
	newID := maxExistingID + 1 // Sequential ID counter for SQLite, starting after existing IDs

	for rows.Next() {
		// Create slice of interface{} to hold ALL column values (including excluded ones)
		allValues := make([]interface{}, len(allMysqlColumns))
		allValuePtrs := make([]interface{}, len(allMysqlColumns))
		for i := range allValues {
			allValuePtrs[i] = &allValues[i]
		}

		if err := rows.Scan(allValuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Extract the old MySQL ID (string)
		oldID := convertValueToString(allValues[idColumnIndex])
		if oldID == "" {
			if cfg.Verbose {
				fmt.Printf("  ⚠ Skipped row with empty ID\n")
			}
			skipped++
			continue
		}

		// Store the ID mapping (old MySQL string ID -> new SQLite int ID)
		idMappings[mapping.SQLiteTable][oldID] = newID

		// Filter to keep only non-excluded columns and convert values
		values := make([]interface{}, len(mysqlColumns))
		for i, idx := range columnIndices {
			// Replace ID column with new sequential ID
			if i == idColumnIndexInFiltered {
				values[i] = newID
			} else if i == statusColumnIndex {
				// Convert status_id to status name string
				statusIDValue := convertValueToString(allValues[idx])
				if statusIDValue != "" {
					if statusName, ok := statusMappings[statusIDValue]; ok {
						values[i] = statusName
					} else {
						// Status ID not found in mapping, use "Estimating" as default
						if cfg.Verbose {
							fmt.Printf("  ⚠ Status ID '%s' not found in mapping, using 'Estimating' as default\n", statusIDValue)
						}
						values[i] = "Estimating"
					}
				} else {
					// NULL status, use "Estimating" as default
					values[i] = "Estimating"
				}
			} else if parentTable, isFk := fkColumns[i]; isFk {
				// Convert foreign key value using parent table's ID mapping
				oldFkValue := convertValueToString(allValues[idx])
				if oldFkValue != "" {
					if newFkValue, ok := idMappings[parentTable][oldFkValue]; ok {
						values[i] = newFkValue
					} else {
						if cfg.Verbose {
							fmt.Printf("  ⚠ Skipped row: FK reference not found (table=%s, oldFK=%s)\n", parentTable, oldFkValue)
						}
						skipped++
						continue
					}
				} else {
					// NULL foreign key
					values[i] = nil
				}
			} else {
				values[i] = convertValue(allValues[idx], columnTypes[idx].DatabaseTypeName())
			}
		}

		// Handle invoice_num special case: if this is the invoice table and invoice_num is in default values,
		// set it to the new ID instead of nil
		finalDefaultValues := make([]interface{}, len(defaultValues))
		copy(finalDefaultValues, defaultValues)
		for i, col := range defaultValueColumns {
			if col == "invoice_num" {
				finalDefaultValues[i] = newID
			}
		}

		// Append default values for columns that don't exist in MySQL
		values = append(values, finalDefaultValues...)

		if !cfg.DryRun {
			// Insert into SQLite using the mapped table name and column names
			if err := insertRow(sqlite, mapping.SQLiteTable, sqliteColumns, values); err != nil {
				// Always show insert errors, not just in verbose mode, as they indicate data issues
				fmt.Printf("  ⚠ Skipped row (ID would be %d): %v\n", newID, err)
				if cfg.Verbose {
					// In verbose mode, show the column names to help debug
					fmt.Printf("    Columns: %v\n", sqliteColumns)
				}
				skipped++
				continue
			}
		}

		inserted++
		newID++ // Increment for next row

		if cfg.Verbose && inserted%100 == 0 {
			fmt.Printf("  Migrated %d rows...\n", inserted)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	if cfg.DryRun {
		fmt.Printf("  ✓ Would migrate %d rows to %s", inserted, mapping.SQLiteTable)
	} else {
		fmt.Printf("  ✓ Migrated %d rows to %s", inserted, mapping.SQLiteTable)
	}
	if skipped > 0 {
		fmt.Printf(" (skipped %d)", skipped)
	}
	fmt.Println()

	return nil
}

func getColumns(db *sql.DB, tableName string) ([]string, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 1", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	return columns, nil
}

func convertValue(v interface{}, mysqlType string) interface{} {
	// Handle NULL values
	if v == nil {
		return nil
	}

	// Convert byte arrays to strings (common for TEXT, VARCHAR, etc.)
	if b, ok := v.([]byte); ok {
		return string(b)
	}

	// MySQL DATETIME to SQLite compatible format
	if t, ok := v.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}

	return v
}

func convertValueToString(v interface{}) string {
	// Handle NULL values
	if v == nil {
		return ""
	}

	// Convert byte arrays to strings
	if b, ok := v.([]byte); ok {
		return string(b)
	}

	// Convert to string using fmt.Sprintf
	return fmt.Sprintf("%v", v)
}

func loadStatusMappings(db *sql.DB) (map[string]string, error) {
	// Query the ftapp_projectstatus table to get status_id -> status_name mappings
	query := "SELECT id, name FROM ftapp_projectstatus"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query status table: %w", err)
	}
	defer rows.Close()

	mappings := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("failed to scan status row: %w", err)
		}
		mappings[id] = name
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating status rows: %w", err)
	}

	return mappings, nil
}

func insertRow(db *sql.DB, tableName string, columns []string, values []interface{}) error {
	// Build INSERT statement
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	columnNames := strings.Join(columns, ", ")
	placeholderStr := strings.Join(placeholders, ", ")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		columnNames,
		placeholderStr)

	_, err := db.Exec(query, values...)
	return err
}
