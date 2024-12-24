package domain

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrNotFound = errors.New("not found")
	ErrPostNotFound = errors.New("post not found")
)

// NotFoundError represents a not found error with context
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id '%s' not found", e.Resource, e.ID)
}

// NewNotFoundError creates a new not found error
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
