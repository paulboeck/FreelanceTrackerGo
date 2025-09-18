package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

type paginationData struct {
	CurrentPage int
	TotalPages  int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	PageSize    int
}

type templateData struct {
	CurrentYear        int
	Client             *models.Client
	Clients            []models.Client
	Project            *models.Project
	Projects           []models.Project
	ProjectsWithClient []models.ProjectWithClient
	Timesheets         []models.Timesheet
	Invoices           []models.Invoice
	Settings           []models.AppSetting
	Form               any
	Pagination         *paginationData
	Flash              string
	IsAuthenticated    bool
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

func currencySymbol(project models.Project) string {
	return project.CurrencySymbol()
}

func currencyDisplayOnInvoice(project models.Project) string {
	return project.CurrencyDisplayOnInvoice()
}

func isPositive(value float64) bool {
	return value > 0
}

func isNonZero(value float64) bool {
	return value != 0
}

func mul(a, b float64) float64 {
	return a * b
}

func formatDiscountPercent(discount *float64) string {
	if discount == nil {
		return "0.0000"
	}
	return fmt.Sprintf("%.4f", *discount)
}

func formatAdjustmentAmount(adjustment *float64) string {
	if adjustment == nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", *adjustment)
}

var functions = template.FuncMap{
	"humanDate":                humanDate,
	"currencySymbol":           currencySymbol,
	"currencyDisplayOnInvoice": currencyDisplayOnInvoice,
	"isPositive":               isPositive,
	"isNonZero":                isNonZero,
	"mul":                      mul,
	"formatDiscountPercent":    formatDiscountPercent,
	"formatAdjustmentAmount":   formatAdjustmentAmount,
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
