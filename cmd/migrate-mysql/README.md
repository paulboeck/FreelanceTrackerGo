# MySQL to SQLite Migration Tool

A command-line tool to migrate your existing MySQL FreelanceTracker database to SQLite.

## Quick Start

```bash
# Build the tool
go build -o migrate-mysql

# Run migration with dry-run first (recommended)
./migrate-mysql \
  -mysql 'user:password@tcp(localhost:3306)/database_name' \
  -sqlite ./freelance_tracker.db \
  -dry-run \
  -verbose

# If dry-run looks good, run actual migration
./migrate-mysql \
  -mysql 'user:password@tcp(localhost:3306)/database_name' \
  -sqlite ./freelance_tracker.db \
  -verbose
```

## Usage

```
migrate-mysql -mysql <mysql-dsn> [-sqlite <sqlite-path>] [-dry-run] [-verbose]
```

### Flags

| Flag | Description | Required | Default |
|------|-------------|----------|---------|
| `-mysql` | MySQL connection string (DSN) | Yes | - |
| `-sqlite` | Path to SQLite database file | No | `./freelance_tracker.db` |
| `-dry-run` | Test migration without writing data | No | `false` |
| `-verbose` | Show detailed progress | No | `false` |

### MySQL DSN Format

```
username:password@tcp(host:port)/database_name
```

**Examples:**

```bash
# Local MySQL
-mysql 'root:password@tcp(localhost:3306)/freelance_db'

# Remote MySQL
-mysql 'user:pass@tcp(192.168.1.100:3306)/freelance_db'

# MySQL with special characters in password (use quotes)
-mysql 'user:p@ssw0rd!@tcp(localhost:3306)/freelance_db'
```

The tool automatically adds `?parseTime=true` to the DSN to handle date/time fields correctly.

## What It Does

1. Connects to your MySQL database
2. Reads data from these MySQL tables and maps them to SQLite tables:
   - `ftapp_client` → `client` - All client records
   - `ftapp_project` → `project` - All project records
   - `ftapp_timesheet` → `timesheet` - All timesheet entries
   - `ftapp_invoice` → `invoice` - All invoice records
3. **Handles ID conversion**: Converts MySQL string IDs to SQLite integer IDs
   - Generates new sequential integer IDs (1, 2, 3...) for SQLite
   - Maintains ID mappings to preserve foreign key relationships
   - Automatically converts foreign key references
3a. **Handles status conversion**: Converts project status foreign keys to status strings
   - Queries `ftapp_projectstatus` table to get status ID → status name mappings
   - Converts `ftapp_project.status_id` (foreign key) to `project.status` (string)
   - Uses "Estimating" as default if status not found
4. Maps MySQL column names to SQLite column names:
   - `created_date` → `created_at` (all tables)
   - `last_modified_date` → `updated_at` (all tables)
   - `status_id` → `status` (project table only - also converts FK ID to status name string)
   - `invoice_cc` → `invoice_cc_email` (project table only)
   - `invoice_cc_desc` → `invoice_cc_description` (project table only)
   - `additional_info_2` → `additional_info2` (project table only)
   - `hours` → `hours_worked` (timesheet table only)
   - All other columns are migrated with their original names
5. Excludes unwanted columns:
   - `imported_project_id` is excluded from all tables
   - `first_name` and `last_name` are excluded from client table
   - `imported_id` and `imported_client_id` are excluded from project table
6. Provides default values for SQLite-only columns:
   - `display_details` (invoice table) → defaults to false
   - `invoice_num` (invoice table) → set to match the new SQLite ID
7. Converts MySQL data types to SQLite-compatible formats
8. Inserts data into SQLite database
9. Reports progress and any issues

**Note:** The `setting` table is not migrated because it doesn't exist in the MySQL schema.

## Migration Process

### Step 1: Backup Your MySQL Database

```bash
mysqldump -u username -p database_name > backup_$(date +%Y%m%d).sql
```

### Step 2: Initialize SQLite Database

The SQLite database must exist with the correct schema before migration:

```bash
# Run the application once to create the database and run migrations
go run ./cmd/web

# Press Ctrl+C after it starts
```

This creates `./freelance_tracker.db` with all the necessary tables.

**IMPORTANT**: If your SQLite database already contains data, you should clear it before migration to avoid ID conflicts:

```bash
# Option 1: Delete and recreate the database
rm ./freelance_tracker.db
go run ./cmd/web  # Recreate with migrations
# Press Ctrl+C after it starts

# Option 2: Clear specific tables (preserves settings and users)
sqlite3 ./freelance_tracker.db << EOF
DELETE FROM invoice;
DELETE FROM timesheet;
DELETE FROM project;
DELETE FROM client;
EOF
```

If you choose not to clear the tables, the migration tool will append new data starting with IDs after the existing maximum ID. This may be useful for merging data, but be aware that this creates a gap in IDs.

### Step 3: Ensure Database is Not in Use

**IMPORTANT:** Before running the migration, ensure the SQLite database is not being used by any other process:

```bash
# Stop the FreelanceTracker application if it's running
# Press Ctrl+C to stop it

# Check for any processes using the database (macOS/Linux)
lsof ./freelance_tracker.db

# Or on macOS:
fuser ./freelance_tracker.db

# If any processes are shown, stop them before proceeding
```

### Step 4: Test with Dry Run

```bash
./migrate-mysql \
  -mysql 'user:password@tcp(localhost:3306)/freelance_db' \
  -sqlite ./freelance_tracker.db \
  -dry-run \
  -verbose
```

This will show:
- How many rows would be migrated
- Any potential issues
- Column information (if verbose)
- Does NOT modify SQLite database

### Step 5: Run Actual Migration

```bash
./migrate-mysql \
  -mysql 'user:password@tcp(localhost:3306)/freelance_db' \
  -sqlite ./freelance_tracker.db \
  -verbose
```

### Step 6: Verify the Migration

```bash
# Check record counts
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM client;"
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM project;"
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM timesheet;"
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM invoice;"

# Start the application and test
go run ./cmd/web
```

Open http://localhost:8080 and verify:
- All clients appear
- Projects are linked correctly
- Timesheets show correct data
- Invoices have correct amounts

## Example Output

```
Connecting to MySQL database...
✓ Connected to MySQL database
Connecting to SQLite database: ./freelance_tracker.db
✓ Connected to SQLite database

Migrating table: ftapp_client -> client
  Found 45 rows in MySQL table ftapp_client
  ✓ Migrated 45 rows to client

Migrating table: ftapp_project -> project
  Found 12 rows in MySQL table ftapp_project
  ✓ Migrated 12 rows to project

Migrating table: ftapp_timesheet -> timesheet
  Found 234 rows in MySQL table ftapp_timesheet
  ✓ Migrated 234 rows to timesheet

Migrating table: ftapp_invoice -> invoice
  Found 18 rows in MySQL table ftapp_invoice
  ✓ Migrated 18 rows to invoice

✓ Migration completed successfully!
```

### Example Verbose Output

With the `-verbose` flag, you can see column mappings:

```
Migrating table: ftapp_client -> client
  Found 45 rows in MySQL table ftapp_client
  MySQL columns: [id name email phone address notes created_date last_modified_date]
  SQLite columns: [id name email phone address notes created_at updated_at]
  ✓ Migrated 45 rows to client

Migrating table: ftapp_timesheet -> timesheet
  Found 234 rows in MySQL table ftapp_timesheet
  MySQL columns: [id project_id work_date hours description created_date last_modified_date]
  SQLite columns: [id project_id work_date hours_worked description created_at updated_at]
  ✓ Migrated 234 rows to timesheet
```

## Data Type Conversions

The tool automatically handles MySQL to SQLite type conversions:

| MySQL Type | SQLite Type | Conversion |
|------------|-------------|------------|
| `VARCHAR(32)` (IDs) | `INTEGER` | New sequential IDs generated (1, 2, 3...) |
| `INT`, `BIGINT` | `INTEGER` | Direct mapping |
| `VARCHAR(n)`, `TEXT` | `TEXT` | Direct mapping |
| `DECIMAL(m,n)` | `REAL` | Direct mapping |
| `DATETIME`, `TIMESTAMP` | `TEXT` | Formatted as ISO8601 |
| `TINYINT(1)` (boolean) | `INTEGER` | 0 or 1 |
| `NULL` | `NULL` | Preserved |

### ID Conversion Details

Since MySQL uses 32-character string UUIDs for IDs and SQLite uses integer IDs, the migration tool:
1. Generates new sequential integer IDs starting from 1 for each table
2. Maintains an internal mapping of old MySQL IDs → new SQLite IDs
3. Automatically converts all foreign key references using the mapping
4. Ensures referential integrity is preserved across tables

Example:
- MySQL client ID: `"abc123def456..."` → SQLite client ID: `1`
- MySQL project with `client_id = "abc123def456..."` → SQLite project with `client_id = 1`

## Migration Order

Tables are migrated in this specific order to respect foreign key relationships:

1. **client** - No dependencies
2. **project** - References client
3. **timesheet** - References project
4. **invoice** - References project
5. **setting** - No dependencies

**Note:** Foreign key constraints are temporarily disabled during migration (`PRAGMA foreign_keys = OFF`) to allow flexible data insertion even if there are reference issues. The migration tool handles FK mapping internally by converting MySQL string IDs to SQLite integer IDs. After successful migration, FK constraints are re-enabled (`PRAGMA foreign_keys = ON`).

## Common Issues and Solutions

### Issue: "database is locked" (SQLITE_BUSY)

**Error:**
```
⚠ Skipped row (ID would be 737): database is locked (5) (SQLITE_BUSY)
```

**Cause:**
Another process has the SQLite database open and has an active lock on it.

**Solutions:**

1. **Stop the FreelanceTracker application:**
   ```bash
   # If running in another terminal, press Ctrl+C to stop it

   # Or find and kill the process
   pkill -f "freelance_tracker"
   pkill -f "go run ./cmd/web"
   ```

2. **Check for any processes using the database:**
   ```bash
   # macOS/Linux
   lsof ./freelance_tracker.db

   # Kill any processes shown
   kill <PID>
   ```

3. **Wait for locks to clear:**
   - The migration tool now waits up to 10 seconds for locks to clear automatically
   - If errors persist after stopping all processes, wait a minute and try again

4. **Check for stale lock files:**
   ```bash
   # Remove SQLite WAL and SHM files if they exist
   rm -f ./freelance_tracker.db-wal
   rm -f ./freelance_tracker.db-shm
   ```

5. **Run migration again:**
   - The migration tool will skip already-migrated rows automatically
   - It will continue from where it left off

### Issue: "Failed to connect to MySQL"

**Error:**
```
failed to connect to MySQL: dial tcp: connection refused
```

**Solutions:**
- Verify MySQL is running: `mysql -u username -p`
- Check host and port are correct
- Ensure MySQL allows connections from localhost
- Test connection: `mysql -h localhost -P 3306 -u username -p`

### Issue: "Table doesn't exist"

**Error:**
```
Table 'client' does not exist in MySQL, skipping
```

**Solutions:**
- Your MySQL database might use different table names
- Check your tables: `mysql -e "SHOW TABLES;" database_name`
- The tool expects tables named: `client`, `project`, `timesheet`, `invoice`, `setting`
- You may need to rename tables or modify the migration tool for custom schema

### Issue: "UNIQUE constraint failed" or Duplicate Primary Key Errors

**Error:**
```
UNIQUE constraint failed: client.id
```
or
```
PRIMARY KEY must be unique
```

**Cause:**
Your SQLite database already has data, and the migration is trying to insert rows with IDs that already exist.

**Solutions:**

**Option 1: Clear the database (Recommended for first-time migration)**
```bash
# Delete database and recreate
rm ./freelance_tracker.db
go run ./cmd/web  # Recreate with migrations
# Press Ctrl+C after it starts
# Then run migration again
```

**Option 2: Clear specific tables (Preserves settings and users)**
```bash
sqlite3 ./freelance_tracker.db << EOF
DELETE FROM invoice;
DELETE FROM timesheet;
DELETE FROM project;
DELETE FROM client;
EOF
# Then run migration again
```

**Option 3: Append to existing data**
The migration tool automatically detects existing data and starts new IDs after the highest existing ID. This works but creates ID gaps:
- Existing: IDs 1-10
- New data: IDs 11-50

This option is useful for merging multiple MySQL databases into one SQLite database.

### Issue: "Access denied for user"

**Error:**
```
failed to ping MySQL: Error 1045: Access denied
```

**Solutions:**
- Verify username and password are correct
- Check user has proper permissions:
  ```sql
  GRANT SELECT ON database_name.* TO 'username'@'localhost';
  FLUSH PRIVILEGES;
  ```
- The migration tool only needs `SELECT` permission on MySQL

### Issue: "Failed to scan row"

**Error:**
```
failed to scan row: sql: Scan error
```

**Solutions:**
- Your MySQL schema might have custom columns
- Run with `-verbose` to see which row fails
- The tool expects the standard FreelanceTracker schema
- You may need to add custom field handling in the code

## Troubleshooting Tips

### Check MySQL Connection

```bash
# Test MySQL connectivity
mysql -h localhost -P 3306 -u username -p database_name -e "SELECT 1;"
```

### Verify MySQL Schema

```bash
# Show table structure
mysql -u username -p database_name << EOF
DESCRIBE client;
DESCRIBE project;
DESCRIBE timesheet;
DESCRIBE invoice;
DESCRIBE setting;
EOF
```

### Check SQLite Schema

```bash
# View SQLite schema
sqlite3 ./freelance_tracker.db << EOF
.schema client
.schema project
.schema timesheet
.schema invoice
.schema setting
EOF
```

### Monitor Progress with Verbose Mode

```bash
./migrate-mysql \
  -mysql 'DSN' \
  -sqlite ./freelance_tracker.db \
  -verbose
```

This shows:
- Column names for each table
- Progress every 100 rows
- Warnings for skipped rows
- Detailed error messages

## Advanced Usage

### Migrate to Different SQLite File

```bash
./migrate-mysql \
  -mysql 'user:pass@tcp(localhost:3306)/freelance_db' \
  -sqlite ./custom_database.db
```

### Test Without Writing (Multiple Times)

```bash
# You can run dry-run multiple times safely
./migrate-mysql -mysql 'DSN' -dry-run -verbose
./migrate-mysql -mysql 'DSN' -dry-run -verbose
# No data written to SQLite
```

### Migrate from Remote MySQL

```bash
./migrate-mysql \
  -mysql 'user:password@tcp(remote-server.com:3306)/freelance_db' \
  -sqlite ./freelance_tracker.db
```

## Security Considerations

### Never Commit Credentials

Add to `.gitignore`:
```
*.db
*.sql
.env
```

### Use Environment Variables

```bash
# Store DSN in environment variable
export MYSQL_DSN='user:password@tcp(localhost:3306)/freelance_db'

# Use in migration
./migrate-mysql -mysql "$MYSQL_DSN" -sqlite ./freelance_tracker.db

# Clear from shell history if needed
history -c
```

### MySQL Permissions

The migration tool only needs `SELECT` permission:

```sql
-- Create read-only user for migration
CREATE USER 'migration'@'localhost' IDENTIFIED BY 'password';
GRANT SELECT ON database_name.* TO 'migration'@'localhost';
FLUSH PRIVILEGES;

-- Use this user for migration
./migrate-mysql -mysql 'migration:password@tcp(localhost:3306)/database_name'
```

## Performance

### Expected Migration Speeds

- **Small databases** (<1,000 rows): ~1-5 seconds
- **Medium databases** (1,000-10,000 rows): ~5-30 seconds
- **Large databases** (10,000-100,000 rows): ~30-300 seconds

Speed depends on:
- Number of rows
- Number of columns
- MySQL server performance
- Disk I/O speed

### Large Database Tips

For very large databases:

1. **Monitor progress with verbose:**
   ```bash
   ./migrate-mysql -mysql 'DSN' -sqlite ./database.db -verbose
   ```

2. **Ensure enough disk space:**
   ```bash
   # SQLite database will be similar size to MySQL data
   du -h /path/to/mysql/data/database_name
   ```

3. **Consider migrating during off-hours** to avoid slowing down your MySQL server

## After Migration

### Verify Data Integrity

```bash
# Compare record counts
mysql -e "SELECT COUNT(*) FROM client;" database_name
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM client;"

mysql -e "SELECT COUNT(*) FROM project;" database_name
sqlite3 ./freelance_tracker.db "SELECT COUNT(*) FROM project;"

# Spot check some records
sqlite3 ./freelance_tracker.db "SELECT * FROM client LIMIT 5;"
```

### Test the Application

```bash
go run ./cmd/web
```

Verify:
- [ ] Login works
- [ ] All clients appear
- [ ] Projects show correct data
- [ ] Timesheets are accurate
- [ ] Invoices calculate correctly
- [ ] Settings are preserved

### Clean Up

Once verified:
```bash
# Keep MySQL backup for 30+ days
# You can decommission the MySQL database when confident
```

## Building from Source

```bash
# Build for your platform
go build -o migrate-mysql ./cmd/migrate-mysql

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o migrate-mysql-linux ./cmd/migrate-mysql
GOOS=darwin GOARCH=arm64 go build -o migrate-mysql-macos ./cmd/migrate-mysql
GOOS=windows GOARCH=amd64 go build -o migrate-mysql.exe ./cmd/migrate-mysql
```

## Getting Help

If you encounter issues:

1. **Run with verbose mode:**
   ```bash
   ./migrate-mysql -mysql 'DSN' -dry-run -verbose
   ```

2. **Check MySQL connectivity:**
   ```bash
   mysql -h localhost -u username -p database_name -e "SELECT 1;"
   ```

3. **Verify table structures match**

4. **Check error messages** - they usually indicate the specific issue

For schema mismatches or custom fields, you may need to modify the tool's source code in `main.go`.
