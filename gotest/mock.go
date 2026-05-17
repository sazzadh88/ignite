package gotest

import (
	"reflect"
	"testing"
)

// Mock provides simple mocking capabilities for testing.
type Mock struct {
	Calls        []MockCall
	expectations map[string]*MockExpectation
}

// MockCall represents a recorded method call.
type MockCall struct {
	Method  string
	Args    []any
	Returns []any
}

// MockExpectation represents an expected method call with return values.
type MockExpectation struct {
	Method  string
	Args    []any
	Returns []any
}

// NewMock creates a new mock instance.
func NewMock() *Mock {
	return &Mock{
		Calls:        make([]MockCall, 0),
		expectations: make(map[string]*MockExpectation),
	}
}

// On sets up an expectation for a method call with optional arguments.
// Returns a MockExpectation that can be configured with return values.
func (m *Mock) On(method string, args ...any) *MockExpectation {
	exp := &MockExpectation{
		Method: method,
		Args:   args,
	}
	m.expectations[method] = exp
	return exp
}

// Return sets the return values for the mock expectation.
func (e *MockExpectation) Return(values ...any) *MockExpectation {
	e.Returns = values
	return e
}

// Called records a method call and returns any configured return values.
// Use this in your mock implementation to track calls.
func (m *Mock) Called(method string, args ...any) []any {
	call := MockCall{
		Method: method,
		Args:   args,
	}

	// Get expected return values if configured
	if exp, ok := m.expectations[method]; ok {
		call.Returns = exp.Returns
	}

	m.Calls = append(m.Calls, call)
	return call.Returns
}

// AssertCalled asserts that the method was called at least once.
func (m *Mock) AssertCalled(t *testing.T, method string) {
	for _, call := range m.Calls {
		if call.Method == method {
			return
		}
	}
	t.Errorf("Expected method %s to be called, but it was not", method)
}

// AssertNotCalled asserts that the method was not called.
func (m *Mock) AssertNotCalled(t *testing.T, method string) {
	for _, call := range m.Calls {
		if call.Method == method {
			t.Errorf("Expected method %s to not be called, but it was", method)
			return
		}
	}
}

// AssertCalledTimes asserts that the method was called exactly the given number of times.
func (m *Mock) AssertCalledTimes(t *testing.T, method string, times int) {
	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	if count != times {
		t.Errorf("Expected method %s to be called %d times, but it was called %d times", method, times, count)
	}
}

// AssertCalledWith asserts that the method was called with the given arguments.
func (m *Mock) AssertCalledWith(t *testing.T, method string, args ...any) {
	for _, call := range m.Calls {
		if call.Method == method && reflect.DeepEqual(call.Args, args) {
			return
		}
	}
	t.Errorf("Expected method %s to be called with args %v, but it was not", method, args)
}

// Reset clears all recorded calls and expectations.
func (m *Mock) Reset() {
	m.Calls = make([]MockCall, 0)
	m.expectations = make(map[string]*MockExpectation)
}
