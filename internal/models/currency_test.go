package models

import (
	"testing"
)

func TestProject_TotalAmountDue(t *testing.T) {
	project := Project{
		HourlyRate: 50.0,
	}

	timesheets := []Timesheet{
		{HoursWorked: 2.0, HourlyRate: 50.0},
		{HoursWorked: 3.0, HourlyRate: 60.0},
		{HoursWorked: 1.5, HourlyRate: 40.0},
	}

	total := project.TotalAmountDue(timesheets)
	expected := 2.0*50.0 + 3.0*60.0 + 1.5*40.0 // 100 + 180 + 60 = 340
	
	if total != expected {
		t.Errorf("Expected total amount due to be %.2f, got %.2f", expected, total)
	}
}

func TestProject_ConvertedTotalAmountDue(t *testing.T) {
	project := Project{
		HourlyRate:             50.0,
		CurrencyConversionRate: 1.2, // Convert USD to EUR with 1.2 rate
	}

	timesheets := []Timesheet{
		{HoursWorked: 2.0, HourlyRate: 50.0},
		{HoursWorked: 3.0, HourlyRate: 60.0},
	}

	total := project.ConvertedTotalAmountDue(timesheets)
	expected := (2.0*50.0 + 3.0*60.0) * 1.2 // (100 + 180) * 1.2 = 336
	
	if total != expected {
		t.Errorf("Expected converted total amount due to be %.2f, got %.2f", expected, total)
	}
}

func TestProject_ConvertedHourlyRate(t *testing.T) {
	project := Project{
		HourlyRate:             50.0,
		CurrencyConversionRate: 1.3,
	}

	convertedRate := project.ConvertedHourlyRate()
	expected := 50.0 * 1.3 // 65.0
	
	if convertedRate != expected {
		t.Errorf("Expected converted hourly rate to be %.2f, got %.2f", expected, convertedRate)
	}
}

func TestProject_DiscountAmount(t *testing.T) {
	discountPercent := 10.0
	project := Project{
		DiscountPercent: &discountPercent,
	}

	subtotal := 1000.0
	discount := project.DiscountAmount(subtotal)
	expected := 100.0 // 10% of 1000
	
	if discount != expected {
		t.Errorf("Expected discount amount to be %.2f, got %.2f", expected, discount)
	}

	// Test with nil discount
	project.DiscountPercent = nil
	discount = project.DiscountAmount(subtotal)
	if discount != 0.0 {
		t.Errorf("Expected discount amount to be 0.0 when no discount, got %.2f", discount)
	}
}

func TestProject_ConvertedAdjustmentAmount(t *testing.T) {
	adjustmentAmount := 50.0
	project := Project{
		AdjustmentAmount:       &adjustmentAmount,
		CurrencyConversionRate: 1.2,
	}

	converted := project.ConvertedAdjustmentAmount()
	expected := 50.0 * 1.2 // 60.0
	
	if converted != expected {
		t.Errorf("Expected converted adjustment amount to be %.2f, got %.2f", expected, converted)
	}

	// Test with nil adjustment
	project.AdjustmentAmount = nil
	converted = project.ConvertedAdjustmentAmount()
	if converted != 0.0 {
		t.Errorf("Expected converted adjustment amount to be 0.0 when no adjustment, got %.2f", converted)
	}
}

func TestProject_AdjustedAmountDue(t *testing.T) {
	discountPercent := 10.0
	adjustmentAmount := -25.0
	project := Project{
		DiscountPercent:  &discountPercent,
		AdjustmentAmount: &adjustmentAmount,
	}

	timesheets := []Timesheet{
		{HoursWorked: 10.0, HourlyRate: 50.0}, // 500.0
	}

	adjusted := project.AdjustedAmountDue(timesheets)
	// 500 - (500 * 0.1) - (-25) = 500 - 50 + 25 = 475
	expected := 475.0
	
	if adjusted != expected {
		t.Errorf("Expected adjusted amount due to be %.2f, got %.2f", expected, adjusted)
	}
}

func TestProject_CurrencySymbol(t *testing.T) {
	tests := []struct {
		currency string
		expected string
	}{
		{"USD", "$"},
		{"EUR", "€"},
		{"GBP", "£"},
		{"CAD", "C$"},
		{"XYZ", "$"}, // Unknown currency defaults to USD
	}

	for _, tt := range tests {
		project := Project{CurrencyDisplay: tt.currency}
		symbol := project.CurrencySymbol()
		if symbol != tt.expected {
			t.Errorf("For currency %s, expected symbol %s, got %s", tt.currency, tt.expected, symbol)
		}
	}
}

func TestProject_CurrencyDisplayOnInvoice(t *testing.T) {
	tests := []struct {
		currency string
		expected string
	}{
		{"USD", "USD $"},
		{"EUR", "EUR "},
		{"GBP", "GBP "},
		{"CAD", "CAD "},
	}

	for _, tt := range tests {
		project := Project{CurrencyDisplay: tt.currency}
		display := project.CurrencyDisplayOnInvoice()
		if display != tt.expected {
			t.Errorf("For currency %s, expected display %s, got %s", tt.currency, tt.expected, display)
		}
	}
}

func TestTimesheet_ConvertedHourlyRate(t *testing.T) {
	project := Project{
		CurrencyConversionRate: 1.5,
	}

	timesheet := Timesheet{
		HourlyRate: 40.0,
	}

	converted := timesheet.ConvertedHourlyRate(project)
	expected := 40.0 * 1.5 // 60.0
	
	if converted != expected {
		t.Errorf("Expected converted hourly rate to be %.2f, got %.2f", expected, converted)
	}
}

func TestTimesheet_ConvertedTotal(t *testing.T) {
	project := Project{
		CurrencyConversionRate: 1.3,
	}

	timesheet := Timesheet{
		HoursWorked: 2.5,
		HourlyRate:  50.0,
	}

	total := timesheet.ConvertedTotal(project)
	expected := 2.5 * (50.0 * 1.3) // 2.5 * 65 = 162.5
	
	if total != expected {
		t.Errorf("Expected converted total to be %.2f, got %.2f", expected, total)
	}
}

func TestTimesheet_Total(t *testing.T) {
	timesheet := Timesheet{
		HoursWorked: 3.0,
		HourlyRate:  75.0,
	}

	total := timesheet.Total()
	expected := 3.0 * 75.0 // 225.0
	
	if total != expected {
		t.Errorf("Expected total to be %.2f, got %.2f", expected, total)
	}
}

func TestInvoice_ConvertedAmountDue(t *testing.T) {
	project := Project{
		CurrencyConversionRate: 1.25,
	}

	invoice := Invoice{
		AmountDue: 800.0,
	}

	converted := invoice.ConvertedAmountDue(project)
	expected := 800.0 * 1.25 // 1000.0
	
	if converted != expected {
		t.Errorf("Expected converted amount due to be %.2f, got %.2f", expected, converted)
	}
}