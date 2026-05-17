# GoTest Package

Laravel-inspired testing utilities for GoFrame applications.

## Overview

The `gotest` package provides HTTP testing helpers, fluent assertions, and simple mocking utilities for testing GoFrame web applications. Named `gotest` to avoid conflict with Go's stdlib `testing` package.

## Features

- **HTTP Test Helpers**: GET, POST, PUT, DELETE, JSON methods
- **Fluent Assertions**: Chain assertions for readable tests
- **Response Assertions**: Status codes, headers, JSON, body content
- **JSON Path Support**: Assert nested JSON values with dot notation
- **Simple Mocking**: Record calls and verify expectations
- **Database Assertions**: Placeholder for future ORM integration
- **Zero External Dependencies**: Uses only Go stdlib

## Package Structure

```
gotest/
├── doc.go                      # Package documentation
├── testcase.go                 # Base test case with SetUp/TearDown
├── http_test_helpers.go        # HTTP request helpers
├── assertions.go               # Response assertions
├── database_assertions.go      # Database assertions (placeholder)
├── mock.go                     # Simple mocking utilities
├── gotest_test.go              # Comprehensive tests
└── README.md                   # This file
```

## Quick Start

```go
package myapp_test

import (
    "testing"
    "github.com/sazzad/goframe/gotest"
)

func TestUserAPI(t *testing.T) {
    tc := gotest.NewTestCase(t)
    tc.SetUp()
    defer tc.TearDown()

    // Test GET request
    tc.Get("/api/users").
        AssertOk().
        AssertJSON(map[string]any{
            "count": 10,
        })

    // Test POST request with authentication
    tc.WithToken("test-token").
        Post("/api/users", map[string]any{
            "name":  "John Doe",
            "email": "john@example.com",
        }).
        AssertCreated().
        AssertJSONPath("user.name", "John Doe")

    // Test DELETE request
    tc.Delete("/api/users/123").
        AssertNoContent()
}
```

## Testing

Run tests:
```bash
go test ./gotest/...
```

Run with coverage:
```bash
go test ./gotest/... -cover
```

## Implementation Notes

- All public types and methods are documented with GoDoc comments
- HTTP helpers currently use `httptest` for request/response handling
- Database assertions are placeholders for future ORM integration
- Mock implementation is simple but sufficient for most use cases
- Future: Router integration for end-to-end HTTP testing

## License

Part of the GoFrame project.
