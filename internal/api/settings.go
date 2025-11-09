package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

// SettingsHandlers handles settings API endpoints
type SettingsHandlers struct {
	settings *models.AppSettingModel
}

// NewSettingsHandlers creates a new SettingsHandlers instance
func NewSettingsHandlers(settings *models.AppSettingModel) *SettingsHandlers {
	return &SettingsHandlers{
		settings: settings,
	}
}

// SettingResponse represents a setting in API responses
type SettingResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// UpdateSettingsRequest represents the request to update multiple settings
type UpdateSettingsRequest struct {
	Settings map[string]string `json:"settings"`
}

// GetAllSettings returns all settings
// GET /api/v1/settings
func (h *SettingsHandlers) GetAllSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	settingsMap, err := h.settings.GetAll()
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Convert to response format
	response := make(map[string]string)
	for key, value := range settingsMap {
		response[key] = value.Value
	}

	WriteJSON(w, http.StatusOK, response, nil)
}

// GetSetting returns a single setting by key
// GET /api/v1/settings/:key
func (h *SettingsHandlers) GetSetting(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	if key == "" {
		ErrorBadRequest(w, "Setting key is required")
		return
	}

	setting, err := h.settings.Get(key)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Setting not found")
			return
		}
		ErrorInternal(w)
		return
	}

	response := SettingResponse{
		Key:   setting.Key,
		Value: setting.Value,
	}

	WriteJSON(w, http.StatusOK, response, nil)
}

// UpdateAllSettings updates multiple settings
// PUT /api/v1/settings
func (h *SettingsHandlers) UpdateAllSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate that we have at least one setting
	if len(req.Settings) == 0 {
		ErrorBadRequest(w, "At least one setting is required")
		return
	}

	// Update each setting
	for key, value := range req.Settings {
		if err := h.settings.UpdateValue(key, value); err != nil {
			ErrorInternal(w)
			return
		}
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Settings updated successfully",
	}, nil)
}

// UpdateSetting updates a single setting
// PUT /api/v1/settings/:key
func (h *SettingsHandlers) UpdateSetting(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	key := ps.ByName("key")
	if key == "" {
		ErrorBadRequest(w, "Setting key is required")
		return
	}

	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.Value), "value", "Value is required")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Update setting
	if err := h.settings.UpdateValue(key, req.Value); err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Setting updated successfully",
	}, nil)
}
