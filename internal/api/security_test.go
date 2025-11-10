package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestSQLInjectionPrevention tests that the API is protected against SQL injection attacks
func TestSQLInjectionPrevention(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create handlers
	clientModel := models.NewClientModel(testDB.DB)
	projectModel := models.NewProjectModel(testDB.DB)
	handlers := NewClientHandlers(clientModel, projectModel)

	// Create a normal client for comparison
	testDB.InsertTestClient(t, "Normal Client")

	sqlInjectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE client; --",
		"' UNION SELECT * FROM user --",
		"admin'--",
		"' OR 1=1--",
		"<script>alert('xss')</script>",
		"'); DELETE FROM client WHERE ('1'='1",
		"1' AND '1'='1",
	}

	t.Run("search parameter SQL injection attempts", func(t *testing.T) {
		for _, payload := range sqlInjectionPayloads {
			t.Run(fmt.Sprintf("payload: %s", payload), func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/clients?search="+url.QueryEscape(payload), nil)
				ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
				req = req.WithContext(ctx)

				rr := httptest.NewRecorder()
				handlers.ListClients(rr, req, nil)

				// Should return OK (empty results or actual results, but not an error)
				// The key is it shouldn't execute the SQL injection
				assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusBadRequest,
					"Expected OK or BadRequest, got %d", rr.Code)

				// Verify the database is still intact by counting clients
				clients, _ := clientModel.GetAll()
				assert.NotEmpty(t, clients, "Clients table should not be dropped")
			})
		}
	})

	t.Run("create client with SQL injection in name", func(t *testing.T) {
		for _, payload := range sqlInjectionPayloads {
			t.Run(fmt.Sprintf("payload: %s", payload), func(t *testing.T) {
				reqBody := ClientRequest{
					Name:       payload,
					Email:      "test@example.com",
					HourlyRate: 100.00,
				}
				body, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
				req = req.WithContext(ctx)

				rr := httptest.NewRecorder()
				handlers.CreateClient(rr, req, nil)

				// Should either succeed (storing the malicious string as data) or fail validation
				// The key is it shouldn't execute the SQL injection
				assert.True(t, rr.Code == http.StatusCreated || rr.Code == http.StatusUnprocessableEntity || rr.Code == http.StatusBadRequest,
					"Expected Created, UnprocessableEntity, or BadRequest, got %d", rr.Code)

				// Verify the database is still intact
				clients, _ := clientModel.GetAll()
				assert.NotEmpty(t, clients, "Clients table should not be affected")
			})
		}
	})

	t.Run("get client with SQL injection in ID parameter", func(t *testing.T) {
		// Create a test client first
		clientID := testDB.InsertTestClient(t, "Test Client")

		sqlIDPayloads := []string{
			"1 OR 1=1",
			"1'; DROP TABLE client; --",
			"1 UNION SELECT * FROM user",
		}

		for _, payload := range sqlIDPayloads {
			t.Run(fmt.Sprintf("payload: %s", payload), func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients/%s", url.PathEscape(payload)), nil)
				ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
				req = req.WithContext(ctx)

				rr := httptest.NewRecorder()

				ps := httprouter.Params{
					{Key: "id", Value: payload},
				}

				handlers.GetClient(rr, req, ps)

				// Should return BadRequest for invalid ID format, not execute SQL
				assert.Equal(t, http.StatusBadRequest, rr.Code,
					"SQL injection in ID should return BadRequest")

				// Verify the actual client still exists
				client, err := clientModel.Get(clientID)
				assert.NoError(t, err)
				assert.Equal(t, "Test Client", client.Name)
			})
		}
	})
}

// TestXSSPrevention tests that the API properly encodes output to prevent XSS
func TestXSSPrevention(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	clientModel := models.NewClientModel(testDB.DB)
	projectModel := models.NewProjectModel(testDB.DB)
	handlers := NewClientHandlers(clientModel, projectModel)

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"javascript:alert('XSS')",
		"<iframe src='javascript:alert(\"XSS\")'></iframe>",
		"<body onload=alert('XSS')>",
	}

	t.Run("XSS in client name is safely stored and returned", func(t *testing.T) {
		for _, payload := range xssPayloads {
			t.Run(fmt.Sprintf("payload: %s", payload), func(t *testing.T) {
				// Create client with XSS payload
				reqBody := ClientRequest{
					Name:       payload,
					Email:      "xss@example.com",
					HourlyRate: 100.00,
				}
				body, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
				req = req.WithContext(ctx)

				rr := httptest.NewRecorder()
				handlers.CreateClient(rr, req, nil)

				if rr.Code == http.StatusCreated {
					var resp Response
					err := json.NewDecoder(rr.Body).Decode(&resp)
					require.NoError(t, err)

					client := resp.Data.(map[string]interface{})
					clientID := int(client["id"].(float64))

					// Retrieve the client and verify JSON encoding
					req2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients/%d", clientID), nil)
					ctx2 := SetAPIKeyContext(req2.Context(), 1, userID, "clients:read", "Test Key")
					req2 = req2.WithContext(ctx2)

					rr2 := httptest.NewRecorder()
					ps := httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", clientID)}}
					handlers.GetClient(rr2, req2, ps)

					assert.Equal(t, http.StatusOK, rr2.Code)
					assert.Equal(t, "application/json", rr2.Header().Get("Content-Type"))

					// Verify the response is valid JSON and contains the escaped payload
					var resp2 Response
					err = json.NewDecoder(rr2.Body).Decode(&resp2)
					require.NoError(t, err, "Response should be valid JSON")

					client2 := resp2.Data.(map[string]interface{})
					// The payload should be safely stored and returned as-is in JSON
					// JSON encoding automatically escapes HTML special characters
					assert.Equal(t, payload, client2["name"])
				}
			})
		}
	})

	t.Run("response content-type is application/json", func(t *testing.T) {
		// Verify all responses use JSON content-type
		clientID := testDB.InsertTestClient(t, "Content Type Test")

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients/%d", clientID), nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		ps := httprouter.Params{{Key: "id", Value: fmt.Sprintf("%d", clientID)}}
		handlers.GetClient(rr, req, ps)

		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"),
			"API should always return JSON, not HTML")
	})
}

// TestInputValidation tests that the API properly validates and sanitizes input
func TestInputValidation(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	clientModel := models.NewClientModel(testDB.DB)
	projectModel := models.NewProjectModel(testDB.DB)
	handlers := NewClientHandlers(clientModel, projectModel)

	t.Run("reject negative hourly rate", func(t *testing.T) {
		reqBody := ClientRequest{
			Name:       "Test Client",
			Email:      "test@example.com",
			HourlyRate: -100.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.CreateClient(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("reject invalid email format", func(t *testing.T) {
		reqBody := ClientRequest{
			Name:       "Test Client",
			Email:      "not-an-email",
			HourlyRate: 100.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.CreateClient(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("reject empty required fields", func(t *testing.T) {
		reqBody := ClientRequest{
			Name:       "",
			Email:      "test@example.com",
			HourlyRate: 100.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.CreateClient(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("handle extremely long strings", func(t *testing.T) {
		// Create a very long string (10KB)
		longString := string(make([]byte, 10000))
		for i := range longString {
			longString = longString[:i] + "A" + longString[i:]
		}

		reqBody := ClientRequest{
			Name:       longString,
			Email:      "test@example.com",
			HourlyRate: 100.00,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.CreateClient(rr, req, nil)

		// Should handle gracefully (either accept or reject with proper error)
		assert.True(t, rr.Code == http.StatusCreated || rr.Code == http.StatusUnprocessableEntity || rr.Code == http.StatusBadRequest)
	})
}

// TestAuthorizationScopes tests scope checking logic
// NOTE: Authorization is enforced at the route level via RequireScopes middleware
// in cmd/web/api_routes.go. These tests verify the scope checking logic works correctly.
func TestAuthorizationScopes(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	t.Run("HasScope function correctly identifies scopes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/clients", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read projects:write", "Test Key")
		req = req.WithContext(ctx)

		// Should have clients:read
		assert.True(t, HasScope(req, "clients:read"))

		// Should have projects:write
		assert.True(t, HasScope(req, "projects:write"))

		// Should not have clients:write
		assert.False(t, HasScope(req, "clients:write"))

		// Should not have invoices:read
		assert.False(t, HasScope(req, "invoices:read"))
	})

	t.Run("wildcard scope grants all permissions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/clients", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "*", "Admin Key")
		req = req.WithContext(ctx)

		// Should have any scope with wildcard
		assert.True(t, HasScope(req, "clients:read"))
		assert.True(t, HasScope(req, "clients:write"))
		assert.True(t, HasScope(req, "projects:read"))
		assert.True(t, HasScope(req, "invoices:write"))
	})

	t.Run("resource wildcard grants all actions on resource", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/clients", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:*", "Client Admin Key")
		req = req.WithContext(ctx)

		// Should have all client permissions
		assert.True(t, HasScope(req, "clients:read"))
		assert.True(t, HasScope(req, "clients:write"))

		// Should not have other resource permissions
		assert.False(t, HasScope(req, "projects:read"))
		assert.False(t, HasScope(req, "invoices:read"))
	})

	t.Run("empty scope has no permissions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/clients", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "", "No Permissions Key")
		req = req.WithContext(ctx)

		// Should not have any permissions
		assert.False(t, HasScope(req, "clients:read"))
		assert.False(t, HasScope(req, "clients:write"))
	})

	t.Run("RequireScopes middleware properly blocks unauthorized access", func(t *testing.T) {
		// Test RequireScopes middleware directly
		middleware := RequireScopes("clients:write")

		// Create a dummy handler that should only be called if authorized
		called := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		// Wrap handler with middleware
		wrappedHandler := middleware(handler)

		// Test with wrong scope (clients:read instead of clients:write)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "clients:read", "Read Only Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		// Should be forbidden and handler should not be called
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.False(t, called, "Handler should not be called without proper scope")

		// Test with correct scope
		called = false
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/clients", nil)
		ctx2 := SetAPIKeyContext(req2.Context(), 1, userID, "clients:write", "Write Key")
		req2 = req2.WithContext(ctx2)

		rr2 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr2, req2)

		// Should succeed and handler should be called
		assert.Equal(t, http.StatusOK, rr2.Code)
		assert.True(t, called, "Handler should be called with proper scope")
	})
}
