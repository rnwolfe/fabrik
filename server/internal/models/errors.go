package models

import "errors"

// Domain-level errors returned by the service layer.
// Handlers map these to appropriate HTTP status codes.
var (
	// ErrNotFound is returned when the requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrConstraintViolation is returned when a business constraint is violated.
	ErrConstraintViolation = errors.New("constraint violation")

	// ErrDuplicate is returned when a unique constraint would be violated.
	ErrDuplicate = errors.New("duplicate")

	// ErrSeedReadOnly is returned when a mutation is attempted on a seed device model.
	ErrSeedReadOnly = errors.New("seed device models are read-only")

	// ErrRUOverflow is returned when a device placement would exceed rack RU capacity (hard limit).
	ErrRUOverflow = errors.New("RU overflow")

	// ErrPositionOverlap is returned when a device placement would overlap an existing device.
	ErrPositionOverlap = errors.New("position overlap")

	// ErrConflict is returned when deleting a resource that is referenced by others.
	ErrConflict = errors.New("conflict")

	// ErrAggPortsFull is returned when an aggregation switch has no free ports for a new rack.
	ErrAggPortsFull = errors.New("aggregation ports full")

	// ErrAggModelDownsize is returned when changing an agg model would orphan existing connections.
	ErrAggModelDownsize = errors.New("aggregation model downsize would orphan connections")
)
