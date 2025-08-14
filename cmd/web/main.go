package main

import (
	"flag"
	"github.com/go-playground/form/v4"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/paulboeck/FreelanceTrackerGo/internal/database"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

type application struct {
	logger        *slog.Logger
	clients       models.ClientModelInterface
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
}

func main() {
	addr := flag.String("addr", ":8080", "http service address")
	
	// Determine database type from environment or default to SQLite
	dbType := database.GetDatabaseTypeFromEnv()
	defaultDSN := database.GetDefaultDSN(dbType)
	
	var dsnDescription string
	switch dbType {
	case database.DatabaseTypeMySQL:
		dsnDescription = "MySQL data source name"
	case database.DatabaseTypeSQLite:
		dsnDescription = "SQLite database file path"
	}
	
	dsn := flag.String("dsn", defaultDSN, dsnDescription)
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Open database with type-aware configuration
	dbConfig := database.Config{
		Type: dbType,
		DSN:  *dsn,
	}
	
	db, err := database.OpenDB(dbConfig)
	if err != nil {
		logger.Error("Failed to open database", "error", err.Error(), "type", dbType)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db, dbType, "./migrations"); err != nil {
		logger.Error("Failed to run migrations", "error", err.Error())
		os.Exit(1)
	}
	
	logger.Info("Database initialized", "type", dbType, "dsn", *dsn)

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("Failed to create template cache", "error", err.Error())
		os.Exit(1)
	}
	formDecoder := form.NewDecoder()

	// Choose client implementation based on database type
	var clientModel models.ClientModelInterface
	switch dbType {
	case database.DatabaseTypeSQLite:
		// Use the new SQLC-generated adapter for SQLite
		clientModel = models.NewClientAdapter(db)
		logger.Info("Using SQLC-generated client adapter")
	case database.DatabaseTypeMySQL:
		// Use the original implementation for MySQL (for now)
		clientModel = &models.ClientModel{DB: db}
		logger.Info("Using original client model")
	default:
		logger.Error("Unsupported database type", "type", dbType)
		os.Exit(1)
	}

	app := &application{
		logger:        logger,
		clients:       clientModel,
		templateCache: templateCache,
		formDecoder:   formDecoder,
	}

	logger.Info("Starting server", slog.String("addr", *addr))

	err = http.ListenAndServe(*addr, app.routes())
	if err != nil {
		logger.Error("error starting server", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

