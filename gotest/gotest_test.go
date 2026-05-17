package gotest

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestNewTestCase(t *testing.T) {
	tc := NewTestCase(t)
	if tc == nil {
		t.Fatal("Expected NewTestCase to return a non-nil TestCase")
	}
	if tc.t != t {
		t.Error("Expected TestCase.t to be set to the testing.T instance")
	}
	if tc.headers == nil {
		t.Error("Expected TestCase.headers to be initialized")
	}
}

func TestTestCaseSetUpTearDown(t *testing.T) {
	tc := NewTestCase(t)
	tc.SetUp()

	if tc.App == nil {
		t.Error("Expected SetUp to initialize the Application")
	}

	tc.TearDown()

	if tc.App != nil {
		t.Error("Expected TearDown to clear the Application")
	}
}

func TestWithHeader(t *testing.T) {
	tc := NewTestCase(t)
	result := tc.WithHeader("X-Test", "value")

	if result != tc {
		t.Error("Expected WithHeader to return the TestCase for chaining")
	}
	if tc.headers["X-Test"] != "value" {
		t.Error("Expected header to be set")
	}
}

func TestWithHeaders(t *testing.T) {
	tc := NewTestCase(t)
	headers := map[string]string{
		"X-Test-1": "value1",
		"X-Test-2": "value2",
	}
	result := tc.WithHeaders(headers)

	if result != tc {
		t.Error("Expected WithHeaders to return the TestCase for chaining")
	}
	if tc.headers["X-Test-1"] != "value1" || tc.headers["X-Test-2"] != "value2" {
		t.Error("Expected headers to be set")
	}
}

func TestWithToken(t *testing.T) {
	tc := NewTestCase(t)
	result := tc.WithToken("test-token")

	if result != tc {
		t.Error("Expected WithToken to return the TestCase for chaining")
	}
	if tc.headers["Authorization"] != "Bearer test-token" {
		t.Error("Expected Authorization header to be set")
	}
}

func TestActingAs(t *testing.T) {
	tc := NewTestCase(t)
	result := tc.ActingAs("user")

	if result != tc {
		t.Error("Expected ActingAs to return the TestCase for chaining")
	}
}

func TestAssertStatus(t *testing.T) {
	tc := NewTestCase(t)
	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       []byte{},
		testCase:   tc,
	}

	result := resp.AssertStatus(http.StatusOK)
	if result != resp {
		t.Error("Expected AssertStatus to return the TestResponse for chaining")
	}
}

func TestStatusAssertions(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		assertFn   func(*TestResponse) *TestResponse
	}{
		{"AssertOk", http.StatusOK, (*TestResponse).AssertOk},
		{"AssertCreated", http.StatusCreated, (*TestResponse).AssertCreated},
		{"AssertNoContent", http.StatusNoContent, (*TestResponse).AssertNoContent},
		{"AssertNotFound", http.StatusNotFound, (*TestResponse).AssertNotFound},
		{"AssertForbidden", http.StatusForbidden, (*TestResponse).AssertForbidden},
		{"AssertUnauthorized", http.StatusUnauthorized, (*TestResponse).AssertUnauthorized},
		{"AssertUnprocessable", http.StatusUnprocessableEntity, (*TestResponse).AssertUnprocessable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := NewTestCase(t)
			resp := &TestResponse{
				StatusCode: tt.statusCode,
				Headers:    http.Header{},
				Body:       []byte{},
				testCase:   tc,
			}
			result := tt.assertFn(resp)
			if result != resp {
				t.Error("Expected assertion to return the TestResponse for chaining")
			}
		})
	}
}

func TestAssertRedirect(t *testing.T) {
	tc := NewTestCase(t)
	resp := &TestResponse{
		StatusCode: http.StatusFound,
		Headers:    http.Header{},
		Body:       []byte{},
		testCase:   tc,
	}

	result := resp.AssertRedirect()
	if result != resp {
		t.Error("Expected AssertRedirect to return the TestResponse for chaining")
	}
}

func TestAssertRedirectTo(t *testing.T) {
	tc := NewTestCase(t)
	headers := http.Header{}
	headers.Set("Location", "/redirect-target")

	resp := &TestResponse{
		StatusCode: http.StatusFound,
		Headers:    headers,
		Body:       []byte{},
		testCase:   tc,
	}

	result := resp.AssertRedirectTo("/redirect-target")
	if result != resp {
		t.Error("Expected AssertRedirectTo to return the TestResponse for chaining")
	}
}

func TestAssertHeader(t *testing.T) {
	tc := NewTestCase(t)
	headers := http.Header{}
	headers.Set("X-Test", "value")

	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       []byte{},
		testCase:   tc,
	}

	result := resp.AssertHeader("X-Test", "value")
	if result != resp {
		t.Error("Expected AssertHeader to return the TestResponse for chaining")
	}
}

func TestAssertHeaderMissing(t *testing.T) {
	tc := NewTestCase(t)
	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       []byte{},
		testCase:   tc,
	}

	result := resp.AssertHeaderMissing("X-Missing")
	if result != resp {
		t.Error("Expected AssertHeaderMissing to return the TestResponse for chaining")
	}
}

func TestAssertJSON(t *testing.T) {
	tc := NewTestCase(t)
	data := map[string]any{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}
	body, _ := json.Marshal(data)

	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       body,
		testCase:   tc,
	}

	expected := map[string]any{
		"name": "John",
		"age":  float64(30), // JSON numbers are float64
	}

	result := resp.AssertJSON(expected)
	if result != resp {
		t.Error("Expected AssertJSON to return the TestResponse for chaining")
	}
}

func TestAssertExactJSON(t *testing.T) {
	tc := NewTestCase(t)
	data := map[string]any{
		"name": "John",
		"age":  float64(30),
	}
	body, _ := json.Marshal(data)

	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       body,
		testCase:   tc,
	}

	result := resp.AssertExactJSON(data)
	if result != resp {
		t.Error("Expected AssertExactJSON to return the TestResponse for chaining")
	}
}

func TestAssertJSONPath(t *testing.T) {
	tc := NewTestCase(t)
	data := map[string]any{
		"user": map[string]any{
			"name": "John",
			"age":  float64(30),
		},
	}
	body, _ := json.Marshal(data)

	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       body,
		testCase:   tc,
	}

	result := resp.AssertJSONPath("user.name", "John")
	if result != resp {
		t.Error("Expected AssertJSONPath to return the TestResponse for chaining")
	}
}

func TestAssertJSONCount(t *testing.T) {
	tc := NewTestCase(t)
	data := map[string]any{
		"items": []any{"a", "b", "c"},
	}
	body, _ := json.Marshal(data)

	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       body,
		testCase:   tc,
	}

	result := resp.AssertJSONCount("items", 3)
	if result != resp {
		t.Error("Expected AssertJSONCount to return the TestResponse for chaining")
	}
}

func TestAssertJSONMissing(t *testing.T) {
	tc := NewTestCase(t)
	data := map[string]any{
		"name": "John",
	}
	body, _ := json.Marshal(data)

	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       body,
		testCase:   tc,
	}

	result := resp.AssertJSONMissing("age")
	if result != resp {
		t.Error("Expected AssertJSONMissing to return the TestResponse for chaining")
	}
}

func TestAssertSee(t *testing.T) {
	tc := NewTestCase(t)
	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       []byte("Hello World"),
		testCase:   tc,
	}

	result := resp.AssertSee("Hello")
	if result != resp {
		t.Error("Expected AssertSee to return the TestResponse for chaining")
	}
}

func TestAssertDontSee(t *testing.T) {
	tc := NewTestCase(t)
	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       []byte("Hello World"),
		testCase:   tc,
	}

	result := resp.AssertDontSee("Goodbye")
	if result != resp {
		t.Error("Expected AssertDontSee to return the TestResponse for chaining")
	}
}

func TestAssertSeeInOrder(t *testing.T) {
	tc := NewTestCase(t)
	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       []byte("First Second Third"),
		testCase:   tc,
	}

	result := resp.AssertSeeInOrder([]string{"First", "Second", "Third"})
	if result != resp {
		t.Error("Expected AssertSeeInOrder to return the TestResponse for chaining")
	}
}

func TestAssertBodyContains(t *testing.T) {
	tc := NewTestCase(t)
	resp := &TestResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       []byte("Hello World"),
		testCase:   tc,
	}

	result := resp.AssertBodyContains("World")
	if result != resp {
		t.Error("Expected AssertBodyContains to return the TestResponse for chaining")
	}
}

func TestDatabaseAssertions(t *testing.T) {
	tc := NewTestCase(t)

	// These are placeholders, so we just verify they don't panic
	result := tc.AssertDatabaseHas("users", map[string]any{"name": "John"})
	if result != tc {
		t.Error("Expected AssertDatabaseHas to return the TestCase for chaining")
	}

	result = tc.AssertDatabaseMissing("users", map[string]any{"name": "Jane"})
	if result != tc {
		t.Error("Expected AssertDatabaseMissing to return the TestCase for chaining")
	}

	result = tc.AssertDatabaseCount("users", 5)
	if result != tc {
		t.Error("Expected AssertDatabaseCount to return the TestCase for chaining")
	}
}

func TestNewMock(t *testing.T) {
	mock := NewMock()
	if mock == nil {
		t.Fatal("Expected NewMock to return a non-nil Mock")
	}
	if mock.Calls == nil {
		t.Error("Expected Mock.Calls to be initialized")
	}
	if mock.expectations == nil {
		t.Error("Expected Mock.expectations to be initialized")
	}
}

func TestMockOn(t *testing.T) {
	mock := NewMock()
	exp := mock.On("TestMethod", "arg1", "arg2")

	if exp == nil {
		t.Fatal("Expected On to return a MockExpectation")
	}
	if exp.Method != "TestMethod" {
		t.Error("Expected MockExpectation.Method to be set")
	}
	if len(exp.Args) != 2 {
		t.Error("Expected MockExpectation.Args to contain 2 arguments")
	}
}

func TestMockReturn(t *testing.T) {
	mock := NewMock()
	exp := mock.On("TestMethod").Return("result1", "result2")

	if len(exp.Returns) != 2 {
		t.Error("Expected Return to set return values")
	}
	if exp.Returns[0] != "result1" || exp.Returns[1] != "result2" {
		t.Error("Expected return values to be set correctly")
	}
}

func TestMockCalled(t *testing.T) {
	mock := NewMock()
	mock.On("TestMethod", "arg1").Return("result")

	returns := mock.Called("TestMethod", "arg1")

	if len(mock.Calls) != 1 {
		t.Fatal("Expected Called to record the call")
	}
	if mock.Calls[0].Method != "TestMethod" {
		t.Error("Expected call to be recorded with correct method")
	}
	if len(returns) != 1 || returns[0] != "result" {
		t.Error("Expected Called to return configured return values")
	}
}

func TestMockAssertCalled(t *testing.T) {
	mock := NewMock()
	mock.Called("TestMethod")

	// This should not fail
	mock.AssertCalled(t, "TestMethod")
}

func TestMockAssertNotCalled(t *testing.T) {
	mock := NewMock()

	// This should not fail
	mock.AssertNotCalled(t, "TestMethod")
}

func TestMockAssertCalledTimes(t *testing.T) {
	mock := NewMock()
	mock.Called("TestMethod")
	mock.Called("TestMethod")
	mock.Called("TestMethod")

	// This should not fail
	mock.AssertCalledTimes(t, "TestMethod", 3)
}

func TestMockAssertCalledWith(t *testing.T) {
	mock := NewMock()
	mock.Called("TestMethod", "arg1", "arg2")

	// This should not fail
	mock.AssertCalledWith(t, "TestMethod", "arg1", "arg2")
}

func TestMockReset(t *testing.T) {
	mock := NewMock()
	mock.On("TestMethod").Return("result")
	mock.Called("TestMethod")

	mock.Reset()

	if len(mock.Calls) != 0 {
		t.Error("Expected Reset to clear all calls")
	}
	if len(mock.expectations) != 0 {
		t.Error("Expected Reset to clear all expectations")
	}
}

func TestGetJSONPath(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "John",
			"profile": map[string]any{
				"age": float64(30),
			},
		},
	}

	tests := []struct {
		path     string
		expected any
	}{
		{"user.name", "John"},
		{"user.profile.age", float64(30)},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getJSONPath(data, tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHTTPHelpers(t *testing.T) {
	tc := NewTestCase(t)

	// Test that HTTP methods return TestResponse
	methods := []struct {
		name string
		fn   func() *TestResponse
	}{
		{"Get", func() *TestResponse { return tc.Get("/test") }},
		{"Post", func() *TestResponse { return tc.Post("/test", map[string]any{"key": "value"}) }},
		{"Put", func() *TestResponse { return tc.Put("/test", map[string]any{"key": "value"}) }},
		{"Delete", func() *TestResponse { return tc.Delete("/test") }},
		{"JSON", func() *TestResponse { return tc.JSON("PATCH", "/test", map[string]any{"key": "value"}) }},
	}

	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			resp := method.fn()
			if resp == nil {
				t.Error("Expected HTTP helper to return a TestResponse")
			}
			if resp.testCase != tc {
				t.Error("Expected TestResponse to reference the TestCase")
			}
		})
	}
}
