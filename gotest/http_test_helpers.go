package gotest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
)

// TestResponse represents an HTTP response for testing.
// It captures status code, headers, and body for assertion.
type TestResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	testCase   *TestCase
}

// Get performs a GET request to the given URL.
func (tc *TestCase) Get(url string) *TestResponse {
	return tc.request(http.MethodGet, url, nil)
}

// Post performs a POST request with the given data.
// Data is automatically JSON-encoded.
func (tc *TestCase) Post(url string, data map[string]any) *TestResponse {
	return tc.request(http.MethodPost, url, data)
}

// Put performs a PUT request with the given data.
// Data is automatically JSON-encoded.
func (tc *TestCase) Put(url string, data map[string]any) *TestResponse {
	return tc.request(http.MethodPut, url, data)
}

// Delete performs a DELETE request to the given URL.
func (tc *TestCase) Delete(url string) *TestResponse {
	return tc.request(http.MethodDelete, url, nil)
}

// JSON performs a request with JSON-encoded data.
// Use this for custom HTTP methods or explicit JSON content type.
func (tc *TestCase) JSON(method, url string, data map[string]any) *TestResponse {
	return tc.request(method, url, data)
}

// WithHeader sets a header for the next request.
// Returns the test case for method chaining.
func (tc *TestCase) WithHeader(key, val string) *TestCase {
	tc.headers[key] = val
	return tc
}

// WithHeaders sets multiple headers for the next request.
// Returns the test case for method chaining.
func (tc *TestCase) WithHeaders(headers map[string]string) *TestCase {
	for k, v := range headers {
		tc.headers[k] = v
	}
	return tc
}

// WithToken sets a Bearer token for the next request.
// Returns the test case for method chaining.
func (tc *TestCase) WithToken(token string) *TestCase {
	tc.headers["Authorization"] = "Bearer " + token
	return tc
}

// ActingAs authenticates as the given user for the next request.
// This is a placeholder for future authentication integration.
// Returns the test case for method chaining.
func (tc *TestCase) ActingAs(user any) *TestCase {
	// Future: integrate with auth system to set authenticated user
	return tc
}

// request performs the actual HTTP request and returns the response.
func (tc *TestCase) request(method, url string, data map[string]any) *TestResponse {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			tc.t.Fatalf("Failed to marshal request data: %v", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req := httptest.NewRequest(method, url, body)

	// Set headers
	for k, v := range tc.headers {
		req.Header.Set(k, v)
	}

	// Set Content-Type for JSON requests
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create response recorder
	rec := httptest.NewRecorder()

	// Future: route the request through the app's router
	// For now, we just return the recorder's response

	// Read response body
	respBody, err := io.ReadAll(rec.Body)
	if err != nil {
		tc.t.Fatalf("Failed to read response body: %v", err)
	}

	// Clear headers for next request
	tc.headers = make(map[string]string)

	return &TestResponse{
		StatusCode: rec.Code,
		Headers:    rec.Header(),
		Body:       respBody,
		testCase:   tc,
	}
}
