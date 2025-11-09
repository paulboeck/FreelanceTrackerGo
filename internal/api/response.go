package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response represents a standard API response
type Response struct {
	Data interface{}     `json:"data,omitempty"`
	Meta *ResponseMeta   `json:"meta,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`
}

// ResponseMeta contains metadata about the response
type ResponseMeta struct {
	Timestamp  time.Time `json:"timestamp"`
	Page       int       `json:"page,omitempty"`
	PageSize   int       `json:"pageSize,omitempty"`
	Total      int       `json:"total,omitempty"`
	TotalPages int       `json:"totalPages,omitempty"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}, meta *ResponseMeta) error {
	if meta == nil {
		meta = &ResponseMeta{
			Timestamp: time.Now().UTC(),
		}
	} else {
		meta.Timestamp = time.Now().UTC()
	}

	response := Response{
		Data: data,
		Meta: meta,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(response)
}

// WriteError writes a JSON error response
func WriteError(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) error {
	response := Response{
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(response)
}

// WritePaginatedJSON writes a paginated JSON response
func WritePaginatedJSON(w http.ResponseWriter, data interface{}, page, pageSize, total int) error {
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 0 {
		totalPages = 0
	}

	meta := &ResponseMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	return WriteJSON(w, http.StatusOK, data, meta)
}
