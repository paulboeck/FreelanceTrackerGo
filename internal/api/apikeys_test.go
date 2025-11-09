package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAPIKeyHandlers_Login(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create test user with known password
	password := "testpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	require.NoError(t, err)

	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create handlers
	apiKeyModel := models.NewAPIKeyModel(testDB.DB)
	userModel := models.NewUserModel(testDB.DB)
	handlers := NewAPIKeyHandlers(apiKeyModel, userModel)

	t.Run("successful login", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: password,
			Name:     "Test API Key",
			Scopes:   "*",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handlers.Login(rr, req, nil)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		// Verify response contains API key
		apiKey, ok := data["apiKey"].(string)
		require.True(t, ok)
		assert.NotEmpty(t, apiKey)
		assert.Contains(t, apiKey, "ftk_")

		// Verify other fields
		assert.Equal(t, float64(userID), data["userId"])
		assert.NotEmpty(t, data["message"])
	})

	t.Run("invalid email", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "nonexistent@example.com",
			Password: password,
			Name:     "Test API Key",
			Scopes:   "*",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handlers.Login(rr, req, nil)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid password", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
			Name:     "Test API Key",
			Scopes:   "*",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handlers.Login(rr, req, nil)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing email", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "",
			Password: password,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handlers.Login(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handlers.Login(rr, req, nil)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAPIKeyHandlers_CreateAPIKey(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create API key model
	apiKeyModel := models.NewAPIKeyModel(testDB.DB)
	userModel := models.NewUserModel(testDB.DB)
	handlers := NewAPIKeyHandlers(apiKeyModel, userModel)

	t.Run("create API key with valid scopes", func(t *testing.T) {
		reqBody := CreateAPIKeyRequest{
			Name:   "Test API Key",
			Scopes: "clients:read clients:write",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/apikeys", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Add API key context (simulating authenticated request)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "*", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateAPIKey(rr, req, nil)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		// Verify API key was created
		apiKey, ok := data["apiKey"].(string)
		require.True(t, ok)
		assert.Contains(t, apiKey, "ftk_")
		assert.NotEmpty(t, data["keyId"])
		assert.NotEmpty(t, data["keyPrefix"])
		assert.NotEmpty(t, data["message"])
	})

	t.Run("create API key with wildcard scope", func(t *testing.T) {
		reqBody := CreateAPIKeyRequest{
			Name:   "Admin API Key",
			Scopes: "*",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/apikeys", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx := SetAPIKeyContext(req.Context(), 1, userID, "*", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateAPIKey(rr, req, nil)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("create API key with invalid scope", func(t *testing.T) {
		reqBody := CreateAPIKeyRequest{
			Name:   "Invalid Scope Key",
			Scopes: "invalid:scope",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/apikeys", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx := SetAPIKeyContext(req.Context(), 1, userID, "*", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateAPIKey(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("create API key with missing name", func(t *testing.T) {
		reqBody := CreateAPIKeyRequest{
			Name:   "",
			Scopes: "clients:read",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/apikeys", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx := SetAPIKeyContext(req.Context(), 1, userID, "*", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.CreateAPIKey(rr, req, nil)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestAPIKeyHandlers_ListAPIKeys(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create some API keys
	apiKeyModel := models.NewAPIKeyModel(testDB.DB)
	key1, _ := models.GenerateAPIKey()
	key1Hash, _ := bcrypt.GenerateFromPassword([]byte(key1), 12)
	testDB.InsertTestAPIKey(t, userID, "Key 1", string(key1Hash), key1[:12], "clients:read")

	key2, _ := models.GenerateAPIKey()
	key2Hash, _ := bcrypt.GenerateFromPassword([]byte(key2), 12)
	testDB.InsertTestAPIKey(t, userID, "Key 2", string(key2Hash), key2[:12], "projects:write")

	userModel := models.NewUserModel(testDB.DB)
	handlers := NewAPIKeyHandlers(apiKeyModel, userModel)

	t.Run("list API keys for user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/apikeys", nil)

		// Add API key context
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "*", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.ListAPIKeys(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		keys, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, keys, 2)

		// Verify keys don't contain full API key values
		key1Data := keys[0].(map[string]interface{})
		_, hasAPIKey := key1Data["apiKey"]
		assert.False(t, hasAPIKey, "Response should not contain full API key")
	})
}

func TestAPIKeyHandlers_DeleteAPIKey(t *testing.T) {
	testDB := testutil.SetupTestSQLite(t)
	defer testDB.Cleanup(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create API key
	apiKeyModel := models.NewAPIKeyModel(testDB.DB)
	key, _ := models.GenerateAPIKey()
	keyHash, _ := bcrypt.GenerateFromPassword([]byte(key), 12)
	keyID := testDB.InsertTestAPIKey(t, userID, "Test Key", string(keyHash), key[:12], "clients:read")

	userModel := models.NewUserModel(testDB.DB)
	handlers := NewAPIKeyHandlers(apiKeyModel, userModel)

	t.Run("delete own API key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/auth/apikeys/%d", keyID), nil)

		// Add API key context
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "*", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: fmt.Sprintf("%d", keyID)},
		}

		handlers.DeleteAPIKey(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("delete nonexistent API key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/apikeys/99999", nil)

		ctx := SetAPIKeyContext(req.Context(), 1, userID, "*", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "id", Value: "99999"},
		}

		handlers.DeleteAPIKey(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
