# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Standard Workflow
1. First think through the problem, read the codebase for relevant files, and write a plan to todo.md.
2. The plan should have a list of todo items that you can check off as you complete them.
3. Before you begin working, check in with me and I will verify the plan.
4. Then, begin working on the todo items, marking them as complete as you go.
5. Please give me a high level explanation of what changes you made with every step of the plan.
6. Make every task and code change you make as simple as possible. I want to avaoid making any massive or complex changes. Every change should impact as little code as possible. Everything is about simplicity.
7. Finally, add a review section to the todo.md file with a summary of the changes you made and any other relevant information.

## Project Overview

FreelanceTrackerGo is a Go web application for tracking freelance clients. It follows a clean MVC architecture with:

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

# Run with MySQL (set environment variable)
DATABASE_TYPE=mysql go run ./cmd/web -dsn="root:root@/freelance_tracker?parseTime=true"
```

### Database Migrations
```bash
# Migrations run automatically on startup, but you can also run manually:
# For SQLite
goose -dir migrations sqlite3 ./freelance_tracker.db up

# For MySQL  
goose -dir migrations mysql "root:root@/freelance_tracker?parseTime=true" up
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
# Run all tests (uses MySQL via testcontainers)
go test ./...

# Test with MySQL explicitly
DATABASE_TYPE=mysql go test ./...

# Run with verbose output
go test -v ./...

# Test specific package
go test ./internal/models -v
```

## Architecture Details

### Application Structure
The main application struct in `cmd/web/main.go` contains:
- `logger`: Structured logging using slog
- `clients`: Database model interface supporting multiple implementations
- `templateCache`: Pre-compiled HTML templates  
- `formDecoder`: Form data decoder for POST requests

### Database Layer (Dual Support)
**SQLite (Default)**:
- Uses `modernc.org/sqlite` (CGO-free) driver
- SQLC-generated type-safe queries in `internal/db/`
- Automatic migrations via Goose
- Single-file database for easy deployment

**MySQL (Legacy Support)**:  
- Uses `github.com/go-sql-driver/mysql` driver
- Original hand-written models in `internal/models/`
- DSN requires `parseTime=true` for proper time handling

**Configuration**:
- Database type determined by `DATABASE_TYPE` environment variable
- Defaults to SQLite if not specified
- Both implementations satisfy `ClientModelInterface`

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

### Web Layer
- Uses Go 1.22+ native HTTP routing with `http.NewServeMux()`
- Middleware chain using `github.com/justinas/alice`
- Template rendering with caching
- Form validation using custom validator package

### Current Routes
- `GET /` - Home page showing all clients
- `GET /client/view/{id}` - View specific client details  
- `GET /client/create` - Show client creation form
- `POST /client/create` - Process new client creation
- `GET /static/` - Serve static files

### Database Schema
The application supports both SQLite and MySQL with a `client` table containing:
- `id` (integer primary key / auto increment)
- `name` (text/varchar, max 255 chars)
- `created_at` (datetime/timestamp)
- `updated_at` (datetime/timestamp)

## Key Patterns

### Error Handling
- Custom `ErrNoRecord` for missing database records
- Structured error responses with `serverError()` and `clientError()` helpers
- Stack trace logging for server errors

### Form Processing
- Form structs with validation tags (see `clientCreateForm` in `cmd/web/handlers.go:15-18`)
- Server-side validation using the custom validator package
- Form data persisted on validation errors for better UX

### Template Rendering
- Template caching for performance
- Template data structure in `cmd/web/templates.go`
- Separation of page templates and partials in `ui/html/`

## Migration History

This application was successfully migrated from MySQL-only to dual MySQL/SQLite support:

**What Changed**:
- Added SQLite support with modernc.org/sqlite driver (CGO-free)
- Implemented Goose migration framework for database versioning
- Added SQLC for type-safe SQL code generation
- Created adapter pattern for gradual migration between implementations
- Established comprehensive integration test suite using testcontainers

**Current State**:
- ✅ SQLite: Default database using SQLC-generated type-safe code
- ✅ MySQL: Legacy support using original hand-written models  
- ✅ Full test coverage for both implementations
- ✅ Automatic migrations on startup
- ✅ Zero-downtime deployment capability

**Benefits Achieved**:
- **Simplified Deployment**: No external database required for new installations
- **Type Safety**: SQLC eliminates SQL runtime errors
- **Maintainability**: Structured migrations and code generation
- **Testing**: Isolated tests with containerized databases
- **Flexibility**: Easy switching between database types