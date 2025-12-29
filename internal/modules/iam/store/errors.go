package store

import "errors"

var (
	// ErrAuthPayloadNotFound is returned when the security context is missing from a request context.
	ErrAuthPayloadNotFound = errors.New("auth payload not found in context")
)
