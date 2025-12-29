// Package apierror provides a structured, production-ready error type for the API.
// It ensures that all API errors are consistent and do not leak internal implementation details.
package apierror

import (
	"fmt"
	"net/http"
)

// APIError is a structured error type for all API responses.
type APIError struct {
	StatusCode    int
	PublicMessage string
	internalError error
}

// Error satisfies the standard error interface.
func (e *APIError) Error() string {
	if e.internalError != nil {
		return e.internalError.Error()
	}
	return e.PublicMessage
}

// Unwrap provides compatibility for Go's standard errors.Is and errors.As functions.
func (e *APIError) Unwrap() error {
	return e.internalError
}

// --- Factory Functions ---

// NewBadRequest creates a new APIError for HTTP 400 Bad Request responses.
func NewBadRequest(message string, internalErr error) *APIError {
	if message == "" {
		message = "The request was invalid or cannot be otherwise served."
	}
	return &APIError{
		StatusCode:    http.StatusBadRequest,
		PublicMessage: message,
		internalError: internalErr,
	}
}

// NewUnauthorized creates a new APIError for HTTP 401 Unauthorized responses.
func NewUnauthorized(message string, internalErr error) *APIError {
	if message == "" {
		message = "Authentication is required and has failed or has not yet been provided."
	}
	return &APIError{
		StatusCode:    http.StatusUnauthorized,
		PublicMessage: message,
		internalError: internalErr,
	}
}

// NewNotFound creates a new APIError for HTTP 404 Not Found responses.
func NewNotFound(resource string, internalErr error) *APIError {
	return &APIError{
		StatusCode:    http.StatusNotFound,
		PublicMessage: fmt.Sprintf("The requested resource '%s' was not found.", resource),
		internalError: internalErr,
	}
}

// NewInternalServer creates a new APIError for HTTP 500 Internal Server Error responses.
// The public message is always generic to avoid leaking information.
func NewInternalServer(internalErr error) *APIError {
	return &APIError{
		StatusCode:    http.StatusInternalServerError,
		PublicMessage: "An unexpected error occurred on the server.",
		internalError: internalErr,
	}
}
