package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Validator holds validation errors for form fields.
// It's used to collect and track validation errors before returning them to the user.
type Validator struct {
	// map[string]string maps field names to error messages
	FieldErrors map[string]string
}

// Valid returns true if there are no validation errors.
// This is a method on *Validator (pointer receiver).
func (v *Validator) Valid() bool {
	// len() returns the number of items in a map
	return len(v.FieldErrors) == 0
}

// AddFieldError adds an error message for a specific form field.
// It only adds the first error for each field (won't overwrite existing errors).
func (v *Validator) AddFieldError(field, message string) {
	// Initialize the map if it's nil (lazy initialization)
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	// Check if an error already exists for this field
	// The underscore (_) discards the value, we only care about exists
	if _, exists := v.FieldErrors[field]; !exists {
		v.FieldErrors[field] = message
	}
}

// CheckField adds an error message for a field if the validation check fails.
// ok is typically the result of a validation function (true = valid, false = invalid).
func (v *Validator) CheckField(ok bool, field, message string) {
	if !ok {
		v.AddFieldError(field, message)
	}
}

// NotBlank checks if a string has non-whitespace content.
// strings.TrimSpace removes leading and trailing whitespace.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars checks if a string has at most n characters.
// utf8.RuneCountInString counts Unicode characters (not bytes) - important for international text.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// EmailRegex is a compiled regular expression for validating email addresses.
// regexp.MustCompile compiles the regex at startup - panics if the regex is invalid.
var EmailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

// Matches checks if a value matches a regular expression pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
