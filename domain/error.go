package domain

import (
	"errors"
	"fmt"
)

// Common domain errors
var (
	// ErrNotFound represents a generic not found error
	ErrNotFound = errors.New("not found")
	
	// ErrInvalidID represents an invalid ID error
	ErrInvalidID = errors.New("invalid ID")
	
	// ErrUnauthorized represents an unauthorized access error
	ErrUnauthorized = errors.New("unauthorized")
	
	// ErrInvalidInput represents an invalid input error
	ErrInvalidInput = errors.New("invalid input")
	
	// ErrInternalError represents an internal server error
	ErrInternalError = errors.New("internal error")
	
	// ErrDuplicate represents a duplicate resource error
	ErrDuplicate = errors.New("duplicate resource")
)

// NotFoundError represents a not found error with context
type NotFoundError struct {
	Resource string
	ID       string
}

// Error returns the error message
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id '%s' not found", e.Resource, e.ID)
}

// NewNotFoundError creates a new not found error with context
func NewNotFoundError(resource, id string) error {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// IsNotFoundError checks if the error is a NotFoundError
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
