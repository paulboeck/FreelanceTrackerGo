package main

import (
	"crypto/tls"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"

	"github.com/paulboeck/FreelanceTrackerGo/internal/database"
	"github.com/paulboeck/FreelanceTrackerGo/internal/email"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

type application struct {
	logger         *slog.Logger
	clients        models.ClientModelInterface
	projects       models.ProjectModelInterface
	timesheets     models.TimesheetModelInterface
	invoices       models.InvoiceModelInterface
	settings       models.AppSettingModelInterface
	users          models.UserModelInterface
	emailService   *email.Service
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":8080", "http service address")
	dsn := flag.String("dsn", "./freelance_tracker.db", "SQLite database file path")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Open SQLite database
	db, err := database.OpenDB(*dsn)
	if err != nil {
		logger.Error("Failed to open database", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if database.RunMigrations(db, "./migrations"); err != nil {
		logger.Error("Failed to run migrations", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("Database initialized", "dsn", *dsn)

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("Failed to create template cache", "error", err.Error())
		os.Exit(1)
	}
	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	// Create encryption seed from DSN for consistent key generation
	encryptionSeed := "FreelanceTrackerGo-" + *dsn
	
	// Create SQLite models
	clientModel := models.NewClientModel(db)
	projectModel := models.NewProjectModel(db)
	timesheetModel := models.NewTimesheetModel(db)
	invoiceModel := models.NewInvoiceModel(db)
	settingModel := models.NewAppSettingModel(db, encryptionSeed)
	userModel := models.NewUserModel(db)
	logger.Info("Using SQLite models")

	// Initialize email service
	emailService, err := email.NewServiceFromSettings(settingModel)
	if err != nil {
		logger.Warn("Failed to initialize email service", "error", err.Error())
		emailService = email.NewService(email.Config{Enabled: false})
	}

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
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("Starting server", slog.String("addr", srv.Addr))

	//err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	err = srv.ListenAndServe()
	if err != nil {
		logger.Error("error starting server", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
