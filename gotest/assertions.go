package gotest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// AssertStatus asserts that the response has the given status code.
func (r *TestResponse) AssertStatus(code int) *TestResponse {
	if r.StatusCode != code {
		r.testCase.t.Errorf("Expected status %d, got %d", code, r.StatusCode)
	}
	return r
}

// AssertOk asserts that the response status is 200 OK.
func (r *TestResponse) AssertOk() *TestResponse {
	return r.AssertStatus(http.StatusOK)
}

// AssertCreated asserts that the response status is 201 Created.
func (r *TestResponse) AssertCreated() *TestResponse {
	return r.AssertStatus(http.StatusCreated)
}

// AssertNoContent asserts that the response status is 204 No Content.
func (r *TestResponse) AssertNoContent() *TestResponse {
	return r.AssertStatus(http.StatusNoContent)
}

// AssertNotFound asserts that the response status is 404 Not Found.
func (r *TestResponse) AssertNotFound() *TestResponse {
	return r.AssertStatus(http.StatusNotFound)
}

// AssertForbidden asserts that the response status is 403 Forbidden.
func (r *TestResponse) AssertForbidden() *TestResponse {
	return r.AssertStatus(http.StatusForbidden)
}

// AssertUnauthorized asserts that the response status is 401 Unauthorized.
func (r *TestResponse) AssertUnauthorized() *TestResponse {
	return r.AssertStatus(http.StatusUnauthorized)
}

// AssertUnprocessable asserts that the response status is 422 Unprocessable Entity.
func (r *TestResponse) AssertUnprocessable() *TestResponse {
	return r.AssertStatus(http.StatusUnprocessableEntity)
}

// AssertRedirect asserts that the response is a redirect (3xx status).
func (r *TestResponse) AssertRedirect() *TestResponse {
	if r.StatusCode < 300 || r.StatusCode >= 400 {
		r.testCase.t.Errorf("Expected redirect status (3xx), got %d", r.StatusCode)
	}
	return r
}

// AssertRedirectTo asserts that the response redirects to the given URL.
func (r *TestResponse) AssertRedirectTo(url string) *TestResponse {
	location := r.Headers.Get("Location")
	if location != url {
		r.testCase.t.Errorf("Expected redirect to %s, got %s", url, location)
	}
	return r.AssertRedirect()
}

// AssertHeader asserts that the response has a header with the given value.
func (r *TestResponse) AssertHeader(key, value string) *TestResponse {
	actual := r.Headers.Get(key)
	if actual != value {
		r.testCase.t.Errorf("Expected header %s to be %s, got %s", key, value, actual)
	}
	return r
}

// AssertHeaderMissing asserts that the response does not have the given header.
func (r *TestResponse) AssertHeaderMissing(key string) *TestResponse {
	if r.Headers.Get(key) != "" {
		r.testCase.t.Errorf("Expected header %s to be missing, but it was present", key)
	}
	return r
}

// AssertJSON asserts that the response JSON contains the expected data (subset match).
func (r *TestResponse) AssertJSON(expected map[string]any) *TestResponse {
	var actual map[string]any
	if err := json.Unmarshal(r.Body, &actual); err != nil {
		r.testCase.t.Fatalf("Failed to parse response JSON: %v", err)
	}

	for key, expectedVal := range expected {
		actualVal, ok := actual[key]
		if !ok {
			r.testCase.t.Errorf("Expected JSON key %s to exist", key)
			continue
		}
		if !reflect.DeepEqual(actualVal, expectedVal) {
			r.testCase.t.Errorf("Expected JSON key %s to be %v, got %v", key, expectedVal, actualVal)
		}
	}
	return r
}

// AssertExactJSON asserts that the response JSON exactly matches the expected data.
func (r *TestResponse) AssertExactJSON(expected map[string]any) *TestResponse {
	var actual map[string]any
	if err := json.Unmarshal(r.Body, &actual); err != nil {
		r.testCase.t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		r.testCase.t.Errorf("Expected exact JSON match.\nExpected: %v\nGot: %v", expected, actual)
	}
	return r
}

// AssertJSONPath asserts that a value at the given dot-notation path matches the expected value.
func (r *TestResponse) AssertJSONPath(path string, value any) *TestResponse {
	var data map[string]any
	if err := json.Unmarshal(r.Body, &data); err != nil {
		r.testCase.t.Fatalf("Failed to parse response JSON: %v", err)
	}

	actual := getJSONPath(data, path)
	if !reflect.DeepEqual(actual, value) {
		r.testCase.t.Errorf("Expected JSON path %s to be %v, got %v", path, value, actual)
	}
	return r
}

// AssertJSONCount asserts that the JSON array at the given path has the expected count.
func (r *TestResponse) AssertJSONCount(path string, count int) *TestResponse {
	var data map[string]any
	if err := json.Unmarshal(r.Body, &data); err != nil {
		r.testCase.t.Fatalf("Failed to parse response JSON: %v", err)
	}

	val := getJSONPath(data, path)
	arr, ok := val.([]any)
	if !ok {
		r.testCase.t.Errorf("Expected JSON path %s to be an array", path)
		return r
	}

	if len(arr) != count {
		r.testCase.t.Errorf("Expected JSON path %s to have %d elements, got %d", path, count, len(arr))
	}
	return r
}

// AssertJSONMissing asserts that the given key is missing from the response JSON.
func (r *TestResponse) AssertJSONMissing(key string) *TestResponse {
	var data map[string]any
	if err := json.Unmarshal(r.Body, &data); err != nil {
		r.testCase.t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if _, ok := data[key]; ok {
		r.testCase.t.Errorf("Expected JSON key %s to be missing, but it was present", key)
	}
	return r
}

// AssertSee asserts that the response body contains the given text.
func (r *TestResponse) AssertSee(text string) *TestResponse {
	if !strings.Contains(string(r.Body), text) {
		r.testCase.t.Errorf("Expected response body to contain %q", text)
	}
	return r
}

// AssertDontSee asserts that the response body does not contain the given text.
func (r *TestResponse) AssertDontSee(text string) *TestResponse {
	if strings.Contains(string(r.Body), text) {
		r.testCase.t.Errorf("Expected response body to not contain %q", text)
	}
	return r
}

// AssertSeeInOrder asserts that the response body contains the given texts in order.
func (r *TestResponse) AssertSeeInOrder(texts []string) *TestResponse {
	body := string(r.Body)
	lastIndex := -1

	for _, text := range texts {
		index := strings.Index(body, text)
		if index == -1 {
			r.testCase.t.Errorf("Expected response body to contain %q", text)
			return r
		}
		if index <= lastIndex {
			r.testCase.t.Errorf("Expected %q to appear after previous text, but it appeared earlier", text)
			return r
		}
		lastIndex = index
	}
	return r
}

// AssertBodyContains asserts that the response body contains the given text.
// This is an alias for AssertSee.
func (r *TestResponse) AssertBodyContains(text string) *TestResponse {
	return r.AssertSee(text)
}

// getJSONPath retrieves a value from a map using dot-notation path.
func getJSONPath(data map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var current any = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		default:
			return nil
		}
	}
	return current
}

// Dump prints the response body for debugging.
// Useful for inspecting the actual response during test development.
func (r *TestResponse) Dump() *TestResponse {
	fmt.Printf("Status: %d\nHeaders: %v\nBody: %s\n", r.StatusCode, r.Headers, string(r.Body))
	return r
}
