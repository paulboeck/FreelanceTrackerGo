package api

import (
	"net/http"

	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

// Error codes following industry standards
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeConflict      = "CONFLICT"
	ErrCodeRateLimit     = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeBadRequest    = "BAD_REQUEST"
	ErrCodeInvalidScope  = "INVALID_SCOPE"
)

// Standard error messages
const (
	MsgValidationFailed   = "Validation failed"
	MsgUnauthorized       = "Authentication required"
	MsgForbidden          = "Insufficient permissions"
	MsgNotFound           = "Resource not found"
	MsgConflict           = "Resource conflict"
	MsgRateLimitExceeded  = "Rate limit exceeded"
	MsgInternalError      = "Internal server error"
	MsgBadRequest         = "Bad request"
	MsgInvalidAPIKey      = "Invalid API key"
	MsgExpiredAPIKey      = "API key has expired"
	MsgRevokedAPIKey      = "API key has been revoked"
)

// ErrorValidation writes a validation error response
func ErrorValidation(w http.ResponseWriter, v validator.Validator) error {
	details := make(map[string]interface{})
	for field, err := range v.FieldErrors {
		details[field] = err
	}
	return WriteError(w, http.StatusUnprocessableEntity, ErrCodeValidation, MsgValidationFailed, details)
}

// ErrorUnauthorized writes an unauthorized error response
func ErrorUnauthorized(w http.ResponseWriter, message string) error {
	if message == "" {
		message = MsgUnauthorized
	}
	return WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, message, nil)
}

// ErrorForbidden writes a forbidden error response
func ErrorForbidden(w http.ResponseWriter, message string) error {
	if message == "" {
		message = MsgForbidden
	}
	return WriteError(w, http.StatusForbidden, ErrCodeForbidden, message, nil)
}

// ErrorNotFound writes a not found error response
func ErrorNotFound(w http.ResponseWriter, message string) error {
	if message == "" {
		message = MsgNotFound
	}
	return WriteError(w, http.StatusNotFound, ErrCodeNotFound, message, nil)
}

// ErrorConflict writes a conflict error response
func ErrorConflict(w http.ResponseWriter, message string) error {
	if message == "" {
		message = MsgConflict
	}
	return WriteError(w, http.StatusConflict, ErrCodeConflict, message, nil)
}

// ErrorRateLimit writes a rate limit error response
func ErrorRateLimit(w http.ResponseWriter, retryAfter int) error {
	w.Header().Set("Retry-After", string(rune(retryAfter)))
	return WriteError(w, http.StatusTooManyRequests, ErrCodeRateLimit, MsgRateLimitExceeded, nil)
}

// ErrorInternal writes an internal error response
func ErrorInternal(w http.ResponseWriter) error {
	return WriteError(w, http.StatusInternalServerError, ErrCodeInternal, MsgInternalError, nil)
}

// ErrorBadRequest writes a bad request error response
func ErrorBadRequest(w http.ResponseWriter, message string) error {
	if message == "" {
		message = MsgBadRequest
	}
	return WriteError(w, http.StatusBadRequest, ErrCodeBadRequest, message, nil)
}
