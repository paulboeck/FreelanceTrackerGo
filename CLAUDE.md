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
# Run with default settings (localhost:8080, default MySQL DSN)
go run ./cmd/web

# Run with custom address and database
go run ./cmd/web -addr=":8081" -dsn="root:root@/freelance_tracker?parseTime=true"
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
# Run tests (when they exist)
go test ./...

# Run tests with verbose output
go test -v ./...
```

## Architecture Details

### Application Structure
The main application struct in `cmd/web/main.go:17-22` contains:
- `logger`: Structured logging using slog
- `clients`: Database model for client operations
- `templateCache`: Pre-compiled HTML templates
- `formDecoder`: Form data decoder for POST requests

### Database Layer
- Uses MySQL with the `go-sql-driver/mysql` driver
- Connection configured via DSN flag with `parseTime=true` for proper time handling
- Models are in `internal/models/` with a repository pattern
- Custom error types defined in `internal/models/errors.go`

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
The application expects a MySQL database named `freelance_tracker` with a `client` table containing:
- `id` (primary key)
- `name` (varchar, max 255 chars)
- `created_at` (timestamp)
- `updated_at` (timestamp)

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