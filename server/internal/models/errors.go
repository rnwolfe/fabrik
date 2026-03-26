package models

import "errors"

// Domain-level errors returned by the service layer.
// Handlers map these to appropriate HTTP status codes.
var (
	// ErrNotFound is returned when the requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrConstraintViolation is returned when a business constraint is violated.
	ErrConstraintViolation = errors.New("constraint violation")
)
