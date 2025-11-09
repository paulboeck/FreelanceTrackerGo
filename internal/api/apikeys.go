package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

// APIKeyHandlers handles API key management endpoints
type APIKeyHandlers struct {
	apiKeys *models.APIKeyModel
	users   *models.UserModel
}

// NewAPIKeyHandlers creates a new APIKeyHandlers instance
func NewAPIKeyHandlers(apiKeys *models.APIKeyModel, users *models.UserModel) *APIKeyHandlers {
	return &APIKeyHandlers{
		apiKeys: apiKeys,
		users:   users,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Scopes   string `json:"scopes"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	APIKey   string `json:"apiKey"`
	KeyID    int    `json:"keyId"`
	UserID   int    `json:"userId"`
	Message  string `json:"message"`
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name   string `json:"name"`
	Scopes string `json:"scopes"`
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	APIKey    string `json:"apiKey"`
	KeyID     int    `json:"keyId"`
	KeyPrefix string `json:"keyPrefix"`
	Message   string `json:"message"`
}

// APIKeyListResponse represents an API key in list responses (without sensitive data)
type APIKeyListResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	KeyPrefix  string `json:"keyPrefix"`
	Scopes     string `json:"scopes"`
	LastUsedAt *string `json:"lastUsedAt"`
	Created    string `json:"created"`
}

// Login authenticates a user and returns an API key
// POST /api/v1/auth/login
func (h *APIKeyHandlers) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate input
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.Email), "email", "Email is required")
	v.CheckField(validator.Matches(req.Email, validator.EmailRegex), "email", "Must be a valid email address")
	v.CheckField(validator.NotBlank(req.Password), "password", "Password is required")
	v.CheckField(validator.NotBlank(req.Name), "name", "API key name is required")
	v.CheckField(validator.NotBlank(req.Scopes), "scopes", "Scopes are required")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Validate scopes
	if err := models.ValidateScopes(req.Scopes); err != nil {
		v.AddFieldError("scopes", err.Error())
		ErrorValidation(w, *v)
		return
	}

	// Authenticate user
	userID, err := h.users.Authenticate(req.Email, req.Password)
	if err != nil {
		if err == models.ErrInvalidCredentials {
			ErrorUnauthorized(w, "Invalid email or password")
			return
		}
		ErrorInternal(w)
		return
	}

	// Create API key
	apiKey, keyID, err := h.apiKeys.Insert(userID, req.Name, req.Scopes)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return the API key (only time it's shown in plaintext)
	response := LoginResponse{
		APIKey:  apiKey,
		KeyID:   keyID,
		UserID:  userID,
		Message: "API key created successfully. Store this key securely, it will not be shown again.",
	}

	WriteJSON(w, http.StatusCreated, response, nil)
}

// CreateAPIKey creates a new API key for the authenticated user
// POST /api/v1/auth/apikeys
func (h *APIKeyHandlers) CreateAPIKey(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get user ID from context (set by auth middleware)
	userID, ok := GetUserID(r)
	if !ok {
		ErrorUnauthorized(w, "")
		return
	}

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate input
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.Name), "name", "API key name is required")
	v.CheckField(validator.NotBlank(req.Scopes), "scopes", "Scopes are required")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Validate scopes
	if err := models.ValidateScopes(req.Scopes); err != nil {
		v.AddFieldError("scopes", err.Error())
		ErrorValidation(w, *v)
		return
	}

	// Create API key
	apiKey, keyID, err := h.apiKeys.Insert(userID, req.Name, req.Scopes)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return the API key (only time it's shown in plaintext)
	response := CreateAPIKeyResponse{
		APIKey:    apiKey,
		KeyID:     keyID,
		KeyPrefix: apiKey[:12],
		Message:   "API key created successfully. Store this key securely, it will not be shown again.",
	}

	WriteJSON(w, http.StatusCreated, response, nil)
}

// ListAPIKeys returns all API keys for the authenticated user
// GET /api/v1/auth/apikeys
func (h *APIKeyHandlers) ListAPIKeys(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get user ID from context (set by auth middleware)
	userID, ok := GetUserID(r)
	if !ok {
		ErrorUnauthorized(w, "")
		return
	}

	// Get API keys for user
	keys, err := h.apiKeys.GetByUserID(userID)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Convert to response format (excluding sensitive data)
	response := make([]APIKeyListResponse, len(keys))
	for i, key := range keys {
		var lastUsed *string
		if key.LastUsedAt != nil {
			t := key.LastUsedAt.Format("2006-01-02T15:04:05Z")
			lastUsed = &t
		}

		response[i] = APIKeyListResponse{
			ID:         key.ID,
			Name:       key.Name,
			KeyPrefix:  key.KeyPrefix,
			Scopes:     key.Scopes,
			LastUsedAt: lastUsed,
			Created:    key.Created.Format("2006-01-02T15:04:05Z"),
		}
	}

	WriteJSON(w, http.StatusOK, response, nil)
}

// DeleteAPIKey revokes an API key
// DELETE /api/v1/auth/apikeys/:id
func (h *APIKeyHandlers) DeleteAPIKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get user ID from context (set by auth middleware)
	userID, ok := GetUserID(r)
	if !ok {
		ErrorUnauthorized(w, "")
		return
	}

	// Get key ID from URL
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid API key ID")
		return
	}

	// Get the API key to verify ownership
	key, err := h.apiKeys.GetByID(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "API key not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Verify the key belongs to the authenticated user
	if key.UserID != userID {
		ErrorForbidden(w, "You can only delete your own API keys")
		return
	}

	// Delete the key
	if err := h.apiKeys.Delete(id); err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "API key revoked successfully",
	}, nil)
}
