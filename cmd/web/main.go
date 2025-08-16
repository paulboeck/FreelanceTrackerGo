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
	if err := database.RunMigrations(db, "./migrations"); err != nil {
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

	// Create SQLite client model
	clientModel := models.NewClientModel(db)
	logger.Info("Using SQLite client model")

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

