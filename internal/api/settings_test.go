package api

import (
	"bytes"
	"encoding/json"
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

func setupSettingsTest(t *testing.T) (*SettingsHandlers, *testutil.TestDatabase, int) {
	testDB := testutil.SetupTestSQLite(t)

	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	userID := testDB.InsertTestUser(t, "Test User", "test@example.com", string(hashedPassword))

	// Create handlers
	settingsModel := models.NewAppSettingModel(testDB.DB, "test-seed")
	handlers := NewSettingsHandlers(settingsModel)

	return handlers, testDB, userID
}

func TestSettingsHandlers_GetAllSettings(t *testing.T) {
	handlers, testDB, userID := setupSettingsTest(t)
	defer testDB.Cleanup(t)

	t.Run("get all settings", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.GetAllSettings(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		settings, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		// Verify some default settings exist (from testutil schema)
		assert.Contains(t, settings, "default_hourly_rate")
		assert.Contains(t, settings, "invoice_title")
		assert.Contains(t, settings, "freelancer_name")
		assert.Contains(t, settings, "email_enabled")
	})
}

func TestSettingsHandlers_GetSetting(t *testing.T) {
	handlers, testDB, userID := setupSettingsTest(t)
	defer testDB.Cleanup(t)

	t.Run("get existing setting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/default_hourly_rate", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "key", Value: "default_hourly_rate"},
		}

		handlers.GetSetting(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		setting, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "default_hourly_rate", setting["key"])
		assert.Equal(t, "85.00", setting["value"])
	})

	t.Run("get nonexistent setting", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/nonexistent_key", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "key", Value: "nonexistent_key"},
		}

		handlers.GetSetting(rr, req, ps)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "Setting not found")
	})

	t.Run("get setting with empty key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/", nil)
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:read", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "key", Value: ""},
		}

		handlers.GetSetting(rr, req, ps)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "Setting key is required")
	})
}

func TestSettingsHandlers_UpdateAllSettings(t *testing.T) {
	handlers, testDB, userID := setupSettingsTest(t)
	defer testDB.Cleanup(t)

	t.Run("update multiple settings", func(t *testing.T) {
		reqBody := UpdateSettingsRequest{
			Settings: map[string]string{
				"default_hourly_rate": "95.00",
				"freelancer_name":     "Updated Name",
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.UpdateAllSettings(rr, req, nil)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data["message"], "updated successfully")

		// Verify the settings were actually updated
		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
		ctx2 := SetAPIKeyContext(req2.Context(), 1, userID, "settings:read", "Test Key")
		req2 = req2.WithContext(ctx2)
		rr2 := httptest.NewRecorder()

		handlers.GetAllSettings(rr2, req2, nil)

		var resp2 Response
		json.NewDecoder(rr2.Body).Decode(&resp2)
		settings := resp2.Data.(map[string]interface{})
		assert.Equal(t, "95.00", settings["default_hourly_rate"])
		assert.Equal(t, "Updated Name", settings["freelancer_name"])
	})

	t.Run("update with empty settings", func(t *testing.T) {
		reqBody := UpdateSettingsRequest{
			Settings: map[string]string{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.UpdateAllSettings(rr, req, nil)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "At least one setting is required")
	})

	t.Run("update with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handlers.UpdateAllSettings(rr, req, nil)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "Invalid JSON")
	})
}

func TestSettingsHandlers_UpdateSetting(t *testing.T) {
	handlers, testDB, userID := setupSettingsTest(t)
	defer testDB.Cleanup(t)

	t.Run("update single setting", func(t *testing.T) {
		reqBody := struct {
			Value string `json:"value"`
		}{
			Value: "100.00",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/default_hourly_rate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "key", Value: "default_hourly_rate"},
		}

		handlers.UpdateSetting(rr, req, ps)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data["message"], "updated successfully")

		// Verify the setting was actually updated
		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/settings/default_hourly_rate", nil)
		ctx2 := SetAPIKeyContext(req2.Context(), 1, userID, "settings:read", "Test Key")
		req2 = req2.WithContext(ctx2)
		rr2 := httptest.NewRecorder()

		handlers.GetSetting(rr2, req2, ps)

		var resp2 Response
		json.NewDecoder(rr2.Body).Decode(&resp2)
		setting := resp2.Data.(map[string]interface{})
		assert.Equal(t, "100.00", setting["value"])
	})

	t.Run("update with empty value", func(t *testing.T) {
		reqBody := struct {
			Value string `json:"value"`
		}{
			Value: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/default_hourly_rate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "key", Value: "default_hourly_rate"},
		}

		handlers.UpdateSetting(rr, req, ps)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
	})

	t.Run("update with empty key", func(t *testing.T) {
		reqBody := struct {
			Value string `json:"value"`
		}{
			Value: "test value",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "key", Value: ""},
		}

		handlers.UpdateSetting(rr, req, ps)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "Setting key is required")
	})

	t.Run("update with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/default_hourly_rate", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		ctx := SetAPIKeyContext(req.Context(), 1, userID, "settings:write", "Test Key")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		ps := httprouter.Params{
			{Key: "key", Value: "default_hourly_rate"},
		}

		handlers.UpdateSetting(rr, req, ps)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp Response
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotNil(t, resp.Error)
		assert.Contains(t, resp.Error.Message, "Invalid JSON")
	})
}
