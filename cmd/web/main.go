package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"

	"github.com/paulboeck/FreelanceTrackerGo/internal/database"
	"github.com/paulboeck/FreelanceTrackerGo/internal/email"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

// application is a struct that holds all the dependencies for our web application.
// In Go, a struct is like a container that groups related data together.
type application struct {
	// The asterisk (*) before a type means it's a pointer - it points to where
	// the actual data is stored in memory, rather than copying the whole thing.
	// Pointers are used for efficiency and to allow modification of the original data.
	logger         *slog.Logger                       // Structured logger for recording application events
	clients        models.ClientModelInterface        // Interface for client database operations
	projects       models.ProjectModelInterface       // Interface for project database operations
	timesheets     models.TimesheetModelInterface     // Interface for timesheet database operations
	invoices       models.InvoiceModelInterface       // Interface for invoice database operations
	settings       models.AppSettingModelInterface    // Interface for application settings operations
	users          models.UserModelInterface          // Interface for user database operations
	roles          models.RoleModelInterface          // Interface for role database operations
	permissions    models.PermissionModelInterface    // Interface for permission database operations
	emailService   *email.Service                     // Email service for sending invoices
	templateCache  map[string]*template.Template      // Cache of compiled HTML templates (map is like a dictionary: key -> value)
	formDecoder    *form.Decoder                      // Decoder for parsing HTML form data into Go structs
	sessionManager *scs.SessionManager                // Session manager for handling user sessions and cookies
}

// getDefaultDatabasePath returns the default database path in ~/FreelanceTracker/
// In Go, functions can return multiple values - here we return both a string and an error.
// The error return pattern is very common in Go for handling potential failures.
func getDefaultDatabasePath() (string, error) {
	// In Go, := declares a new variable and assigns a value in one step (short declaration).
	// Many functions return (result, error) - you check if err != nil to see if something went wrong.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// %w wraps the error, preserving the original error while adding context
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// filepath.Join combines path segments in a cross-platform way (handles / vs \ on different OS)
	dbDir := filepath.Join(homeDir, "FreelanceTracker")
	dbPath := filepath.Join(dbDir, "freelance_tracker.db")

	return dbPath, nil
}

// ensureDirectoryExists creates the directory if it doesn't exist
func ensureDirectoryExists(path string) error {
	// filepath.Dir extracts the directory part of a path (e.g., "/home/user/file.db" -> "/home/user")
	dir := filepath.Dir(path)
	// MkdirAll creates the directory and any necessary parent directories (like mkdir -p)
	// 0755 sets permissions: owner can read/write/execute, others can read/execute
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}

// main is the entry point of the program. Every Go program must have a main function.
func main() {
	// Get default database path
	defaultDBPath, err := getDefaultDatabasePath()
	if err != nil {
		// Fallback to current directory if we can't get home dir
		defaultDBPath = "./freelance_tracker.db"
	}

	// flag.String creates a command-line flag (like --addr=:8080 or --dsn=/path/to/db)
	// It returns a pointer (*string) to where the flag value will be stored
	addr := flag.String("addr", ":8080", "http service address")
	dsn := flag.String("dsn", defaultDBPath, "SQLite database file path")
	// flag.Parse() reads the command-line arguments and populates the flag variables
	flag.Parse()

	// Create a structured JSON logger that writes to standard output (console)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Ensure the database directory exists
	// *dsn dereferences the pointer to get the actual string value
	if err := ensureDirectoryExists(*dsn); err != nil {
		logger.Error("Failed to create database directory", "error", err.Error())
		// os.Exit(1) terminates the program with exit code 1 (indicating an error)
		os.Exit(1)
	}

	// Log configuration info - *dsn dereferences the pointer to get the value
	logger.Info("Database configuration", "path", *dsn)

	// Open SQLite database connection
	db, err := database.OpenDB(*dsn)
	if err != nil {
		logger.Error("Failed to open database", "error", err.Error())
		os.Exit(1)
	}
	// defer schedules db.Close() to run when main() exits (cleanup pattern)
	// This ensures the database connection is properly closed even if there's an error
	defer db.Close()

	// Run database migrations to set up or update the database schema
	if database.RunMigrations(db, "./migrations"); err != nil {
		logger.Error("Failed to run migrations", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("Database initialized", "dsn", *dsn)

	// Pre-compile all HTML templates into a cache for faster rendering
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("Failed to create template cache", "error", err.Error())
		os.Exit(1)
	}
	// Create a decoder to parse HTML form data into Go structs
	formDecoder := form.NewDecoder()

	// Set up session management for user login/logout
	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)           // Store sessions in SQLite database
	sessionManager.Lifetime = 12 * time.Hour               // Sessions expire after 12 hours
	sessionManager.Cookie.Name = "session"                 // Explicit cookie name
	sessionManager.Cookie.HttpOnly = true                  // Prevent JavaScript access to cookie (security)
	sessionManager.Cookie.Persist = true                   // Sessions persist across browser restarts
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode // Allow cookie in same-site and safe cross-site requests
	// Note: Secure flag is auto-detected from X-Forwarded-Proto header (set by fly.io)

	// Create encryption seed from DSN for consistent key generation
	encryptionSeed := "FreelanceTrackerGo-" + *dsn

	// Create model instances for each database table
	// These wrap the database connection and provide methods for database operations
	clientModel := models.NewClientModel(db)
	projectModel := models.NewProjectModel(db)
	timesheetModel := models.NewTimesheetModel(db)
	invoiceModel := models.NewInvoiceModel(db)
	settingModel := models.NewAppSettingModel(db, encryptionSeed)
	userModel := models.NewUserModel(db)
	roleModel := models.NewRoleModel(db)
	permissionModel := models.NewPermissionModel(db)
	logger.Info("Using SQLite models")

	// Initialize email service for sending invoices
	emailService, err := email.NewServiceFromSettings(settingModel)
	if err != nil {
		// If email setup fails, create a disabled service so the app can still run
		logger.Warn("Failed to initialize email service", "error", err.Error())
		emailService = email.NewService(email.Config{Enabled: false})
	}

	// &application{} creates a new application struct and returns a pointer to it
	// This struct holds all our dependencies (dependency injection pattern)
	app := &application{
		logger:         logger,
		clients:        clientModel,
		projects:       projectModel,
		timesheets:     timesheetModel,
		invoices:       invoiceModel,
		settings:       settingModel,
		emailService:   emailService,
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		users:          userModel,
		roles:          roleModel,
		permissions:    permissionModel,
	}

	// Configure TLS (HTTPS) settings for security
	// []tls.CurveID is a slice (dynamic array) of curve IDs for encryption
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256}, // Preferred encryption curves
	}

	// Configure the HTTP server with our settings
	srv := &http.Server{
		Addr:         *addr,                                          // Address to listen on (e.g., ":8080")
		Handler:      app.routes(),                                   // Router that handles all HTTP requests
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError), // Log server errors
		TLSConfig:    tlsConfig,                                      // TLS configuration for HTTPS
		IdleTimeout:  time.Minute,                                    // Close idle connections after 1 minute
		ReadTimeout:  5 * time.Second,                                // Max time to read request (prevents slow clients from holding connections)
		WriteTimeout: 10 * time.Second,                               // Max time to write response
	}

	logger.Info("Starting server", slog.String("addr", srv.Addr))

	// Start the HTTP server (uncomment line below for HTTPS with TLS certificates)
	//err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	err = srv.ListenAndServe() // Start HTTP server (blocking call - program waits here)
	if err != nil {
		logger.Error("error starting server", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
