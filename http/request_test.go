package http

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNewRequest(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test", nil)
	req := NewRequest(rawReq)

	if req == nil {
		t.Fatal("NewRequest returned nil")
	}

	if req.Raw() != rawReq {
		t.Error("Raw() did not return original *http.Request")
	}
}

func TestRequestInput(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		formData url.Values
		key      string
		expected string
	}{
		{
			name:     "query parameter",
			url:      "/test?name=John",
			key:      "name",
			expected: "John",
		},
		{
			name:     "form data",
			url:      "/test",
			formData: url.Values{"email": {"test@example.com"}},
			key:      "email",
			expected: "test@example.com",
		},
		{
			name:     "missing key",
			url:      "/test",
			key:      "missing",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawReq *http.Request
			if tt.formData != nil {
				rawReq = httptest.NewRequest("POST", tt.url, strings.NewReader(tt.formData.Encode()))
				rawReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				rawReq = httptest.NewRequest("GET", tt.url, nil)
			}

			req := NewRequest(rawReq)
			result := req.InputString(tt.key)

			if result != tt.expected {
				t.Errorf("InputString(%s) = %s; want %s", tt.key, result, tt.expected)
			}
		})
	}
}

func TestRequestInputInt(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test?age=25&invalid=abc", nil)
	req := NewRequest(rawReq)

	if age := req.InputInt("age", 0); age != 25 {
		t.Errorf("InputInt(age) = %d; want 25", age)
	}

	if invalid := req.InputInt("invalid", 18); invalid != 18 {
		t.Errorf("InputInt(invalid) = %d; want 18 (default)", invalid)
	}

	if missing := req.InputInt("missing", 30); missing != 30 {
		t.Errorf("InputInt(missing) = %d; want 30 (default)", missing)
	}
}

func TestRequestAll(t *testing.T) {
	formData := url.Values{
		"name":  {"John"},
		"email": {"john@example.com"},
	}
	rawReq := httptest.NewRequest("POST", "/test?extra=value", strings.NewReader(formData.Encode()))
	rawReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req := NewRequest(rawReq)
	all := req.All()

	if len(all) != 3 {
		t.Errorf("All() returned %d items; want 3", len(all))
	}

	if all["name"] != "John" {
		t.Errorf("All()[name] = %v; want John", all["name"])
	}

	if all["extra"] != "value" {
		t.Errorf("All()[extra] = %v; want value", all["extra"])
	}
}

func TestRequestOnly(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test?name=John&email=test@example.com&age=25", nil)
	req := NewRequest(rawReq)

	result := req.Only("name", "email")

	if len(result) != 2 {
		t.Errorf("Only() returned %d items; want 2", len(result))
	}

	if result["name"] != "John" {
		t.Error("Only() missing name")
	}

	if result["email"] != "test@example.com" {
		t.Error("Only() missing email")
	}

	if _, exists := result["age"]; exists {
		t.Error("Only() should not include age")
	}
}

func TestRequestExcept(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test?name=John&password=secret&email=test@example.com", nil)
	req := NewRequest(rawReq)

	result := req.Except("password")

	if len(result) != 2 {
		t.Errorf("Except() returned %d items; want 2", len(result))
	}

	if _, exists := result["password"]; exists {
		t.Error("Except() should not include password")
	}

	if result["name"] != "John" {
		t.Error("Except() should include name")
	}
}

func TestRequestHasFilledMissing(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test?name=John&empty=", nil)
	req := NewRequest(rawReq)

	if !req.Has("name") {
		t.Error("Has(name) should be true")
	}

	if !req.Filled("name") {
		t.Error("Filled(name) should be true")
	}

	if req.Missing("name") {
		t.Error("Missing(name) should be false")
	}

	if !req.Has("empty") {
		t.Error("Has(empty) should be true")
	}

	if req.Filled("empty") {
		t.Error("Filled(empty) should be false")
	}

	if !req.Missing("nothere") {
		t.Error("Missing(nothere) should be true")
	}
}

func TestRequestQuery(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test?name=John", nil)
	req := NewRequest(rawReq)

	if name := req.Query("name"); name != "John" {
		t.Errorf("Query(name) = %s; want John", name)
	}

	if missing := req.Query("missing", "default"); missing != "default" {
		t.Errorf("Query(missing) = %s; want default", missing)
	}
}

func TestRequestRouteParam(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/users/123", nil)
	req := NewRequest(rawReq)
	req.SetRouteParam("id", "123")

	if id := req.RouteParam("id"); id != "123" {
		t.Errorf("RouteParam(id) = %s; want 123", id)
	}
}

func TestRequestHeader(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test", nil)
	rawReq.Header.Set("Authorization", "Bearer token123")
	req := NewRequest(rawReq)

	if auth := req.Header("Authorization"); auth != "Bearer token123" {
		t.Errorf("Header(Authorization) = %s; want Bearer token123", auth)
	}
}

func TestRequestIP(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*http.Request)
		expected string
	}{
		{
			name: "X-Forwarded-For",
			setup: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
			},
			expected: "192.168.1.1",
		},
		{
			name: "X-Real-IP",
			setup: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "192.168.1.2")
			},
			expected: "192.168.1.2",
		},
		{
			name: "RemoteAddr",
			setup: func(r *http.Request) {
				r.RemoteAddr = "192.168.1.3:12345"
			},
			expected: "192.168.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawReq := httptest.NewRequest("GET", "/test", nil)
			tt.setup(rawReq)
			req := NewRequest(rawReq)

			if ip := req.IP(); ip != tt.expected {
				t.Errorf("IP() = %s; want %s", ip, tt.expected)
			}
		})
	}
}

func TestRequestMethodPathURL(t *testing.T) {
	rawReq := httptest.NewRequest("POST", "http://example.com/api/users?page=1", nil)
	req := NewRequest(rawReq)

	if method := req.Method(); method != "POST" {
		t.Errorf("Method() = %s; want POST", method)
	}

	if path := req.Path(); path != "/api/users" {
		t.Errorf("Path() = %s; want /api/users", path)
	}

	expectedURL := "http://example.com/api/users"
	if reqURL := req.URL(); reqURL != expectedURL {
		t.Errorf("URL() = %s; want %s", reqURL, expectedURL)
	}

	expectedFullURL := "http://example.com/api/users?page=1"
	if fullURL := req.FullURL(); fullURL != expectedFullURL {
		t.Errorf("FullURL() = %s; want %s", fullURL, expectedFullURL)
	}
}

func TestRequestWantsJSON(t *testing.T) {
	tests := []struct {
		name     string
		accept   string
		expected bool
	}{
		{"JSON", "application/json", true},
		{"HTML", "text/html", false},
		{"Any", "*/*", false},
		{"JSON with charset", "application/json; charset=utf-8", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawReq := httptest.NewRequest("GET", "/test", nil)
			rawReq.Header.Set("Accept", tt.accept)
			req := NewRequest(rawReq)

			if result := req.WantsJSON(); result != tt.expected {
				t.Errorf("WantsJSON() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestRequestAjax(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test", nil)
	rawReq.Header.Set("X-Requested-With", "XMLHttpRequest")
	req := NewRequest(rawReq)

	if !req.Ajax() {
		t.Error("Ajax() should be true for XMLHttpRequest")
	}
}

func TestRequestBearerToken(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test", nil)
	rawReq.Header.Set("Authorization", "Bearer mytoken123")
	req := NewRequest(rawReq)

	if token := req.BearerToken(); token != "mytoken123" {
		t.Errorf("BearerToken() = %s; want mytoken123", token)
	}
}

func TestRequestMerge(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test?name=John", nil)
	req := NewRequest(rawReq)

	req.Merge(map[string]any{
		"age":   25,
		"email": "john@example.com",
	})

	if age := req.Input("age"); age != 25 {
		t.Errorf("Input(age) after Merge = %v; want 25", age)
	}

	if email := req.InputString("email"); email != "john@example.com" {
		t.Errorf("InputString(email) after Merge = %s; want john@example.com", email)
	}
}

func TestRequestMergeIfMissing(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test?name=John", nil)
	req := NewRequest(rawReq)

	req.MergeIfMissing(map[string]any{
		"name":  "Jane", // Should not override
		"email": "jane@example.com",
	})

	if name := req.InputString("name"); name != "John" {
		t.Errorf("InputString(name) = %s; want John (should not be overridden)", name)
	}

	if email := req.InputString("email"); email != "jane@example.com" {
		t.Errorf("InputString(email) = %s; want jane@example.com", email)
	}
}

func TestRequestFile(t *testing.T) {
	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add a file
	fileWriter, err := writer.CreateFormFile("avatar", "test.jpg")
	if err != nil {
		t.Fatal(err)
	}
	fileWriter.Write([]byte("fake image content"))

	writer.Close()

	rawReq := httptest.NewRequest("POST", "/upload", body)
	rawReq.Header.Set("Content-Type", writer.FormDataContentType())

	req := NewRequest(rawReq)
	file := req.File("avatar")

	if file == nil {
		t.Fatal("File(avatar) returned nil")
	}

	if file.GetClientOriginalName() != "test.jpg" {
		t.Errorf("GetClientOriginalName() = %s; want test.jpg", file.GetClientOriginalName())
	}
}

func TestRequestFiles(t *testing.T) {
	// Create multipart form data with multiple files
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add multiple files
	for i := 1; i <= 3; i++ {
		fileWriter, err := writer.CreateFormFile("images", "test.jpg")
		if err != nil {
			t.Fatal(err)
		}
		fileWriter.Write([]byte("fake image content"))
	}

	writer.Close()

	rawReq := httptest.NewRequest("POST", "/upload", body)
	rawReq.Header.Set("Content-Type", writer.FormDataContentType())

	req := NewRequest(rawReq)
	files := req.Files("images")

	if len(files) != 3 {
		t.Errorf("Files() returned %d files; want 3", len(files))
	}
}

func TestRequestCookie(t *testing.T) {
	rawReq := httptest.NewRequest("GET", "/test", nil)
	rawReq.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})

	req := NewRequest(rawReq)
	cookie, err := req.Cookie("session")

	if err != nil {
		t.Fatalf("Cookie(session) returned error: %v", err)
	}

	if cookie.Value != "abc123" {
		t.Errorf("Cookie value = %s; want abc123", cookie.Value)
	}
}
