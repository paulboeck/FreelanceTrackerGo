package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

// paginationData holds information needed to render pagination controls in templates.
type paginationData struct {
	CurrentPage int
	TotalPages  int
	HasPrev     bool // Whether there's a previous page
	HasNext     bool // Whether there's a next page
	PrevPage    int
	NextPage    int
	PageSize    int  // Number of items per page
}

// templateData is a struct that holds all the dynamic data passed to HTML templates.
// Not all fields are used in every template - only the relevant ones are populated.
type templateData struct {
	CurrentYear        int
	Client             *models.Client                // Pointer allows nil (no client) or a single client
	Clients            []models.Client               // Slice (dynamic array) of multiple clients
	Project            *models.Project
	Projects           []models.Project
	ProjectsWithClient []models.ProjectWithClient
	Timesheets         []models.Timesheet
	Invoices           []models.Invoice
	InvoicesWithProject []models.InvoiceWithProject
	Settings           []models.AppSetting
	Form               any                          // 'any' (interface{}) accepts any type - for form data
	Pagination         *paginationData
	Flash              string                       // One-time message to display (e.g., "Client created successfully")
	IsAuthenticated    bool
	SearchTerm         string
	// Additional fields for income report
	Total              float64
	Year               int
}

// humanDate formats a time.Time into a human-readable string.
// Go's time formatting uses a reference date (Jan 2, 2006 at 3:04 PM) as a template.
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// currencySymbol returns the currency symbol for a project (e.g., "$", "€").
// These functions are helpers that can be called from HTML templates.
func currencySymbol(project models.Project) string {
	return project.CurrencySymbol()
}

// currencyDisplayOnInvoice returns the full currency display for invoices.
func currencyDisplayOnInvoice(project models.Project) string {
	return project.CurrencyDisplayOnInvoice()
}

// isPositive checks if a number is greater than zero.
func isPositive(value float64) bool {
	return value > 0
}

// isNonZero checks if a number is not equal to zero.
func isNonZero(value float64) bool {
	return value != 0
}

// mul multiplies two numbers (used in templates for calculations).
func mul(a, b float64) float64 {
	return a * b
}

// formatDiscountPercent formats a discount percentage for display.
// The *float64 parameter is a pointer, which can be nil if no discount exists.
func formatDiscountPercent(discount *float64) string {
	if discount == nil {
		return "0.0000"
	}
	// fmt.Sprintf formats a string (like printf in C)
	// %.4f means float with 4 decimal places, *discount dereferences the pointer
	return fmt.Sprintf("%.4f", *discount)
}

// formatAdjustmentAmount formats an adjustment amount for display.
func formatAdjustmentAmount(adjustment *float64) string {
	if adjustment == nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", *adjustment)
}

// functions is a FuncMap that registers custom functions to use in templates.
// var declares a package-level variable that's accessible throughout this package.
// template.FuncMap is a map[string]interface{} that maps function names to functions.
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

// newTemplateCache creates a cache of all compiled HTML templates.
// Caching templates on startup improves performance vs. parsing them on every request.
func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize an empty map to store compiled templates
	// map[string]*template.Template means: string keys -> template pointers values
	cache := map[string]*template.Template{}

	// filepath.Glob finds all files matching a pattern (like *.html)
	// Returns a slice of file paths
	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	// For each page, create a template set containing the base html, all partials, and the page itself
	// range iterates over a slice; underscore (_) discards the index, page is the value
	for _, page := range pages {
		// Extract just the filename from the full path (e.g., "/path/to/home.html" -> "home.html")
		name := filepath.Base(page)

		// Create a new template with custom functions, then parse the base template
		// Method chaining: New() returns a template, Funcs() adds functions, ParseFiles() parses files
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		// Parse all partial templates (header, footer, etc.) into this template set
		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			// Empty error handler - probably should handle this error
		}
		// Parse the specific page template
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Store the compiled template set in the cache with the page name as key
		cache[name] = ts
	}

	return cache, nil
}
