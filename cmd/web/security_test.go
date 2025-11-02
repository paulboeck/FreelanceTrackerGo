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

// TestOWASP_SQLInjection tests protection against SQL injection attacks (OWASP A03:2021)
func TestOWASP_SQLInjection(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	sqlInjectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE client; --",
		"' UNION SELECT * FROM user --",
		"admin'--",
		"' OR 1=1--",
		"1' AND '1'='1",
		"'; DELETE FROM project WHERE '1'='1",
		"1 OR 1=1",
		"' OR 'x'='x",
		"1; UPDATE client SET name='hacked' WHERE 1=1--",
	}

	t.Run("SQL injection in login email field", func(t *testing.T) {
		// Create a test user first
		_, err := app.users.Insert("Test User", "security@test.com", "password123")
		require.NoError(t, err)

		for _, payload := range sqlInjectionPayloads {
			form := url.Values{}
			form.Add("email", payload)
			form.Add("password", "testpass")

			req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLoginPost))
			handler.ServeHTTP(rr, req)

			// Should not succeed with SQL injection
			assert.NotEqual(t, http.StatusSeeOther, rr.Code, "SQL injection payload should not bypass authentication: %s", payload)
			// Should not cause server error
			assert.NotEqual(t, http.StatusInternalServerError, rr.Code, "SQL injection should not crash server: %s", payload)
		}
	})

	t.Run("SQL injection in client name", func(t *testing.T) {
		for _, payload := range sqlInjectionPayloads {
			form := url.Values{}
			form.Add("name", payload)
			form.Add("email", "test@example.com")
			form.Add("hourly_rate", "50.00")

			req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientCreatePost))
			handler.ServeHTTP(rr, req)

			// Should either accept it as data or reject with validation error, but not crash
			assert.NotEqual(t, http.StatusInternalServerError, rr.Code, "SQL injection should not cause server error: %s", payload)

			// If accepted, verify it's stored safely
			if rr.Code == http.StatusSeeOther {
				// Check that database wasn't corrupted
				var count int
				err := testDB.DB.QueryRow("SELECT COUNT(*) FROM client").Scan(&count)
				require.NoError(t, err, "Database should still be queryable after SQL injection attempt")
			}
		}
	})
}

// TestOWASP_XSSAttacks tests protection against Cross-Site Scripting (OWASP A03:2021)
func TestOWASP_XSSAttacks(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"<svg/onload=alert('XSS')>",
		"javascript:alert('XSS')",
		"<iframe src='javascript:alert(`XSS`)'></iframe>",
		"<body onload=alert('XSS')>",
		"\"><script>alert('XSS')</script>",
		"'><script>alert(String.fromCharCode(88,83,83))</script>",
	}

	t.Run("XSS in client name is escaped in HTML output", func(t *testing.T) {
		testDB.TruncateTable(t, "client")

		for _, payload := range xssPayloads {
			// Create client with XSS payload
			form := url.Values{}
			form.Add("name", payload)
			form.Add("email", "xss@test.com")
			form.Add("hourly_rate", "50.00")

			req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientCreatePost))
			handler.ServeHTTP(rr, req)

			if rr.Code == http.StatusSeeOther {
				// Now retrieve the page and check output is escaped
				req = httptest.NewRequest(http.MethodGet, "/", nil)
				rr = httptest.NewRecorder()

				handler = app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientsList))
				handler.ServeHTTP(rr, req)

				body := rr.Body.String()
				// XSS payload should not be executable (script tags should be escaped)
				assert.NotContains(t, body, "<script>alert", "Script tags should be escaped, not executable: %s", payload)

				// Verify HTML entities are used
				if strings.Contains(payload, "<script>") {
					// Should be escaped as &lt;script&gt;
					assert.Contains(t, body, "&lt;", "HTML should be entity-encoded")
				}
			}

			testDB.TruncateTable(t, "client")
		}
	})
}

// TestOWASP_BrokenAccessControl tests authorization (OWASP A01:2021)
func TestOWASP_BrokenAccessControl(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("cannot delete resources by guessing IDs", func(t *testing.T) {
		// Try to delete client with non-existent ID
		req := httptest.NewRequest(http.MethodPost, "/client/delete/99999", nil)
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientDelete))
		handler.ServeHTTP(rr, req)

		// Should handle gracefully (404 or redirect), not crash
		assert.NotEqual(t, http.StatusInternalServerError, rr.Code, "Should handle non-existent IDs gracefully")
	})

	t.Run("cannot access resources with invalid ID formats", func(t *testing.T) {
		invalidIDs := []string{
			"-1",
			"abc",
			"999999999999999999999", // Overflow
			"0",
			url.PathEscape("1 OR 1=1"),
			url.PathEscape("'; DROP TABLE client--"),
		}

		for _, id := range invalidIDs {
			req := httptest.NewRequest(http.MethodGet, "/client/view/"+id, nil)
			rr := httptest.NewRecorder()

			handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientView))
			handler.ServeHTTP(rr, req)

			// Should not crash, should return error page
			assert.NotEqual(t, http.StatusInternalServerError, rr.Code, "Invalid ID should not cause server error: %s", id)
		}
	})
}

// TestOWASP_AuthenticationFailures tests authentication security (OWASP A07:2021)
func TestOWASP_AuthenticationFailures(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("passwords are hashed using bcrypt", func(t *testing.T) {
		// Create a user
		userID, err := app.users.Insert("Hash Test", "hash@test.com", "plaintextpassword")
		require.NoError(t, err)

		// Read password hash directly from database
		var hashedPassword string
		err = testDB.DB.QueryRow("SELECT hashed_password FROM user WHERE id = ?", userID).Scan(&hashedPassword)
		require.NoError(t, err)

		// Should be bcrypt hash
		assert.True(t, strings.HasPrefix(hashedPassword, "$2a$"), "Password should be bcrypt hashed")
		assert.NotEqual(t, "plaintextpassword", hashedPassword, "Password should not be stored in plain text")
		assert.Greater(t, len(hashedPassword), 50, "Bcrypt hash should be long")
	})

	t.Run("login rejects invalid credentials", func(t *testing.T) {
		// Create a user
		_, err := app.users.Insert("Valid User", "valid@test.com", "correctpassword123")
		require.NoError(t, err)

		// Try to login with wrong password
		form := url.Values{}
		form.Add("email", "valid@test.com")
		form.Add("password", "wrongpassword")

		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLoginPost))
		handler.ServeHTTP(rr, req)

		// Should not redirect (successful login would redirect)
		assert.NotEqual(t, http.StatusSeeOther, rr.Code, "Wrong password should not succeed")
		assert.Contains(t, rr.Body.String(), "incorrect", "Should show error message")
	})

	t.Run("session cookies have security attributes", func(t *testing.T) {
		// Create and login user
		_, err := app.users.Insert("Cookie Test", "cookie@test.com", "password123")
		require.NoError(t, err)

		form := url.Values{}
		form.Add("email", "cookie@test.com")
		form.Add("password", "password123")

		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLoginPost))
		handler.ServeHTTP(rr, req)

		cookies := rr.Result().Cookies()
		require.NotEmpty(t, cookies, "Should have session cookie")

		sessionCookie := cookies[0]
		// Check security attributes
		assert.True(t, sessionCookie.HttpOnly, "Session cookie should be HttpOnly to prevent XSS attacks")
		// Note: Secure flag is typically only set in production with HTTPS
		// SameSite should be set to prevent CSRF
		t.Logf("Session cookie SameSite: %v (Lax or Strict recommended for CSRF protection)", sessionCookie.SameSite)
	})
}

// TestOWASP_SensitiveDataExposure tests for sensitive data leakage (OWASP A02:2021)
func TestOWASP_SensitiveDataExposure(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("passwords never exposed in responses", func(t *testing.T) {
		// Create user
		_, err := app.users.Insert("Sensitive User", "sensitive@test.com", "mysecretpassword123")
		require.NoError(t, err)

		// Check login page
		req := httptest.NewRequest(http.MethodGet, "/user/login", nil)
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLogin))
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		// Should never contain the password
		assert.NotContains(t, body, "mysecretpassword123", "Plain text password should not be in response")
		// Should not contain bcrypt hashes either
		assert.NotContains(t, body, "$2a$", "Password hash should not be exposed")
	})

	t.Run("SMTP passwords are encrypted in database", func(t *testing.T) {
		// Set SMTP password
		err := app.settings.UpdateValue("smtp_password", "plain-text-smtp-password")
		require.NoError(t, err)

		// Read directly from database
		var storedValue string
		err = testDB.DB.QueryRow("SELECT value FROM setting WHERE key = 'smtp_password'").Scan(&storedValue)
		require.NoError(t, err)

		// Should not be plain text
		assert.NotEqual(t, "plain-text-smtp-password", storedValue, "SMTP password should be encrypted in database")

		// But should be decryptable
		decrypted, err := app.settings.GetDecryptedSMTPPassword()
		require.NoError(t, err)
		assert.Equal(t, "plain-text-smtp-password", decrypted, "Should be able to decrypt SMTP password")
	})

	t.Run("error messages don't leak implementation details", func(t *testing.T) {
		// Try to access non-existent client
		req := httptest.NewRequest(http.MethodGet, "/client/view/99999", nil)
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientView))
		handler.ServeHTTP(rr, req)

		body := rr.Body.String()
		// Should not expose SQL errors or internal details
		assert.NotContains(t, strings.ToLower(body), "sql", "Should not expose SQL errors")
		assert.NotContains(t, strings.ToLower(body), "panic", "Should not expose panic messages")
		assert.NotContains(t, strings.ToLower(body), "goroutine", "Should not expose stack traces")
	})
}

// TestOWASP_InputValidation tests comprehensive input validation
func TestOWASP_InputValidation(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("extremely long input is handled safely", func(t *testing.T) {
		longString := strings.Repeat("A", 10000)

		form := url.Values{}
		form.Add("name", longString)
		form.Add("email", "test@example.com")
		form.Add("hourly_rate", "50.00")

		req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientCreatePost))
		handler.ServeHTTP(rr, req)

		// Should either reject or accept, but not crash
		assert.NotEqual(t, http.StatusInternalServerError, rr.Code, "Extremely long input should not crash server")
	})

	t.Run("null bytes are handled", func(t *testing.T) {
		form := url.Values{}
		form.Add("name", "Test\x00Client")
		form.Add("email", "test@example.com")
		form.Add("hourly_rate", "50.00")

		req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientCreatePost))
		handler.ServeHTTP(rr, req)

		assert.NotEqual(t, http.StatusInternalServerError, rr.Code, "Null bytes should not crash server")
	})

	t.Run("special characters are handled safely", func(t *testing.T) {
		specialChars := []string{
			"../../../etc/passwd",
			"..\\..\\windows\\system32",
			"$(whoami)",
			"`ls -la`",
			"${7*7}",
			"{{7*7}}",
			"%00",
		}

		for _, chars := range specialChars {
			form := url.Values{}
			form.Add("name", chars)
			form.Add("email", "test@example.com")
			form.Add("hourly_rate", "50.00")

			req := httptest.NewRequest(http.MethodPost, "/client/create", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientCreatePost))
			handler.ServeHTTP(rr, req)

			assert.NotEqual(t, http.StatusInternalServerError, rr.Code, "Special characters should not crash server: %s", chars)
		}
	})

	t.Run("negative and zero IDs are handled", func(t *testing.T) {
		invalidIDs := []string{"-1", "0", "-999"}

		for _, id := range invalidIDs {
			req := httptest.NewRequest(http.MethodGet, "/client/view/"+id, nil)
			rr := httptest.NewRecorder()

			handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientView))
			handler.ServeHTTP(rr, req)

			assert.NotEqual(t, http.StatusInternalServerError, rr.Code, "Negative/zero ID should not crash: %s", id)
		}
	})
}

// TestOWASP_SecurityMisconfiguration tests for security configuration issues (OWASP A05:2021)
func TestOWASP_SecurityMisconfiguration(t *testing.T) {
	app, testDB := createTestApp(t)
	defer testDB.Cleanup(t)

	t.Run("application handles missing templates gracefully", func(t *testing.T) {
		// This test verifies the app doesn't crash with missing resources
		// The test setup already creates minimal templates, so we just verify it works
		req := httptest.NewRequest(http.MethodGet, "/user/login", nil)
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLogin))
		handler.ServeHTTP(rr, req)

		// Should respond without crashing
		assert.NotEqual(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("application doesn't expose version information", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler := app.sessionManager.LoadAndSave(http.HandlerFunc(app.clientsList))
		handler.ServeHTTP(rr, req)

		headers := rr.Result().Header
		body := rr.Body.String()

		// Should not expose version info in headers or body
		serverHeader := headers.Get("Server")
		if serverHeader != "" {
			t.Logf("Server header: %s (consider removing or obfuscating version)", serverHeader)
		}

		// Should not expose Go version or framework details
		assert.NotContains(t, body, "Go/", "Should not expose Go version")
		assert.NotContains(t, strings.ToLower(body), "powered by", "Should not expose framework info")
	})
}
