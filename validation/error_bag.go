// Package validation provides request validation utilities.
package validation

import (
	"encoding/json"
	"sync"
)

// ErrorBag collects validation error messages.
type ErrorBag struct {
	errors map[string][]string
	mu     sync.RWMutex
}

// NewErrorBag creates a new ErrorBag.
func NewErrorBag() *ErrorBag {
	return &ErrorBag{
		errors: make(map[string][]string),
	}
}

// Add appends an error message for the given field.
func (e *ErrorBag) Add(field, message string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errors[field] = append(e.errors[field], message)
}

// Has checks if the given field has any errors.
func (e *ErrorBag) Has(field string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	msgs, ok := e.errors[field]
	return ok && len(msgs) > 0
}

// Get returns all error messages for the given field.
func (e *ErrorBag) Get(field string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if msgs, ok := e.errors[field]; ok {
		result := make([]string, len(msgs))
		copy(result, msgs)
		return result
	}
	return []string{}
}

// First returns the first error message for the given field.
func (e *ErrorBag) First(field string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if msgs, ok := e.errors[field]; ok && len(msgs) > 0 {
		return msgs[0]
	}
	return ""
}

// All returns all error messages for all fields.
func (e *ErrorBag) All() map[string][]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make(map[string][]string, len(e.errors))
	for k, v := range e.errors {
		cpy := make([]string, len(v))
		copy(cpy, v)
		result[k] = cpy
	}
	return result
}

// ToMap is an alias for All.
func (e *ErrorBag) ToMap() map[string][]string {
	return e.All()
}

// ToJSON serializes the error bag to JSON.
func (e *ErrorBag) ToJSON() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return json.Marshal(e.errors)
}

// IsEmpty returns true if there are no errors.
func (e *ErrorBag) IsEmpty() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.errors) == 0
}

// Count returns the total number of fields with errors.
func (e *ErrorBag) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.errors)
}
