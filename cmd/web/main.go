package main

import (
	"database/sql"
	"flag"
	"github.com/go-playground/form/v4"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	logger        *slog.Logger
	clients       *models.ClientModel
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
}

func main() {
	addr := flag.String("addr", ":8080", "http service address")
	// Note: we need parseTime=true so the MySQL driver converts TIME and DATE fields correctly to time.Time
	dsn := flag.String("dsn", "root:root@/freelance_tracker?parseTime=true", "MySQL data source name")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	formDecoder := form.NewDecoder()

	app := &application{
		logger:        logger,
		clients:       &models.ClientModel{DB: db},
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

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
