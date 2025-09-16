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
	model := NewAppSettingModel(testDB.DB, "test-encryption-seed")

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
	model := NewAppSettingModel(testDB.DB, "test-encryption-seed")

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
	model := NewAppSettingModel(testDB.DB, "test-encryption-seed")

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
	model := NewAppSettingModel(testDB.DB, "test-encryption-seed")

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

func TestAppSettingModel_BooleanPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	model := NewAppSettingModel(testDB.DB, "test-encryption-seed")

	// Test setting boolean to true
	err := model.UpdateValue("email_enabled", "true")
	if err != nil {
		t.Fatalf("Expected to update boolean setting to true, got error: %v", err)
	}

	enabled, err := model.GetBool("email_enabled")
	if err != nil {
		t.Fatalf("Expected to get boolean setting, got error: %v", err)
	}

	if !enabled {
		t.Error("Expected email_enabled to be true")
	}

	// Test setting boolean to false
	err = model.UpdateValue("email_enabled", "false")
	if err != nil {
		t.Fatalf("Expected to update boolean setting to false, got error: %v", err)
	}

	enabled, err = model.GetBool("email_enabled")
	if err != nil {
		t.Fatalf("Expected to get boolean setting, got error: %v", err)
	}

	if enabled {
		t.Error("Expected email_enabled to be false")
	}
}

func TestAppSettingModel_SMTPPasswordEncryption(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	model := NewAppSettingModel(testDB.DB, "test-encryption-seed")

	plainPassword := "my-secret-gmail-password"

	// Update SMTP password
	err := model.UpdateValue("smtp_password", plainPassword)
	if err != nil {
		t.Fatalf("Expected to update SMTP password, got error: %v", err)
	}

	// Verify the raw stored value is encrypted (not plaintext)
	storedPassword, err := model.GetString("smtp_password")
	if err != nil {
		t.Fatalf("Expected to get stored password, got error: %v", err)
	}

	if storedPassword == plainPassword {
		t.Error("Expected stored password to be encrypted, but it matches plaintext")
	}

	if storedPassword == "" {
		t.Error("Expected stored password to be non-empty encrypted value")
	}

	// Verify we can decrypt it back to the original
	decryptedPassword, err := model.GetDecryptedSMTPPassword()
	if err != nil {
		t.Fatalf("Expected to decrypt SMTP password, got error: %v", err)
	}

	if decryptedPassword != plainPassword {
		t.Errorf("Expected decrypted password to be %q, got %q", plainPassword, decryptedPassword)
	}
}

func TestAppSettingModel_SMTPPasswordEmptyHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	model := NewAppSettingModel(testDB.DB, "test-encryption-seed")

	// Set an initial password
	originalPassword := "initial-password"
	err := model.UpdateValue("smtp_password", originalPassword)
	if err != nil {
		t.Fatalf("Expected to set initial password, got error: %v", err)
	}

	// Verify password is set
	decrypted, err := model.GetDecryptedSMTPPassword()
	if err != nil {
		t.Fatalf("Expected to decrypt initial password, got error: %v", err)
	}
	if decrypted != originalPassword {
		t.Errorf("Expected initial password %q, got %q", originalPassword, decrypted)
	}

	// Try to update with empty password (should be ignored)
	err = model.UpdateValue("smtp_password", "")
	if err != nil {
		t.Fatalf("Expected empty password update to succeed (as no-op), got error: %v", err)
	}

	// Verify password is unchanged
	decrypted, err = model.GetDecryptedSMTPPassword()
	if err != nil {
		t.Fatalf("Expected to decrypt password after empty update, got error: %v", err)
	}
	if decrypted != originalPassword {
		t.Errorf("Expected password to remain %q after empty update, got %q", originalPassword, decrypted)
	}
}

func TestAppSettingModel_SMTPPasswordNoDoubleEncryption(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	model := NewAppSettingModel(testDB.DB, "test-encryption-seed")

	originalPassword := "test-password-123"

	// Set password first time
	err := model.UpdateValue("smtp_password", originalPassword)
	if err != nil {
		t.Fatalf("Expected to set password first time, got error: %v", err)
	}

	// Update with the same password again
	err = model.UpdateValue("smtp_password", originalPassword)
	if err != nil {
		t.Fatalf("Expected to update password again, got error: %v", err)
	}

	// Both should decrypt correctly to the original password
	decrypted, err := model.GetDecryptedSMTPPassword()
	if err != nil {
		t.Fatalf("Expected to decrypt password, got error: %v", err)
	}

	if decrypted != originalPassword {
		t.Errorf("Expected decrypted password to be %q, got %q", originalPassword, decrypted)
	}

	// Password should still be correctly decryptable after multiple updates
}

func TestAppSettingModel_MultipleSettingsUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)
	model := NewAppSettingModel(testDB.DB, "test-encryption-seed")

	// Update multiple settings
	settings := map[string]string{
		"email_enabled":  "true",
		"smtp_host":      "smtp.gmail.com",
		"smtp_port":      "587",
		"smtp_username":  "freelancer@example.com",
		"smtp_password":  "secret-app-password",
		"smtp_from_name": "FreelanceTracker",
		"smtp_use_tls":   "true",
	}

	// Update all settings
	for key, value := range settings {
		err := model.UpdateValue(key, value)
		if err != nil {
			t.Fatalf("Expected to update setting %s, got error: %v", key, err)
		}
	}

	// Verify all settings were persisted correctly
	emailEnabled, err := model.GetBool("email_enabled")
	if err != nil {
		t.Fatalf("Expected to get email_enabled, got error: %v", err)
	}
	if !emailEnabled {
		t.Error("Expected email_enabled to be true")
	}

	smtpHost, err := model.GetString("smtp_host")
	if err != nil {
		t.Fatalf("Expected to get smtp_host, got error: %v", err)
	}
	if smtpHost != "smtp.gmail.com" {
		t.Errorf("Expected smtp_host to be 'smtp.gmail.com', got %q", smtpHost)
	}

	smtpPort, err := model.GetInt("smtp_port")
	if err != nil {
		t.Fatalf("Expected to get smtp_port, got error: %v", err)
	}
	if smtpPort != 587 {
		t.Errorf("Expected smtp_port to be 587, got %d", smtpPort)
	}

	smtpUsername, err := model.GetString("smtp_username")
	if err != nil {
		t.Fatalf("Expected to get smtp_username, got error: %v", err)
	}
	if smtpUsername != "freelancer@example.com" {
		t.Errorf("Expected smtp_username to be 'freelancer@example.com', got %q", smtpUsername)
	}

	// SMTP password should be decryptable
	smtpPassword, err := model.GetDecryptedSMTPPassword()
	if err != nil {
		t.Fatalf("Expected to decrypt SMTP password, got error: %v", err)
	}
	if smtpPassword != "secret-app-password" {
		t.Errorf("Expected SMTP password to be 'secret-app-password', got %q", smtpPassword)
	}

	fromName, err := model.GetString("smtp_from_name")
	if err != nil {
		t.Fatalf("Expected to get smtp_from_name, got error: %v", err)
	}
	if fromName != "FreelanceTracker" {
		t.Errorf("Expected smtp_from_name to be 'FreelanceTracker', got %q", fromName)
	}

	useTLS, err := model.GetBool("smtp_use_tls")
	if err != nil {
		t.Fatalf("Expected to get smtp_use_tls, got error: %v", err)
	}
	if !useTLS {
		t.Error("Expected smtp_use_tls to be true")
	}
}
