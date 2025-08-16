package main

import (
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"html/template"
	"path/filepath"
	"time"
)

type templateData struct {
	CurrentYear int
	Client      *models.Client
	Clients     []models.Client
	Project     *models.Project
	Projects    []models.Project
	Form        any
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	// For each page, create a template set containing the base html, all partials, and the page itself
	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {

		}
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
