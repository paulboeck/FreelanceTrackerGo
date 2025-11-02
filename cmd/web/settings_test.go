package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createCompleteSettingsForm creates a form with all required settings fields
// This helper ensures tests send complete form data that matches what the handler expects
func createCompleteSettingsForm() url.Values {
	form := url.Values{}
	// Required settings that must be in every form submission
	form.Add("default_hourly_rate", "85.00")
	form.Add("invoice_title", "Invoice for Academic Editing")
	form.Add("freelancer_name", "Your Name Here")
	form.Add("freelancer_address", "Your Address")
	form.Add("freelancer_phone", "Your Phone")
	form.Add("freelancer_email", "your.email@example.com")
	form.Add("email_enabled", "false")
	form.Add("smtp_host", "")
	form.Add("smtp_port", "587")
	form.Add("smtp_username", "")
	form.Add("smtp_password", "")
	form.Add("smtp_from_name", "Freelance Tracker")
	form.Add("smtp_use_tls", "true")
	return form
}

func TestSettingsView(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rr := httptest.NewRecorder()

	// Use the session middleware to wrap the handler
	handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsView))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, "Application Settings")
	assert.Contains(t, body, "Edit Setting Values")
}

func TestSettingsEdit(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	req := httptest.NewRequest(http.MethodGet, "/settings/edit", nil)
	rr := httptest.NewRecorder()

	// Use the session middleware to wrap the handler
	handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEdit))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(t, body, "Email Settings")
	assert.Contains(t, body, "Other Settings")
	assert.Contains(t, body, "<form action=\"/settings/edit\"")
}

func TestSettingsEditPost_BooleanPersistence(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("boolean setting persists correctly", func(t *testing.T) {
		// Test updating a boolean setting to true
		form := createCompleteSettingsForm()
		form.Set("email_enabled", "true")
		// When email is enabled, SMTP fields are required
		form.Set("smtp_host", "smtp.gmail.com")
		form.Set("smtp_username", "test@gmail.com")

		req := httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to wrap the handler
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		// Should redirect after successful update
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		location := rr.Header().Get("Location")
		assert.Equal(t, "/settings", location)

		// Verify the setting was actually updated in the database
		enabled, err := app.settings.GetBool("email_enabled")
		require.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("boolean setting can be set to false", func(t *testing.T) {
		// First set it to true
		err := app.settings.UpdateValue("email_enabled", "true")
		require.NoError(t, err)

		// Now test updating to false
		form := createCompleteSettingsForm()
		form.Set("email_enabled", "false")

		req := httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to wrap the handler
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		// Should redirect after successful update
		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// Verify the setting was actually updated to false
		enabled, err := app.settings.GetBool("email_enabled")
		require.NoError(t, err)
		assert.False(t, enabled)
	})
}

func TestSettingsEditPost_SMTPPasswordEncryption(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("SMTP password is encrypted when stored", func(t *testing.T) {
		plainPassword := "my-secret-gmail-app-password"

		form := createCompleteSettingsForm()
		form.Set("email_enabled", "true")
		form.Set("smtp_password", plainPassword)
		form.Set("smtp_username", "test@gmail.com")
		form.Set("smtp_host", "smtp.gmail.com")

		req := httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to wrap the handler
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		// Should redirect after successful update
		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// Verify the password is stored encrypted (not plain text)
		storedPassword, err := app.settings.GetString("smtp_password")
		require.NoError(t, err)
		assert.NotEmpty(t, storedPassword)
		assert.NotEqual(t, plainPassword, storedPassword) // Should be encrypted, not plaintext

		// Verify we can decrypt it back to original
		decryptedPassword, err := app.settings.GetDecryptedSMTPPassword()
		require.NoError(t, err)
		assert.Equal(t, plainPassword, decryptedPassword)
	})

	t.Run("empty SMTP password field preserves existing password", func(t *testing.T) {
		// First set a password
		originalPassword := "original-password"
		form := createCompleteSettingsForm()
		form.Set("email_enabled", "true")
		form.Set("smtp_password", originalPassword)
		form.Set("smtp_username", "test@gmail.com")
		form.Set("smtp_host", "smtp.gmail.com")

		req := httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusSeeOther, rr.Code)

		// Verify original password is stored
		decryptedPassword, err := app.settings.GetDecryptedSMTPPassword()
		require.NoError(t, err)
		assert.Equal(t, originalPassword, decryptedPassword)

		// Now submit form with empty password field
		form = createCompleteSettingsForm()
		form.Set("email_enabled", "true")
		form.Set("smtp_password", "") // Empty password - should preserve existing
		form.Set("smtp_username", "updated@gmail.com")
		form.Set("smtp_host", "smtp.gmail.com")

		req = httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr = httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusSeeOther, rr.Code)

		// Verify password is unchanged (kept current password)
		decryptedPassword, err = app.settings.GetDecryptedSMTPPassword()
		require.NoError(t, err)
		assert.Equal(t, originalPassword, decryptedPassword)

		// But username should be updated
		username, err := app.settings.GetString("smtp_username")
		require.NoError(t, err)
		assert.Equal(t, "updated@gmail.com", username)
	})

	t.Run("SMTP password no double encryption", func(t *testing.T) {
		// Set initial password
		originalPassword := "test-password-123"
		form := createCompleteSettingsForm()
		form.Set("email_enabled", "true")
		form.Set("smtp_password", originalPassword)
		form.Set("smtp_username", "test@gmail.com")
		form.Set("smtp_host", "smtp.gmail.com")

		req := httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusSeeOther, rr.Code)

		// Get the encrypted value from database
		encryptedValue1, err := app.settings.GetString("smtp_password")
		require.NoError(t, err)

		// Submit same form again (which would previously cause double encryption)
		form = createCompleteSettingsForm()
		form.Set("email_enabled", "true")
		form.Set("smtp_password", originalPassword)
		form.Set("smtp_username", "test@gmail.com")
		form.Set("smtp_host", "smtp.gmail.com")

		req = httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr = httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusSeeOther, rr.Code)

		// Get encrypted value again
		encryptedValue2, err := app.settings.GetString("smtp_password")
		require.NoError(t, err)

		// Values should be different (due to random nonces in AES-GCM) but...
		// ...decryption should still work correctly
		decryptedPassword, err := app.settings.GetDecryptedSMTPPassword()
		require.NoError(t, err)
		assert.Equal(t, originalPassword, decryptedPassword)

		// Both encrypted values should be different due to nonces, but both valid
		assert.NotEqual(t, encryptedValue1, encryptedValue2)
	})
}

func TestSettingsEditPost_MultipleSettings(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("multiple settings updated simultaneously", func(t *testing.T) {
		form := createCompleteSettingsForm()
		form.Set("email_enabled", "true")
		form.Set("smtp_host", "smtp.gmail.com")
		form.Set("smtp_port", "587")
		form.Set("smtp_username", "freelancer@example.com")
		form.Set("smtp_password", "app-password-123")
		form.Set("smtp_from_name", "My Freelance Business")
		form.Set("smtp_use_tls", "true")

		req := httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to wrap the handler
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		// Should redirect after successful update
		assert.Equal(t, http.StatusSeeOther, rr.Code)

		// Verify all settings were updated
		emailEnabled, err := app.settings.GetBool("email_enabled")
		require.NoError(t, err)
		assert.True(t, emailEnabled)

		smtpHost, err := app.settings.GetString("smtp_host")
		require.NoError(t, err)
		assert.Equal(t, "smtp.gmail.com", smtpHost)

		smtpPort, err := app.settings.GetInt("smtp_port")
		require.NoError(t, err)
		assert.Equal(t, 587, smtpPort)

		smtpUsername, err := app.settings.GetString("smtp_username")
		require.NoError(t, err)
		assert.Equal(t, "freelancer@example.com", smtpUsername)

		// Password should be decryptable
		smtpPassword, err := app.settings.GetDecryptedSMTPPassword()
		require.NoError(t, err)
		assert.Equal(t, "app-password-123", smtpPassword)

		fromName, err := app.settings.GetString("smtp_from_name")
		require.NoError(t, err)
		assert.Equal(t, "My Freelance Business", fromName)

		useTLS, err := app.settings.GetBool("smtp_use_tls")
		require.NoError(t, err)
		assert.True(t, useTLS)
	})
}

func TestSettingsEditPost_ValidationErrors(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("invalid port number", func(t *testing.T) {
		form := createCompleteSettingsForm()
		form.Set("smtp_port", "invalid-port")

		req := httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to wrap the handler
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Must be a valid integer")
	})

	t.Run("invalid boolean value", func(t *testing.T) {
		form := createCompleteSettingsForm()
		form.Set("email_enabled", "maybe") // Invalid boolean

		req := httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to wrap the handler
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		// Should return form with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Must be true or false")
	})

	t.Run("empty SMTP password is allowed", func(t *testing.T) {
		form := createCompleteSettingsForm()
		form.Set("email_enabled", "true")
		form.Set("smtp_username", "test@example.com")
		form.Set("smtp_host", "smtp.gmail.com")
		form.Set("smtp_password", "") // Empty password should be allowed

		req := httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Use the session middleware to wrap the handler
		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		// Should succeed without validation error
		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, "/settings", rr.Header().Get("Location"))
	})
}

func TestSettingsIntegration(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("complete settings workflow", func(t *testing.T) {
		// 1. View settings page
		req := httptest.NewRequest(http.MethodGet, "/settings", nil)
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsView))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// 2. Edit settings form
		req = httptest.NewRequest(http.MethodGet, "/settings/edit", nil)
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEdit))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body := rr.Body.String()
		assert.Contains(t, body, "Email Settings") // Grouped settings

		// 3. Update settings
		form := createCompleteSettingsForm()
		form.Set("email_enabled", "true")
		form.Set("smtp_username", "freelancer@gmail.com")
		form.Set("smtp_password", "gmail-app-password-123")
		form.Set("smtp_host", "smtp.gmail.com")

		req = httptest.NewRequest(http.MethodPost, "/settings/edit", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsEditPost))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusSeeOther, rr.Code)
		assert.Equal(t, "/settings", rr.Header().Get("Location"))

		// 4. View updated settings
		req = httptest.NewRequest(http.MethodGet, "/settings", nil)
		rr = httptest.NewRecorder()

		handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.settingsView))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		body = rr.Body.String()
		assert.Contains(t, body, "freelancer@gmail.com") // Username should be visible
		assert.Contains(t, body, "[Password Set]")       // Password should show as set, not plaintext

		// 5. Verify settings work with email service
		emailEnabled, err := app.settings.GetBool("email_enabled")
		require.NoError(t, err)
		assert.True(t, emailEnabled)

		decryptedPassword, err := app.settings.GetDecryptedSMTPPassword()
		require.NoError(t, err)
		assert.Equal(t, "gmail-app-password-123", decryptedPassword)
	})
}