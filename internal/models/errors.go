package models

import (
	"errors"
)

// Package-level error variables that can be checked with errors.Is()
// var (...) is a block declaration - declares multiple variables at once
// These are sentinel errors - predefined errors used to signal specific conditions
var (
	// ErrNoRecord is returned when a database query finds no matching record.
	// errors.New creates a new error with the given message.
	ErrNoRecord = errors.New("models: no matching record found")

	// ErrInvalidCredentials is returned when a user tries to log in with
	// incorrect email address or password.
	ErrInvalidCredentials = errors.New("models: invalid credentials")

	// ErrDuplicateEmail is returned when a user tries to sign up with an
	// email address that's already in use.
	ErrDuplicateEmail = errors.New("models: duplicate email")
)
