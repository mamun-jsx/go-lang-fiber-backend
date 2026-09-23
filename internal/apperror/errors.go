// Package apperror defines typed application errors that map to HTTP status codes.
package apperror

import (
	"errors"   // errors is used for errors.As
	"net/http" // net/http provides standard HTTP status codes
)

// AppError is a domain error that carries an HTTP status and a public message.
type AppError struct {
	Code    int    // Code is the HTTP status code to return
	Message string // Message is a safe, client-facing message
	Details any    // Details holds optional extra info (e.g. validation errors)
	Err     error  // Err is the underlying internal error (never shown to clients)
}

// Error implements the built-in error interface.
func (e *AppError) Error() string {
	// Include the internal error if present for better logs
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	// Otherwise just return the public message
	return e.Message
}

// Unwrap lets errors.Is / errors.As inspect the wrapped error.
func (e *AppError) Unwrap() error {
	// Return the underlying error
	return e.Err
}

// NotFound returns a 404 error with the given message.
func NotFound(msg string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg}
}

// BadRequest returns a 400 error with the given message.
func BadRequest(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}

// Validation returns a 422 error carrying field-level validation details.
func Validation(details any) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: "validation failed", Details: details}
}

// Internal returns a 500 error that wraps the real cause for logging.
func Internal(err error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: "internal server error", Err: err}
}

// As extracts an *AppError from any error chain (returns nil, false if not found).
func As(err error) (*AppError, bool) {
	// Target variable that errors.As will fill
	var appErr *AppError
	// Walk the error chain looking for an *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	// Not an application error
	return nil, false
}
