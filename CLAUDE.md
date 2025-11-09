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

# Clean up any orphaned processes after testing
pkill -f "web -addr"
```

## Architecture Details

### Application Structure
The main application struct in `cmd/web/main.go` contains:
- `logger`: Structured logging using slog
- `clients`: Client database model interface
- `projects`: Project database model interface
- `timesheets`: Timesheet database model interface
- `invoices`: Invoice database model interface
- `settings`: Application settings database model interface
- `templateCache`: Pre-compiled HTML templates  
- `formDecoder`: Form data decoder for POST requests
- `sessionManager`: Session management using SCS with SQLite store

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
- always remember to disable browser interference from native HTML5 validation
- check for and remove unused imports

## REST API

FreelanceTrackerGo includes a comprehensive REST API for programmatic access to all features. The API is designed for building integrations, automation, and conversational interfaces (like MCP servers).

### API Features
- **Versioned API**: All endpoints under `/api/v1/` for future compatibility
- **API Key Authentication**: Secure bearer token authentication with `ftk_*` format keys
- **OAuth 2.0-style Scopes**: Granular permissions (e.g., `clients:read`, `invoices:write`)
- **Rate Limiting**: 100 requests per minute per API key (token bucket algorithm)
- **CORS Support**: Configurable for browser-based clients
- **Standardized Responses**: Consistent JSON format with metadata and error handling
- **Full CRUD**: Complete create, read, update, delete operations for all resources
- **Pagination & Search**: List endpoints support pagination and search queries
- **PDF Generation**: Generate invoice PDFs programmatically
- **Email Delivery**: Send invoices via email through the API

### API Authentication

#### Getting an API Key

**Default Credentials** (created on first run):
- Email: `admin@example.com`
- Password: `changeme`

**Request an API Key:**
```bash
curl -X POST http://localhost:4000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "changeme",
    "name": "My API Key",
    "scopes": "*"
  }'
```

**Response:**
```json
{
  "data": {
    "apiKey": "ftk_xxxxxxxxxxxxxxxxxxxxxxxxxx",
    "keyId": 1,
    "userId": 1,
    "message": "API key created successfully. Store this key securely, it will not be shown again."
  },
  "meta": {
    "timestamp": "2025-11-09T10:30:00Z"
  }
}
```

⚠️ **Important**: Save the API key immediately - it's only shown once!

#### Using the API Key

Include the API key in the `Authorization` header of all requests:

```bash
curl http://localhost:4000/api/v1/clients \
  -H "Authorization: Bearer ftk_xxxxxxxxxxxxxxxxxxxxxxxxxx"
```

### Available Scopes

API keys use OAuth 2.0-style scopes for granular permissions:

- `*` - Full access to all resources
- `clients:read`, `clients:write`, `clients:*` - Client operations
- `projects:read`, `projects:write`, `projects:*` - Project operations
- `timesheets:read`, `timesheets:write`, `timesheets:*` - Timesheet operations
- `invoices:read`, `invoices:write`, `invoices:*` - Invoice operations
- `reports:read` - Report access
- `settings:read`, `settings:write`, `settings:*` - Settings management
- `apikeys:read`, `apikeys:write`, `apikeys:*` - API key management

**Example with specific scopes:**
```json
{
  "scopes": "clients:read projects:read invoices:write"
}
```

### API Endpoints Overview

**Authentication** (4 endpoints):
- `POST /api/v1/auth/login` - Get API key (no auth required)
- `POST /api/v1/auth/apikeys` - Create API key
- `GET /api/v1/auth/apikeys` - List user's API keys
- `DELETE /api/v1/auth/apikeys/:id` - Revoke API key

**Clients** (7 endpoints):
- `GET /api/v1/clients` - List with pagination & search
- `GET /api/v1/clients/:id` - Get single client
- `POST /api/v1/clients` - Create client
- `PUT /api/v1/clients/:id` - Update client
- `DELETE /api/v1/clients/:id` - Delete client
- `GET /api/v1/clients/:id/projects` - Get client's projects
- `GET /api/v1/clients/:id/hourlyrate` - Get hourly rate

**Projects** (8 endpoints):
- `GET /api/v1/projects` - List with pagination & search
- `GET /api/v1/projects/:id` - Get single project
- `POST /api/v1/projects` - Create project
- `PUT /api/v1/projects/:id` - Update project
- `PATCH /api/v1/projects/:id/status` - Update status only
- `DELETE /api/v1/projects/:id` - Delete project
- `GET /api/v1/projects/:id/timesheets` - Get project timesheets
- `GET /api/v1/projects/:id/invoices` - Get project invoices

**Timesheets** (4 endpoints):
- `GET /api/v1/timesheets/:id` - Get single timesheet
- `POST /api/v1/projects/:id/timesheets` - Create timesheet for project
- `PUT /api/v1/timesheets/:id` - Update timesheet
- `DELETE /api/v1/timesheets/:id` - Delete timesheet

**Invoices** (7 endpoints):
- `GET /api/v1/invoices/:id` - Get single invoice
- `POST /api/v1/projects/:id/invoices` - Create invoice for project
- `PUT /api/v1/invoices/:id` - Update invoice
- `DELETE /api/v1/invoices/:id` - Delete invoice
- `GET /api/v1/invoices/:id/pdf` - Generate PDF (returns binary)
- `POST /api/v1/invoices/:id/email` - Email invoice to client

**Reports** (1 endpoint):
- `GET /api/v1/reports/income?year=2024` - Income report by year

**Settings** (4 endpoints):
- `GET /api/v1/settings` - Get all settings
- `GET /api/v1/settings/:key` - Get single setting
- `PUT /api/v1/settings` - Update multiple settings
- `PUT /api/v1/settings/:key` - Update single setting

### API Examples

#### List Clients with Pagination
```bash
curl "http://localhost:4000/api/v1/clients?page=1&pageSize=20" \
  -H "Authorization: Bearer ftk_xxx"
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Acme Corp",
      "email": "contact@acme.com",
      "hourlyRate": 150.00,
      "created": "2025-01-01T00:00:00Z",
      "updated": "2025-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 1,
    "totalPages": 1,
    "timestamp": "2025-11-09T10:30:00Z"
  }
}
```

#### Create a Client
```bash
curl -X POST http://localhost:4000/api/v1/clients \
  -H "Authorization: Bearer ftk_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "email": "contact@acme.com",
    "hourlyRate": 150.00,
    "phone": "555-1234",
    "notes": "Important client"
  }'
```

#### Create a Timesheet
```bash
curl -X POST http://localhost:4000/api/v1/projects/1/timesheets \
  -H "Authorization: Bearer ftk_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "workDate": "2025-01-15",
    "hoursWorked": 5.5,
    "hourlyRate": 150.00,
    "description": "Bug fixes and feature implementation"
  }'
```

#### Generate Invoice PDF
```bash
curl http://localhost:4000/api/v1/invoices/1/pdf \
  -H "Authorization: Bearer ftk_xxx" \
  --output invoice.pdf
```

#### Email Invoice
```bash
curl -X POST http://localhost:4000/api/v1/invoices/1/email \
  -H "Authorization: Bearer ftk_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "client@example.com",
    "subject": "Invoice for Project X",
    "body": "Please find your invoice attached."
  }'
```

#### Search Clients
```bash
curl "http://localhost:4000/api/v1/clients?search=acme" \
  -H "Authorization: Bearer ftk_xxx"
```

#### Get Income Report
```bash
curl "http://localhost:4000/api/v1/reports/income?year=2024" \
  -H "Authorization: Bearer ftk_xxx"
```

**Response:**
```json
{
  "data": {
    "year": 2024,
    "total": 45000.00,
    "months": [
      {"month": 1, "amount": 5000.00},
      {"month": 2, "amount": 4500.00},
      ...
    ]
  }
}
```

### Error Handling

All API errors follow a consistent format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "email": "Invalid email format",
      "hourlyRate": "Must be greater than 0"
    }
  }
}
```

**Common Error Codes:**
- `VALIDATION_ERROR` (422) - Invalid input data
- `UNAUTHORIZED` (401) - Missing or invalid API key
- `FORBIDDEN` (403) - Insufficient permissions (scopes)
- `NOT_FOUND` (404) - Resource not found
- `RATE_LIMIT_EXCEEDED` (429) - Too many requests
- `INTERNAL_ERROR` (500) - Server error

### Rate Limiting

The API enforces rate limits to prevent abuse:

- **Limit**: 100 requests per minute per API key
- **Algorithm**: Token bucket with automatic refill
- **Response**: `429 Too Many Requests` when exceeded
- **Header**: `Retry-After` header indicates when to retry

### Testing the API

```bash
# Run API tests
go test ./internal/api -v

# Run all tests including API
go test ./...
```

### API Implementation

**Location**: `internal/api/`
- `response.go` - JSON response helpers
- `errors.go` - Error handling
- `middleware.go` - API key authentication
- `ratelimit.go` - Rate limiting (token bucket)
- `cors.go` - CORS support
- `scopes.go` - Scope-based authorization
- `apikeys.go` - API key management endpoints
- `clients.go` - Client endpoints
- `projects_full.go` - Project endpoints
- `timesheets.go` - Timesheet endpoints
- `invoices.go` - Invoice endpoints (with PDF/email)
- `settings.go` - Settings endpoints
- `reports.go` - Report endpoints

**Routes**: `cmd/web/api_routes.go` - All API routes with middleware chains

### Future: MCP Server

The REST API is designed to power an MCP (Model Context Protocol) server that will enable conversational interaction with FreelanceTracker through Claude. The MCP server will:
- Connect to the deployed FreelanceTracker instance
- Authenticate using API keys
- Transform natural language requests into API calls
- Enable interactions like "Show me all active projects" or "Create a timesheet for Project X"

See `docs/rest_api_mcp_context.md` for complete API architecture and MCP server plans.