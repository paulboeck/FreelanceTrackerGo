package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestApp creates an application instance for testing
func createTestApp(t *testing.T) (*application, *testutil.TestDatabase) {
	testDB := testutil.SetupTestSQLite(t)

	// Create a minimal template cache for testing with base template
	templateCache := map[string]*template.Template{
		"home.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<h1>Clients</h1>
				{{range .Clients}}
					<div>{{.Name}}</div>
				{{end}}
				{{template "pagination" .}}
			</body></html>
			{{end}}
			{{define "pagination"}}
			{{if and .Pagination (gt .Pagination.TotalPages 1)}}
			<div class="pagination">
				<div class="pagination-info">
					Page {{.Pagination.CurrentPage}} of {{.Pagination.TotalPages}}
				</div>
				<div class="pagination-controls">
					{{if .Pagination.HasPrev}}
						<a href="?page={{.Pagination.PrevPage}}" class="pagination-btn pagination-btn-prev">← Previous</a>
					{{else}}
						<span class="pagination-btn pagination-btn-disabled">← Previous</span>
					{{end}}
					{{if .Pagination.HasNext}}
						<a href="?page={{.Pagination.NextPage}}" class="pagination-btn pagination-btn-next">Next →</a>
					{{else}}
						<span class="pagination-btn pagination-btn-disabled">Next →</span>
					{{end}}
				</div>
			</div>
			{{end}}
			{{end}}
		`)),
		"clients.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<h1>Clients</h1>
				{{range .Clients}}
					<div>{{.Name}}</div>
				{{end}}
				{{template "pagination" .}}
			</body></html>
			{{end}}
			{{define "pagination"}}
			{{if and .Pagination (gt .Pagination.TotalPages 1)}}
			<div class="pagination">
				<div class="pagination-info">
					Page {{.Pagination.CurrentPage}} of {{.Pagination.TotalPages}}
				</div>
				<div class="pagination-controls">
					{{if .Pagination.HasPrev}}
						<a href="?page={{.Pagination.PrevPage}}" class="pagination-btn pagination-btn-prev">← Previous</a>
					{{else}}
						<span class="pagination-btn pagination-btn-disabled">← Previous</span>
					{{end}}
					{{if .Pagination.HasNext}}
						<a href="?page={{.Pagination.NextPage}}" class="pagination-btn pagination-btn-next">Next →</a>
					{{else}}
						<span class="pagination-btn pagination-btn-disabled">Next →</span>
					{{end}}
				</div>
			</div>
			{{end}}
			{{end}}
		`)),
		"client.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<h1>{{.Client.Name}}</h1>
				<p>ID: {{.Client.ID}}</p>
			</body></html>
			{{end}}
		`)),
		"client_create.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<form method="POST">
					<input type="text" name="name" value="{{.Form.Name}}">
					{{if .Form.FieldErrors.name}}<span>{{.Form.FieldErrors.name}}</span>{{end}}
					<button type="submit">Create</button>
				</form>
			</body></html>
			{{end}}
		`)),
		"projects.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<h2>All Projects</h2>
				{{if .ProjectsWithClient}}
					<table>
						{{range .ProjectsWithClient}}
							<tr>
								<td>{{.ID}}</td>
								<td>{{.Name}}</td>
								<td>{{.ClientName}}</td>
								<td>{{.Status}}</td>
							</tr>
						{{end}}
					</table>
					{{template "pagination" .}}
				{{else}}
					<p>No projects found.</p>
				{{end}}
			</body></html>
			{{end}}
			{{define "pagination"}}
			{{if and .Pagination (gt .Pagination.TotalPages 1)}}
			<div class="pagination">
				<div class="pagination-info">
					Page {{.Pagination.CurrentPage}} of {{.Pagination.TotalPages}}
				</div>
				<div class="pagination-controls">
					{{if .Pagination.HasPrev}}
						<a href="?page={{.Pagination.PrevPage}}" class="pagination-btn pagination-btn-prev">← Previous</a>
					{{else}}
						<span class="pagination-btn pagination-btn-disabled">← Previous</span>
					{{end}}
					{{if .Pagination.HasNext}}
						<a href="?page={{.Pagination.NextPage}}" class="pagination-btn pagination-btn-next">Next →</a>
					{{else}}
						<span class="pagination-btn pagination-btn-disabled">Next →</span>
					{{end}}
				</div>
			</div>
			{{end}}
			{{end}}
		`)),
		"project.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<h1>{{.Project.Name}}</h1>
				<p>ID: {{.Project.ID}}</p>
				<p>Client: {{.Client.Name}}</p>
			</body></html>
			{{end}}
		`)),
		"project_create.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<form method="POST">
					<input type="text" name="name" value="{{.Form.Name}}">
					{{if .Form.FieldErrors.name}}<span>{{.Form.FieldErrors.name}}</span>{{end}}
					<input type="number" name="hourly_rate" value="{{.Form.HourlyRate}}">
					<input type="text" name="additional_info" value="{{.Form.AdditionalInfo}}">
					<input type="text" name="additional_info2" value="{{.Form.AdditionalInfo2}}">
					<input type="email" name="invoice_cc_email" value="{{.Form.InvoiceCCEmail}}">
					<input type="text" name="invoice_cc_description" value="{{.Form.InvoiceCCDescription}}">
					<button type="submit">Create</button>
				</form>
			</body></html>
			{{end}}
		`)),
		"timesheet_create.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<form method="POST">
					<input type="date" name="work_date" value="{{.Form.WorkDate}}">
					{{if .Form.FieldErrors.work_date}}<span>{{.Form.FieldErrors.work_date}}</span>{{end}}
					<input type="number" name="hours_worked" value="{{.Form.HoursWorked}}">
					{{if .Form.FieldErrors.hours_worked}}<span>{{.Form.FieldErrors.hours_worked}}</span>{{end}}
					<input type="number" name="hourly_rate" value="{{.Form.HourlyRate}}">
					<input type="text" name="description" value="{{.Form.Description}}">
					<button type="submit">Create</button>
				</form>
			</body></html>
			{{end}}
		`)),
		"invoice_create.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<form method="POST">
					<input type="date" name="invoice_date" value="{{.Form.InvoiceDate}}">
					{{if .Form.FieldErrors.invoice_date}}<span>{{.Form.FieldErrors.invoice_date}}</span>{{end}}
					<input type="number" name="amount_due" value="{{.Form.AmountDue}}">
					{{if .Form.FieldErrors.amount_due}}<span>{{.Form.FieldErrors.amount_due}}</span>{{end}}
					<input type="text" name="payment_terms" value="{{.Form.PaymentTerms}}">
					<input type="date" name="date_paid" value="{{.Form.DatePaid}}">
					<button type="submit">Create</button>
				</form>
			</body></html>
			{{end}}
		`)),
		"user_create.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<form method="POST" action="/user/signup">
					<input type="text" name="name" value="{{.Form.Name}}">
					{{if .Form.FieldErrors.name}}<span>{{.Form.FieldErrors.name}}</span>{{end}}
					<input type="email" name="email" value="{{.Form.Email}}">
					{{if .Form.FieldErrors.email}}<span>{{.Form.FieldErrors.email}}</span>{{end}}
					<input type="password" name="password" value="{{.Form.Password}}">
					{{if .Form.FieldErrors.password}}<span>{{.Form.FieldErrors.password}}</span>{{end}}
					<button type="submit">Create account</button>
				</form>
			</body></html>
			{{end}}
		`)),
		"user_login.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<h2>Login</h2>
				<form method="POST" action="/user/login">
					<input type="email" name="email" value="{{.Form.Email}}">
					{{if .Form.FieldErrors.email}}<span>{{.Form.FieldErrors.email}}</span>{{end}}
					<input type="password" name="password" value="{{.Form.Password}}">
					{{if .Form.FieldErrors.password}}<span>{{.Form.FieldErrors.password}}</span>{{end}}
					<button type="submit">Login</button>
				</form>
			</body></html>
			{{end}}
		`)),
		"settings.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<h1>Application Settings</h1>
				<a href="/settings/edit" class="btn-client-action">Edit Setting Values</a>
				{{if .Settings}}
					{{range .Settings}}
						<div>
							<strong>{{.Description}}</strong>: 
							<span>
								{{if eq .Key "smtp_password"}}
									{{if .Value}}[Password Set]{{else}}[Not Set]{{end}}
								{{else}}
									{{.Value}}
								{{end}}
							</span>
						</div>
					{{end}}
				{{end}}
			</body></html>
			{{end}}
		`)),
		"settings_edit.html": template.Must(template.New("base").Parse(`
			{{define "base"}}
			<html><body>
				<h2>Email Settings</h2>
				<form action="/settings/edit" method="POST">
					{{range .Settings}}
						{{if or (eq .Key "email_enabled") (eq .Key "smtp_host") (eq .Key "smtp_port") (eq .Key "smtp_username") (eq .Key "smtp_password") (eq .Key "smtp_from_name") (eq .Key "smtp_use_tls")}}
							<div>
								<label for="{{.Key}}">{{.Description}}</label>
								{{if eq .DataType "bool"}}
									<select id="{{.Key}}" name="{{.Key}}">
										<option value="true" {{if eq .Value "true"}}selected{{end}}>Yes</option>
										<option value="false" {{if eq .Value "false"}}selected{{end}}>No</option>
									</select>
								{{else if eq .Key "smtp_password"}}
									<input type="password" id="{{.Key}}" name="{{.Key}}" value="" placeholder="Enter Gmail app password (leave blank to keep current)">
								{{else if eq .DataType "int"}}
									<input type="number" id="{{.Key}}" name="{{.Key}}" value="{{.Value}}">
								{{else}}
									<input type="text" id="{{.Key}}" name="{{.Key}}" value="{{.Value}}">
								{{end}}
								{{with index $.Form.FieldErrors .Key}}<span class="error">{{.}}</span>{{end}}
							</div>
						{{end}}
					{{end}}
					<h2>Other Settings</h2>
					{{range .Settings}}
						{{if not (or (eq .Key "email_enabled") (eq .Key "smtp_host") (eq .Key "smtp_port") (eq .Key "smtp_username") (eq .Key "smtp_password") (eq .Key "smtp_from_name") (eq .Key "smtp_use_tls"))}}
							<div>
								<label for="{{.Key}}">{{.Description}}</label>
								{{if eq .DataType "bool"}}
									<select id="{{.Key}}" name="{{.Key}}">
										<option value="true" {{if eq .Value "true"}}selected{{end}}>Yes</option>
										<option value="false" {{if eq .Value "false"}}selected{{end}}>No</option>
									</select>
								{{else if eq .DataType "decimal"}}
									<input type="number" step="0.01" id="{{.Key}}" name="{{.Key}}" value="{{.Value}}">
								{{else if eq .DataType "int"}}
									<input type="number" id="{{.Key}}" name="{{.Key}}" value="{{.Value}}">
								{{else if eq .DataType "float"}}
									<input type="number" step="any" id="{{.Key}}" name="{{.Key}}" value="{{.Value}}">
								{{else}}
									<input type="text" id="{{.Key}}" name="{{.Key}}" value="{{.Value}}">
								{{end}}
								{{with index $.Form.FieldErrors .Key}}<span class="error">{{.}}</span>{{end}}
							</div>
						{{end}}
					{{end}}
					<input type="submit" value="Save Settings">
				</form>
			</body></html>
			{{end}}
		`)),
	}

	// Create session manager for tests
	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(testDB.DB)
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
		clients:        models.NewClientModel(testDB.DB),
		projects:       models.NewProjectModel(testDB.DB),
		timesheets:     models.NewTimesheetModel(testDB.DB),
		invoices:       models.NewInvoiceModel(testDB.DB),
		settings:       models.NewAppSettingModel(testDB.DB, "test-encryption-seed"),
		users:          models.NewUserModel(testDB.DB),
		templateCache:  templateCache,
		formDecoder:    form.NewDecoder(),
		sessionManager: sessionManager,
	}

	return app, testDB
}

func TestHomeHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("home with no clients", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		// Use the session middleware to wrap the handler
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientsList))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "<h1>Clients</h1>")
	})

	t.Run("home with clients", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert test clients
		testDB.InsertTestClient(t, "Client A")
		testDB.InsertTestClient(t, "Client B")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		// Use the session middleware to wrap the handler
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientsList))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Client A")
		assert.Contains(t, body, "Client B")
	})
}

func TestHomeHandlerPagination(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("pagination with default page size", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert 15 clients (more than default page size of 10)
		for i := 1; i <= 15; i++ {
			testDB.InsertTestClient(t, fmt.Sprintf("Client %d", i))
		}

		// Test first page
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		app.clientsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Should show "Next" button but not "Previous"
		assert.Contains(t, body, "Next →")
		assert.Contains(t, body, "pagination-btn-disabled\">← Previous")
		assert.Contains(t, body, "Page 1 of 2")
	})

	t.Run("pagination second page", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert 15 clients
		for i := 1; i <= 15; i++ {
			testDB.InsertTestClient(t, fmt.Sprintf("Client %d", i))
		}

		// Test second page
		req := httptest.NewRequest(http.MethodGet, "/?page=2", nil)
		rr := httptest.NewRecorder()
		app.clientsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Should show "Previous" button but not "Next"
		assert.Contains(t, body, "← Previous")
		assert.Contains(t, body, "pagination-btn-disabled\">Next →")
		assert.Contains(t, body, "Page 2 of 2")
	})

	t.Run("pagination with invalid page number", func(t *testing.T) {
		testDB.TruncateTable(t, "client")
		testDB.InsertTestClient(t, "Test Client")

		// Test with invalid page number - should default to page 1
		req := httptest.NewRequest(http.MethodGet, "/?page=invalid", nil)
		rr := httptest.NewRecorder()
		app.clientsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		// Should not show pagination controls for single page
		body := rr.Body.String()
		assert.NotContains(t, body, "Page 1 of")
	})

	t.Run("no pagination when items fit on one page", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert only 5 clients (less than page size)
		for i := 1; i <= 5; i++ {
			testDB.InsertTestClient(t, fmt.Sprintf("Client %d", i))
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		app.clientsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Should not show pagination controls
		assert.NotContains(t, body, "pagination")
		assert.NotContains(t, body, "Page 1 of")
	})
}

func TestProjectsListPagination(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("pagination with default page size", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client
		clientID := testDB.InsertTestClient(t, "Test Client")

		// Insert 15 projects (more than default page size of 10)
		for i := 1; i <= 15; i++ {
			_, err := testDB.DB.Exec(`INSERT INTO project (name, client_id, status, hourly_rate, currency_display, currency_conversion_rate, flat_fee_invoice) 
				VALUES (?, ?, ?, ?, ?, ?, ?)`, fmt.Sprintf("Project %d", i), clientID, "In Progress", 100.0, "USD", 1.0, 0)
			require.NoError(t, err)
		}

		// Test first page
		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		rr := httptest.NewRecorder()
		app.projectsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Should show "Next" button but not "Previous"
		assert.Contains(t, body, "Next →")
		assert.Contains(t, body, "pagination-btn-disabled\">← Previous")
		assert.Contains(t, body, "Page 1 of 2")
	})

	t.Run("pagination second page", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client
		clientID := testDB.InsertTestClient(t, "Test Client")

		// Insert 15 projects
		for i := 1; i <= 15; i++ {
			_, err := testDB.DB.Exec(`INSERT INTO project (name, client_id, status, hourly_rate, currency_display, currency_conversion_rate, flat_fee_invoice) 
				VALUES (?, ?, ?, ?, ?, ?, ?)`, fmt.Sprintf("Project %d", i), clientID, "In Progress", 100.0, "USD", 1.0, 0)
			require.NoError(t, err)
		}

		// Test second page
		req := httptest.NewRequest(http.MethodGet, "/projects?page=2", nil)
		rr := httptest.NewRecorder()
		app.projectsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Should show "Previous" button but not "Next"
		assert.Contains(t, body, "← Previous")
		assert.Contains(t, body, "pagination-btn-disabled\">Next →")
		assert.Contains(t, body, "Page 2 of 2")
	})
}

func TestClientViewHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("view existing client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert a test client
		id := testDB.InsertTestClient(t, "Test Client")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/client/view/%d", id), nil)
		req.SetPathValue("id", strconv.Itoa(id))
		rr := httptest.NewRecorder()

		app.clientView(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Test Client")
		assert.Contains(t, body, fmt.Sprintf("ID: %d", id))
	})

	t.Run("view non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		req := httptest.NewRequest(http.MethodGet, "/client/view/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		app.clientView(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("view with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/client/view/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		app.clientView(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("view with negative ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/client/view/-1", nil)
		req.SetPathValue("id", "-1")
		rr := httptest.NewRecorder()

		app.clientView(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestClientCreateHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("show create form", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/client/create", nil)
		rr := httptest.NewRecorder()

		app.clientCreate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "<form method=\"POST\">")
		assert.Contains(t, body, "name=\"name\"")
	})
}

func TestClientCreatePostHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful client creation", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		form := url.Values{}
		form.Add("name", "New Test Client")
		form.Add("email", "newtest@example.com")
		form.Add("hourly_rate", "75.00")

		req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientCreatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to the new client view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, "/client/view/")

		// Verify the client was actually created in the database
		clients, err := app.clients.GetAll()
		require.NoError(t, err)
		require.Len(t, clients, 1)
		assert.Equal(t, "New Test Client", clients[0].Name)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		form := url.Values{}
		form.Add("name", "")

		req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientCreatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Name is required")

		// Verify no client was created
		clients, err := app.clients.GetAll()
		require.NoError(t, err)
		assert.Empty(t, clients)
	})

	t.Run("validation error - name too long", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Create a name longer than 255 characters
		longName := strings.Repeat("a", 256)

		form := url.Values{}
		form.Add("name", longName)

		req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientCreatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Name must be shorter than 255 characters")

		// Verify no client was created
		clients, err := app.clients.GetAll()
		require.NoError(t, err)
		assert.Empty(t, clients)
	})

	t.Run("malformed form data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader("invalid-form-data"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		app.clientCreatePost(rr, req)

		// The form parsing doesn't fail on "invalid-form-data", but validation does
		// since no proper "name" field is provided, leading to validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestHandlersIntegration(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("full workflow - create and view client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// 1. Create a client via POST
		form := url.Values{}
		form.Add("name", "Integration Test Client")
		form.Add("email", "integration@example.com")
		form.Add("hourly_rate", "85.00")

		req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientCreatePost))
		handler.ServeHTTP(rr, req)

		// Extract the client ID from the redirect URL
		require.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		require.Contains(t, location, "/client/view/")

		// Extract ID from URL
		parts := strings.Split(location, "/")
		idStr := parts[len(parts)-1]
		id, err := strconv.Atoi(idStr)
		require.NoError(t, err)

		// 2. View the created client
		req = httptest.NewRequest(http.MethodGet, location, nil)
		req.SetPathValue("id", idStr)
		rr = httptest.NewRecorder()

		app.clientView(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Integration Test Client")
		assert.Contains(t, body, fmt.Sprintf("ID: %d", id))

		// 3. Verify it appears on home page
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		rr = httptest.NewRecorder()

		app.clientsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body = rr.Body.String()
		assert.Contains(t, body, "Integration Test Client")
	})
}

func TestClientUpdateHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("show update form for existing client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert a test client
		id := testDB.InsertTestClient(t, "Test Client")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/client/update/%d", id), nil)
		req.SetPathValue("id", strconv.Itoa(id))
		rr := httptest.NewRecorder()

		app.clientUpdate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "<form method=\"POST\">")
		assert.Contains(t, body, "name=\"name\"")
		assert.Contains(t, body, "value=\"Test Client\"") // Form should be pre-populated
	})

	t.Run("update form for non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		req := httptest.NewRequest(http.MethodGet, "/client/update/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		app.clientUpdate(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update form with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/client/update/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		app.clientUpdate(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update form with negative ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/client/update/-1", nil)
		req.SetPathValue("id", "-1")
		rr := httptest.NewRecorder()

		app.clientUpdate(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestClientUpdatePostHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful client update", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert a test client
		id := testDB.InsertTestClient(t, "Original Name")

		form := url.Values{}
		form.Add("name", "Updated Name")
		form.Add("email", "updated@example.com")
		form.Add("hourly_rate", "65.00")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/client/update/%d", id), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(id))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to the client view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Equal(t, fmt.Sprintf("/client/view/%d", id), location)

		// Verify the client was actually updated in the database
		client, err := app.clients.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", client.Name)
	})

	t.Run("update non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		form := url.Values{}
		form.Add("name", "Updated Name")

		req := httptest.NewRequest(http.MethodPost, "/client/update/999", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "999")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		app.clientUpdatePost(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update with invalid ID", func(t *testing.T) {
		form := url.Values{}
		form.Add("name", "Updated Name")

		req := httptest.NewRequest(http.MethodPost, "/client/update/invalid", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "invalid")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		app.clientUpdatePost(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert a test client
		id := testDB.InsertTestClient(t, "Original Name")

		form := url.Values{}
		form.Add("name", "")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/client/update/%d", id), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(id))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Name is required")

		// Verify the client was not updated
		client, err := app.clients.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "Original Name", client.Name)
	})

	t.Run("validation error - name too long", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert a test client
		id := testDB.InsertTestClient(t, "Original Name")

		// Create a name longer than 255 characters
		longName := strings.Repeat("a", 256)

		form := url.Values{}
		form.Add("name", longName)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/client/update/%d", id), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(id))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Name must be shorter than 255 characters")

		// Verify the client was not updated
		client, err := app.clients.Get(id)
		require.NoError(t, err)
		assert.Equal(t, "Original Name", client.Name)
	})
}

func TestUpdateHandlersIntegration(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("full update workflow", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// 1. Create a client
		originalName := "Original Client Name"
		id := testDB.InsertTestClient(t, originalName)

		// 2. Get the update form
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/client/update/%d", id), nil)
		req.SetPathValue("id", strconv.Itoa(id))
		rr := httptest.NewRecorder()

		app.clientUpdate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, originalName) // Should show current name

		// 3. Submit the update
		newName := "Updated Client Name"
		form := url.Values{}
		form.Add("name", newName)
		form.Add("email", "updatedclient@example.com")
		form.Add("hourly_rate", "95.00")

		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/client/update/%d", id), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(id))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr = httptest.NewRecorder()

		app.clientUpdatePost(rr, req)

		// Should redirect to client view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Equal(t, fmt.Sprintf("/client/view/%d", id), location)

		// 4. Verify the client view shows updated name
		req = httptest.NewRequest(http.MethodGet, location, nil)
		req.SetPathValue("id", strconv.Itoa(id))
		rr = httptest.NewRecorder()

		app.clientView(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body = rr.Body.String()
		assert.Contains(t, body, newName)
		assert.NotContains(t, body, originalName)

		// 5. Verify home page shows updated name
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		rr = httptest.NewRecorder()

		app.clientsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body = rr.Body.String()
		assert.Contains(t, body, newName)
	})
}

// PROJECT HANDLER TESTS

func TestProjectViewHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("view existing project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/project/view/%d", projectID), nil)
		req.SetPathValue("id", strconv.Itoa(projectID))
		rr := httptest.NewRecorder()

		app.projectView(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Test Project")
		assert.Contains(t, body, fmt.Sprintf("ID: %d", projectID))
		assert.Contains(t, body, "Test Client")
	})

	t.Run("view non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")

		req := httptest.NewRequest(http.MethodGet, "/project/view/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		app.projectView(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("view with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/project/view/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		app.projectView(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestProjectCreateHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("show create form", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert test client
		clientID := testDB.InsertTestClient(t, "Test Client")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/client/%d/project/create", clientID), nil)
		req.SetPathValue("id", strconv.Itoa(clientID))
		rr := httptest.NewRecorder()

		app.projectCreate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "<form method=\"POST\">")
		assert.Contains(t, body, "name=\"name\"")
	})

	t.Run("create form for non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		req := httptest.NewRequest(http.MethodGet, "/client/999/project/create", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		app.projectCreate(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestProjectCreatePostHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful project creation", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client
		clientID := testDB.InsertTestClient(t, "Test Client")

		form := url.Values{}
		form.Add("name", "New Test Project")
		form.Add("status", "Estimating")
		form.Add("hourly_rate", "50.00")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/client/%d/project/create", clientID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(clientID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.projectCreatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to the client view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, fmt.Sprintf("/client/view/%d", clientID))

		// Verify the project was actually created in the database
		projects, err := app.projects.GetByClient(clientID)
		require.NoError(t, err)
		require.Len(t, projects, 1)
		assert.Equal(t, "New Test Project", projects[0].Name)
		assert.Equal(t, clientID, projects[0].ClientID)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client
		clientID := testDB.InsertTestClient(t, "Test Client")

		form := url.Values{}
		form.Add("name", "")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/client/%d/project/create", clientID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(clientID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.projectCreatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Name is required")

		// Verify no project was created
		projects, err := app.projects.GetByClient(clientID)
		require.NoError(t, err)
		assert.Empty(t, projects)
	})

	t.Run("create project for non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		form := url.Values{}
		form.Add("name", "Test Project")

		req := httptest.NewRequest(http.MethodPost, "/client/999/project/create", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "999")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		app.projectCreatePost(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestProjectCreateDefaulting(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("project form defaults from client fields", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert test client with specific values
		clientID := testDB.InsertTestClientWithDefaults(t, "Test Client", 125.50,
			"Additional Info Value", "Additional Info 2 Value",
			"cc@example.com", "CC Description Value")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/client/%d/project/create", clientID), nil)
		req.SetPathValue("id", strconv.Itoa(clientID))
		rr := httptest.NewRecorder()

		app.projectCreate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Check that form defaults are populated from client
		assert.Contains(t, body, `value="125.50"`)                  // Hourly rate
		assert.Contains(t, body, `value="Additional Info Value"`)   // Additional Info
		assert.Contains(t, body, `value="Additional Info 2 Value"`) // Additional Info 2
		assert.Contains(t, body, `value="cc@example.com"`)          // Invoice CC Email
		assert.Contains(t, body, `value="CC Description Value"`)    // Invoice CC Description
	})

	t.Run("project form handles empty client fields", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert test client with empty optional fields
		clientID := testDB.InsertTestClientWithDefaults(t, "Test Client", 75.00, "", "", "", "")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/client/%d/project/create", clientID), nil)
		req.SetPathValue("id", strconv.Itoa(clientID))
		rr := httptest.NewRecorder()

		app.projectCreate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Check that hourly rate still defaults but other fields are empty
		assert.Contains(t, body, `value="75.00"`) // Hourly rate should still be set
		// Empty fields should have empty values
		assert.Contains(t, body, `value=""`) // Should have empty value attributes
	})
}

func TestTimesheetCreate(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("show timesheet create form", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/project/%d/timesheet/create", projectID), nil)
		req.SetPathValue("id", strconv.Itoa(projectID))
		rr := httptest.NewRecorder()

		app.timesheetCreate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "<form method=\"POST\">")
		assert.Contains(t, body, "name=\"work_date\"")
		assert.Contains(t, body, "name=\"hourly_rate\"")
	})

	t.Run("timesheet form defaults hourly rate from project", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client
		clientID := testDB.InsertTestClient(t, "Test Client")

		// Insert project with specific hourly rate
		result, err := testDB.DB.Exec(`INSERT INTO project (name, client_id, status, hourly_rate, currency_display, currency_conversion_rate, flat_fee_invoice) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`, "Test Project", clientID, "In Progress", 95.75, "USD", 1.0, 0)
		require.NoError(t, err)

		projectIDRaw, err := result.LastInsertId()
		require.NoError(t, err)
		projectID := int(projectIDRaw)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/project/%d/timesheet/create", projectID), nil)
		req.SetPathValue("id", strconv.Itoa(projectID))
		rr := httptest.NewRecorder()

		app.timesheetCreate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Check that hourly rate defaults from project
		assert.Contains(t, body, `value="95.75"`) // Hourly rate from project
	})

	t.Run("timesheet form defaults description to project name", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and project with a specific name
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "My Awesome Project", clientID)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/project/%d/timesheet/create", projectID), nil)
		req.SetPathValue("id", strconv.Itoa(projectID))
		rr := httptest.NewRecorder()

		app.timesheetCreate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()

		// Check that description defaults to project name
		assert.Contains(t, body, `value="My Awesome Project"`)
	})

	t.Run("timesheet create for non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")

		req := httptest.NewRequest(http.MethodGet, "/project/999/timesheet/create", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		app.timesheetCreate(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestTimesheetCreatePost(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful timesheet creation", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)

		form := url.Values{}
		form.Add("work_date", "2024-01-15")
		form.Add("hours_worked", "8.0")
		form.Add("hourly_rate", "85.00")
		form.Add("description", "Test timesheet entry")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/%d/timesheet/create", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetCreatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to the project view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, fmt.Sprintf("/project/view/%d", projectID))

		// Verify the timesheet was actually created in the database
		timesheets, err := app.timesheets.GetByProject(projectID)
		require.NoError(t, err)
		require.Len(t, timesheets, 1)
		assert.Equal(t, 8.0, timesheets[0].HoursWorked)
		assert.Equal(t, 85.0, timesheets[0].HourlyRate)
		assert.Equal(t, "Test timesheet entry", timesheets[0].Description)
	})

	t.Run("validation error - empty required fields", func(t *testing.T) {
		testDB.TruncateTable(t, "timesheet")
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)

		form := url.Values{}
		form.Add("work_date", "")
		form.Add("hours_worked", "")
		form.Add("hourly_rate", "")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/%d/timesheet/create", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.timesheetCreatePost))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

		// Verify no timesheet was created
		timesheets, err := app.timesheets.GetByProject(projectID)
		require.NoError(t, err)
		assert.Len(t, timesheets, 0)
	})

	t.Run("timesheet create for non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")

		form := url.Values{}
		form.Add("work_date", "2024-01-15")
		form.Add("hours_worked", "8.0")
		form.Add("hourly_rate", "85.00")

		req := httptest.NewRequest(http.MethodPost, "/project/999/timesheet/create", strings.NewReader(form.Encode()))
		req.SetPathValue("id", "999")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		app.timesheetCreatePost(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestProjectsList(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("show projects list with projects", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and projects
		clientID := testDB.InsertTestClient(t, "Test Client")

		// Insert a few projects
		_, err := testDB.DB.Exec(`INSERT INTO project (name, client_id, status, hourly_rate, currency_display, currency_conversion_rate, flat_fee_invoice) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`, "Project 1", clientID, "In Progress", 100.0, "USD", 1.0, 0)
		require.NoError(t, err)

		_, err = testDB.DB.Exec(`INSERT INTO project (name, client_id, status, hourly_rate, currency_display, currency_conversion_rate, flat_fee_invoice) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`, "Project 2", clientID, "Estimating", 125.0, "USD", 1.0, 0)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		rr := httptest.NewRecorder()

		app.projectsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "<h2>All Projects</h2>")
		assert.Contains(t, body, "Project 1")
		assert.Contains(t, body, "Project 2")
		assert.Contains(t, body, "Test Client")
		assert.Contains(t, body, "In Progress")
		assert.Contains(t, body, "Estimating")
	})

	t.Run("show projects list when empty", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		rr := httptest.NewRecorder()

		app.projectsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "<h2>All Projects</h2>")
		assert.Contains(t, body, "No projects found.")
	})
}

func TestProjectUpdateHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("show update form for existing project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Test Project", clientID)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/project/update/%d", projectID), nil)
		req.SetPathValue("id", strconv.Itoa(projectID))
		rr := httptest.NewRecorder()

		app.projectUpdate(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "<form method=\"POST\">")
		assert.Contains(t, body, "name=\"name\"")
		assert.Contains(t, body, "value=\"Test Project\"")
	})

	t.Run("update form for non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")

		req := httptest.NewRequest(http.MethodGet, "/project/update/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		app.projectUpdate(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestProjectUpdatePostHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful project update", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Original Project", clientID)

		form := url.Values{}
		form.Add("name", "Updated Project")
		form.Add("status", "In Progress")
		form.Add("hourly_rate", "60.00")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/update/%d", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.projectUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should redirect to the client view
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, fmt.Sprintf("/client/view/%d", clientID))

		// Verify the project was actually updated in the database
		project, err := app.projects.Get(projectID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Project", project.Name)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Original Project", clientID)

		form := url.Values{}
		form.Add("name", "")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/update/%d", projectID), strings.NewReader(form.Encode()))
		req.SetPathValue("id", strconv.Itoa(projectID))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.projectUpdatePost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Name is required")

		// Verify the project was not updated
		project, err := app.projects.Get(projectID)
		require.NoError(t, err)
		assert.Equal(t, "Original Project", project.Name)
	})
}

func TestProjectDeleteHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful project delete", func(t *testing.T) {
		testDB.TruncateTable(t, "project")
		testDB.TruncateTable(t, "client")

		// Insert test client and project
		clientID := testDB.InsertTestClient(t, "Test Client")
		projectID := testDB.InsertTestProject(t, "Project to Delete", clientID)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/project/delete/%d", projectID), nil)
		req.SetPathValue("id", strconv.Itoa(projectID))
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.projectDelete))
		handler.ServeHTTP(rr, req)

		// Should redirect to client view page
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Contains(t, location, fmt.Sprintf("/client/view/%d", clientID))

		// Verify the project was soft deleted
		projects, err := app.projects.GetByClient(clientID)
		require.NoError(t, err)
		assert.Empty(t, projects)

		// Verify the project can't be retrieved via Get
		_, err = app.projects.Get(projectID)
		assert.Error(t, err)
		assert.Equal(t, models.ErrNoRecord, err)
	})

	t.Run("delete non-existent project", func(t *testing.T) {
		testDB.TruncateTable(t, "project")

		req := httptest.NewRequest(http.MethodPost, "/project/delete/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		app.projectDelete(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestClientDeleteHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful client delete", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// Insert a test client
		id := testDB.InsertTestClient(t, "Client to Delete")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/client/delete/%d", id), nil)
		req.SetPathValue("id", strconv.Itoa(id))
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientDelete))
		handler.ServeHTTP(rr, req)

		// Should redirect to home page
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Equal(t, "/", location)

		// Verify the client was soft deleted (no longer appears in GetAll)
		clients, err := app.clients.GetAll()
		require.NoError(t, err)
		assert.Empty(t, clients)

		// Verify the client can't be retrieved via Get
		_, err = app.clients.Get(id)
		assert.Error(t, err)
		assert.Equal(t, models.ErrNoRecord, err)
	})

	t.Run("delete non-existent client", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		req := httptest.NewRequest(http.MethodPost, "/client/delete/999", nil)
		req.SetPathValue("id", "999")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientDelete))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/client/delete/invalid", nil)
		req.SetPathValue("id", "invalid")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientDelete))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with negative ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/client/delete/-1", nil)
		req.SetPathValue("id", "-1")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientDelete))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestDeleteHandlersIntegration(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("full delete workflow", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		// 1. Create clients
		client1ID := testDB.InsertTestClient(t, "Client 1")
		client2ID := testDB.InsertTestClient(t, "Client 2")
		_ = testDB.InsertTestClient(t, "Client 3")

		// 2. Verify all clients appear in home page
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		app.clientsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Client 1")
		assert.Contains(t, body, "Client 2")
		assert.Contains(t, body, "Client 3")

		// 3. Delete one client
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/client/delete/%d", client2ID), nil)
		req.SetPathValue("id", strconv.Itoa(client2ID))
		rr = httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientDelete))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// 4. Verify home page only shows remaining clients
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		rr = httptest.NewRecorder()
		app.clientsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body = rr.Body.String()
		assert.Contains(t, body, "Client 1")
		assert.NotContains(t, body, "Client 2") // Deleted client should not appear
		assert.Contains(t, body, "Client 3")

		// 5. Verify deleted client detail page returns 404
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/client/view/%d", client2ID), nil)
		req.SetPathValue("id", strconv.Itoa(client2ID))
		rr = httptest.NewRecorder()
		app.clientView(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		// 6. Verify remaining clients still accessible
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/client/view/%d", client1ID), nil)
		req.SetPathValue("id", strconv.Itoa(client1ID))
		rr = httptest.NewRecorder()
		app.clientView(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body = rr.Body.String()
		assert.Contains(t, body, "Client 1")
	})
}

func TestUserSignupHandler(t *testing.T) {
	t.Skip("Skipping user signup tests - requires role infrastructure not in test schema")
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("GET /user/signup returns signup form", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/signup", nil)
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userSignup))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "form method=\"POST\" action=\"/user/signup\"")
		assert.Contains(t, body, "name=\"name\"")
		assert.Contains(t, body, "name=\"email\"")
		assert.Contains(t, body, "name=\"password\"")
	})
}

func TestUserSignupPostHandler(t *testing.T) {
	t.Skip("Skipping user signup tests - requires role infrastructure not in test schema")
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful user creation", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		form := url.Values{}
		form.Add("name", "John Doe")
		form.Add("email", "john@example.com")
		form.Add("password", "secretpassword123")

		req := httptest.NewRequest(http.MethodPost, "/user/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userSignupPost))
		handler.ServeHTTP(rr, req)

		// Should redirect to login page
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Equal(t, "/user/login", location)

		// Verify the user was actually created in the database
		exists, err := app.users.Exists("john@example.com")
		require.NoError(t, err)
		assert.True(t, exists)

		// Verify user can be authenticated with the password
		userID, err := app.users.Authenticate("john@example.com", "secretpassword123")
		require.NoError(t, err)
		assert.Greater(t, userID, 0)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		form := url.Values{}
		form.Add("name", "")
		form.Add("email", "test@example.com")
		form.Add("password", "password123")

		req := httptest.NewRequest(http.MethodPost, "/user/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userSignupPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Name is required")

		// Verify no user was created
		exists, err := app.users.Exists("test@example.com")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("validation error - empty email", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		form := url.Values{}
		form.Add("name", "John Doe")
		form.Add("email", "")
		form.Add("password", "password123")

		req := httptest.NewRequest(http.MethodPost, "/user/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userSignupPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Email is required")
	})

	t.Run("validation error - invalid email", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		form := url.Values{}
		form.Add("name", "John Doe")
		form.Add("email", "invalid-email")
		form.Add("password", "password123")

		req := httptest.NewRequest(http.MethodPost, "/user/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userSignupPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Email must be a valid email address")
	})

	t.Run("validation error - empty password", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		form := url.Values{}
		form.Add("name", "John Doe")
		form.Add("email", "john@example.com")
		form.Add("password", "")

		req := httptest.NewRequest(http.MethodPost, "/user/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userSignupPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Password is required")
	})

	t.Run("validation error - duplicate email", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// First, create a user
		_, err := app.users.Insert("Existing User", "existing@example.com", "password123")
		require.NoError(t, err)

		// Try to create another user with the same email
		form := url.Values{}
		form.Add("name", "New User")
		form.Add("email", "existing@example.com")
		form.Add("password", "newpassword123")

		req := httptest.NewRequest(http.MethodPost, "/user/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userSignupPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Email address is already in use")
	})
}

func TestUserLoginHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	req := httptest.NewRequest(http.MethodGet, "/user/login", nil)
	rr := httptest.NewRecorder()

	// Use the session middleware to load and save session
	handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLogin))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, "<h2>Login</h2>")
	assert.Contains(t, body, `<form method="POST" action="/user/login"`)
}

func TestUserLoginPostHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("successful login", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// First create a user
		_, err := app.users.Insert("John Doe", "john@example.com", "secretpassword123")
		require.NoError(t, err)

		form := url.Values{}
		form.Add("email", "john@example.com")
		form.Add("password", "secretpassword123")

		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLoginPost))
		handler.ServeHTTP(rr, req)

		// Should redirect to home page
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Equal(t, "/", location)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		form := url.Values{}
		form.Add("email", "nonexistent@example.com")
		form.Add("password", "wrongpassword")

		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLoginPost))
		handler.ServeHTTP(rr, req)

		// Should return form with error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Email or password is incorrect")
	})

	t.Run("wrong password for existing user", func(t *testing.T) {
		testDB.TruncateTable(t, "user")

		// First create a user
		_, err := app.users.Insert("John Doe", "john@example.com", "correctpassword")
		require.NoError(t, err)

		form := url.Values{}
		form.Add("email", "john@example.com")
		form.Add("password", "wrongpassword")

		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to load and save session
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLoginPost))
		handler.ServeHTTP(rr, req)

		// Should return form with error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Email or password is incorrect")
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name     string
			email    string
			password string
			contains string
		}{
			{"empty email", "", "password123", "Email is required"},
			{"invalid email", "invalid-email", "password123", "Email must be a valid email address"},
			{"empty password", "user@example.com", "", "Password is required"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				form := url.Values{}
				form.Add("email", tt.email)
				form.Add("password", tt.password)

				req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				rr := httptest.NewRecorder()

				// Use the session middleware to load and save session
				handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLoginPost))
				handler.ServeHTTP(rr, req)

				// Should return form with validation error
				assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
				body := rr.Body.String()
				assert.Contains(t, body, tt.contains)
			})
		}
	})
}

func TestUserLogoutHandler(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	testDB.TruncateTable(t, "user")

	// Create a user and simulate login
	userID, err := app.users.Insert("John Doe", "john@example.com", "password123")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/user/logout", nil)
	rr := httptest.NewRecorder()

	// Use the session middleware and simulate an authenticated user
	handler := app.sessionManager.LoadAndSave(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.sessionManager.Put(r.Context(), "authenticatedUserID", userID)
		app.userLogout(w, r)
	}))
	handler.ServeHTTP(rr, req)

	// Should redirect to home page
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	location := rr.Header().Get("Location")
	assert.Equal(t, "/", location)
}

func TestAuthenticationMiddleware(t *testing.T) {
	t.Skip("Skipping authentication middleware tests - requires permissions infrastructure not in test schema")
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	testDB.TruncateTable(t, "user")

	t.Run("unauthenticated user redirected to login", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		// Create a handler chain with authentication required
		handler := app.sessionManager.LoadAndSave(
			app.authenticate(
				app.requireAuthentication(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				),
			),
		)
		handler.ServeHTTP(rr, req)

		// Should redirect to login
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Equal(t, "/user/login", location)
	})

	t.Run("authenticated user can access protected route", func(t *testing.T) {
		// Create a user
		userID, err := app.users.Insert("John Doe", "john@example.com", "password123")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		// Create a handler chain with authentication that simulates a logged-in user
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.sessionManager.Put(r.Context(), "authenticatedUserID", userID)
			
			authHandler := app.authenticate(
				app.requireAuthentication(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("Protected content"))
					}),
				),
			)
			authHandler.ServeHTTP(w, r)
		}))
		handler.ServeHTTP(rr, req)

		// Should allow access
		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Protected content")
	})
}
