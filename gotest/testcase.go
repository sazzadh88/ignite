package gotest

import (
	"testing"

	"github.com/sazzadh88/ignite/foundation"
)

// TestCase provides a base test case for Ignite applications.
// It manages test application lifecycle and provides HTTP testing helpers.
type TestCase struct {
	App     *foundation.Application
	t       *testing.T
	headers map[string]string
}

// NewTestCase creates a new test case instance.
func NewTestCase(t *testing.T) *TestCase {
	return &TestCase{
		t:       t,
		headers: make(map[string]string),
	}
}

// SetUp initializes the test application.
// Call this in your test setup to prepare the application for testing.
func (tc *TestCase) SetUp() {
	tc.App = foundation.NewApplication(".")
	tc.App.Bootstrap()
}

// TearDown cleans up the test application.
// Call this in your test cleanup to release resources.
func (tc *TestCase) TearDown() {
	// Future: cleanup database connections, temp files, etc.
	tc.App = nil
	tc.headers = make(map[string]string)
}
