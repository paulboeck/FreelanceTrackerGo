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
- `GET /client/update/{id}` - Show client update form
- `POST /client/update/{id}` - Process client updates
- `GET /static/` - Serve static files

### Database Schema
The application uses SQLite with a `client` table containing:
- `id` (integer primary key autoincrement)
- `name` (text, max 255 chars)
- `created_at` (datetime)
- `updated_at` (datetime)

## Key Patterns

### Error Handling
- Custom `ErrNoRecord` for missing database records
- Structured error responses with `serverError()` and `clientError()` helpers
- Stack trace logging for server errors

### Form Processing
- Form structs with validation tags (see `clientForm` in `cmd/web/handlers.go` - reused for create and update)
- Server-side validation using the custom validator package
- Form data persisted on validation errors for better UX

### Template Rendering
- Template caching for performance
- Template data structure in `cmd/web/templates.go`
- Separation of page templates and partials in `ui/html/`

## Migration History

This application was migrated from MySQL to SQLite-only for simplified deployment:

**What Changed**:
- Replaced MySQL with SQLite using modernc.org/sqlite driver (CGO-free)
- Implemented Goose migration framework for database versioning
- Added SQLC for type-safe SQL code generation
- Removed dual database support for simplicity
- Updated test suite to use in-memory SQLite databases

**Current State**:
- ✅ SQLite: Single database solution using SQLC-generated type-safe code
- ✅ Full test coverage with fast SQLite tests
- ✅ Automatic migrations on startup
- ✅ Zero-dependency deployment (no external database required)
- ✅ Client CRUD operations (Create, Read, Update)

**Benefits Achieved**:
- **Simplified Deployment**: No external database required
- **Type Safety**: SQLC eliminates SQL runtime errors
- **Maintainability**: Single database implementation to maintain
- **Fast Testing**: In-memory SQLite tests run quickly
- **Easy Distribution**: Single binary + database file deployment