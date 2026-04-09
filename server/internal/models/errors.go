package models

import "errors"

// ErrRef identifies a related resource referenced by an error.
type ErrRef struct {
	Kind string `json:"kind"`
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// DomainError is a domain-level error that carries a stable code, a human
// message, and optional references to related resources.
type DomainError struct {
	code    string
	message string
	Refs    []ErrRef
}

// Error implements the error interface.
func (e *DomainError) Error() string { return e.message }

// Code returns the stable machine-readable error code.
func (e *DomainError) Code() string { return e.code }

// Is reports whether e matches target by comparing error codes.
// This allows errors.Is to work correctly when refs differ between
// the sentinel and a wrapped instance, avoiding data races on shared
// package-level sentinels.
func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.Code() == t.Code()
}

// newDomainErr constructs a DomainError with the given code and message.
func newDomainErr(code, message string) *DomainError {
	return &DomainError{code: code, message: message}
}

// Domain-level errors returned by the service layer.
// Handlers map these to appropriate HTTP status codes.
var (
	// ErrNotFound is returned when the requested resource does not exist.
	ErrNotFound = newDomainErr("not_found", "not found")

	// ErrConstraintViolation is returned when a business constraint is violated.
	ErrConstraintViolation = newDomainErr("constraint_violation", "constraint violation")

	// ErrDuplicate is returned when a unique constraint would be violated.
	ErrDuplicate = newDomainErr("duplicate", "duplicate")

	// ErrSeedReadOnly is returned when a mutation is attempted on a seed device model.
	ErrSeedReadOnly = newDomainErr("seed_read_only", "seed device models are read-only")

	// ErrRUOverflow is returned when a device placement would exceed rack RU capacity (hard limit).
	ErrRUOverflow = newDomainErr("ru_overflow", "RU overflow")

	// ErrPositionOverlap is returned when a device placement would overlap an existing device.
	ErrPositionOverlap = newDomainErr("position_overlap", "position overlap")

	// ErrConflict is returned when deleting a resource that is referenced by others.
	ErrConflict = newDomainErr("conflict", "conflict")

	// ErrAggPortsFull is returned when an aggregation switch has no free ports for a new rack.
	ErrAggPortsFull = newDomainErr("agg_ports_full", "aggregation ports full")

	// ErrAggModelDownsize is returned when changing an agg model would orphan existing connections.
	ErrAggModelDownsize = newDomainErr("agg_model_downsize", "aggregation model downsize would orphan connections")
)

// IsDomainError reports whether err is a DomainError.
func IsDomainError(err error) (*DomainError, bool) {
	var de *DomainError
	if errors.As(err, &de) {
		return de, true
	}
	return nil, false
}
