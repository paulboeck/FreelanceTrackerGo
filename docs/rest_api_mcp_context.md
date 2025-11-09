# REST API & MCP Server Project Context

**Date Started**: 2025-11-09
**Status**: Planning phase - ready to implement

## Project Goal

Create an MCP (Model Context Protocol) server that allows conversational interaction with FreelanceTracker, enabling users to manage clients, projects, timesheets, and invoices through natural language with Claude.

Example interactions:
- "Show me all my active projects"
- "Create a timesheet for Project X: 5 hours today for bug fixes"
- "Generate and email the invoice for client ABC Corp"
- "What's my total income for 2024?"

## Architecture Decision

**Chosen Approach**: Two-part solution
1. **REST API in FreelanceTracker** - Add comprehensive JSON API endpoints
2. **MCP Server as API Client** - Lightweight server that calls the REST API

### Why This Approach?

- ✅ Reuses all existing authentication, validation, and permissions
- ✅ Works with deployed Fly.io instance (no direct database access needed)
- ✅ More secure (goes through existing security layers)
- ✅ Bonus: API can be used by other clients (future React UI, mobile apps, etc.)
- ✅ Separates concerns: FreelanceTracker handles business logic, MCP handles conversational interface

### Rejected Approach

Initially considered having MCP server connect directly to the database and use the model layer, but this wouldn't work with the deployed Fly.io instance and would bypass authentication/validation.

## Implementation Decisions

### 1. API Scope
**Decision**: Comprehensive REST API covering all CRUD operations
- Not just minimal endpoints for MCP
- Future-proof for potential React UI or other clients
- Full feature parity with web interface

### 2. Authentication
**Decision**: API Key authentication (not session-based)
- Format: `ftk_<random_32_chars>` (FreelanceTracker Key)
- Stored as bcrypt hash (like passwords)
- Key prefix stored for identification
- Header: `Authorization: Bearer ftk_abc123...`

### 3. API Route Prefix
**Decision**: `/api/v1/*` prefix for all REST endpoints
- Example: `/api/v1/clients`, `/api/v1/projects/123/timesheets`
- Versioning allows future breaking changes without disrupting existing clients

### 4. Code Organization
**Decision**: Separate package `internal/api/` for REST API code
- Keeps API logic separate from web handlers
- Clean separation of concerns

## Package Structure

```
internal/api/
├── middleware.go      # API key auth middleware
├── ratelimit.go       # Rate limiting middleware (token bucket)
├── cors.go            # CORS middleware
├── scopes.go          # Scope validation and authorization
├── response.go        # Standard JSON response helpers
├── errors.go          # Error handling and error responses
├── clients.go         # Client CRUD endpoints
├── projects.go        # Project CRUD endpoints
├── timesheets.go      # Timesheet CRUD endpoints
├── invoices.go        # Invoice CRUD endpoints
├── reports.go         # Report endpoints
├── settings.go        # Settings endpoints
└── apikeys.go         # API key management endpoints

internal/models/
└── apikeys.go         # API key model with CRUD operations

queries/
└── api_keys.sql       # SQLC queries for API keys

migrations/
└── 00X_create_api_keys.sql  # New migration
```

## Database Schema for API Keys

```sql
CREATE TABLE api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,              -- e.g., "MCP Server", "Mobile App"
    key_hash TEXT NOT NULL UNIQUE,   -- bcrypt hash of the API key
    key_prefix TEXT NOT NULL,        -- First 8 chars for identification
    scopes TEXT NOT NULL,            -- Space-separated scopes (e.g., "clients:read clients:write invoices:*")
    last_used_at TIMESTAMP,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Available scopes follow OAuth 2.0 patterns:
-- clients:read, clients:write
-- projects:read, projects:write
-- timesheets:read, timesheets:write
-- invoices:read, invoices:write
-- reports:read
-- settings:read, settings:write
-- apikeys:read, apikeys:write
-- Wildcard: * (full access) or clients:* (all client operations)
```

## REST API Endpoints (Complete List)

All endpoints require the `Authorization: Bearer ftk_...` header and appropriate scopes.

### Authentication (`/api/v1/auth`)
- `POST /api/v1/auth/login` - Get API key (email/password → returns API key) - No auth required
- `POST /api/v1/auth/apikeys` - Create new API key - Requires `apikeys:write`
- `GET /api/v1/auth/apikeys` - List user's API keys - Requires `apikeys:read`
- `DELETE /api/v1/auth/apikeys/{id}` - Revoke API key - Requires `apikeys:write`

### Clients (`/api/v1/clients`)
- `GET /api/v1/clients` - List with pagination & search - Requires `clients:read`
- `GET /api/v1/clients/{id}` - Get single client - Requires `clients:read`
- `POST /api/v1/clients` - Create client - Requires `clients:write`
- `PUT /api/v1/clients/{id}` - Update client - Requires `clients:write`
- `DELETE /api/v1/clients/{id}` - Delete client - Requires `clients:write`
- `GET /api/v1/clients/{id}/projects` - Get client's projects - Requires `clients:read projects:read`
- `GET /api/v1/clients/{id}/hourlyrate` - Get hourly rate - Requires `clients:read`

### Projects (`/api/v1/projects`)
- `GET /api/v1/projects` - List with pagination & search - Requires `projects:read`
- `GET /api/v1/projects/{id}` - Get single project with timesheets/invoices - Requires `projects:read`
- `POST /api/v1/projects` - Create project - Requires `projects:write`
- `PUT /api/v1/projects/{id}` - Update project - Requires `projects:write`
- `PATCH /api/v1/projects/{id}/status` - Update status only (triggers emails) - Requires `projects:write`
- `DELETE /api/v1/projects/{id}` - Delete project - Requires `projects:write`
- `GET /api/v1/projects/{id}/timesheets` - Get project timesheets - Requires `projects:read timesheets:read`
- `GET /api/v1/projects/{id}/invoices` - Get project invoices - Requires `projects:read invoices:read`

### Timesheets (`/api/v1/timesheets`)
- `GET /api/v1/timesheets/{id}` - Get single timesheet - Requires `timesheets:read`
- `POST /api/v1/projects/{id}/timesheets` - Create timesheet for project - Requires `timesheets:write`
- `PUT /api/v1/timesheets/{id}` - Update timesheet - Requires `timesheets:write`
- `DELETE /api/v1/timesheets/{id}` - Delete timesheet - Requires `timesheets:write`

### Invoices (`/api/v1/invoices`)
- `GET /api/v1/invoices/{id}` - Get single invoice - Requires `invoices:read`
- `POST /api/v1/projects/{id}/invoices` - Create invoice for project - Requires `invoices:write`
- `PUT /api/v1/invoices/{id}` - Update invoice - Requires `invoices:write`
- `DELETE /api/v1/invoices/{id}` - Delete invoice - Requires `invoices:write`
- `GET /api/v1/invoices/{id}/pdf` - Generate PDF (returns binary) - Requires `invoices:read`
- `POST /api/v1/invoices/{id}/email` - Email invoice to client - Requires `invoices:write`

### Reports (`/api/v1/reports`)
- `GET /api/v1/reports/income?year={year}` - Income report by year - Requires `reports:read`

### Settings (`/api/v1/settings`)
- `GET /api/v1/settings` - Get all settings - Requires `settings:read`
- `GET /api/v1/settings/{key}` - Get single setting - Requires `settings:read`
- `PUT /api/v1/settings` - Update multiple settings - Requires `settings:write`
- `PUT /api/v1/settings/{key}` - Update single setting - Requires `settings:write`

## Response Formats

### Success Response
```json
{
  "data": { ... },
  "meta": {
    "timestamp": "2025-11-09T10:30:00Z"
  }
}
```

### Paginated Response
```json
{
  "data": [ ... ],
  "meta": {
    "page": 1,
    "pageSize": 20,
    "total": 150,
    "totalPages": 8
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": {
      "email": "Invalid email format",
      "hourlyRate": "Must be greater than 0"
    }
  }
}
```

## Implementation Plan (Todo List)

1. ⏳ Create internal/api package structure with separate files per entity
2. ⏳ Add database migration for api_keys table
3. ⏳ Create SQLC queries for API key operations
4. ⏳ Generate SQLC code for API keys
5. ⏳ Implement API key model in internal/models/api_keys.go
6. ⏳ Create API middleware (auth, JSON helpers, error handling)
7. ⏳ Implement Clients REST API endpoints
8. ⏳ Implement Projects REST API endpoints
9. ⏳ Implement Timesheets REST API endpoints
10. ⏳ Implement Invoices REST API endpoints
11. ⏳ Implement Reports REST API endpoints
12. ⏳ Implement Settings REST API endpoints
13. ⏳ Implement API Keys management endpoints
14. ⏳ Add API routes to cmd/web/routes.go under /api prefix
15. ⏳ Add tests for API endpoints
16. ⏳ Update CLAUDE.md with API documentation

## Implementation Decisions - Answers

1. **Permissions**: ✅ Separate API-specific permissions using scopes
   - API keys have scope-based permissions (e.g., `clients:read`, `clients:write`, `invoices:*`)
   - More granular control than user roles
   - Follows OAuth 2.0 scope patterns for familiarity

2. **Rate Limiting**: ✅ Implement now
   - In-memory token bucket algorithm
   - Default: 100 requests per minute per API key
   - Configurable per key if needed in future
   - Returns 429 status with Retry-After header

3. **CORS**: ✅ Add CORS support
   - Configurable allowed origins (default: none for security)
   - Support for preflight OPTIONS requests
   - Appropriate headers for browser-based clients

4. **Versioning**: ✅ Use `/api/v1/` prefix
   - All endpoints under `/api/v1/clients`, `/api/v1/projects`, etc.
   - Allows for future breaking changes in v2 without disrupting existing clients
   - Industry standard pattern

## MCP Server (Phase 2 - Not Started)

After REST API is complete, we'll implement the MCP server that:
- Connects to deployed FreelanceTracker at Fly.io
- Authenticates using API key
- Exposes ~25-30 MCP tools (list_clients, create_timesheet, etc.)
- Transforms natural language requests into REST API calls
- Returns formatted responses

**MCP Tools to Implement**:
- Clients: list, get, create, update, delete (5 tools)
- Projects: list, get, create, update, update_status, delete (6 tools)
- Timesheets: list, create, update, delete (4 tools)
- Invoices: list, get, create, update, generate_pdf, email, delete (7 tools)
- Reports: get_income_report (1 tool)
- Settings: get, update (2 tools)

## Next Steps

When resuming:
1. Answer the 4 open questions above
2. Start implementing the REST API following the todo list order
3. Begin with API keys migration and model
4. Then implement middleware and authentication
5. Then implement endpoints entity by entity (Clients → Projects → Timesheets → Invoices → Reports/Settings)
6. Add tests throughout
7. Document as we go

## Key Technical Notes

- All API endpoints will reuse existing model layer business logic
- Validation logic from web forms will be reused
- Permissions system will be preserved (users need appropriate permissions)
- Email notifications and PDF generation will work through existing code
- SQLite with WAL mode supports concurrent access (web app + API)
- API key bcrypt hashing uses same cost factor as passwords (12)

## Reference: Existing FreelanceTracker Features

Backend has complete CRUD for:
- Clients (with hourly rates, addresses, invoice preferences)
- Projects (with multi-currency, discounts, adjustments)
- Timesheets (with auto-recalculation of unpaid invoices)
- Invoices (with PDF generation and email delivery)
- Settings (SMTP config, branding, payment terms)
- Users & Authentication (bcrypt, sessions, role-based permissions)
- Reports (income by year)

All operations use soft deletes for data preservation.
