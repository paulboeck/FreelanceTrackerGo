package models

import (
	"testing"

	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
)

func TestAppSettingModel_GetString(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	model := NewAppSettingModel(testDB.DB)

	// Test getting a string setting that should exist after migration
	value, err := model.GetString("invoice_title")
	if err != nil {
		t.Fatalf("Expected to get invoice_title setting, got error: %v", err)
	}

	if value == "" {
		t.Error("Expected non-empty invoice_title, got empty string")
	}
}

func TestAppSettingModel_GetDecimal(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	model := NewAppSettingModel(testDB.DB)

	// Test getting a decimal setting
	rate, err := model.GetDecimal("default_hourly_rate")
	if err != nil {
		t.Fatalf("Expected to get default_hourly_rate setting, got error: %v", err)
	}

	if rate <= 0 {
		t.Errorf("Expected positive hourly rate, got %f", rate)
	}
}

func TestAppSettingModel_GetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	model := NewAppSettingModel(testDB.DB)

	// Test getting all settings
	settings, err := model.GetAll()
	if err != nil {
		t.Fatalf("Expected to get all settings, got error: %v", err)
	}

	if len(settings) == 0 {
		t.Error("Expected at least some settings, got none")
	}

	// Verify we can access default_hourly_rate from the map
	if rate, exists := settings["default_hourly_rate"]; exists {
		rateValue, err := rate.AsDecimal()
		if err != nil {
			t.Errorf("Expected to convert rate to decimal, got error: %v", err)
		}
		if rateValue <= 0 {
			t.Errorf("Expected positive rate, got %f", rateValue)
		}
	} else {
		t.Error("Expected default_hourly_rate in settings map")
	}
}

func TestAppSettingModel_UpdateValue(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	model := NewAppSettingModel(testDB.DB)

	// Update a setting value
	err := model.UpdateValue("default_hourly_rate", "95.00")
	if err != nil {
		t.Fatalf("Expected to update setting value, got error: %v", err)
	}

	// Verify the update
	rate, err := model.GetDecimal("default_hourly_rate")
	if err != nil {
		t.Fatalf("Expected to get updated setting, got error: %v", err)
	}

	if rate != 95.00 {
		t.Errorf("Expected rate to be 95.00, got %f", rate)
	}
}