package api

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

// ReportsHandlers handles reports API endpoints
type ReportsHandlers struct {
	invoices *models.InvoiceModel
}

// NewReportsHandlers creates a new ReportsHandlers instance
func NewReportsHandlers(invoices *models.InvoiceModel) *ReportsHandlers {
	return &ReportsHandlers{
		invoices: invoices,
	}
}

// IncomeReportResponse represents the income report response
type IncomeReportResponse struct {
	Year   int                        `json:"year"`
	Total  float64                    `json:"total"`
	Months []MonthlyIncomeResponse    `json:"months"`
}

// MonthlyIncomeResponse represents income for a single month
type MonthlyIncomeResponse struct {
	Month  int     `json:"month"`
	Amount float64 `json:"amount"`
}

// GetIncomeReport returns income report for a specific year
// GET /api/v1/reports/income?year=2024
func (h *ReportsHandlers) GetIncomeReport(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get year from query parameter
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		ErrorBadRequest(w, "Year parameter is required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1900 || year > 2100 {
		ErrorBadRequest(w, "Invalid year format")
		return
	}

	// Get paid invoices for the year
	invoices, err := h.invoices.GetPaidInvoicesForYear(year)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Calculate monthly income
	monthlyIncome := make(map[int]float64)
	var totalIncome float64

	for _, invoice := range invoices {
		if invoice.DatePaid != nil {
			month := int(invoice.DatePaid.Month())
			monthlyIncome[month] += invoice.AmountDue
			totalIncome += invoice.AmountDue
		}
	}

	// Convert to response format
	months := make([]MonthlyIncomeResponse, 0, 12)
	for month := 1; month <= 12; month++ {
		months = append(months, MonthlyIncomeResponse{
			Month:  month,
			Amount: monthlyIncome[month],
		})
	}

	response := IncomeReportResponse{
		Year:   year,
		Total:  totalIncome,
		Months: months,
	}

	WriteJSON(w, http.StatusOK, response, nil)
}
