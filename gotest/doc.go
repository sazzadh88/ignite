// Package gotest provides testing utilities for Ignite applications.
//
// This package offers Laravel-style HTTP testing helpers, assertions,
// and simple mocking capabilities for testing Go web applications built
// with the Ignite framework.
//
// # Basic Usage
//
// Create a test case and set up the test application:
//
//	func TestExample(t *testing.T) {
//	    tc := gotest.NewTestCase(t)
//	    tc.SetUp()
//	    defer tc.TearDown()
//
//	    // Your tests here
//	}
//
// # HTTP Testing
//
// Perform HTTP requests and assert responses:
//
//	func TestHTTP(t *testing.T) {
//	    tc := gotest.NewTestCase(t)
//	    tc.SetUp()
//	    defer tc.TearDown()
//
//	    tc.Get("/api/users").
//	        AssertOk().
//	        AssertJSON(map[string]any{
//	            "count": 10,
//	        })
//
//	    tc.Post("/api/users", map[string]any{
//	        "name": "John",
//	        "email": "john@example.com",
//	    }).AssertCreated().
//	        AssertJSONPath("user.name", "John")
//	}
//
// # Authentication
//
// Test authenticated requests using WithToken or ActingAs:
//
//	tc.WithToken("test-token").
//	    Get("/api/profile").
//	    AssertOk()
//
//	tc.ActingAs(user).
//	    Delete("/api/posts/123").
//	    AssertNoContent()
//
// # Mocking
//
// Create mocks to test components in isolation:
//
//	mock := gotest.NewMock()
//	mock.On("SendEmail", "john@example.com").Return(nil)
//
//	// Use mock in your code
//	results := mock.Called("SendEmail", "john@example.com")
//
//	// Assert expectations
//	mock.AssertCalled(t, "SendEmail")
//	mock.AssertCalledWith(t, "SendEmail", "john@example.com")
//	mock.AssertCalledTimes(t, "SendEmail", 1)
//
// # Database Assertions
//
// Database assertions are placeholders for future integration:
//
//	tc.AssertDatabaseHas("users", map[string]any{
//	    "email": "john@example.com",
//	})
//
//	tc.AssertDatabaseCount("users", 5)
package gotest
