package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/paulboeck/FreelanceTrackerGo/internal/email"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/require"
)

// E2ETestContext holds all the resources needed for e2e tests
type E2ETestContext struct {
	Browser    *rod.Browser
	Page       *rod.Page
	ServerURL  string
	DB         *sql.DB
	TestDB     *testutil.TestDatabase
	Server     *http.Server
	ServerDone chan error
	App        *application
	T          *testing.T
}

// SetupE2ETest creates a new test context with browser, database, and running server
func SetupE2ETest(t *testing.T) *E2ETestContext {
	// Create test database
	testDB := testutil.SetupTestSQLite(t)

	// Create application instance for testing
	app := createE2EApplication(t, testDB.DB)

	// Start test server on port 9876
	serverURL := "http://localhost:9876"
	serverDone := make(chan error, 1)
	server := startE2EServer(t, app, serverURL, serverDone)

	// Wait for server to be ready
	waitForServer(t, serverURL)

	// Launch headless browser with proper configuration
	launcher := launcher.New().
		Headless(true).
		NoSandbox(true).
		Set("disable-gpu").
		Set("no-first-run").
		Set("disable-dev-shm-usage")

	launcherURL, err := launcher.Launch()
	require.NoError(t, err)

	// Create browser with timeout configuration
	browser := rod.New().
		ControlURL(launcherURL).
		Timeout(30 * time.Second). // Set overall timeout
		MustConnect()

	// Create page with proper timeout
	page := browser.MustPage("").
		Timeout(10 * time.Second) // Set page-level timeout

	ctx := &E2ETestContext{
		Browser:    browser,
		Page:       page,
		ServerURL:  serverURL,
		DB:         testDB.DB,
		TestDB:     testDB,
		Server:     server,
		ServerDone: serverDone,
		App:        app,
		T:          t,
	}

	// Register cleanup
	t.Cleanup(func() {
		ctx.Cleanup()
	})

	return ctx
}

// createE2EApplication creates a real application instance for e2e testing
func createE2EApplication(t *testing.T, db *sql.DB) *application {
	// Save current directory and change to project root for template loading
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// Change to project root (two directories up from cmd/web)
	err = os.Chdir("../..")
	require.NoError(t, err)

	// Restore directory after template loading
	defer func() {
		os.Chdir(originalDir)
	}()

	// Create logger (quiet for tests)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Create template cache (this requires being in project root)
	templateCache, err := newTemplateCache()
	require.NoError(t, err)

	// Create form decoder
	formDecoder := form.NewDecoder()

	// Setup session manager with SQLite store
	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	// Create all model instances
	clientModel := models.NewClientModel(db)
	projectModel := models.NewProjectModel(db)
	timesheetModel := models.NewTimesheetModel(db)
	invoiceModel := models.NewInvoiceModel(db)
	settingsModel := models.NewAppSettingModel(db, "test-encryption-seed")
	userModel := models.NewUserModel(db)

	// For testing, we'll use a mock permission model that returns empty permissions
	// This allows the middleware to work without complex permission setup
	var roleModel models.RoleModelInterface = nil
	var permissionModel models.PermissionModelInterface = &mockPermissionModel{}

	// Create disabled email service for testing
	emailService := &email.Service{}

	app := &application{
		logger:         logger,
		clients:        clientModel,
		projects:       projectModel,
		timesheets:     timesheetModel,
		invoices:       invoiceModel,
		settings:       settingsModel,
		users:          userModel,
		roles:          roleModel,
		permissions:    permissionModel,
		emailService:   emailService,
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	return app
}

// startE2EServer starts the HTTP server for testing
func startE2EServer(t *testing.T, app *application, serverURL string, serverDone chan error) *http.Server {
	// Create server
	srv := &http.Server{
		Addr:         ":9876",
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in background
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverDone <- err
		}
	}()

	return srv
}

// waitForServer waits for the server to be ready
func waitForServer(t *testing.T, serverURL string) {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(serverURL + "/user/login")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("Server did not start in time")
}

// Cleanup cleans up all test resources
func (ctx *E2ETestContext) Cleanup() {
	// Close browser
	if ctx.Browser != nil {
		ctx.Browser.MustClose()
	}

	// Stop server
	if ctx.Server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ctx.Server.Shutdown(shutdownCtx)
	}

	// Close database
	if ctx.TestDB != nil {
		ctx.TestDB.Cleanup(ctx.T)
	}
}

// Helper methods for interacting with the browser

// Navigate navigates to a path on the test server
func (ctx *E2ETestContext) Navigate(path string) {
	ctx.Page.Timeout(10 * time.Second).MustNavigate(ctx.ServerURL + path)
	ctx.Page.Timeout(10 * time.Second).MustWaitLoad()
}

// Login performs a login action
func (ctx *E2ETestContext) Login(email, password string) error {
	ctx.Navigate("/user/login")

	// Fill in login form with timeout
	ctx.Page.Timeout(5 * time.Second).MustElement("input[name='email']").MustInput(email)
	ctx.Page.Timeout(5 * time.Second).MustElement("input[name='password']").MustInput(password)

	// Submit form by clicking submit button (it's an input, not a button)
	ctx.Page.Timeout(5 * time.Second).MustElement("input[type='submit']").MustClick()
	ctx.Page.Timeout(10 * time.Second).MustWaitLoad()

	return nil
}

// TakeScreenshot takes a screenshot for debugging
func (ctx *E2ETestContext) TakeScreenshot(name string) error {
	screenshotDir := "../../tmp/screenshots"
	os.MkdirAll(screenshotDir, 0755)

	filename := fmt.Sprintf("%s/%s.png", screenshotDir, name)
	data, err := ctx.Page.Screenshot(false, nil)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// WaitForElement waits for an element to appear
func (ctx *E2ETestContext) WaitForElement(selector string, timeout time.Duration) error {
	return rod.Try(func() {
		ctx.Page.Timeout(timeout).MustElement(selector)
	})
}

// GetText gets the text content of an element
func (ctx *E2ETestContext) GetText(selector string) string {
	element := ctx.Page.Timeout(5 * time.Second).MustElement(selector)
	return element.MustText()
}

// ClickElement clicks an element
func (ctx *E2ETestContext) ClickElement(selector string) error {
	element := ctx.Page.Timeout(5 * time.Second).MustElement(selector)
	element.MustClick()
	return nil
}

// FillInput fills an input field
func (ctx *E2ETestContext) FillInput(selector, value string) error {
	element := ctx.Page.Timeout(5 * time.Second).MustElement(selector)
	return element.Input(value)
}

// SubmitForm submits a form by clicking its submit button
func (ctx *E2ETestContext) SubmitForm(formSelector string) error {
	// Find submit button within the form and click it
	form := ctx.Page.Timeout(5 * time.Second).MustElement(formSelector)
	submitBtn := form.MustElement("button[type='submit'], input[type='submit']")
	submitBtn.MustClick()
	ctx.Page.Timeout(10 * time.Second).MustWaitLoad()
	return nil
}

// AssertURL asserts the current URL path
func (ctx *E2ETestContext) AssertURL(expectedPath string) {
	info := ctx.Page.Timeout(5 * time.Second).MustInfo()
	expectedURL := ctx.ServerURL + expectedPath
	require.Equal(ctx.T, expectedURL, info.URL, "URL should match expected path")
}

// AssertElementExists asserts that an element exists
func (ctx *E2ETestContext) AssertElementExists(selector string) {
	_, err := ctx.Page.Timeout(5 * time.Second).Element(selector)
	require.NoError(ctx.T, err, "Element %s should exist", selector)
}

// AssertElementContainsText asserts that an element contains specific text
func (ctx *E2ETestContext) AssertElementContainsText(selector, text string) {
	element := ctx.Page.Timeout(5 * time.Second).MustElement(selector)
	actualText := element.MustText()
	require.Contains(ctx.T, actualText, text, "Element %s should contain text: %s", selector, text)
}

// CreateTestUser creates a test user in the database and disables password change requirement
func (ctx *E2ETestContext) CreateTestUser(name, email, password string) (int, error) {
	userID, err := ctx.App.users.Insert(name, email, password)
	if err != nil {
		return 0, err
	}

	// Disable password change requirement for e2e tests
	_, err = ctx.DB.Exec("UPDATE user SET require_password_change = 0 WHERE id = ?", userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// mockPermissionModel is a simple mock that returns empty permissions for e2e tests
type mockPermissionModel struct{}

func (m *mockPermissionModel) Get(id int) (models.Permission, error) {
	return models.Permission{}, models.ErrNoRecord
}

func (m *mockPermissionModel) GetByName(name string) (models.Permission, error) {
	return models.Permission{}, models.ErrNoRecord
}

func (m *mockPermissionModel) GetAll() ([]models.Permission, error) {
	return []models.Permission{}, nil
}

func (m *mockPermissionModel) GetRolePermissions(roleID int) ([]models.Permission, error) {
	return []models.Permission{}, nil
}

func (m *mockPermissionModel) GetUserPermissions(userID int) ([]models.Permission, error) {
	return []models.Permission{}, nil
}

func (m *mockPermissionModel) GetUserPermissionNames(userID int) ([]string, error) {
	// Return empty permissions for all users in e2e tests
	return []string{}, nil
}

func (m *mockPermissionModel) UserHasPermission(userID int, permissionName string) (bool, error) {
	// All users have all permissions in e2e tests
	return true, nil
}
