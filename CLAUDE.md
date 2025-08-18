# CLAUDE.md

## Workflow
1. First think through the problem, read the codebase for relevant files, and give me a plan.
2. The plan must have a list of todo items that you can check off as you complete them. When making code changes, always consider updating existing automated tests where needed and/or adding new automated tests.
3. Before you begin working, check in with me and I will verify the plan.
4. Then, begin working on the todo items, marking them as complete as you go.
5. Please give me a high level explanation of what changes you made with every step of the plan.
6. You must make every task and code change you make as simple as possible. I want to avaoid making any massive or complex changes. Every change should impact as little code as possible. Everything is about simplicity.
7. Always execute automated tests to validate any code change that you make.

## Project Overview

FreelanceTrackerGo is a Go web application for tracking freelance clients, projects, timesheets, and invoices. It follows a clean MVC architecture with:

- **Web Application**: Located in `cmd/web/` - contains the main HTTP server, handlers, routes, and middleware
- **Models**: Located in `internal/models/` - contains database models and business logic 
- **UI**: Located in `ui/` - contains HTML templates and static assets
- **Validation**: Located in `internal/validator/` - contains form validation utilities

## Development Commands

### Running the Application
```bash
# Run with SQLite (default)
go run ./cmd/web

# Run with SQLite on custom port and database file
go run ./cmd/web -addr=":8081" -dsn="./my_database.db"
```

### Database Migrations
```bash
# Migrations run automatically on startup, but you can also run manually:
goose -dir migrations sqlite3 ./freelance_tracker.db up
```

### Code Generation
```bash
# Generate type-safe Go code from SQL queries (run when queries change)
sqlc generate

# Or using go run
go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
```

### Building
```bash
# Build the web application
go build ./cmd/web

# Build and install dependencies
go mod tidy
```

### Testing
```bash
# Run all tests (uses SQLite)
go test ./...

# Run with verbose output
go test -v ./...

# Test specific package
go test ./internal/models -v
```

## Architecture Details

### Application Structure
The main application struct in `cmd/web/main.go` contains:
- `logger`: Structured logging using slog
- `clients`: SQLite database model using SQLC-generated code
- `templateCache`: Pre-compiled HTML templates  
- `formDecoder`: Form data decoder for POST requests

### Database Layer
**SQLite**:
- Uses `modernc.org/sqlite` (CGO-free) driver
- SQLC-generated type-safe queries in `internal/db/`
- Automatic migrations via Goose
- Single-file database for easy deployment
- Client model implementation in `internal/models/clients.go`

### Modern Code Generation
**Migrations**: 
- Located in `migrations/` directory
- Managed by Goose migration framework
- Automatic execution on application startup

**SQLC Integration**:
- SQL queries in `queries/` directory  
- Type-safe Go code generated in `internal/db/`
- Configuration in `sqlc.yaml`
- Run `sqlc generate` after query changes
